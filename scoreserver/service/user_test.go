package service

import (
	"testing"
)

func TestRegister(t *testing.T) {
	app := newApp(t)

	var testCases = []struct {
		username    string
		email       string
		password    string
		teamname    string
		countrycode string
		hasError    bool
	}{
		{"testregister", "testregister@example.com", "password", "team-testregister", "JPN", false},
		{"testregister", "testregister@example.com", "password", "team-testregister", "JPN", true},
		{"testregister", "mogumogu@example.com", "password", "mogumogu-team", "JPN", true},
		{"mogumogu", "testregister@example.com", "password", "mogumogu-team", "JPN", true},
		{"mogumogu", "mogumogu@example.com", "password", "team-testregister", "JPN", true},
		{"", "mogumogu@example.com", "password", "mogumogu-team", "JPN", true},
		{"mogumogu", "", "password", "mogumogu-team", "JPN", true},
		{"mogumogu", "mogumogu@example.com", "", "mogumogu-team", "JPN", true},
		{"mogumogu", "mogumogu@example.com", "password", "", "JPN", true},
		{"mogumogu", "mogumogu@example.com", "password", "mogumogu-team", "japan", true},
		{"mogumogu", "mogumogu@example.com", "password", "mogumogu-team", "", false},
		{"あいうえお", "aiueo@example.com", "password", "aiueo", "", true},
	}

	for _, c := range testCases {
		err := app.RegisterUserCreateTeam(c.username, c.email, c.password, c.teamname, c.countrycode)
		if c.hasError != (err != nil) {
			t.Errorf("case %v, err: %v", c, err)
		}
	}
}

func TestLogin(t *testing.T) {
	app := newApp(t)

	err := app.RegisterUserCreateTeam("testlogin", "testlogin@example.com", "password", "team-testlogin", "JPN")
	if err != nil {
		t.Error(err)
	}

	var testCases = []struct {
		username string
		password string
		hasError bool
	}{
		{"testlogin", "password", false},
		{"", "password", true},
		{"testlogin", "", true},
		{"", "", true},
		{"testlogin", "wrongpassword", true},
		{"wrogusername", "password", true},
	}
	for _, c := range testCases {
		_, _, err := app.LoginUser(c.username, c.password)
		if c.hasError != (err != nil) {
			t.Errorf("case %v, err: %v", c, err)
		}
	}
}

func TestToken(t *testing.T) {
	app := newApp(t)

	err := app.RegisterUserCreateTeam("testtoken", "testtoken@example.com", "password", "team-testtoken", "JPN")
	if err != nil {
		t.Error(err)
	}
	user, token, err := app.LoginUser("testtoken", "password")
	if err != nil {
		t.Error(err)
	}

	user2, err := app.GetLoginUser(token)
	if err != nil {
		t.Error(err)
	}
	if user.ID != user2.ID {
		t.Error("wrong user found")
	}
}

func TestLogout(t *testing.T) {
	app := newApp(t)

	err := app.RegisterUserCreateTeam("testlogout", "testlogout@example.com", "password", "team-testlogout", "JPN")
	if err != nil {
		t.Error(err)
	}
	user, token, err := app.LoginUser("testlogout", "password")
	if err != nil {
		t.Error(err)
	}

	_, err = app.GetLoginUser(token)
	if err != nil {
		t.Error(err)
	}

	err = app.LogoutUser(user.ID)
	if err != nil {
		t.Error(err)
	}

	_, err = app.GetLoginUser(token)
	if err == nil {
		t.Error("logout failed")
	}
}
