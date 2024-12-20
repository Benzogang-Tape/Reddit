package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

type UserStorage interface {
	RegisterUser(ctx context.Context, authData users.AuthUserInfo) (*users.User, error)
	Authorize(ctx context.Context, authData users.AuthUserInfo) (*users.User, error)
}

type UserHandler struct {
	Repo UserStorage
}

func NewUserHandler(u UserStorage) *UserHandler {
	return &UserHandler{
		Repo: u,
	}
}

func (h *UserHandler) Register(ctx context.Context, authData users.AuthUserInfo) (*jwt.TokenPayload, error) {
	source := "Register"
	newUser, err := h.Repo.RegisterUser(ctx, authData)
	if err != nil {
		err = errors.Wrap(err, source)
		return nil, err
	}

	return &jwt.TokenPayload{
		Login: newUser.Username,
		ID:    newUser.ID,
	}, nil
}

func (h *UserHandler) Authorize(ctx context.Context, authData users.AuthUserInfo) (*jwt.TokenPayload, error) {
	source := "Authorize"
	user, err := h.Repo.Authorize(ctx, authData)
	if err != nil {
		err = errors.Wrap(err, source)
		return nil, err
	}

	return &jwt.TokenPayload{
		Login: user.Username,
		ID:    user.ID,
	}, nil
}
