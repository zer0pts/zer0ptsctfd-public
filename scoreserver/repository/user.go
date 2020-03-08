package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

type UserRepository interface {
	RegisterUser(username, email, passwordHash string, tid uint32) (uint32, error)
	UpdateUserPassword(uid uint32, passwordHash string) error

	FindUserByID(id uint32) (*model.User, error)
	FindUserByName(username string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	FindUserByToken(token string) (*model.User, error)

	RevokeToken(token string) error
	RevokeTokenByUserID(uid uint32) error
	NewToken(uid uint32, token string, expiresAt uint64) error

	FindUserByPasswordResetToken(token string) (*model.User, error)
	RevokePasswordResetTokenByUserID(uid uint32) error
	NewPasswordResetToken(uid uint32, token string, expiresAt uint64) error
}

func (r *repository) RegisterUser(username, email, passwordHash string, tid uint32) (uint32, error) {
	id := r.newID()
	_, err := r.db.Exec(
		`INSERT INTO
		users(id, username, email, password_hash, icon_path, team_id, is_hidden, is_admin)
		VALUES (?, ?, ?, ?, NULL, ?, FALSE, FALSE)`,
		id, username, email, passwordHash, tid,
	)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok && mysqlerr.Number == 1062 {
			return 0, model.DuplicateError("user")
		}
		return 0, err
	}
	return id, nil
}

func (r *repository) UpdateUserPassword(uid uint32, passwordHash string) error {
	_, err := r.db.Exec(
		`UPDATE users
		SET password_hash = ?
		WHERE id = ?`,
		passwordHash, uid,
	)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (r *repository) FindUserByID(uid uint32) (*model.User, error) {
	var user model.User
	err := r.db.Get(
		&user,
		`SELECT id, username, email, password_hash, team_id, is_hidden, is_admin, icon_path
		FROM users
		WHERE id = ?
		LIMIT 1`,
		uid,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("user")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUserByName(username string) (*model.User, error) {
	var user model.User
	err := r.db.Get(
		&user,
		`SELECT id, username, email, password_hash, team_id, is_hidden, is_admin, icon_path
		FROM users
		WHERE username = ?
		LIMIT 1`,
		username,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("user")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Get(
		&user,
		`SELECT id, username, email, password_hash, team_id, is_hidden, is_admin, icon_path
		FROM users
		WHERE email = ?
		LIMIT 1`,
		email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("user")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUserByToken(token string) (*model.User, error) {
	var user model.User
	now := time.Now().Unix()

	err := r.db.Get(
		&user,
		`SELECT id, username, email, password_hash, team_id, is_hidden, is_admin, icon_path 
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		AND tokens.token = ?
		AND tokens.expires_at > ?
		AND tokens.revoked = FALSE
		LIMIT 1`,
		token, now,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("user")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) RevokeToken(token string) error {
	_, err := r.db.Exec(
		`UPDATE tokens
		SET revoked = TRUE
		WHERE token = ?`,
		token,
	)
	return err
}

func (r *repository) RevokeTokenByUserID(uid uint32) error {
	_, err := r.db.Exec(
		`UPDATE tokens
		SET revoked = TRUE
		WHERE user_id = ?`,
		uid,
	)
	return err
}

func (r *repository) NewToken(uid uint32, token string, expiresAt uint64) error {
	_, err := r.db.Exec(
		`INSERT INTO tokens(user_id, token, expires_at, revoked)
		VALUES (?, ?, ?, FALSE)`,
		uid, token, expiresAt,
	)
	return err
}

func (r *repository) FindUserByPasswordResetToken(token string) (*model.User, error) {
	var user model.User
	now := time.Now().Unix()

	err := r.db.Get(
		&user,
		`SELECT users.*
		FROM users
		INNER JOIN password_reset_tokens
		ON users.id = password_reset_tokens.user_id
		AND password_reset_tokens.token = ?
		AND password_reset_tokens.expires_at > ?
		AND password_reset_tokens.revoked = FALSE
		LIMIT 1`,
		token, now,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NotFoundError("user")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) RevokePasswordResetTokenByUserID(uid uint32) error {
	_, err := r.db.Exec(
		`UPDATE password_reset_tokens
		SET revoked = TRUE
		WHERE user_id = ?`,
		uid,
	)
	return err
}

func (r *repository) NewPasswordResetToken(uid uint32, token string, expiresAt uint64) error {
	_, err := r.db.Exec(
		`INSERT INTO password_reset_tokens(user_id, token, expires_at, revoked)
		VALUES (?, ?, ?, FALSE)`,
		uid, token, expiresAt,
	)
	return err
}
