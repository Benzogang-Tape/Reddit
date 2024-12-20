package rest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
	"github.com/Benzogang-Tape/Reddit/internal/storage/mocks"
	"github.com/Benzogang-Tape/Reddit/internal/transport/rest"
)

var (
	rawCredentials = `{"username":"admin","password":"rootroot"}` //nolint:gosec
	session        = &jwt.Session{
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzExMjUzMjMsImlhdCI6MTczMDUyMDUyMywidXNlciI6eyJ1c2VybmFtZSI6ImFkbWluIiwiaWQiOiJyb290cm9vdCJ9fQ.SHa_cgEHVKzDfIawE1Rtn7A6gBOyauLtem2G-3-iKaQ",
	}
	credentials = &users.AuthUserInfo{
		Login:    "admin",
		Password: "rootroot",
	}
	payload = &jwt.TokenPayload{
		Login: "admin",
		ID:    "ffffffff-ffff-ffff-ffff-ffffffffffff",
	}
)

type fakeBody struct {
	data string
}

func (f *fakeBody) Read(p []byte) (int, error) {
	return 0, errors.New("planned read error")
}

func (f *fakeBody) Close() error {
	return errors.New("planned close error")
}

func TestLoginUser(t *testing.T) { //nolint:funlen
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mocks.NewMockSessionAPI(ctrl)

	st := mocks.NewMockUserAPI(ctrl)
	handler := rest.NewUserHandler(st, sm, zap.NewNop().Sugar())

	expectedCtx := context.WithValue(context.Background(), jwt.Payload, *payload)

	// Success
	st.EXPECT().Authorize(context.Background(), *credentials).Return(payload, nil)
	sm.EXPECT().New(expectedCtx).Return(session, nil)

	r := httptest.NewRequest("POST", "/api/login", strings.NewReader(rawCredentials))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.LoginUser(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), session.Token)

	// Read body error
	r = httptest.NewRequest("POST", "/api/login", bytes.NewReader(nil))
	r.Body = &fakeBody{data: rawCredentials}
	w = httptest.NewRecorder()
	handler.LoginUser(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Unmarshal body error
	r = httptest.NewRequest("POST", "/api/login", bytes.NewReader(nil))
	w = httptest.NewRecorder()
	handler.LoginUser(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// No user found
	st.EXPECT().Authorize(context.Background(), *credentials).Return(nil, errs.ErrNoUser)
	r = httptest.NewRequest("POST", "/api/login", strings.NewReader(rawCredentials))
	w = httptest.NewRecorder()

	handler.LoginUser(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body) //nolint:errcheck

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrBadPass.Error())

	// Bad password
	st.EXPECT().Authorize(context.Background(), *credentials).Return(nil, errs.ErrBadPass)
	r = httptest.NewRequest("POST", "/api/login", strings.NewReader(rawCredentials))
	w = httptest.NewRecorder()

	handler.LoginUser(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body) //nolint:errcheck

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrBadPass.Error())

	// Unknown error
	st.EXPECT().Authorize(context.Background(), *credentials).Return(nil, errs.ErrUnknownError)
	r = httptest.NewRequest("POST", "/api/login", strings.NewReader(rawCredentials))
	w = httptest.NewRecorder()

	handler.LoginUser(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body) //nolint:errcheck

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mocks.NewMockSessionAPI(ctrl)

	st := mocks.NewMockUserAPI(ctrl)
	handler := rest.NewUserHandler(st, sm, zap.NewNop().Sugar())

	expectedCtx := context.WithValue(context.Background(), jwt.Payload, *payload)

	// Success
	st.EXPECT().Register(context.Background(), *credentials).Return(payload, nil)
	sm.EXPECT().New(expectedCtx).Return(session, nil)

	r := httptest.NewRequest("POST", "/api/register", strings.NewReader(rawCredentials))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.RegisterUser(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, string(body), session.Token)

	// Read body error
	r = httptest.NewRequest("POST", "/api/register", bytes.NewReader(nil))
	r.Body = &fakeBody{data: rawCredentials}
	w = httptest.NewRecorder()
	handler.RegisterUser(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Unmarshal body error
	r = httptest.NewRequest("POST", "/api/register", bytes.NewReader(nil))
	w = httptest.NewRecorder()
	handler.RegisterUser(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Username already exists
	st.EXPECT().Register(context.Background(), *credentials).Return(nil, errs.ErrUserExists)
	r = httptest.NewRequest("POST", "/api/register", strings.NewReader(rawCredentials))
	w = httptest.NewRecorder()

	handler.RegisterUser(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body) //nolint:errcheck

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Contains(t, string(body), `already exists`)

	// Unknown error
	st.EXPECT().Register(context.Background(), *credentials).Return(nil, errs.ErrUnknownError)
	r = httptest.NewRequest("POST", "/api/register", strings.NewReader(rawCredentials))
	w = httptest.NewRecorder()

	handler.RegisterUser(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body) //nolint:errcheck

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestNewSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mocks.NewMockSessionAPI(ctrl)

	st := mocks.NewMockUserAPI(ctrl)
	handler := rest.NewUserHandler(st, sm, zap.NewNop().Sugar())

	// Bad Content-Type
	st.EXPECT().Authorize(context.Background(), *credentials).Return(payload, nil)

	r := httptest.NewRequest("POST", "/api/login", strings.NewReader(rawCredentials))
	r.Header.Set("Content-Type", "plain/text")
	w := httptest.NewRecorder()

	handler.LoginUser(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownPayload.Error())

	// New session error
	expectedCtx := context.WithValue(context.Background(), jwt.Payload, *payload)
	st.EXPECT().Authorize(context.Background(), *credentials).Return(payload, nil)
	sm.EXPECT().New(expectedCtx).Return(nil, errs.ErrUnknownError)

	r = httptest.NewRequest("POST", "/api/login", strings.NewReader(rawCredentials))
	r.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler.LoginUser(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
