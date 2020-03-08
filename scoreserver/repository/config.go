package repository

import (
	"database/sql"
	"fmt"

	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

type ConfigRepository interface {
	SetCTFName(ctfName string) error
	SetStartAt(t int64) error
	SetEndAt(t int64) error
	SetLock(second, count, duration int) error
	SetSolves(easy, medium int) error
	SetMinScore(score int) error
	GetConfig() (*model.Config, error)
}

func (r *repository) SetCTFName(ctfName string) error {
	_, err := r.db.Exec(
		`UPDATE config
		SET ctf_name = ?`,
		ctfName,
	)
	return err
}

func (r *repository) SetStartAt(t int64) error {
	_, err := r.db.Exec(
		`UPDATE config
		SET start_at = from_unixtime(?)`,
		t,
	)
	return err
}

func (r *repository) SetEndAt(t int64) error {
	_, err := r.db.Exec(
		`UPDATE config
		SET end_at = from_unixtime(?)`,
		t,
	)
	return err
}

func (r *repository) SetLock(second, count, duration int) error {
	_, err := r.db.Exec(
		`UPDATE config
		SET lock_second = ?, lock_duration = ?, lock_count = ?`,
		second, duration, count,
	)
	return err
}
func (r *repository) SetSolves(easy, medium int) error {
	_, err := r.db.Exec(
		`UPDATE config
		SET easy_solves = ?, medium_solves = ?`,
		easy, medium,
	)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (r *repository) SetMinScore(score int) error {
	_, err := r.db.Exec(
		`UPDATE config
		SET min_score = ?`,
		score,
	)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (r *repository) GetConfig() (*model.Config, error) {
	var config model.Config
	err := r.db.Get(
		&config,
		`SELECT ctf_name, unix_timestamp(start_at) as start_at, unix_timestamp(end_at) as end_at, lock_second, lock_duration, lock_count, easy_solves, medium_solves, min_score 
		FROM config
		LIMIT 1`,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("config")
		}
		return nil, err
	}
	return &config, nil
}
