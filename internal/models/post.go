package models

import (
	"github.com/hashicorp/go-uuid"
	"github.com/pkg/errors"
	"slices"
	"time"
)

type Post struct {
	Score            int            `json:"score"`
	Views            uint           `json:"views"`
	Type             PostType       `json:"type"`
	Title            string         `json:"title"`
	URL              string         `json:"url,omitempty"`
	Author           TokenPayload   `json:"author"`
	Category         PostCategory   `json:"category"`
	Text             string         `json:"text,omitempty"`
	Votes            []*PostVote    `json:"votes"`
	Comments         []*PostComment `json:"comments"`
	Created          string         `json:"created"`
	UpvotePercentage int            `json:"upvotePercentage"`
	ID               ID             `json:"id"`
}

type PostPayload struct {
	Type     PostType     `json:"type"`
	Title    string       `json:"title"`
	URL      string       `json:"url,omitempty"`
	Category PostCategory `json:"category"`
	Text     string       `json:"text,omitempty"`
}

func NewPost(author TokenPayload, payload PostPayload) (*Post, error) {
	newPost := &Post{
		Score:            1,
		Views:            0,
		Type:             payload.Type,
		Title:            payload.Title,
		Author:           author,
		Category:         payload.Category,
		Text:             payload.Text,
		Votes:            append(make([]*PostVote, 0, 42), NewPostVote(author.ID, upVote)),
		Comments:         make([]*PostComment, 0, 42),
		Created:          time.Now().Format(time.RFC3339Nano),
		UpvotePercentage: 100,
	}
	newPostID, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	newPost.ID = ID(newPostID)
	if newPost.Type == WithLink {
		newPost.URL = payload.URL
	}
	return newPost, nil
}

func (p *Post) AddComment(author TokenPayload, commentBody string) error {
	newComment, err := NewPostComment(author, commentBody)
	if err != nil {
		return errors.Wrap(err, "AddComment: ")
	}
	p.Comments = append(p.Comments, newComment)
	return nil
}

func (p *Post) DeleteComment(commentID ID) error {
	lenBeforeDelete := len(p.Comments)
	p.Comments = slices.DeleteFunc(p.Comments, func(comment *PostComment) bool {
		return commentID == comment.ID
	})
	if lenBeforeDelete == len(p.Comments) {
		return ErrCommentNotFound
	}
	return nil
}

func (p *Post) Upvote(userID ID) error {
	vote, err := p.getVoteByUserID(userID)
	if errors.Is(err, ErrVoteNotFound) {
		p.Votes = append(p.Votes, NewPostVote(userID, upVote))
		p.Score++
	} else if err == nil {
		if vote.Vote == downVote {
			vote.Vote = upVote
			p.Score += 2
		}
	}
	p.updateUpvotePercentage()
	return nil
}

func (p *Post) Downvote(userID ID) error {
	vote, err := p.getVoteByUserID(userID)
	if errors.Is(err, ErrVoteNotFound) {
		p.Votes = append(p.Votes, NewPostVote(userID, downVote))
		p.Score--
	} else if err == nil {
		if vote.Vote == upVote {
			vote.Vote = downVote
			p.Score -= 2
		}
	}
	p.updateUpvotePercentage()
	return nil
}

func (p *Post) Unvote(userID ID) error {
	voteIdx := slices.IndexFunc(p.Votes, func(vote *PostVote) bool {
		return vote.UserID == userID
	})
	if voteIdx == -1 {
		return ErrVoteNotFound
	}

	if p.Votes[voteIdx].Vote == upVote {
		p.Score--
	} else {
		p.Score++
	}
	p.Votes = slices.Delete(p.Votes, voteIdx, voteIdx+1)
	p.updateUpvotePercentage()
	return nil
}

func (p *Post) updateUpvotePercentage() {
	totalVotes := len(p.Votes)
	if totalVotes == 0 {
		p.UpvotePercentage = 0
		return
	}
	p.UpvotePercentage = (p.Score + totalVotes) / (totalVotes * 2) * 100
}

func (p *Post) UpdateViews() *Post {
	p.Views++
	return p
}

func (p *Post) getVoteByUserID(userID ID) (*PostVote, error) {
	voteIdx := slices.IndexFunc(p.Votes, func(vote *PostVote) bool {
		return vote.UserID == userID
	})
	if voteIdx == -1 {
		return nil, ErrVoteNotFound
	}
	return p.Votes[voteIdx], nil
}
