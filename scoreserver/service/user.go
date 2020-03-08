package service

import (
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
	"golang.org/x/crypto/bcrypt"
)

const UsernameMaxLength = 64
const TokenLimit = time.Hour * 24 * 7
const PasswordResetTokenLimit = time.Hour * 1

type UserApp interface {
	JoinUserToTeam(username, email, password, token string) error
	RegisterUserCreateTeam(username, email, password, teamName, countryCode string) error
	LoginUser(username, password string) (*model.User, string, error)
	LogoutUser(uid uint32) error
	LogoutUserByToken(token string) error
	GetLoginUser(token string) (*model.User, error)

	IssuePasswordResetToken(email string) error
	ResetPassword(token, password string) error
}

func (app *app) checkUserAvailable(username, email, password string) error {
	if username == "" {
		return ErrorMessage("username is required")
	}
	if email == "" {
		return ErrorMessage("email is required")
	}
	if password == "" {
		return ErrorMessage("password is required")
	}

	//(username must be matched to ^[0-9A-Za-z_-]{1, 32}$
	if len(username) > UsernameMaxLength {
		return ErrorMessage("username must follow the regex: ^[0-9A-Za-z_-]{1, 32}$")
	}
	for _, c := range username {
		if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '_' || c == '-') {
			return ErrorMessage("username must follow the regex: ^[0-9A-Za-z_-]{1, 32}$")
		}
	}
	_, err := app.repo.FindUserByName(username)
	if err != nil && !model.IsNotFound(err) {
		return err
	}
	if err == nil {
		return ErrorMessage("username already used")
	}

	_, err = app.repo.FindUserByEmail(email)
	if err != nil && !model.IsNotFound(err) {
		return err
	}
	if err == nil {
		return ErrorMessage("email already used")
	}
	return nil
}

func (app *app) registerUser(username, email, password string, tid uint32) (uint32, error) {
	sha256password := sha256.Sum256([]byte(password))
	passwordHash, err := bcrypt.GenerateFromPassword(sha256password[:], bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	uid, err := app.repo.RegisterUser(username, email, string(passwordHash), tid)
	if err != nil {
		if model.IsDuplicated(err) {
			return 0, ErrorMessage("username already used")
		}
		return 0, err
	}
	return uid, nil
}

func (app *app) JoinUserToTeam(username, email, password, token string) error {
	t, err := app.repo.FindTeamByToken(token)
	if err != nil {
		if model.IsNotFound(err) {
			return ErrorMessage("invalid token")
		}
		return err
	}

	err = app.checkUserAvailable(username, email, password)
	if err != nil {
		return err
	}
	_, err = app.registerUser(username, email, password, t.ID)
	if err != nil {
		return err
	}

	return nil
}

func (app *app) RegisterUserCreateTeam(username, email, password, teamName, countryCode string) error {
	err := app.checkUserAvailable(username, email, password)
	if err != nil {
		return err
	}
	err = app.checkTeamAvailable(teamName)
	if err != nil {
		return err
	}
	code, err := app.validateCountryCode(countryCode)
	if err != nil {
		return err
	}

	tid, err := app.createTeam(teamName, code)
	if err != nil {
		return err
	}

	_, err = app.registerUser(username, email, password, tid)
	if err != nil {
		return err
	}
	return nil
}

func (app *app) LoginUser(username, password string) (*model.User, string, error) {
	user, err := app.repo.FindUserByName(username)
	if err != nil {
		return nil, "", ErrorMessage("wrong username")
	}

	sha256password := sha256.Sum256([]byte(password))
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), sha256password[:]); err != nil {
		return nil, "", ErrorMessage("wrong password")
	}

	token := app.newToken()
	if err := app.repo.NewToken(user.ID, token, uint64(time.Now().Add(TokenLimit).Unix())); err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (app *app) LogoutUser(uid uint32) error {
	return app.repo.RevokeTokenByUserID(uid)
}
func (app *app) LogoutUserByToken(token string) error {
	return app.repo.RevokeToken(token)
}

func (app *app) GetLoginUser(token string) (*model.User, error) {
	user, err := app.repo.FindUserByToken(token)
	if err != nil {
		if model.IsNotFound(err) {
			return nil, ErrorMessage("invalid token")
		}
		return nil, err
	}
	return user, nil
}
func (app *app) IssuePasswordResetToken(email string) error {
	user, err := app.repo.FindUserByEmail(email)
	if err != nil {
		if model.IsNotFound(err) {
			return ErrorMessage("invalid email")
		}
		return err
	}

	token := uuid.New().String()
	err = app.repo.NewPasswordResetToken(user.ID, token, uint64(time.Now().Add(PasswordResetTokenLimit).Unix()))
	if err != nil {
		return err
	}

	go func() {
		err := app.mailer.Send(email, "password reset token", fmt.Sprintf("your password reset token is: %s", token))
		if err != nil {
			log.Println(err)
		}
	}()
	return nil
}

func (app *app) ResetPassword(token, password string) error {
	if password == "" {
		return ErrorMessage("password is required")
	}

	user, err := app.repo.FindUserByPasswordResetToken(token)
	if err != nil {
		if model.IsNotFound(err) {
			return ErrorMessage("invalid token")
		}
		return err
	}

	sha256password := sha256.Sum256([]byte(password))
	passwordHash, err := bcrypt.GenerateFromPassword(sha256password[:], bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = app.repo.UpdateUserPassword(user.ID, string(passwordHash))
	if err != nil {
		return err
	}

	err = app.repo.RevokePasswordResetTokenByUserID(user.ID)
	if err != nil {
		return err
	}

	return nil
}
