package inmem

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

type UserRepo struct {
	storage map[users.Username]*users.User
	mu      *sync.RWMutex
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		storage: make(map[users.Username]*users.User, 42),
		mu:      &sync.RWMutex{},
	}
}

func (repo *UserRepo) Authorize(ctx context.Context, authData users.AuthUserInfo) (*users.User, error) { //nolint:unparam
	source := "Authorize"
	repo.mu.RLock()
	user, ok := repo.storage[authData.Login]
	repo.mu.RUnlock()

	if !ok {
		return nil, errors.Wrap(errs.ErrNoUser, source)
	}
	if user.Password != authData.Password {
		return nil, errors.Wrap(errs.ErrBadPass, source)
	}

	return user, nil
}

func (repo *UserRepo) RegisterUser(ctx context.Context, authData users.AuthUserInfo) (*users.User, error) { //nolint:unparam
	source := "RegisterUser"
	repo.mu.RLock()
	_, ok := repo.storage[authData.Login]
	repo.mu.RUnlock()
	if ok {
		return nil, errors.Wrap(errs.ErrUserExists, source)
	}

	newUser := repo.createUser(authData)

	return newUser, nil
}

func (repo *UserRepo) createUser(authData users.AuthUserInfo) *users.User {
	newUser := users.NewUser(authData)
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.storage[newUser.Username] = newUser

	return newUser
}
