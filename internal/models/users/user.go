package users

import (
	"github.com/google/uuid"
)

type Username string
type ID string

type User struct {
	ID       ID       `schema:"-" json:"-"`
	Username Username `schema:"username,required" json:"username"`
	Password string   `schema:"password,required" json:"password" minLength:"8" format:"password"`
}

// AuthUserInfo model info
//
// @Description AuthUserInfo stores User credentials contained in the JWT Session token.
type AuthUserInfo struct {
	Login    Username `json:"username" example:"Valery_Albertovich"`
	Password string   `json:"password" example:"want_pizza" minLength:"8" format:"password"`
}

func NewUser(authInfo AuthUserInfo) *User {
	return &User{
		ID:       ID(uuid.New().String()),
		Username: authInfo.Login,
		Password: authInfo.Password,
	}
}
