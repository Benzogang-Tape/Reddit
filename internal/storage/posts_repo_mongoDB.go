package storage

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/posts"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

type PostRepoMongoDB struct {
	collection AbstractCollection
}

func NewPostRepoMongoDB(collection AbstractCollection) *PostRepoMongoDB {
	return &PostRepoMongoDB{
		collection: collection,
	}
}

func (p *PostRepoMongoDB) GetAllPosts(ctx context.Context) ([]*posts.Post, error) {
	posts := make(posts.Posts, 0)
	sort := bson.D{{Key: "score", Value: -1}}
	cur, err := p.collection.Find(ctx, bson.M{}, options.Find().SetSort(sort))
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (p *PostRepoMongoDB) GetPostsByCategory(ctx context.Context, postCategory posts.PostCategory) ([]*posts.Post, error) {
	posts := make(posts.Posts, 0)
	filter := bson.M{"category": postCategory}
	sort := bson.D{{Key: "score", Value: -1}}
	cur, err := p.collection.Find(ctx, filter, options.Find().SetSort(sort))
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (p *PostRepoMongoDB) GetPostsByUser(ctx context.Context, userLogin users.Username) ([]*posts.Post, error) {
	postList := make([]*posts.Post, 0)
	filter := bson.M{"author.username": userLogin}
	sort := bson.D{{Key: "created", Value: -1}}
	cur, err := p.collection.Find(ctx, filter, options.Find().SetSort(sort))
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &postList); err != nil {
		return nil, err
	}

	return postList, nil
}

func (p *PostRepoMongoDB) GetPostByID(ctx context.Context, postID users.ID) (*posts.Post, error) {
	filter := bson.M{"uuid": postID}
	res := p.collection.FindOne(ctx, filter)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, errs.ErrPostNotFound
	}

	post := new(posts.Post)
	if err := res.Decode(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (p *PostRepoMongoDB) CreatePost(ctx context.Context, postPayload posts.PostPayload) (*posts.Post, error) {
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	newPost := posts.NewPost(*author, postPayload)

	if _, err := p.collection.InsertOne(ctx, newPost); err != nil {
		fmt.Println("\n\n\n\n", err.Error())
		return nil, err
	}

	return newPost, nil
}

func (p *PostRepoMongoDB) DeletePost(ctx context.Context, postID users.ID) error {
	filter := bson.M{"uuid": postID}
	deletedCount, err := p.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if deletedCount == 0 {
		return errs.ErrPostNotFound
	}

	return nil
}

func (p *PostRepoMongoDB) AddComment(ctx context.Context, post *posts.Post, comment posts.Comment) (*posts.Post, error) {
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	newComment := post.AddComment(*author, comment.Body)
	if _, err := p.collection.UpdateOne(
		ctx,
		bson.M{"uuid": post.ID},
		bson.M{"$push": bson.M{"comments": newComment}},
	); err != nil {
		return nil, err
	}

	return post, nil
}

func (p *PostRepoMongoDB) DeleteComment(ctx context.Context, post *posts.Post, commentID users.ID) (*posts.Post, error) {
	source := "DeleteComment"
	if err := post.DeleteComment(commentID); err != nil {
		return nil, errors.Wrap(err, source)
	}

	if _, err := p.collection.UpdateOne(
		ctx,
		bson.M{"uuid": post.ID},
		bson.M{"$pull": bson.M{"comments": bson.M{"uuid": commentID}}},
	); err != nil {
		return nil, errors.Wrap(err, source)
	}

	return post, nil
}

func (p *PostRepoMongoDB) Upvote(ctx context.Context, post *posts.Post) (*posts.Post, error) {
	source := "Upvote"
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	newVote, created := post.Upvote(author.ID)
	if !created {
		filter := bson.M{
			"uuid":       post.ID,
			"votes.user": newVote.UserID,
		}
		update := bson.M{
			"$set": bson.M{
				"votes.$.vote":     newVote.Vote,
				"upvotePercentage": post.UpvotePercentage,
				"score":            post.Score,
			},
		}
		if _, err := p.collection.UpdateOne(ctx, filter, update); err != nil {
			return nil, errors.Wrap(err, source)
		}

		return post, nil
	}

	filter := bson.M{"uuid": post.ID}
	update := bson.M{
		"$push": bson.M{
			"votes": newVote,
		},
		"$set": bson.M{
			"upvotePercentage": post.UpvotePercentage,
			"score":            post.Score,
		},
	}
	if _, err := p.collection.UpdateOne(ctx, filter, update); err != nil {
		return nil, errors.Wrap(err, source)
	}

	return post, nil
}

func (p *PostRepoMongoDB) Downvote(ctx context.Context, post *posts.Post) (*posts.Post, error) {
	source := "Downvote"
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	filter := bson.M{"uuid": post.ID}
	newVote, created := post.Downvote(author.ID)
	if !created {
		filter["votes.user"] = newVote.UserID
		update := bson.M{
			"$set": bson.M{
				"votes.$.vote":     newVote.Vote,
				"upvotePercentage": post.UpvotePercentage,
				"score":            post.Score,
			},
		}

		if _, err := p.collection.UpdateOne(ctx, filter, update); err != nil {
			return nil, errors.Wrap(err, source)
		}

		return post, nil
	}

	update := bson.M{
		"$push": bson.M{
			"votes": newVote,
		},
		"$set": bson.M{
			"upvotePercentage": post.UpvotePercentage,
			"score":            post.Score,
		},
	}
	if _, err := p.collection.UpdateOne(ctx, filter, update); err != nil {
		return nil, errors.Wrap(err, source)
	}

	return post, nil
}

func (p *PostRepoMongoDB) Unvote(ctx context.Context, post *posts.Post) (*posts.Post, error) {
	source := "Unvote"
	author, ok := ctx.Value(jwt.Payload).(*jwt.TokenPayload)
	if !ok {
		return nil, errs.ErrBadPayload
	}

	if err := post.Unvote(author.ID); err != nil {
		return nil, errors.Wrap(err, source)
	}

	filter := bson.M{"uuid": post.ID}
	update := bson.M{
		"$pull": bson.M{
			"votes": bson.M{
				"user": author.ID,
			},
		},
		"$set": bson.M{
			"upvotePercentage": post.UpvotePercentage,
			"score":            post.Score,
		},
	}
	if _, err := p.collection.UpdateOne(ctx, filter, update); err != nil {
		return nil, errors.Wrap(err, source)
	}

	return post, nil
}

func (p *PostRepoMongoDB) UpdateViews(ctx context.Context, postID users.ID) error {
	source := "UpdateViews"
	filter := bson.M{"uuid": postID}
	update := bson.M{"$inc": bson.M{"views": 1}}
	if _, err := p.collection.UpdateOne(ctx, filter, update); err != nil {
		return errors.Wrap(err, source)
	}

	return nil
}
