package service

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

type ChallengeApp interface {
	GetChallenge(id uint32) (*model.Challenge, error)
	ListAllChallenges() ([]*model.Challenge, error)
	ListOpenChallenges() ([]*model.Challenge, error)

	OpenChallenge(id uint32) error
	CloseChallenge(id uint32) error

	SubmitFlag(user *model.User, flag string) (*model.Challenge, bool, error)
	RecalcScore(min, max, e, m int, cid uint32) error

	TeamSolvedChallengeIDs(tid uint32) ([]uint32, error)

	CheckSubmittable(tid uint32) (bool, error)
}

func (app *app) GetChallenge(id uint32) (*model.Challenge, error) {
	chal, err := app.repo.FindChallengeByID(id)
	if err != nil {
		if model.IsNotFound(err) {
			return nil, ErrorMessage("challenge not found")
		}
		return nil, err
	}
	return chal, nil
}

func (app *app) ListAllChallenges() ([]*model.Challenge, error) {
	chals, err := app.repo.ListAllChallenges(false)
	if err != nil {
		return nil, err
	}
	return chals, nil
}
func (app *app) ListOpenChallenges() ([]*model.Challenge, error) {
	chals, err := app.repo.ListAllChallenges(true)
	if err != nil {
		return nil, err
	}
	return chals, nil
}
func (app *app) OpenChallenge(id uint32) error {
	return app.repo.OpenChallenge(id)
}
func (app *app) CloseChallenge(id uint32) error {
	return app.repo.CloseChallenge(id)
}

func (app *app) SubmitFlag(user *model.User, flag string) (*model.Challenge, bool, error) {
	t := time.Now()
	started, err := app.CTFStarted(t)
	if err != nil {
		return nil, false, err
	}
	if !started {
		return nil, false, ErrorMessage(CTFNotStartedYetMessage)
	}

	team, err := app.repo.FindUserTeam(user.ID)
	if err != nil {
		// team should be found
		return nil, false, err
	}

	var (
		cid     sql.NullInt64 = sql.NullInt64{Int64: 0, Valid: false}
		uid     sql.NullInt64 = sql.NullInt64{Int64: int64(user.ID), Valid: true}
		tid     sql.NullInt64 = sql.NullInt64{Int64: int64(team.ID), Valid: true}
		correct bool          = false
		valid   bool          = false
	)

	// check flag is correct
	finished, err := app.CTFFinished(t)
	if err != nil {
		return nil, false, err
	}

	flag = strings.Trim(flag, " \t")
	chal, err := app.repo.FindOpenChallengeByFlag(flag)
	if err != nil && !model.IsNotFound(err) {
		return nil, false, err
	}
	if !model.IsNotFound(err) {
		correct = true
		cid = sql.NullInt64{Int64: int64(chal.ID), Valid: true}
	}

	// check validity only when flag is correct
	if correct && !user.IsHidden && !team.IsHidden && !finished {
		_, err = app.repo.FindValidSubmission(team.ID, chal.ID)
		if err != nil && !model.IsNotFound(err) {
			return nil, false, err
		}
		valid = model.IsNotFound(err)
	}

	err = app.repo.InsertSubmission(cid, uid, tid, flag, t.Unix(), correct, valid)
	if err != nil {
		return nil, false, err
	}

	if !valid && !correct {
		if err := app.webhook.Send(fmt.Sprintf("`%s@%s` send flag `%s`, but wrong", user.Username, team.Teamname, flag)); err != nil {
			log.Println(err)
		}
	} else if !valid && correct {
		/*
			if err := app.webhook.Send(fmt.Sprintf("`%s@%s` send flag `%s` and solved `%s` but already solved", user.Username, team.Teamname, flag, chal.Name)); err != nil {
				log.Println(err)
			}
		*/
	} else {
		if err := app.webhook.Send(fmt.Sprintf("`%s@%s` send flag `%s` and solved `%s` :100:", user.Username, team.Teamname, flag, chal.Name)); err != nil {
			log.Println(err)
		}
		app.repo.AddSolvedChallenge(uint32(tid.Int64), uint32(cid.Int64))
	}

	return chal, valid, nil
}

func (app *app) RecalcScore(min, max, e, m int, cid uint32) error {
	submissions, err := app.repo.ListValidSubmission(cid)
	if err != nil {
		return err
	}

	s := float64(len(submissions))
	v := float64(e-m*m) / float64(2*m-e-1)
	k := 450.0 * math.Log(2.0) / math.Log((float64(m)+v)/(1.0+v))
	p := int(math.Min(math.Max(float64(min), float64(max)-k*math.Log2(float64(s+float64(v))/(1.0+v))), float64(max)))

	return app.repo.UpdateScore(cid, p)
}

func (app *app) TeamSolvedChallengeIDs(tid uint32) ([]uint32, error) {
	return app.repo.TeamSolvedChallenges(tid)
}

func (app *app) CheckSubmittable(tid uint32) (bool, error) {
	return app.repo.CheckSubmitAvailable(tid)
}
