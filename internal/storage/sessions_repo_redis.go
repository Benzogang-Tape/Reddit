package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"

	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
)

type SessionRepoRedis struct {
	rdb *redis.Client
}

func NewSessionRepoRedis(client *redis.Client) *SessionRepoRedis {
	return &SessionRepoRedis{
		rdb: client,
	}
}

func (s *SessionRepoRedis) CreateSession(ctx context.Context, session *jwt.Session, payload *jwt.TokenPayload) (*jwt.Session, error) { //nolint:unparam
	key := session.Token
	value, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	result, err := s.rdb.Set(key, value, jwt.SessLifespan).Result()
	if err != nil {
		return nil, err
	}
	if result != "OK" {
		return nil, fmt.Errorf("result is not OK. Actual value: %s", result)
	}

	return session, nil
}

func (s *SessionRepoRedis) CheckSession(ctx context.Context, sess *jwt.Session) (*jwt.TokenPayload, error) { //nolint:unparam
	key := sess.Token
	val, err := s.rdb.Get(key).Result()
	if err != nil {
		return nil, err
	}

	payload := &jwt.TokenPayload{}
	if err = json.Unmarshal([]byte(val), payload); err != nil {
		return nil, err
	}

	return payload, nil
}
