package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
)

type SessionManager interface {
	CreateSession(ctx context.Context, session *jwt.Session, payload *jwt.TokenPayload) (*jwt.Session, error)
	CheckSession(ctx context.Context, session *jwt.Session) (*jwt.TokenPayload, error)
}

//go:generate mockgen -source=session.go -destination=../storage/mocks/sessions_repo_redis_mock.go -package=mocks SessionAPI
type SessionAPI interface {
	New(ctx context.Context) (*jwt.Session, error)
	Verify(ctx context.Context, session *jwt.Session) (*jwt.TokenPayload, error)
}

type SessionHandler struct {
	manager SessionManager
}

func NewSessionHandler(mngr SessionManager) *SessionHandler {
	return &SessionHandler{
		manager: mngr,
	}
}

func (s *SessionHandler) New(ctx context.Context) (*jwt.Session, error) {
	source := "New session"
	payload, ok := ctx.Value(jwt.Payload).(jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	sess, err := jwt.NewSession(payload)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	sess, err = s.manager.CreateSession(ctx, sess, &payload)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	return sess, nil
}

func (s *SessionHandler) Verify(ctx context.Context, session *jwt.Session) (*jwt.TokenPayload, error) {
	source := "Verify session"
	payload, err := s.manager.CheckSession(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	return payload, nil
}
