package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	redis "github.com/go-redis/redis/v7"
	"github.com/go-sql-driver/mysql"

	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

const (
	redisTimeout          = 10 * time.Second
	challengeScoreHashKey = "SCORES"
)

type ChallengeRepository interface {
	RegisterChallenge(name, flag, desc, category, difficulty, author string, tags []string, baseScore int, isDynamic, isQuestionary bool, host, port *string) (uint32, error)
	UpdateChallengeByName(name, flag, desc, category, difficulty, author string, baseScore int, isDynamic, isQuestionary bool, host, port *string) error
	AddAttachment(cid uint32, url string) error

	OpenChallenge(id uint32) error
	CloseChallenge(id uint32) error

	FindChallengeIDByName(name string) (uint32, error)
	FindChallengeByID(id uint32) (*model.Challenge, error)
	FindOpenChallengeByFlag(flag string) (*model.Challenge, error)
	ListAllChallenges(opened bool) ([]*model.Challenge, error)

	AddSolvedChallenge(tid, cid uint32) error
	TeamSolvedChallenges(tid uint32) ([]uint32, error)
	UpdateScore(cid uint32, score int) error
}

func (r *repository) RegisterChallenge(name, flag, desc, category, difficulty, author string, tags []string, baseScore int, isDynamic, isQuestionary bool, host, port *string) (uint32, error) {
	id := r.newID()
	_, err := r.db.Exec(
		`INSERT INTO
		challenges(id, name, flag, description, category, difficulty, author, base_score, is_dynamic, is_questionary, is_open, host, port)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, name, flag, desc, category, difficulty, author, baseScore, isDynamic, isQuestionary, false, host, port,
	)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok && mysqlerr.Number == 1062 {
			return 0, model.DuplicateError("challenge")
		}
		return 0, err
	}
	if len(tags) == 0 {
		return id, nil
	}

	tagValues := make([]model.Tag, 0, len(tags))
	for _, t := range tags {
		tagValues = append(tagValues, model.Tag{
			ID:          r.newID(),
			ChallengeID: id,
			Tag:         t,
		})
	}

	_, err = r.db.NamedExec(
		`INSERT INTO
		challenge_tags (id, challenge_id, tag)
		VALUE (:id, :challenge_id, :tag)`,
		tagValues,
	)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok && mysqlerr.Number == 1062 {
			return 0, model.DuplicateError("tag")
		}
		return 0, err
	}

	return id, nil
}

func (r *repository) UpdateChallengeByName(name, flag, desc, category, difficulty, author string, baseScore int, isDynamic, isQuestionary bool, host, port *string) error {
	_, err := r.db.Exec(
		`UPDATE challenges
		SET flag = ?, description = ?, category = ?, difficulty = ?, author = ?, base_score = ?, is_dynamic = ?, is_questionary = ?, host = ?, port = ?
		WHERE name = ?`,
		flag, desc, category, difficulty, author, baseScore, isDynamic, isQuestionary, host, port, name,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) AddAttachment(cid uint32, url string) error {
	id := r.newID()
	_, err := r.db.Exec(
		`INSERT INTO
		challenge_attachments (id, challenge_id, url)
		VALUES (?, ?, ?)
		`,
		id, cid, url,
	)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok && mysqlerr.Number == 1062 {
			return model.DuplicateError("url")
		}
		return err
	}
	return nil
}

func (r *repository) OpenChallenge(id uint32) error {
	_, err := r.db.Exec(
		`UPDATE challenges
		SET is_open = TRUE
		WHERE id = ?`,
		id,
	)
	return err
}

func (r *repository) CloseChallenge(id uint32) error {
	_, err := r.db.Exec(
		`UPDATE challenges
		SET is_open = FALSE
		WHERE id = ?`,
		id,
	)
	return err
}

func (r *repository) FindChallengeIDByName(name string) (uint32, error) {
	var chal model.Challenge
	err := r.db.Get(
		&chal,
		`SELECT *
		FROM challenges
		WHERE name = ?
		LIMIT 1`,
		name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, model.NotFoundError("challenge")
		}
		return 0, err
	}

	return chal.ID, nil
}

func (r *repository) FindChallengeByID(id uint32) (*model.Challenge, error) {
	var chal model.Challenge
	err := r.db.Get(
		&chal,
		`SELECT *
		FROM challenges
		WHERE id = ?
		LIMIT 1`,
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("challenge")
		}
		return nil, err
	}

	chal, err = r.setChallengeScore(chal)
	if err != nil {
		return nil, err
	}

	chal, err = r.setTagAttachment(chal)
	if err != nil {
		return nil, err
	}

	chal, err = r.setChallengesolveTeam(chal)
	if err != nil {
		return nil, err
	}
	return &chal, nil
}

func (r *repository) FindOpenChallengeByFlag(flag string) (*model.Challenge, error) {
	var chal model.Challenge
	err := r.db.Get(
		&chal,
		`SELECT *
		FROM challenges
		WHERE flag = ? AND is_open = TRUE
		LIMIT 1`,
		flag,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("challenge")
		}
		return nil, err
	}

	return &chal, nil
}

func (r *repository) ListAllChallenges(opened bool) ([]*model.Challenge, error) {
	chals := make([]*model.Challenge, 0)
	err := r.db.Select(
		&chals,
		`SELECT *
		FROM challenges
		WHERE ? OR is_open
		ORDER BY created_at ASC
		`,
		!opened,
	)
	if err != nil {
		// sqlx.select does not return errnorows
		return nil, fmt.Errorf("%w", err)
	}

	chals, err = r.setChallengeScores(chals)
	if err != nil {
		return nil, err
	}

	chals, err = r.setTagsAttachments(chals)
	if err != nil {
		return nil, err
	}

	chals, err = r.setChallengeSolveTeams(chals)
	if err != nil {
		return nil, err
	}

	return chals, err
}

func (r *repository) AddSolvedChallenge(tid, cid uint32) error {
	key := solvedChallengesKey(tid)
	err := r.redis.SAdd(key, cid).Err()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (r *repository) TeamSolvedChallenges(tid uint32) ([]uint32, error) {
	key := solvedChallengesKey(tid)
	chals, err := r.redis.SMembers(key).Result()
	if err == redis.Nil {
		return []uint32{}, nil
	} else if err != nil {
		return nil, err
	}

	ids, err := convIDList(chals)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *repository) UpdateScore(cid uint32, score int) error {
	key := challengeScoreKey(cid)
	res := r.redis.HSet(challengeScoreHashKey, key, strconv.Itoa(score))
	return res.Err()
}

func (r *repository) setTagAttachment(chal model.Challenge) (model.Challenge, error) {
	tags := make([]model.Tag, 0)
	err := r.db.Select(
		&tags,
		`SELECT tag
		FROM challenge_tags
		WHERE challenge_id = ?
		`,
		chal.ID,
	)
	if err != nil {
		return model.Challenge{}, err
	}

	urls := make([]model.Attachment, 0)
	err = r.db.Select(
		&urls,
		`SELECT url
		FROM challenge_attachments
		WHERE challenge_id = ?
		`,
		chal.ID,
	)
	if err != nil {
		return model.Challenge{}, err
	}

	if chal.Tags == nil {
		chal.Tags = make([]string, 0)
	}
	for i := 0; i < len(tags); i++ {
		chal.Tags = append(chal.Tags, tags[i].Tag)
	}

	if chal.Attachments == nil {
		chal.Attachments = make([]string, 0)
	}
	for i := 0; i < len(urls); i++ {
		chal.Attachments = append(chal.Attachments, urls[i].URL)
	}

	return chal, nil
}

func (r *repository) setTagsAttachments(chals []*model.Challenge) ([]*model.Challenge, error) {
	// set tags / attachments
	tags := make([]model.Tag, 0)
	err := r.db.Select(
		&tags,
		`SELECT challenge_id, tag
		FROM challenge_tags`,
	)
	if err != nil {
		return nil, err
	}

	urls := make([]model.Attachment, 0)
	err = r.db.Select(
		&urls,
		`SELECT challenge_id, url
		FROM challenge_attachments`,
	)
	if err != nil {
		return nil, err
	}

	// create map for cache
	tagMap := make(map[uint32][]string)
	urlMap := make(map[uint32][]string)
	for i := 0; i < len(chals); i++ {
		tagMap[chals[i].ID] = make([]string, 0)
		urlMap[chals[i].ID] = make([]string, 0)
	}

	for _, t := range tags {
		tagMap[t.ChallengeID] = append(tagMap[t.ChallengeID], t.Tag)
	}
	for _, u := range urls {
		urlMap[u.ChallengeID] = append(urlMap[u.ChallengeID], u.URL)
	}

	// assign into the struct
	for i := 0; i < len(chals); i++ {
		chals[i].Tags = tagMap[chals[i].ID]
		chals[i].Attachments = urlMap[chals[i].ID]
	}

	return chals, nil
}

func (r *repository) setChallengesolveTeam(chal model.Challenge) (model.Challenge, error) {
	submissions := make([]model.Submission, 0)
	err := r.db.Select(
		&submissions,
		`SELECT team_id
		FROM submissions
		WHERE is_valid = TRUE AND challenge_id = ?
		`,
		chal.ID,
	)
	if err != nil {
		return model.Challenge{}, err
	}

	if chal.SolveTeams == nil {
		chal.SolveTeams = make([]uint32, 0)
	}
	for i := 0; i < len(submissions); i++ {
		if submissions[i].TeamID != nil {
			chal.SolveTeams = append(chal.SolveTeams, *submissions[i].TeamID)
		}
	}
	return chal, nil
}

func (r *repository) setChallengeSolveTeams(chals []*model.Challenge) ([]*model.Challenge, error) {
	submissions := make([]model.Submission, 0)
	err := r.db.Select(
		&submissions,
		`SELECT team_id, challenge_id
		FROM submissions
		WHERE is_valid = TRUE
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	chalMap := make(map[uint32][]uint32)
	for i := 0; i < len(chals); i++ {
		chalMap[chals[i].ID] = make([]uint32, 0)
	}
	for _, s := range submissions {
		if s.ChallengeID != nil && s.TeamID != nil {
			chalMap[*s.ChallengeID] = append(chalMap[*s.ChallengeID], *s.TeamID)
		}
	}

	for i := 0; i < len(chals); i++ {
		chals[i].SolveTeams = chalMap[chals[i].ID]
	}
	return chals, nil
}

func (r *repository) setChallengeScore(chal model.Challenge) (model.Challenge, error) {
	key := challengeScoreKey(chal.ID)
	scoreStr, err := r.redis.HGet(challengeScoreHashKey, key).Result()
	if err == redis.Nil {
		chal.Score = chal.BaseScore
		return chal, nil
	}
	if err != nil {
		return model.Challenge{}, err
	}
	score, err := strconv.Atoi(scoreStr)
	if err != nil {
		return model.Challenge{}, err
	}
	chal.Score = score
	return chal, err
}

func (r *repository) setChallengeScores(chals []*model.Challenge) ([]*model.Challenge, error) {
	scores, err := r.redis.HGetAll(challengeScoreHashKey).Result()
	if err == redis.Nil {
		for i := 0; i < len(chals); i++ {
			chals[i].Score = chals[i].BaseScore
		}
		return chals, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	for i := 0; i < len(chals); i++ {
		scoreStr, ok := scores[challengeScoreKey(chals[i].ID)]
		if ok {
			chals[i].Score, err = strconv.Atoi(scoreStr)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}
		} else {
			chals[i].Score = chals[i].BaseScore
		}
	}
	return chals, nil
}

func solvedChallengesKey(id uint32) string {
	return fmt.Sprintf("TEAM%d", id)
}

func challengeScoreKey(id uint32) string {
	return fmt.Sprintf("%d", id)
}

func convIDList(idstrs []string) ([]uint32, error) {
	ids := make([]uint32, len(idstrs))

	for i := 0; i < len(idstrs); i++ {
		id, err := strconv.ParseUint(idstrs[i], 10, 32)
		if err != nil {
			return nil, err
		}
		ids[i] = uint32(id)
	}
	return ids, nil
}

func convScoreList(scoreStr []string) ([]int, error) {
	scores := make([]int, len(scoreStr))

	for i := 0; i < len(scoreStr); i++ {
		score, err := strconv.ParseInt(scoreStr[i], 10, 32)
		if err != nil {
			return nil, err
		}
		scores[i] = int(score)
	}
	return scores, nil
}
