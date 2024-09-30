package models

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"time"
)

type favContextKey string

type Session struct {
	Token string `json:"token"`
}

type TokenPayload struct {
	Login Username `json:"username,required"`
	ID    ID       `json:"id,required"`
}

const (
	Payload = favContextKey("payload")
)

var (
	secretKey = []byte("super secret key")
)

func NewSession(payload TokenPayload) (*Session, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": payload,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().AddDate(0, 0, 7).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return nil, err
	}
	return &Session{
		Token: tokenString,
	}, nil
}

func (s *Session) InitWithToken(token string) {
	s.Token = token
}

func (s *Session) ValidateToken() (*TokenPayload, error) {
	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, fmt.Errorf("bad sign method")
		}
		return secretKey, nil
	}
	token, err := jwt.Parse(s.Token, hashSecretGetter)
	if err != nil || !token.Valid {
		return nil, ErrBadToken
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrNoPayload
	}
	dataFromToken, ok := payload["user"].(map[string]interface{})
	if !ok {
		return nil, ErrBadToken
	}

	return &TokenPayload{
		Login: Username(dataFromToken["username"].(string)),
		ID:    ID(dataFromToken["id"].(string)),
	}, nil
}
