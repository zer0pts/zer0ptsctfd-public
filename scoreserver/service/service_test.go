package service

import (
	"os"
	"testing"

	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/repository"
)

func newApp(t *testing.T) App {
	t.Helper()

	dbdsn := os.Getenv("DBDSN")
	if dbdsn == "" {
		t.Fatal("DBDSN not set")
	}

	repo, err := repository.New(dbdsn, nil)
	if err != nil {
		t.Fatal(err)
	}
	return New(repo, nil)
}
