package posts

import (
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

// Post model info
//
// @Description Post Contains all the information about a particular post in the app
type Post struct {
	ID               users.ID         `json:"id" bson:"uuid" example:"12345678-9abc-def1-2345-6789abcdef12" minLength:"36" maxLength:"36"`
	Score            int              `json:"score" bson:"score" example:"-1"` // The overall balance of the post's votes
	Views            uint             `json:"views" bson:"views" example:"1"`  // How many times the post has been viewed by users
	Type             PostType         `json:"type" bson:"type" example:"1"`    // Post with text(1) or with a link(0)
	Title            string           `json:"title" bson:"title" example:"Awesome title"`
	URL              string           `json:"url,omitempty" bson:"url,omitempty" example:"http://localhost:8080/"`
	Author           jwt.TokenPayload `json:"author" bson:"author"`                                                            // User who created the Post
	Category         PostCategory     `json:"category" bson:"category" example:"0"`                                            // Number of the category to which the Post belongs
	Text             string           `json:"text,omitempty" bson:"text,omitempty" example:"Awesome text" minLength:"4"`       // Content of the Post
	Votes            Votes            `json:"votes" bson:"votes"`                                                              // List of all the votes put by users on the post
	Comments         []*PostComment   `json:"comments" bson:"comments"`                                                        // List of all comments left by users under the post
	Created          string           `json:"created" bson:"created" example:"2006-01-02T15:04:05.999Z" format:"date-time"`    // Date the Post was created
	UpvotePercentage int              `json:"upvotePercentage" bson:"upvotePercentage" example:"75" minimum:"0" maximum:"100"` // Percentage of positive Votes to Post
}

type Posts []*Post

// PostPayload model info
//
// @Description PostPayload contains the necessary information to create a post
type PostPayload struct {
	Type     PostType     `json:"type"` // link or text
	Title    string       `json:"title" example:"Awesome title"`
	URL      string       `json:"url,omitempty" example:"http://localhost:8080/"`
	Category PostCategory `json:"category" example:"0"`                                // Number of the category to which the Post belongs
	Text     string       `json:"text,omitempty" example:"Awesome text" minLength:"4"` // Content of the Post
}

func NewPost(author jwt.TokenPayload, payload PostPayload) *Post {
	newPost := &Post{
		ID:               users.ID(uuid.New().String()),
		Score:            1,
		Views:            1,
		Type:             payload.Type,
		Title:            payload.Title,
		Author:           author,
		Category:         payload.Category,
		Text:             payload.Text,
		Votes:            Votes{author.ID: NewPostVote(author.ID, upVote)},
		Comments:         make([]*PostComment, 0),
		Created:          time.Now().Format(TimeFormat),
		UpvotePercentage: 100,
	}
	if newPost.Type == WithLink {
		newPost.URL = payload.URL
	}

	return newPost
}

func (p *Post) AddComment(author jwt.TokenPayload, commentBody string) *PostComment {
	newComment := NewPostComment(author, commentBody)
	p.Comments = append(p.Comments, newComment)

	return newComment
}

func (p *Post) DeleteComment(commentID users.ID) error {
	lenBeforeDelete := len(p.Comments)
	p.Comments = slices.DeleteFunc(p.Comments, func(comment *PostComment) bool {
		return commentID == comment.ID
	})
	if lenBeforeDelete == len(p.Comments) {
		return errs.ErrCommentNotFound
	}

	return nil
}

func (p *Post) Upvote(userID users.ID) (*PostVote, bool) {
	defer p.updateUpvotePercentage()
	vote, ok := p.getVoteByUserID(userID)
	if !ok {
		vote = NewPostVote(userID, upVote)
		p.Votes[userID] = vote
		p.Score++
		return vote, true
	}
	if vote.Vote == downVote {
		vote.Vote = upVote
		p.Score += 2
	}

	return vote, false
}

func (p *Post) Downvote(userID users.ID) (*PostVote, bool) {
	defer p.updateUpvotePercentage()
	vote, ok := p.getVoteByUserID(userID)
	if !ok {
		vote = NewPostVote(userID, downVote)
		p.Votes[userID] = vote
		p.Score--
		return vote, true
	}
	if vote.Vote == upVote {
		vote.Vote = downVote
		p.Score -= 2
	}

	return vote, false
}

func (p *Post) Unvote(userID users.ID) error {
	vote, ok := p.getVoteByUserID(userID)
	if !ok {
		return errs.ErrVoteNotFound
	}

	if vote.Vote == upVote {
		p.Score--
	} else {
		p.Score++
	}

	delete(p.Votes, userID)
	p.updateUpvotePercentage()

	return nil
}

func (p *Post) updateUpvotePercentage() {
	totalVotes := len(p.Votes)
	if totalVotes == 0 {
		p.UpvotePercentage = 0
		return
	}
	p.UpvotePercentage = ((p.Score + totalVotes) * 100) / (totalVotes * 2)
}

func (p *Post) UpdateViews() *Post {
	p.Views++
	return p
}

func (p *Post) getVoteByUserID(userID users.ID) (*PostVote, bool) {
	postVote, ok := p.Votes[userID]
	if !ok {
		return nil, false
	}

	return postVote, true
}
