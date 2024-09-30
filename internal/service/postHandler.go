package service

import (
	"context"
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"github.com/pkg/errors"
)

type PostStorage interface {
	GetAllPosts(ctx context.Context) ([]models.Post, error)
	GetPostsByCategory(ctx context.Context, postCategory models.PostCategory) ([]models.Post, error)
	GetPostsByUser(ctx context.Context, userLogin models.Username) ([]models.Post, error)
	GetPostByID(ctx context.Context, postID models.ID) (models.Post, error)
	CreatePost(ctx context.Context, postPayload models.PostPayload) (models.Post, error)
	DeletePost(ctx context.Context, postID models.ID) error
}

type PostActions interface {
	AddComment(ctx context.Context, postID models.ID, comment models.Comment) (models.Post, error)
	DeleteComment(ctx context.Context, postID, commentID models.ID) (models.Post, error)
	Upvote(ctx context.Context, postID models.ID) (models.Post, error)
	Downvote(ctx context.Context, postID models.ID) (models.Post, error)
	Unvote(ctx context.Context, postID models.ID) (models.Post, error)
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

func (p *PostHandler) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	postList, err := p.repo.GetAllPosts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetAllPosts: ")
	}
	return postList, nil
}

func (p *PostHandler) GetPostsByCategory(ctx context.Context, postCategory models.PostCategory) ([]models.Post, error) {
	postList, err := p.repo.GetPostsByCategory(ctx, postCategory)
	if err != nil {
		return nil, errors.Wrap(err, "GetPostsByCategory: ")
	}
	return postList, nil
}

func (p *PostHandler) GetPostsByUser(ctx context.Context, userLogin models.Username) ([]models.Post, error) {
	postList, err := p.repo.GetPostsByUser(ctx, userLogin)
	if err != nil {
		return nil, errors.Wrap(err, "GetPostsByUser: ")
	}
	return postList, nil
}

func (p *PostHandler) GetPostByID(ctx context.Context, postID models.ID) (models.Post, error) {
	post, err := p.repo.GetPostByID(ctx, postID)
	if err != nil {
		return post, errors.Wrap(err, "GetPostByID: ")
	}
	return post, nil
}

func (p *PostHandler) CreatePost(ctx context.Context, postPayload models.PostPayload) (models.Post, error) {
	if postPayload.Type == models.WithLink && !models.URLTemplate.MatchString(postPayload.URL) {
		return models.Post{}, errors.Wrap(models.ErrInvalidURL, "CreatePost: ")
	}
	return p.repo.CreatePost(ctx, postPayload)
}

func (p *PostHandler) DeletePost(ctx context.Context, postID models.ID) error {
	if err := p.repo.DeletePost(ctx, postID); err != nil {
		return errors.Wrap(err, "DeletePost: ")
	}
	return nil
}

func (p *PostHandler) Upvote(ctx context.Context, postID models.ID) (models.Post, error) {
	post, err := p.actionController.Upvote(ctx, postID)
	if err != nil {
		return post, errors.Wrap(err, "Upvote: ")
	}
	return post, nil
}

func (p *PostHandler) Downvote(ctx context.Context, postID models.ID) (models.Post, error) {
	post, err := p.actionController.Downvote(ctx, postID)
	if err != nil {
		return post, errors.Wrap(err, "Downvote: ")
	}
	return post, nil
}

func (p *PostHandler) Unvote(ctx context.Context, postID models.ID) (models.Post, error) {
	post, err := p.actionController.Unvote(ctx, postID)
	if err != nil {
		return post, errors.Wrap(err, "Unvote: ")
	}
	return post, nil
}

func (p *PostHandler) AddComment(ctx context.Context, postID models.ID, comment models.Comment) (models.Post, error) {
	post, err := p.actionController.AddComment(ctx, postID, comment)
	if err != nil {
		return post, errors.Wrap(err, "AddComment: ")
	}
	return post, nil
}

func (p *PostHandler) DeleteComment(ctx context.Context, postID, commentID models.ID) (models.Post, error) {
	post, err := p.actionController.DeleteComment(ctx, postID, commentID)
	if err != nil {
		return post, errors.Wrap(err, "DeleteComment: ")
	}
	return post, nil
}
