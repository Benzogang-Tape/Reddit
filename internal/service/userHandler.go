package service

import (
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"github.com/pkg/errors"
)

type UserStorage interface {
	RegisterUser(models.AuthUserInfo) (*models.User, error)
	Authorize(models.AuthUserInfo) (*models.User, error)
}

type UserHandler struct {
	Repo UserStorage
}

func NewUserHandler(u UserStorage) *UserHandler {
	return &UserHandler{
		Repo: u,
	}
}

func (h *UserHandler) Register(authData models.AuthUserInfo) (models.TokenPayload, error) {
	user, err := h.Repo.RegisterUser(authData)
	if err != nil {
		err = errors.Wrap(err, "Register: ")
		return models.TokenPayload{}, err
	}
	return models.TokenPayload{
		Login: user.Username,
		ID:    user.ID,
	}, err
}

func (h *UserHandler) Authorize(authData models.AuthUserInfo) (models.TokenPayload, error) {
	user, err := h.Repo.Authorize(authData)
	if err != nil {
		err = errors.Wrap(err, "Authorize: ")
		return models.TokenPayload{}, err
	}
	return models.TokenPayload{
		Login: user.Username,
		ID:    user.ID,
	}, nil
}
