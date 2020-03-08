package repository

import (
	redis "github.com/go-redis/redis/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	UserRepository
	TeamRepository
	ChallengeRepository
	ConfigRepository
	SubmissionRepository
}

type repository struct {
	db    *sqlx.DB
	redis *redis.Client
}

func New(dbDsn string, redis *redis.Client) (Repository, error) {
	db, err := sqlx.Open("mysql", dbDsn)
	if err != nil {
		return nil, err
	}

	return &repository{
		db:    db,
		redis: redis,
	}, nil
}

func (r *repository) newID() uint32 {
	return uuid.New().ID()
}
