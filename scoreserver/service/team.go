package service

import (
	"time"

	"golang.org/x/exp/utf8string"

	"github.com/pariz/gountries"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

const TeamNameMaxLength = 32

type TeamApp interface {
	GetUserTeam(uid uint32) (*model.Team, error)
	GetTeam(id uint32) (*model.Team, error)

	AddWrongCount(tid uint32, expire time.Duration) (int, error)
	GetWrongCount(tid uint32) (int, error)
	LockSubmission(tid uint32, duration time.Duration) error

	UpdateTeamName(tid uint32, newName string) error
	UpdateTeamCountry(tid uint32, countryCode string) error

	GetTeams() ([]*model.Team, error)
}

func (app *app) GetUserTeam(uid uint32) (*model.Team, error) {
	t, err := app.repo.FindUserTeam(uid)
	if err != nil {
		if model.IsNotFound(err) {
			return nil, ErrorMessage("team not found")
		}
		return nil, err
	}
	return t, nil
}

func (app *app) GetTeam(id uint32) (*model.Team, error) {
	team, err := app.repo.FindTeamByID(id)
	if err != nil {
		if model.IsNotFound(err) {
			return nil, ErrorMessage("team not found")
		}
		return nil, err
	}
	return team, nil
}

func (app *app) UpdateTeamName(tid uint32, newName string) error {
	err := app.checkTeamAvailable(newName)
	if err != nil {
		return err
	}
	err = app.repo.UpdateTeamName(tid, newName)
	if err != nil {
		if model.IsDuplicated(err) {
			return ErrorMessage("teamname already used")
		}
		return err
	}
	return nil
}

func (app *app) UpdateTeamCountry(tid uint32, countryCode string) error {
	code, err := app.validateCountryCode(countryCode)
	if err != nil {
		return err
	}
	err = app.repo.SetCountryCode(tid, code)
	if err != nil {
		return err
	}
	return nil
}

func (app *app) AddWrongCount(tid uint32, expire time.Duration) (int, error) {
	return app.repo.IncrementWrong(tid, expire)
}

func (app *app) GetWrongCount(tid uint32) (int, error) {
	return app.repo.GetWrongCount(tid)
}

func (app *app) LockSubmission(tid uint32, duration time.Duration) error {
	return app.repo.LockSubmission(tid, duration)
}

func (app *app) GetTeams() ([]*model.Team, error) {
	teams, err := app.repo.ListTeams(true)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (app *app) checkTeamAvailable(teamName string) error {
	if teamName == "" {
		return ErrorMessage("teamname is required")
	}

	if len(teamName) > TeamNameMaxLength {
		return ErrorMessage("teamname too long")
	}
	if !utf8string.NewString(teamName).IsASCII() {
		return ErrorMessage("teamname is not ASCII")
	}

	_, err := app.repo.FindTeamByName(teamName)
	if err != nil && !model.IsNotFound(err) {
		return err
	}
	if err == nil {
		return ErrorMessage("teamname already used")
	}

	return nil
}

func (app *app) validateCountryCode(countryCode string) (string, error) {
	if countryCode == "" {
		return "", nil
	}
	q := gountries.New()
	c, err := q.FindCountryByAlpha(countryCode)
	if err != nil {
		return "", ErrorMessage("invalid country code. please follow ISO 3166-1 alpha-3")
	}
	return c.Alpha3, nil
}

func (app *app) createTeam(teamName, countryCode string) (uint32, error) {
	tid, err := app.repo.CreateTeam(teamName, app.newToken(), countryCode)
	if err != nil {
		if model.IsDuplicated(err) {
			return 0, ErrorMessage("teamname already used")
		}
		return 0, err
	}
	return tid, nil
}
