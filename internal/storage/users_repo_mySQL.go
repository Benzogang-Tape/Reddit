package storage

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

type UserRepoMySQL struct {
	db *sql.DB
}

func NewUserRepoMySQL(db *sql.DB) *UserRepoMySQL {
	return &UserRepoMySQL{
		db: db,
	}
}

func (repo *UserRepoMySQL) Authorize(ctx context.Context, authData users.AuthUserInfo) (*users.User, error) { //nolint:unparam
	source := "Authorize"
	user := &users.User{}
	err := repo.db.
		QueryRow(
			"SELECT uuid, login, password FROM users WHERE login = ?",
			authData.Login,
		).Scan(&user.ID, &user.Username, &user.Password)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, errors.Wrap(errs.ErrNoUser, source)
	case err != nil:
		return nil, err
	}

	if user.Password != authData.Password {
		return nil, errors.Wrap(errs.ErrBadPass, source)
	}

	return user, nil
}

func (repo *UserRepoMySQL) RegisterUser(ctx context.Context, authData users.AuthUserInfo) (*users.User, error) { //nolint:unparam
	source := "RegisterUser"
	var userExists bool
	err := repo.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)",
		authData.Login,
	).Scan(&userExists)

	switch {
	case err != nil:
		return nil, err
	case userExists:
		return nil, errors.Wrap(errs.ErrUserExists, source)
	}

	newUser, err := repo.createUser(authData)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (repo *UserRepoMySQL) createUser(credentials users.AuthUserInfo) (*users.User, error) {
	newUser := users.NewUser(credentials)
	if _, err := repo.db.Exec(
		"INSERT INTO users (`uuid`, `login`, `password`) VALUES (?, ?, ?)",
		newUser.ID,
		newUser.Username,
		newUser.Password,
	); err != nil {
		return nil, err
	}

	return newUser, nil
}
