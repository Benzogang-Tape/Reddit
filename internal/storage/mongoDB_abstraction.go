package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate mockgen -source=mongoDB_abstraction.go -destination=./mocks/mongoDB_abstraction_mock.go -package=mocks AbstractCollection AbstractCursor AbstractSingleResult
type AbstractCollection interface {
	Find(ctx context.Context, filter any, opts ...*options.FindOptions) (AbstractCursor, error)
	FindOne(ctx context.Context, filter any, opts ...*options.FindOneOptions) AbstractSingleResult
	InsertOne(ctx context.Context, document any, opts ...*options.InsertOneOptions) (any, error)
	UpdateOne(ctx context.Context, filter any, update any, opts ...*options.UpdateOptions) (int64, error)
	DeleteOne(ctx context.Context, filter any, opts ...*options.DeleteOptions) (int64, error)
}

type AbstractCursor interface {
	All(ctx context.Context, result any) error
}

type AbstractSingleResult interface {
	Decode(v any) error
	Err() error
}

type mongoCollection struct {
	collection *mongo.Collection
}

type mongoCursor struct {
	cursor *mongo.Cursor
}

type mongoSingleResult struct { //nolint:unused
	sr *mongo.SingleResult
}

func NewMongoCollection(collection *mongo.Collection) *mongoCollection {
	return &mongoCollection{
		collection: collection,
	}
}

func (c *mongoCollection) Find(ctx context.Context, filter any, opts ...*options.FindOptions) (AbstractCursor, error) {
	cursor, err := c.collection.Find(ctx, filter, opts...)
	return &mongoCursor{
		cursor: cursor,
	}, err
}

func (c *mongoCollection) FindOne(ctx context.Context, filter any, opts ...*options.FindOneOptions) AbstractSingleResult {
	return c.collection.FindOne(ctx, filter, opts...)
}

func (c *mongoCollection) InsertOne(ctx context.Context, document any, opts ...*options.InsertOneOptions) (any, error) {
	return c.collection.InsertOne(ctx, document, opts...)
}

func (c *mongoCollection) UpdateOne(ctx context.Context, filter any, update any, opts ...*options.UpdateOptions) (int64, error) {
	result, err := c.collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return 0, err
	}

	return result.MatchedCount, nil
}

func (c *mongoCollection) DeleteOne(ctx context.Context, filter any, opts ...*options.DeleteOptions) (int64, error) {
	result, err := c.collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (c *mongoCursor) All(ctx context.Context, result any) error {
	return c.cursor.All(ctx, result)
}

func (sr *mongoSingleResult) Decode(v any) error { //nolint:unused
	return sr.sr.Decode(v)
}

func (sr *mongoSingleResult) Err() error { //nolint:unused
	return sr.sr.Err()
}
