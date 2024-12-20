package storage

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
	"github.com/Benzogang-Tape/Reddit/internal/storage"
)

var (
	expectedUsers = []*users.User{
		{
			ID:       "ffffffff-ffff-ffff-ffff-ffffffffffff",
			Username: "admin",
			Password: "rootroot",
		},
	}
	authData = users.AuthUserInfo{
		Login:    "admin",
		Password: "rootroot",
	}
)

func TestAuthorize(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	userRepoMySQLMock := storage.NewUserRepoMySQL(db)

	rows := sqlmock.NewRows([]string{"uuid", "login", "password"})
	for _, row := range expectedUsers {
		rows.AddRow(row.ID, row.Username, row.Password)
	}

	// Success
	mock.ExpectQuery("SELECT uuid, login, password FROM users WHERE login = ?").
		WithArgs(authData.Login).
		WillReturnRows(rows)

	user, err := userRepoMySQLMock.Authorize(context.Background(), authData)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, expectedUsers[0], user)

	// No rows
	mock.ExpectQuery("SELECT uuid, login, password FROM users WHERE login = ?").
		WithArgs(authData.Login).
		WillReturnError(sql.ErrNoRows)

	_, err = userRepoMySQLMock.Authorize(context.Background(), authData)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.ErrorIs(t, err, errs.ErrNoUser)

	// Invalid password
	rows = sqlmock.NewRows([]string{"uuid", "login", "password"})
	for _, row := range expectedUsers {
		rows.AddRow(row.ID, row.Username, row.Password)
	}

	authData.Password = "Bad password"
	mock.ExpectQuery("SELECT uuid, login, password FROM users WHERE login = ?").
		WithArgs(authData.Login).
		WillReturnRows(rows)

	_, err = userRepoMySQLMock.Authorize(context.Background(), authData)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.ErrorIs(t, err, errs.ErrBadPass)
	authData.Password = "rootroot"

	// Row scan error
	rows = sqlmock.NewRows([]string{"id", "uuid"}).
		AddRow(1, "54")

	mock.ExpectQuery("SELECT uuid, login, password FROM users WHERE login = ?").
		WithArgs(authData.Login).
		WillReturnRows(rows)

	_, err = userRepoMySQLMock.Authorize(context.Background(), authData)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, err, errors.New("sql: expected 2 destination arguments in Scan, not 3"))
}

func TestRegisterUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	userRepoMySQLMock := storage.NewUserRepoMySQL(db)

	rows := sqlmock.NewRows([]string{"uuid", "login", "password"})
	for _, row := range expectedUsers {
		rows.AddRow(row.ID, row.Username, row.Password)
	}

	// Already exists
	response := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
		WithArgs(authData.Login).
		WillReturnRows(response)

	_, err = userRepoMySQLMock.RegisterUser(context.Background(), authData)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.ErrorIs(t, err, errs.ErrUserExists)

	// Row scan error
	rows = sqlmock.NewRows([]string{"id", "uuid"}).AddRow(1, "54")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
		WithArgs(authData.Login).
		WillReturnRows(rows)

	_, err = userRepoMySQLMock.RegisterUser(context.Background(), authData)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, err, errors.New("sql: expected 2 destination arguments in Scan, not 1"))

	rows = sqlmock.NewRows([]string{"uuid", "login", "password"})
	response = sqlmock.NewRows([]string{"exists"}).AddRow(false)

	// createUser error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
		WithArgs(authData.Login).
		WillReturnRows(response)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (`uuid`, `login`, `password`) VALUES (?, ?, ?)")).
		WithArgs(sqlmock.AnyArg(), authData.Login, authData.Password).
		WillReturnError(errors.New("db_error"))

	_, err = userRepoMySQLMock.RegisterUser(context.Background(), authData)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.EqualError(t, err, "db_error")

	rows = sqlmock.NewRows([]string{"uuid", "login", "password"})
	response = sqlmock.NewRows([]string{"exists"}).AddRow(false)

	// Success
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
		WithArgs(authData.Login).
		WillReturnRows(response)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (`uuid`, `login`, `password`) VALUES (?, ?, ?)")).
		WithArgs(sqlmock.AnyArg(), authData.Login, authData.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := userRepoMySQLMock.RegisterUser(context.Background(), authData)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, authData.Login, user.Username)
	assert.Equal(t, authData.Password, user.Password)
}

//func TestNewUserRepoMySQL(t *testing.T) {
//	db, _, err := sqlmock.New()
//	if err != nil {
//		t.Fatalf("cant create mock: %s", err)
//	}
//	defer db.Close()
//
//	userRepoMySQLMock := storage.NewUserRepoMySQL(db)
//	assert.NotNil(t, userRepoMySQLMock)
//	assert.Equal(t, userRepoMySQLMock, &storage.UserRepoMySQL{db: db})
//}
