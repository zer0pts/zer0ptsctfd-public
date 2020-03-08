package service

import (
	"testing"
	"time"
)

func TestGetConfig(t *testing.T) {
	app := newApp(t)

	_, err := app.GetConfig()
	if err != nil {
		t.Error(err)
	}
}

func TestSetCTFName(t *testing.T) {
	app := newApp(t)

	err := app.SetCTFName("testsetctfname")
	if err != nil {
		t.Error(err)
	}

	conf, err := app.GetConfig()
	if err != nil {
		t.Error(err)
	}
	if conf.CTFName != "testsetctfname" {
		t.Error("failed to set ctf_name")
	}
}

func TestSetStartAt(t *testing.T) {
	app := newApp(t)

	now := time.Now().Unix()
	err := app.SetStartAt(now)
	if err != nil {
		t.Error(err)
	}

	conf, err := app.GetConfig()
	if err != nil {
		t.Error(err)
	}
	if conf.StartAt != now {
		t.Error("failed to set start_at")
	}
}

func TestSetEndAt(t *testing.T) {
	app := newApp(t)

	now := time.Now().Unix()
	err := app.SetEndAt(now)
	if err != nil {
		t.Error(err)
	}

	conf, err := app.GetConfig()
	if err != nil {
		t.Error(err)
	}
	if conf.EndAt != now {
		t.Error("failed to set end_at")
	}
}
