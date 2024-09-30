package rest

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type UserAPI interface {
	Register(models.AuthUserInfo) (models.TokenPayload, error)
	Authorize(models.AuthUserInfo) (models.TokenPayload, error)
}

type UserHandler struct {
	logger  *zap.SugaredLogger
	service UserAPI
}

func NewUserHandler(u UserAPI, logger *zap.SugaredLogger) *UserHandler {
	return &UserHandler{
		logger:  logger,
		service: u,
	}
}

func (h *UserHandler) registerUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	credentials := models.AuthUserInfo{}
	if err = json.Unmarshal(body, &credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := h.service.Register(credentials)
	if errors.Is(err, models.ErrUserExists) {
		jsonComplexErr(w, http.StatusUnprocessableEntity, models.NewComplexErr(models.ComplexErr{
			Location: `body`,
			Param:    `username`,
			Value:    `1`,
			Msg:      `already exists`,
		}))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newSession(w, r.WithContext(context.WithValue(r.Context(), models.Payload, payload)), http.StatusCreated)
	h.logger.Infow("New user has registered",
		"login", credentials.Login,
		"remote_addr", r.RemoteAddr,
		"url", r.URL.Path,
	)
}

func (h *UserHandler) loginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	credentials := models.AuthUserInfo{}
	if err = json.Unmarshal(body, &credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := h.service.Authorize(credentials)
	if errors.Is(err, models.ErrNoUser) {
		jsonSimpleErr(w, http.StatusUnauthorized, models.NewSimpleErr(models.ErrNoUser.Error()))
		return
	}
	if errors.Is(err, models.ErrBadPass) {
		jsonSimpleErr(w, http.StatusUnauthorized, models.NewSimpleErr(models.ErrBadPass.Error()))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	newSession(w, r.WithContext(context.WithValue(r.Context(), models.Payload, payload)), http.StatusOK)
	h.logger.Infow("New log in",
		"login", credentials.Login,
		"remote_addr", r.RemoteAddr,
		"url", r.URL.Path,
	)
}
