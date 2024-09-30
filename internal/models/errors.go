package models

import (
	"errors"
)

var (
	ErrNoUser              = errors.New("user not found")
	ErrInternalServerError = errors.New("internal server error")
	ErrBadPass             = errors.New("invalid password")
	ErrUserExists          = errors.New("username already exist")
	ErrBadToken            = errors.New("bad token")
	ErrNoPayload           = errors.New("no payload")
	ErrBadPayload          = errors.New("bad payload")
	ErrInvalidURL          = errors.New("url is invalid")
	ErrResponseError       = errors.New("response generation error")
	ErrPostNotFound        = errors.New("post not found")
	ErrCommentNotFound     = errors.New("comment not found")
	ErrInvalidPostID       = errors.New("invalid post id")
	ErrInvalidCommentID    = errors.New("invalid comment id")
	ErrInvalidCategory     = errors.New("invalid category")
	ErrInvalidPostType     = errors.New("invalid post type")
	ErrVoteNotFound        = errors.New("no votes from the requested user")
	ErrBadCommentBody      = errors.New("comment body is required")
	ErrUnknownPayload      = errors.New("unknown payload")
	ErrUnknownError        = errors.New("unknown error")
)

type SimpleErr struct {
	Message interface{} `json:"message,required"`
}

type ComplexErr struct {
	Location interface{} `json:"location"`
	Param    interface{} `json:"param"`
	Value    interface{} `json:"value,omitempty"`
	Msg      interface{} `json:"msg"`
}

type ComplexErrArr struct {
	Errs []ComplexErr `json:"errors"`
}

func NewSimpleErr(message interface{}) SimpleErr {
	return SimpleErr{
		Message: message,
	}
}

func NewComplexErr(err ...ComplexErr) ComplexErrArr {
	return ComplexErrArr{
		Errs: err,
	}
}
