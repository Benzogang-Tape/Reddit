package models

import (
	"encoding/json"
	"github.com/hashicorp/go-uuid"
	"regexp"
	"time"
)

type Vote int
type PostCategory int
type PostType int

type Comment struct {
	Body string `json:"comment"`
}

type PostComment struct {
	Created string       `json:"created"`
	Author  TokenPayload `json:"author"`
	Body    string       `json:"body"`
	ID      ID           `json:"id"`
}

type PostVote struct {
	UserID ID   `json:"user"`
	Vote   Vote `json:"vote"`
}

const (
	downVote      Vote = iota - 1
	upVote        Vote = iota
	withLink           = "link"
	withText           = "text"
	music              = "music"
	funny              = "funny"
	videos             = "videos"
	programming        = "programming"
	news               = "news"
	fashion            = "fashion"
	CategoryCount int  = 6
	UUIDLength    int  = 36
)

const (
	Music PostCategory = iota
	Funny
	Videos
	Programming
	News
	Fashion
)

const (
	WithLink PostType = iota
	WithText
)

var (
	URLTemplate = regexp.MustCompile(`^((([A-Za-z]{3,9}:(?://)?)(?:[-;:&=+$,\w]+@)?[A-Za-z0-9.-]+(:[0-9]+)?|(?:www.|[-;:&=+$,\w]+@)[A-Za-z0-9.-]+)((?:/[+~%/.\w-_]*)?\??(?:[-+=&;%@.\w_]*)#?(?:\w*))?)$`)
)

func (pc PostCategory) String() string {
	return [...]string{music, funny, videos, programming, news, fashion}[pc]
}

func (pc *PostCategory) UnmarshalJSON(category []byte) error {
	var s string
	if err := json.Unmarshal(category, &s); err != nil {
		return err
	}

	ctgry, err := StringToPostCategory(s)
	if err != nil {
		return err
	}
	*pc = ctgry
	return nil
}

func (pc PostCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(pc.String())
}

func StringToPostCategory(s string) (PostCategory, error) {
	var category PostCategory
	switch s {
	case music:
		category = Music
	case funny:
		category = Funny
	case videos:
		category = Videos
	case programming:
		category = Programming
	case news:
		category = News
	case fashion:
		category = Fashion
	default:
		return category, ErrInvalidCategory
	}
	return category, nil
}

func (pt PostType) String() string {
	return [...]string{withLink, withText}[pt]
}

func (pt *PostType) UnmarshalJSON(postType []byte) error {
	var s string
	if err := json.Unmarshal(postType, &s); err != nil {
		return err
	}

	switch s {
	case withLink:
		*pt = WithLink
	case withText:
		*pt = WithText
	default:
		return ErrInvalidPostType
	}
	return nil
}

func (pt PostType) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

func NewPostComment(author TokenPayload, commentBody string) (*PostComment, error) {
	if commentBody == "" {
		return nil, ErrBadCommentBody
	}

	newComment := &PostComment{
		Created: time.Now().Format(time.RFC3339Nano),
		Author:  author,
		Body:    commentBody,
	}
	newCommentID, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	newComment.ID = ID(newCommentID)
	return newComment, nil
}

func NewPostVote(userID ID, vote Vote) *PostVote {
	return &PostVote{
		UserID: userID,
		Vote:   vote,
	}
}
