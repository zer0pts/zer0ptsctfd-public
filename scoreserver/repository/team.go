package repository

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

type TeamRepository interface {
	FindTeamByName(teamName string) (*model.Team, error)
	FindTeamByToken(token string) (*model.Team, error)
	FindUserTeam(uid uint32) (*model.Team, error)

	FindTeamByID(id uint32) (*model.Team, error)
	UpdateTeamName(tid uint32, teamName string) error
	ListTeams(visibleOnly bool) ([]*model.Team, error)

	CreateTeam(teamName, token, countryCode string) (uint32, error)
	SetCountryCode(tid uint32, counrtyCode string) error
}

func (r *repository) FindTeamByName(teamName string) (*model.Team, error) {
	var team model.Team
	err := r.db.Get(
		&team,
		`SELECT *
		FROM teams
		WHERE teamname = ?`,
		teamName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("team")
		}
		return nil, err
	}
	return &team, nil
}

func (r *repository) FindTeamByToken(token string) (*model.Team, error) {
	var team model.Team
	err := r.db.Get(
		&team,
		`SELECT *
		FROM teams
		WHERE token = ?`,
		token,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("team")
		}
		return nil, err
	}
	return &team, nil
}

func (r *repository) FindUserTeam(uid uint32) (*model.Team, error) {
	var team model.Team
	err := r.db.Get(
		&team,
		`SELECT teams.*
		FROM teams
		INNER JOIN users
		ON users.id = ? AND users.team_id = teams.id
		`,
		uid,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("team")
		}
		return nil, err
	}
	return &team, nil
}

func (r *repository) FindTeamByID(id uint32) (*model.Team, error) {
	var team model.Team
	err := r.db.Get(
		&team,
		`SELECT teams.*
		FROM teams
		WHERE id = ?
		`,
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("team")
		}
		return nil, err
	}

	team, err = r.setUsers(team)
	if err != nil {
		return nil, err
	}

	team, err = r.setValidSubmission(team)
	if err != nil {
		return nil, err
	}

	return &team, nil
}

func (r *repository) ListTeams(visibleOnly bool) ([]*model.Team, error) {
	teams := make([]*model.Team, 0)
	err := r.db.Select(
		&teams,
		`
		SELECT *
		FROM teams
		WHERE ? OR (NOT is_hidden)
		`,
		!visibleOnly,
	)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	teams, err = r.setValidSubmissions(teams)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *repository) CreateTeam(teamName, token, countryCode string) (uint32, error) {
	id := r.newID()
	_, err := r.db.Exec(
		`INSERT INTO
		teams(id, teamname, token, country_code)
		VALUES (?, ?, ?, ?)`,
		id, teamName, token, countryCode,
	)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok && mysqlerr.Number == 1062 {
			return 0, model.DuplicateError("team")
		}
		return 0, err
	}
	return id, nil
}

func (r *repository) SetCountryCode(tid uint32, countryCode string) error {
	_, err := r.db.Exec(
		`UPDATE teams
		SET country_code = ?
		WHERE id = ?`,
		countryCode, tid,
	)
	return err
}

func (r *repository) setUsers(team model.Team) (model.Team, error) {
	users := make([]*model.User, 0)
	err := r.db.Select(
		&users,
		`SELECT *
		FROM users
		WHERE team_id = ?`,
		team.ID,
	)

	if err != nil {
		return model.Team{}, fmt.Errorf("%w", err)
	}
	team.Users = users
	return team, nil
}

func (r *repository) setValidSubmission(team model.Team) (model.Team, error) {
	submissions := make([]*model.Submission, 0)
	err := r.db.Select(
		&submissions,
		`SELECT *
		FROM submissions
		WHERE team_id = ? AND is_valid = TRUE
		`,
		team.ID,
	)

	if err != nil {
		return model.Team{}, fmt.Errorf("%w", err)
	}
	team.Submissions = submissions
	return team, nil
}

func (r *repository) setValidSubmissions(teams []*model.Team) ([]*model.Team, error) {
	submissions := make([]*model.Submission, 0)
	err := r.db.Select(
		&submissions,
		`SELECT *
		FROM submissions
		WHERE is_valid = TRUE
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	submissionsMap := make(map[uint32][]*model.Submission)
	for i := 0; i < len(teams); i++ {
		submissionsMap[teams[i].ID] = make([]*model.Submission, 0)
	}

	for _, s := range submissions {
		submissionsMap[*s.TeamID] = append(submissionsMap[*s.TeamID], s)
	}

	for i := 0; i < len(teams); i++ {
		teams[i].Submissions = submissionsMap[teams[i].ID]
	}

	return teams, nil
}

func (r *repository) UpdateTeamName(tid uint32, teamName string) error {
	_, err := r.db.Exec(
		`UPDATE teams
		SET teamname = ?
		WHERE id = ?`,
		teamName, tid,
	)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok && mysqlerr.Number == 1062 {
			return model.DuplicateError("team")
		}
		return err
	}
	return nil
}
