package service

import (
	redis "github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/mailer"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/repository"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/webhook"
)

type App interface {
	UserApp
	TeamApp
	CTFApp
	ChallengeApp
	MessageApp
}

type app struct {
	*messageApp
	repo    repository.Repository
	redis   *redis.Client
	mailer  mailer.Mailer
	webhook webhook.Webhook
}

func New(repo repository.Repository, redis *redis.Client, mailer mailer.Mailer, webhook webhook.Webhook) App {
	return &app{
		repo:       repo,
		redis:      redis,
		messageApp: newMessageApp(),
		mailer:     mailer,
		webhook:    webhook,
	}
}

func (app *app) newToken() string {
	return uuid.New().String()
}
