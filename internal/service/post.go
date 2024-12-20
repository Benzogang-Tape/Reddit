package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/posts"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

type PostStorage interface {
	GetAllPosts(ctx context.Context) ([]*posts.Post, error)
	GetPostsByCategory(ctx context.Context, postCategory posts.PostCategory) ([]*posts.Post, error)
	GetPostsByUser(ctx context.Context, userLogin users.Username) ([]*posts.Post, error)
	GetPostByID(ctx context.Context, postID users.ID) (*posts.Post, error)
	CreatePost(ctx context.Context, postPayload posts.PostPayload) (*posts.Post, error)
	DeletePost(ctx context.Context, postID users.ID) error
}

type PostActions interface {
	AddComment(ctx context.Context, post *posts.Post, comment posts.Comment) (*posts.Post, error)
	DeleteComment(ctx context.Context, post *posts.Post, commentID users.ID) (*posts.Post, error)
	Upvote(ctx context.Context, post *posts.Post) (*posts.Post, error)
	Downvote(ctx context.Context, post *posts.Post) (*posts.Post, error)
	Unvote(ctx context.Context, post *posts.Post) (*posts.Post, error)
	UpdateViews(ctx context.Context, postID users.ID) error
}

type PostHandler struct {
	repo             PostStorage
	actionController PostActions
}

func NewPostHandler(storage PostStorage, actions PostActions) *PostHandler {
	return &PostHandler{
		repo:             storage,
		actionController: actions,
	}
}

func (p *PostHandler) GetAllPosts(ctx context.Context) ([]*posts.Post, error) {
	source := "GetAllPosts"
	postList, err := p.repo.GetAllPosts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	return postList, nil
}

func (p *PostHandler) GetPostsByCategory(ctx context.Context, postCategory posts.PostCategory) ([]*posts.Post, error) {
	source := "GetPostsByCategory"
	postList, err := p.repo.GetPostsByCategory(ctx, postCategory)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	return postList, nil
}

func (p *PostHandler) GetPostsByUser(ctx context.Context, userLogin users.Username) ([]*posts.Post, error) {
	source := "GetPostsByUser"
	postList, err := p.repo.GetPostsByUser(ctx, userLogin)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	return postList, nil
}

func (p *PostHandler) GetPostByID(ctx context.Context, postID users.ID) (*posts.Post, error) {
	source := "GetPostByID"
	post, err := p.repo.GetPostByID(ctx, postID)
	if err != nil {
		return post, errors.Wrap(err, source)
	}

	if err = p.actionController.UpdateViews(ctx, postID); err != nil {
		return nil, errors.Wrap(err, source)
	}

	return post.UpdateViews(), nil
}

func (p *PostHandler) CreatePost(ctx context.Context, postPayload posts.PostPayload) (*posts.Post, error) {
	source := "CreatePost"
	if postPayload.Type == posts.WithLink && !posts.URLTemplate.MatchString(postPayload.URL) {
		return nil, errors.Wrap(errs.ErrInvalidURL, source)
	}

	return p.repo.CreatePost(ctx, postPayload)
}

func (p *PostHandler) DeletePost(ctx context.Context, postID users.ID) error {
	source := "DeletePost"
	if err := p.repo.DeletePost(ctx, postID); err != nil {
		return errors.Wrap(err, source)
	}

	return nil
}

func (p *PostHandler) Upvote(ctx context.Context, postID users.ID) (*posts.Post, error) {
	source := "Upvote"
	post, err := p.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	post, err = p.actionController.Upvote(ctx, post)
	if err != nil {
		return post, errors.Wrap(err, source)
	}

	return post, nil
}

func (p *PostHandler) Downvote(ctx context.Context, postID users.ID) (*posts.Post, error) {
	source := "Downvote"
	post, err := p.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	post, err = p.actionController.Downvote(ctx, post)
	if err != nil {
		return post, errors.Wrap(err, source)
	}

	return post, nil
}

func (p *PostHandler) Unvote(ctx context.Context, postID users.ID) (*posts.Post, error) {
	source := "Unvote"
	post, err := p.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	post, err = p.actionController.Unvote(ctx, post)
	if err != nil {
		return post, errors.Wrap(err, source)
	}

	return post, nil
}

func (p *PostHandler) AddComment(ctx context.Context, postID users.ID, comment posts.Comment) (*posts.Post, error) {
	source := "AddComment"
	if comment.Body == "" {
		return nil, errs.ErrBadCommentBody
	}

	post, err := p.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	post, err = p.actionController.AddComment(ctx, post, comment)
	if err != nil {
		return post, errors.Wrap(err, source)
	}

	return post, nil
}

func (p *PostHandler) DeleteComment(ctx context.Context, postID, commentID users.ID) (*posts.Post, error) {
	source := "DeleteComment"
	post, err := p.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	post, err = p.actionController.DeleteComment(ctx, post, commentID)
	if err != nil {
		return post, errors.Wrap(err, source)
	}

	return post, nil
}
