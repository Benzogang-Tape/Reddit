package jwt

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

// Session model info
//
// @Description Session stores the JWT token of the session
type Session struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzUyMzU3ODAsImlhdCI6MTczNDYzMDk4MCwidXNlciI6eyJ1c2VybmFtZSI6InRlc3RfdXNlciIsImlkIjoiZDNkNzc1YmEtYTFlZS00MTEwLTkwOTktMTA0ZDVkYzFkYzQ2In19.I_3_yHlH1QUuKavtx8xVN_IRFMYXg3dYumzSrImA_NM"`
}

// TokenPayload model info
//
// @Description TokenPayload stores the User payload contained in the JWT Session token
type TokenPayload struct {
	// User login
	Login users.Username `json:"username" bson:"username" example:"test_user"`
	// User id
	ID users.ID `json:"id" bson:"uuid" example:"12345678-9abc-def1-2345-6789abcdef12" minLength:"36" maxLength:"36"`
}

const (
	SessLifespan = 24 * time.Hour * 7
)

func NewSession(payload TokenPayload) (*Session, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": payload,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(SessLifespan).Unix(),
	})

	tokenString, err := token.SignedString(secretKeyProvider())
	if err != nil {
		return nil, err
	}
	return &Session{
		Token: tokenString,
	}, nil
}

func (s *Session) ValidateToken() (*TokenPayload, error) {
	hashSecretGetter := func(token *jwt.Token) (any, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, fmt.Errorf("bad sign method")
		}
		return secretKeyProvider(), nil
	}
	token, err := jwt.Parse(s.Token, hashSecretGetter)
	if err != nil || !token.Valid {
		return nil, errs.ErrBadToken
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errs.ErrNoPayload
	}
	dataFromToken, ok := payload["user"].(map[string]any)
	if !ok {
		return nil, errs.ErrBadToken
	}

	return &TokenPayload{
		Login: users.Username(dataFromToken["username"].(string)),
		ID:    users.ID(dataFromToken["id"].(string)),
	}, nil
}
