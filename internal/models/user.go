package models

import "github.com/hashicorp/go-uuid"

type Username string
type ID string

type User struct {
	ID       ID       `schema:"-" json:"-"`
	Username Username `schema:"username,required" json:"username,required"`
	Password string   `schema:"password,required" json:"password,required"`
}

type AuthUserInfo struct {
	Login    Username `json:"username,required"`
	Password string   `json:"password,required"`
}

func NewUser(authInfo AuthUserInfo) (*User, error) {
	newUserID, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	return &User{
		ID:       ID(newUserID),
		Username: authInfo.Login,
		Password: authInfo.Password,
	}, nil
}
