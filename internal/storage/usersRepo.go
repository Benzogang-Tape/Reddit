package storage

import (
	"database/sql"
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"github.com/pkg/errors"
)

//type UserRepo struct {
//	storage map[models.Username]*models.User
//	mu      *sync.RWMutex
//}

type UserRepo struct {
	db *sql.DB
}

//func NewUserRepo() *UserRepo {
//	return &UserRepo{
//		storage: make(map[models.Username]*models.User, 42),
//		mu:      &sync.RWMutex{},
//	}
//}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

//func (repo *UserRepo) Authorize(authData models.AuthUserInfo) (*models.User, error) {
//	repo.mu.RLock()
//	user, ok := repo.storage[authData.Login]
//	repo.mu.RUnlock()
//
//	if !ok {
//		return nil, errors.Wrap(models.ErrNoUser, "Authorize: ")
//	}
//	if user.Password != authData.Password {
//		return nil, errors.Wrap(models.ErrBadPass, "Authorize: ")
//	}
//	return user, nil
//}

func (repo *UserRepo) Authorize(authData models.AuthUserInfo) (*models.User, error) {
	user := &models.User{}
	err := repo.db.
		QueryRow("SELECT uuid, login, password FROM users WHERE login = ?", authData.Login).
		Scan(&user.ID, &user.Username, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(models.ErrNoUser, "Authorize: ")
	}
	if err != nil {
		return nil, err
	}

	if user.Password != authData.Password {
		return nil, errors.Wrap(models.ErrBadPass, "Authorize: ")
	}
	return user, nil
}

//func (repo *UserRepo) RegisterUser(authData models.AuthUserInfo) (*models.User, error) {
//	repo.mu.RLock()
//	_, ok := repo.storage[authData.Login]
//	repo.mu.RUnlock()
//	if ok {
//		return nil, errors.Wrap(models.ErrUserExists, "Register: ")
//	}
//
//	newUser, err := repo.createUser(authData)
//	if err != nil {
//		return nil, err
//	}
//	return newUser, nil
//}

func (repo *UserRepo) RegisterUser(authData models.AuthUserInfo) (*models.User, error) {
	user := &models.User{}
	err := repo.db.
		QueryRow("SELECT uuid, login, password FROM users WHERE login = ?", authData.Login).
		Scan(&user.ID, &user.Username, &user.Password)
	if err == nil {
		return nil, errors.Wrap(models.ErrUserExists, "Register: ")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	newUser, err := repo.createUser(authData)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

//func (repo *UserRepo) createUser(authData models.AuthUserInfo) (*models.User, error) {
//	repo.mu.Lock()
//	defer repo.mu.Unlock()
//	newUser, err := models.NewUser(authData)
//	if err != nil {
//		return nil, err
//	}
//	repo.storage[newUser.Username] = newUser
//	return newUser, nil
//}

func (repo *UserRepo) createUser(authData models.AuthUserInfo) (*models.User, error) {
	newUser, err := models.NewUser(authData)
	if err != nil {
		return nil, err
	}
	_, err = repo.db.Exec("INSERT INTO users (`uuid`, `login` `password`) VALUES (?, ?)", newUser.ID, newUser.Username, newUser.Password)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}
