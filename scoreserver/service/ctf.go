package service

import (
	"time"

	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

const (
	CTFNotStartedYetMessage = "CTF has not started yet"
	CTFFinishedMessage      = "CTF has been finished"
)

type CTFApp interface {
	GetConfig() (*model.Config, error)
	SetCTFName(name string) error
	SetStartAt(t int64) error
	SetEndAt(t int64) error
	SetLock(second, count, duration int) error
	SetSolves(easy, medium int) error
	SetMinScore(score int) error
	CTFStarted(t time.Time) (bool, error)
	CTFFinished(t time.Time) (bool, error)
	CTFNowRunning(t time.Time) (bool, error)
}

func (app *app) GetConfig() (*model.Config, error) {
	config, err := app.repo.GetConfig()
	if err != nil {
		if model.IsNotFound(err) {
			return nil, ErrorMessage("ctf is not configured")
		}
		return nil, err
	}
	return config, nil
}

func (app *app) SetCTFName(name string) error {
	return app.repo.SetCTFName(name)
}

func (app *app) SetStartAt(t int64) error {
	return app.repo.SetStartAt(t)
}

func (app *app) SetEndAt(t int64) error {
	return app.repo.SetEndAt(t)
}

func (app *app) SetLock(second, count, duration int) error {
	return app.repo.SetLock(second, count, duration)
}
func (app *app) SetSolves(easy, medium int) error {
	return app.repo.SetSolves(easy, medium)
}
func (app *app) SetMinScore(score int) error {
	return app.repo.SetMinScore(score)
}

func (app *app) CTFStarted(t time.Time) (bool, error) {
	conf, err := app.GetConfig()
	if err != nil {
		return false, err
	}
	return conf.StartAt <= t.Unix(), nil
}

func (app *app) CTFFinished(t time.Time) (bool, error) {
	conf, err := app.GetConfig()
	if err != nil {
		return false, err
	}
	return conf.EndAt <= t.Unix(), nil
}

func (app *app) CTFNowRunning(t time.Time) (bool, error) {
	started, err := app.CTFStarted(t)
	if err != nil {
		return false, err
	}
	finished, err := app.CTFFinished(t)
	if err != nil {
		return false, err
	}
	return started && !finished, nil
}
