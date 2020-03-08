package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	redis "github.com/go-redis/redis/v7"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

type SubmissionRepository interface {
	FindValidSubmission(tid, cid uint32) (*model.Submission, error)
	InsertSubmission(cid, uid, tid sql.NullInt64, flag string, submit_at int64, is_correct, is_valid bool) error

	ListValidSubmission(cid uint32) ([]*model.Submission, error)

	IncrementWrong(tid uint32, expire time.Duration) (int, error)
	GetWrongCount(tid uint32) (int, error)

	LockSubmission(tid uint32, duration time.Duration) error
	CheckSubmitAvailable(tid uint32) (bool, error)
}

func (r *repository) FindValidSubmission(tid, cid uint32) (*model.Submission, error) {
	var submission model.Submission

	err := r.db.Get(
		&submission,
		`SELECT *
		FROM submissions
		WHERE team_id = ? AND challenge_id = ? AND is_valid = TRUE
		LIMIT 1`,
		tid, cid,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("submission")
		}
		return nil, err
	}

	return &submission, nil
}

func (r *repository) InsertSubmission(cid, uid, tid sql.NullInt64, flag string, submit_at int64, is_correct, is_valid bool) error {
	id := r.newID()
	_, err := r.db.Exec(
		`INSERT INTO
		submissions(id, user_id, team_id, challenge_id, flag, submitted_at, is_correct, is_valid)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, uid, tid, cid, flag, submit_at, is_correct, is_valid,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) ListValidSubmission(cid uint32) ([]*model.Submission, error) {
	submissons := make([]*model.Submission, 0)
	err := r.db.Select(
		&submissons,
		`SELECT *
		FROM submissions
		WHERE challenge_id = ? AND is_valid = TRUE
		ORDER BY created_at ASC
		`,
		cid,
	)
	if err != nil {
		// sqlx.select does not return errnorows
		return nil, err
	}
	return submissons, nil
}

func (r *repository) IncrementWrong(tid uint32, expire time.Duration) (int, error) {
	key := wrongCountKey(tid)
	cnt, err := r.redis.Incr(key).Result()
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}

	err = r.redis.Expire(key, expire).Err()
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}

	return int(cnt), nil
}

func (r *repository) GetWrongCount(tid uint32) (int, error) {
	key := wrongCountKey(tid)
	countStr, err := r.redis.Get(key).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("%w", err)
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}

	return count, nil
}

func (r *repository) LockSubmission(tid uint32, duration time.Duration) error {
	key := lockSubmissionKey(tid)
	err := r.redis.Set(key, "1", duration).Err()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (r *repository) CheckSubmitAvailable(tid uint32) (bool, error) {
	key := lockSubmissionKey(tid)
	err := r.redis.Get(key).Err()
	if err == redis.Nil {
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("%w", err)
	}
	return false, nil
}

func wrongCountKey(id uint32) string {
	return fmt.Sprintf("WRONG%d", id)
}

func lockSubmissionKey(id uint32) string {
	return fmt.Sprintf("LOCK%d", id)
}
