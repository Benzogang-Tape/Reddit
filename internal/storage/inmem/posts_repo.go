package inmem

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"github.com/pkg/errors"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/posts"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

type PostRepo struct {
	storage []*posts.Post
	mu      *sync.RWMutex
}

func NewPostRepo() *PostRepo {
	return &PostRepo{
		storage: make([]*posts.Post, 0),
		mu:      &sync.RWMutex{},
	}
}

func (p *PostRepo) GetAllPosts(ctx context.Context) ([]*posts.Post, error) { //nolint:unparam
	postList := make([]*posts.Post, 0, len(p.storage))
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, post := range p.storage {
		postList = append(postList, &(*post))
	}

	return postList, nil
}

func (p *PostRepo) GetPostsByCategory(ctx context.Context, postCategory posts.PostCategory) ([]*posts.Post, error) { //nolint:unparam
	postList := make([]*posts.Post, 0, len(p.storage)/posts.CategoryCount)
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, post := range p.storage {
		if post.Category == postCategory {
			postList = append(postList, &(*post))
		}
	}

	return postList, nil
}

func (p *PostRepo) GetPostsByUser(ctx context.Context, userLogin users.Username) ([]*posts.Post, error) { //nolint:unparam
	postList := make([]*posts.Post, 0)
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, post := range p.storage {
		if post.Author.Login == userLogin {
			postList = append(postList, &(*post))
		}
	}

	return postList, nil
}

func (p *PostRepo) GetPostByID(ctx context.Context, postID users.ID) (*posts.Post, error) { //nolint:unparam
	source := "GetPostByID"
	post, err := p.getPostByID(postID)
	if err != nil {
		return nil, errors.Wrap(err, source)
	}

	//original
	//return &(*post.UpdateViews()), nil

	// fix
	//pst := &(*post)
	//post.UpdateViews()
	//return pst, nil

	// another fix
	return post, nil
}

func (p *PostRepo) CreatePost(ctx context.Context, postPayload posts.PostPayload) (*posts.Post, error) {
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	defer p.sortPosts()

	newPost := posts.NewPost(*author, postPayload)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.storage = append(p.storage, newPost)

	return &(*newPost), nil
}

func (p *PostRepo) DeletePost(ctx context.Context, postID users.ID) error { //nolint:unparam
	lenBeforeDelete := len(p.storage)
	p.mu.Lock()
	p.storage = slices.DeleteFunc(p.storage, func(post *posts.Post) bool {
		return post.ID == postID
	})
	p.mu.Unlock()
	if lenBeforeDelete == len(p.storage) {
		return errs.ErrPostNotFound
	}

	return nil
}

//func (p *PostRepo) AddComment(ctx context.Context, postID models.ID, comment models.Comment) (*models.Post, error) {
//	source := "AddComment"
//	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
//	if !ok {
//		return nil, errs.ErrBadPayload
//	}
//
//	post, err := p.getPostByID(postID)
//	if err != nil {
//		return nil, errors.Wrap(err, source)
//	}
//	post.AddComment(*author, comment.Body)
//
//	return &(*post), nil
//}

func (p *PostRepo) AddComment(ctx context.Context, post *posts.Post, comment posts.Comment) (*posts.Post, error) {
	//source := "AddComment"
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	post.AddComment(*author, comment.Body)

	return &(*post), nil
}

//func (p *PostRepo) DeleteComment(ctx context.Context, postID, commentID models.ID) (*models.Post, error) {
//	source := "DeleteComment"
//	post, err := p.getPostByID(postID)
//	if err != nil {
//		return nil, errors.Wrap(err, source)
//	}
//	if err = post.DeleteComment(commentID); err != nil {
//		return nil, errors.Wrap(err, source)
//	}
//
//	return &(*post), nil
//}

func (p *PostRepo) DeleteComment(ctx context.Context, post *posts.Post, commentID users.ID) (*posts.Post, error) { //nolint:unparam
	source := "DeleteComment"
	if err := post.DeleteComment(commentID); err != nil {
		return nil, errors.Wrap(err, source)
	}

	return &(*post), nil
}

//func (p *PostRepo) Upvote(ctx context.Context, postID models.ID) (*models.Post, error) {
//	source := "Upvote"
//	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
//	if !ok {
//		return nil, errs.ErrBadPayload
//	}
//
//	post, err := p.getPostByID(postID)
//	if err != nil {
//		return nil, errors.Wrap(err, source)
//	}
//	post.Upvote(author.ID)
//	p.sortPosts()
//
//	return &(*post), nil
//}

func (p *PostRepo) Upvote(ctx context.Context, post *posts.Post) (*posts.Post, error) {
	//source := "Upvote"
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	post.Upvote(author.ID)
	p.sortPosts()

	return &(*post), nil
}

//func (p *PostRepo) Downvote(ctx context.Context, postID models.ID) (*models.Post, error) {
//	source := "Downvote"
//	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
//	if !ok {
//		return nil, errs.ErrBadPayload
//	}
//
//	post, err := p.getPostByID(postID)
//	if err != nil {
//		return nil, errors.Wrap(err, source)
//	}
//	post.Downvote(author.ID)
//	p.sortPosts()
//
//	return &(*post), nil
//}

func (p *PostRepo) Downvote(ctx context.Context, post *posts.Post) (*posts.Post, error) {
	//source := "Downvote"
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	post.Downvote(author.ID)
	p.sortPosts()

	return &(*post), nil
}

//func (p *PostRepo) Unvote(ctx context.Context, postID models.ID) (*models.Post, error) {
//	source := "Unvote"
//	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
//	if !ok {
//		return nil, errs.ErrBadPayload
//	}
//
//	post, err := p.getPostByID(postID)
//	if err != nil {
//		return nil, errors.Wrap(err, source)
//	}
//	if err = post.Unvote(author.ID); err != nil {
//		return nil, errors.Wrap(err, source)
//	}
//	p.sortPosts()
//
//	return &(*post), nil
//}

func (p *PostRepo) Unvote(ctx context.Context, post *posts.Post) (*posts.Post, error) {
	source := "Unvote"
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	if err := post.Unvote(author.ID); err != nil {
		return nil, errors.Wrap(err, source)
	}

	p.sortPosts()

	return &(*post), nil
}

func (p *PostRepo) getPostByID(postID users.ID) (*posts.Post, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	postIdx := slices.IndexFunc(p.storage, func(post *posts.Post) bool {
		return post.ID == postID
	})
	if postIdx == -1 {
		return nil, errs.ErrPostNotFound
	}

	return &(*p.storage[postIdx]), nil
}

func (p *PostRepo) UpdateViews(ctx context.Context, postID users.ID) error { //nolint:unparam
	// stub
	return nil
}

func (p *PostRepo) sortPosts() {
	p.mu.Lock()
	defer p.mu.Unlock()
	slices.SortStableFunc(p.storage, func(a, b *posts.Post) int {
		return -cmp.Compare(a.Score, b.Score)
	})
}
