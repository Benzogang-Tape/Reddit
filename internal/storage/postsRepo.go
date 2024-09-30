package storage

import (
	"cmp"
	"context"
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"github.com/pkg/errors"
	"slices"
	"sync"
)

type PostRepo struct {
	storage []*models.Post
	mu      *sync.RWMutex
}

func NewPostRepo() *PostRepo {
	return &PostRepo{
		storage: make([]*models.Post, 0, 42),
		mu:      &sync.RWMutex{},
	}
}

func (p *PostRepo) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	postList := make([]models.Post, 0, len(p.storage))
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, post := range p.storage {
		postList = append(postList, *post)
	}
	return postList, nil
}

func (p *PostRepo) GetPostsByCategory(ctx context.Context, postCategory models.PostCategory) ([]models.Post, error) {
	postList := make([]models.Post, 0, len(p.storage)/models.CategoryCount)
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, post := range p.storage {
		if post.Category == postCategory {
			postList = append(postList, *post)
		}
	}
	return postList, nil
}

func (p *PostRepo) GetPostsByUser(ctx context.Context, userLogin models.Username) ([]models.Post, error) {
	postList := make([]models.Post, 0, 42)
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, post := range p.storage {
		if post.Author.Login == userLogin {
			postList = append(postList, *post)
		}
	}
	return postList, nil
}

func (p *PostRepo) GetPostByID(ctx context.Context, postID models.ID) (models.Post, error) {
	post, err := p.getPostByID(postID)
	if err != nil {
		return models.Post{}, errors.Wrap(err, "GetPostByID: ")
	}
	return *post.UpdateViews(), nil
}

func (p *PostRepo) CreatePost(ctx context.Context, postPayload models.PostPayload) (models.Post, error) {
	author, ok := ctx.Value(models.Payload).(*models.TokenPayload)
	if !ok {
		return models.Post{}, models.ErrBadPayload
	}

	defer p.sortPosts()
	newPost, err := models.NewPost(*author, postPayload)
	if err != nil {
		return models.Post{}, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.storage = append(p.storage, newPost)
	return *newPost, nil
}

func (p *PostRepo) DeletePost(ctx context.Context, postID models.ID) error {
	lenBeforeDelete := len(p.storage)
	p.mu.Lock()
	p.storage = slices.DeleteFunc(p.storage, func(post *models.Post) bool {
		return post.ID == postID
	})
	p.mu.Unlock()
	if lenBeforeDelete == len(p.storage) {
		return models.ErrPostNotFound
	}
	return nil
}

func (p *PostRepo) AddComment(ctx context.Context, postID models.ID, comment models.Comment) (models.Post, error) {
	author, ok := ctx.Value(models.Payload).(*models.TokenPayload)
	if !ok {
		return models.Post{}, models.ErrBadPayload
	}

	post, err := p.getPostByID(postID)
	if err != nil {
		return models.Post{}, errors.Wrap(err, "AddComment: ")
	}
	if err = post.AddComment(*author, comment.Body); err != nil {
		return models.Post{}, errors.Wrap(err, "AddComment: ")
	}
	return *post, nil
}

func (p *PostRepo) DeleteComment(ctx context.Context, postID, commentID models.ID) (models.Post, error) {
	post, err := p.getPostByID(postID)
	if err != nil {
		return models.Post{}, errors.Wrap(err, "DeleteComment: ")
	}
	if err = post.DeleteComment(commentID); err != nil {
		return models.Post{}, errors.Wrap(err, "DeleteComment: ")
	}
	return *post, nil
}

func (p *PostRepo) Upvote(ctx context.Context, postID models.ID) (models.Post, error) {
	author, ok := ctx.Value(models.Payload).(*models.TokenPayload)
	if !ok {
		return models.Post{}, models.ErrBadPayload
	}

	post, err := p.getPostByID(postID)
	if err != nil {
		return models.Post{}, errors.Wrap(err, "Upvote: ")
	}
	if err = post.Upvote(author.ID); err != nil {
		return models.Post{}, errors.Wrap(err, "Upvote: ")
	}
	p.sortPosts()
	return *post, nil
}

func (p *PostRepo) Downvote(ctx context.Context, postID models.ID) (models.Post, error) {
	author, ok := ctx.Value(models.Payload).(*models.TokenPayload)
	if !ok {
		return models.Post{}, models.ErrBadPayload
	}

	post, err := p.getPostByID(postID)
	if err != nil {
		return models.Post{}, errors.Wrap(err, "Downvote: ")
	}
	if err = post.Downvote(author.ID); err != nil {
		return models.Post{}, errors.Wrap(err, "Downvote: ")
	}
	p.sortPosts()
	return *post, nil
}

func (p *PostRepo) Unvote(ctx context.Context, postID models.ID) (models.Post, error) {
	author, ok := ctx.Value(models.Payload).(*models.TokenPayload)
	if !ok {
		return models.Post{}, models.ErrBadPayload
	}

	post, err := p.getPostByID(postID)
	if err != nil {
		return models.Post{}, errors.Wrap(err, "Unvote: ")
	}
	if err = post.Unvote(author.ID); err != nil {
		return models.Post{}, errors.Wrap(err, "Unvote: ")
	}
	p.sortPosts()
	return *post, nil
}

func (p *PostRepo) getPostByID(postID models.ID) (*models.Post, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	postIdx := slices.IndexFunc(p.storage, func(post *models.Post) bool {
		return post.ID == postID
	})
	if postIdx == -1 {
		return nil, models.ErrPostNotFound
	}
	return p.storage[postIdx], nil
}

func (p *PostRepo) sortPosts() {
	p.mu.Lock()
	defer p.mu.Unlock()
	slices.SortStableFunc(p.storage, func(a, b *models.Post) int {
		return -cmp.Compare(a.Score, b.Score)
	})
}
