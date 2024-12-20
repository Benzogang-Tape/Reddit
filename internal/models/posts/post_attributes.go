package posts

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

// Vote type
//
// @Description Vote is an integer(1 or -1) representing the user's reaction to the Post
type Vote int

// PostCategory type
//
// @Description PostCategory is an integer representing the category to which post belongs
type PostCategory int

// PostType type
//
// @Description PostType is an integer(0 or 1) representing the type of the Post
type PostType int
type Votes map[users.ID]*PostVote

// Comment model info
//
// @Description Comment contains the text of the comment on Post
type Comment struct {
	Body string `json:"comment" example:"Some comment body example" minLength:"4"`
}

// PostComment model info
//
// @Description PostComment contains all information about a specific comment on a Post
type PostComment struct {
	Created string           `json:"created" bson:"created" example:"2006-01-02T15:04:05.999Z" format:"date-time"` // Date the comment was created
	Author  jwt.TokenPayload `json:"author" bson:"author"`
	Body    string           `json:"body" bson:"body" example:"Some comment body example" minLength:"4"` // Content of the comment
	ID      users.ID         `json:"id" bson:"uuid" example:"12345678-9abc-def1-2345-6789abcdef12" minLength:"36" maxLength:"36"`
}

// PostVote model info
//
// @Description PostVote is a structure storing user id and his/her Vote
type PostVote struct {
	// ID of the user who left the comment
	UserID users.ID `json:"user" bson:"user" example:"12345678-9abc-def1-2345-6789abcdef12" minLength:"36" maxLength:"36"`
	Vote   Vote     `json:"vote" bson:"vote" example:"-1"`
}

const (
	downVote Vote = iota - 1
	upVote   Vote = iota

	withLink    = "link"
	withText    = "text"
	music       = "music"
	funny       = "funny"
	videos      = "videos"
	programming = "programming"
	news        = "news"
	fashion     = "fashion"

	CategoryCount int = 6
	UUIDLength    int = 36

	TimeFormat = "2006-01-02T15:04:05.999Z"
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

var (
	postCategories = map[PostCategory]string{
		0: music,
		1: funny,
		2: videos,
		3: programming,
		4: news,
		5: fashion,
	}
	postTypes = map[PostType]string{
		0: withLink,
		1: withText,
	}
)

func (pc PostCategory) String() string {
	return postCategories[pc]
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

func (pc *PostCategory) UnmarshalBSONValue(bt bsontype.Type, category []byte) error {
	if bt != bson.TypeString {
		return fmt.Errorf("invalid bson postCategory type '%s'", bt.String())
	}
	cat, _, ok := bsoncore.ReadString(category)
	if !ok {
		return fmt.Errorf("invalid bson postCategory value")
	}

	ctgry, err := StringToPostCategory(cat)
	if err != nil {
		return err
	}
	*pc = ctgry

	return nil
}

func (pc PostCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(pc.String())
}

func (pc PostCategory) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(pc.String())
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
		return category, errs.ErrInvalidCategory
	}

	return category, nil
}

func (pt PostType) String() string {
	return postTypes[pt]
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
		return errs.ErrInvalidPostType
	}

	return nil
}

func (pt *PostType) UnmarshalBSONValue(bt bsontype.Type, postType []byte) error {
	if bt != bson.TypeString {
		return fmt.Errorf("invalid bson postType type '%s'", bt.String())
	}
	tp, _, ok := bsoncore.ReadString(postType)
	if !ok {
		return fmt.Errorf("invalid bson postType value")
	}

	switch tp {
	case withLink:
		*pt = WithLink
	case withText:
		*pt = WithText
	default:
		return errs.ErrInvalidPostType
	}

	return nil
}

func (pt PostType) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

func (pt PostType) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(pt.String())
}

func (v Votes) MarshalJSON() ([]byte, error) {
	votes := make([]*PostVote, 0, len(v))
	for _, postVote := range v {
		votes = append(votes, postVote)
	}

	return json.Marshal(votes)
}

func (v Votes) MarshalBSONValue() (bsontype.Type, []byte, error) {
	postVotes := make([]*PostVote, 0, len(v))
	for _, vote := range v {
		postVotes = append(postVotes, vote)
	}

	return bson.MarshalValue(postVotes)
}

func (v *Votes) UnmarshalJSONValue(votes []byte) error {
	postVotes := make([]*PostVote, 0, len(votes))
	if err := json.Unmarshal(votes, &postVotes); err != nil {
		return err
	}

	for _, postVote := range postVotes {
		(*v)[postVote.UserID] = postVote
	}

	return nil
}

func (v *Votes) UnmarshalBSONValue(bt bsontype.Type, votes []byte) error {
	if bt != bson.TypeArray {
		return fmt.Errorf("invalid bson votes type '%s'", bt.String())
	}

	postVotes := make([]*PostVote, 0)
	if err := bson.UnmarshalValue(bson.TypeArray, votes, &postVotes); err != nil {
		return err
	}

	*v = map[users.ID]*PostVote{}
	for _, postVote := range postVotes {
		(*v)[postVote.UserID] = postVote
	}

	return nil
}

func NewPostComment(author jwt.TokenPayload, commentBody string) *PostComment {
	return &PostComment{
		ID:      users.ID(uuid.New().String()),
		Created: time.Now().Format(TimeFormat),
		Author:  author,
		Body:    commentBody,
	}
}

func NewPostVote(userID users.ID, vote Vote) *PostVote {
	return &PostVote{
		UserID: userID,
		Vote:   vote,
	}
}
