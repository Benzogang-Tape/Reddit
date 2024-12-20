package inmem

import (
	"context"
	"sync"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
)

type SessionRepo struct {
	storage map[string]*jwt.TokenPayload
	mu      *sync.RWMutex
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{
		storage: make(map[string]*jwt.TokenPayload),
		mu:      &sync.RWMutex{},
	}
}

// Need TTL

func (s *SessionRepo) CreateSession(ctx context.Context, session *jwt.Session, payload *jwt.TokenPayload) (*jwt.Session, error) { //nolint:unparam
	key := session.Token
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[key] = payload
	return session, nil
}

func (s *SessionRepo) CheckSession(ctx context.Context, sess *jwt.Session) (*jwt.TokenPayload, error) { //nolint:unparam
	key := sess.Token
	s.mu.RLock()
	defer s.mu.RUnlock()
	payload, ok := s.storage[key]
	if !ok {
		return nil, errs.ErrNoSession
	}

	return payload, nil
}
