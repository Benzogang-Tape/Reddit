package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/httpresp"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
	"github.com/Benzogang-Tape/Reddit/internal/service"
)

//go:generate mockgen -source=user.go -destination=../../storage/mocks/users_repo_mySQL_mock.go -package=mocks UserAPI
type UserAPI interface {
	Register(ctx context.Context, authData users.AuthUserInfo) (*jwt.TokenPayload, error)
	Authorize(ctx context.Context, authData users.AuthUserInfo) (*jwt.TokenPayload, error)
}

type UserHandler struct {
	logger   *zap.SugaredLogger
	service  UserAPI
	sessMngr service.SessionAPI
}

func NewUserHandler(u UserAPI, s service.SessionAPI, logger *zap.SugaredLogger) *UserHandler {
	return &UserHandler{
		logger:   logger,
		service:  u,
		sessMngr: s,
	}
}

// RegisterUser godoc
//
//	@Summary		Register a new user
//	@Description	Register in reddit-clone app
//	@Tags			auth
//	@ID				register-user
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		users.AuthUserInfo	true	"User credentials for registration"
//	@Success		201			{object}	jwt.Session			"User registered successfully"
//	@Failure		400			"Bad request"
//	@Failure		422			{object}	errs.ComplexErrArr	"User already exists"
//	@Failure		500			{object}	errs.SimpleErr		"Internal server error"
//	@Router			/register [post]
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	credentials := users.AuthUserInfo{}
	if err = json.Unmarshal(body, &credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := h.service.Register(r.Context(), credentials)
	switch {
	case errors.Is(err, errs.ErrUserExists):
		sendErrorResponse(w, http.StatusUnprocessableEntity, errs.NewComplexErrArr(errs.ComplexErr{
			Location: `body`,
			Param:    `username`,
			Value:    `1`,
			Msg:      `already exists`,
		}))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	h.newSession(w, r.WithContext(context.WithValue(r.Context(), jwt.Payload, *payload)), http.StatusCreated)
	h.logger.Infow("New user has registered",
		"login", credentials.Login,
		"remote_addr", r.RemoteAddr,
		"url", r.URL.Path,
	)
}

// LoginUser godoc
//
//	@Summary		Login to your account
//	@Description	Login via login and password in reddit-clone app
//	@Tags			auth
//	@ID				login-user
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		users.AuthUserInfo	true	"User credentials for authentication"
//	@Success		200			{object}	jwt.Session			"User authorized successfully"
//	@Failure		400			"Bad request"
//	@Failure		401			{object}	errs.SimpleErr	"Bad login or password"
//	@Failure		500			{object}	errs.SimpleErr	"Internal server error"
//	@Router			/login [post]
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	credentials := users.AuthUserInfo{}
	if err = json.Unmarshal(body, &credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := h.service.Authorize(r.Context(), credentials)
	switch {
	case errors.Is(err, errs.ErrNoUser), errors.Is(err, errs.ErrBadPass):
		sendErrorResponse(w, http.StatusUnauthorized, errs.NewSimpleErr(errs.ErrBadPass.Error()))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	h.newSession(w, r.WithContext(context.WithValue(r.Context(), jwt.Payload, *payload)), http.StatusOK)
	h.logger.Infow("New log in",
		"login", credentials.Login,
		"remote_addr", r.RemoteAddr,
		"url", r.URL.Path,
	)
}

func (h *UserHandler) newSession(w http.ResponseWriter, r *http.Request, statusCode int) {
	if r.Header.Get("Content-Type") != "application/json" {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrUnknownPayload.Error()))
		return
	}

	sess, err := h.sessMngr.New(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendResponse(sess, w, httpresp.WithStatusCode(statusCode))
}
