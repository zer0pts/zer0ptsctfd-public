package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	redis "github.com/go-redis/redis/v7"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/mailer"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/repository"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/server"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/service"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/webhook"
)

func run() error {
	dbdsn := os.Getenv("DBDSN")
	if dbdsn == "" {
		return fmt.Errorf("Environmental variable 'DBDSN' is required")
	}

	raddr := os.Getenv("REDIS")
	if raddr == "" {
		return fmt.Errorf("Environmental variable 'REDIS' is required")
	}

	redis := redis.NewClient(&redis.Options{
		Addr: raddr,
	})

	email := os.Getenv("EMAIL")
	mailAccount := strings.Split(email, "/")
	if len(mailAccount) != 3 {
		return fmt.Errorf("Environmental vairable 'EMAIL' is required and format is <smtp server>:<port>/<email>/<password>")
	}
	mailer, err := mailer.New(mailAccount[0], mailAccount[1], mailAccount[2])
	if err != nil {
		return err
	}

	endpoint := os.Getenv("WEBHOOK")
	if endpoint == "" {
		return fmt.Errorf("Environmental variable 'WEBHOOK' is required")
	}
	webhook := webhook.New(endpoint)

	frontOrigin := os.Getenv("FRONT")
	if frontOrigin == "" {
		return fmt.Errorf("Environmental variable 'FRONT' is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	repo, err := repository.New(dbdsn, redis)
	if err != nil {
		return err
	}
	app := service.New(repo, redis, mailer, webhook)
	srv := server.New(app, []string{frontOrigin})
	go app.HandleMessage()
	return srv.Start(":" + port)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
