package storage

import (
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"github.com/pkg/errors"
	"sync"
)

type UserRepo struct {
	storage map[models.Username]*models.User
	mu      *sync.RWMutex
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		storage: make(map[models.Username]*models.User, 42),
		mu:      &sync.RWMutex{},
	}
}

func (repo *UserRepo) Authorize(authData models.AuthUserInfo) (*models.User, error) {
	repo.mu.RLock()
	user, ok := repo.storage[authData.Login]
	repo.mu.RUnlock()

	if !ok {
		return nil, errors.Wrap(models.ErrNoUser, "Authorize: ")
	}
	if user.Password != authData.Password {
		return nil, errors.Wrap(models.ErrBadPass, "Authorize: ")
	}
	return user, nil
}

func (repo *UserRepo) RegisterUser(authData models.AuthUserInfo) (*models.User, error) {
	repo.mu.RLock()
	_, ok := repo.storage[authData.Login]
	repo.mu.RUnlock()
	if ok {
		return nil, errors.Wrap(models.ErrUserExists, "Register: ")
	}

	newUser, err := repo.createUser(authData)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (repo *UserRepo) createUser(authData models.AuthUserInfo) (*models.User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	newUser, err := models.NewUser(authData)
	if err != nil {
		return nil, err
	}
	repo.storage[newUser.Username] = newUser
	return newUser, nil
}
