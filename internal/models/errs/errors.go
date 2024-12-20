package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoUser              = errors.New("user not found")
	ErrNoSession           = errors.New("session not found")
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
	ErrBadID               = errors.New("bad id")
	ErrInvalidPostID       = errors.New("invalid post id")
	ErrInvalidCommentID    = errors.New("invalid comment id")
	ErrInvalidCategory     = errors.New("invalid category")
	ErrInvalidPostType     = errors.New("invalid post type")
	ErrVoteNotFound        = errors.New("no votes from the requested user")
	ErrBadCommentBody      = errors.New("comment body is required")
	ErrUnknownPayload      = errors.New("unknown payload")
	ErrUnknownError        = errors.New("unknown error")
)

type RespError interface {
	Marshal() ([]byte, error)
	Error() string
}

// SimpleErr model info
//
// @Description SimpleErr stores a brief description of an error
type SimpleErr struct {
	Message any `json:"message"` // Any type
}

// ComplexErr model info
//
// @Description ComplexErr contains a more detailed description of the error, including the location and cause of the error
type ComplexErr struct {
	Location any `json:"location"` // Any type
	Param    any `json:"param"`    // Any type
	Value    any `json:"value"`    // Any type
	Msg      any `json:"msg"`      // Any type
}

// ComplexErrArr model info
//
// @Description ComplexErrArr is an array of ComplexErr returned in case of a non-obvious error
type ComplexErrArr struct {
	Errs []ComplexErr `json:"errors"`
}

func (s SimpleErr) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

func (s SimpleErr) Error() string {
	return s.Message.(string)
}

func (c ComplexErr) Marshal() ([]byte, error) {
	complexErrs := ComplexErrArr{
		Errs: []ComplexErr{c},
	}

	return json.Marshal(complexErrs)
}

func (c ComplexErr) Error() string {
	return fmt.Sprintf("location: %s\nparam: %s\nvalue: %s\nmsg: %s\n",
		c.Location.(string),
		c.Param.(string),
		c.Value.(string),
		c.Msg.(string),
	)
}

func (ca ComplexErrArr) Marshal() ([]byte, error) {
	return json.Marshal(ca)
}

func (ca ComplexErrArr) Error() string {
	b := strings.Builder{}
	for _, c := range ca.Errs {
		fmt.Fprintf(&b, "%s\n", c.Error())
	}

	return b.String()
}

func NewSimpleErr(message any) SimpleErr {
	return SimpleErr{
		Message: message,
	}
}

func NewComplexErrArr(err ...ComplexErr) ComplexErrArr {
	return ComplexErrArr{
		Errs: err,
	}
}
