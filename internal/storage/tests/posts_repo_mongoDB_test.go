package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/posts"
	"github.com/Benzogang-Tape/Reddit/internal/storage"
	"github.com/Benzogang-Tape/Reddit/internal/storage/mocks"
)

var (
	expectedPosts = []*posts.Post{
		{
			ID:    "11111111-1111-1111-1111-111111111111",
			Score: 1,
			Views: 1,
			Type:  posts.WithText,
			Title: "TEST POST",
			URL:   "",
			Author: jwt.TokenPayload{
				Login: "admin",
				ID:    "rootroot",
			},
			Category: posts.Music,
			Text:     "I love music of Kartik.",
			Votes: posts.Votes{
				"ffffffff-ffff-ffff-ffff-ffffffffffff": &posts.PostVote{
					UserID: "ffffffff-ffff-ffff-ffff-ffffffffffff",
					Vote:   1,
				},
			},
			Comments: []*posts.PostComment{
				{
					Created: "2024-02-20T10:21:54.716Z",
					Author: jwt.TokenPayload{
						Login: "admin",
						ID:    "ffffffff-ffff-ffff-ffff-ffffffffffff",
					},
					Body: "We love music of Kartik.\nThanks.",
					ID:   "22222222-2222-2222-2222-222222222222",
				},
			},
			Created:          "2024-02-20T10:21:04.716Z",
			UpvotePercentage: 100,
		},
		{
			ID:    "33333333-3333-3333-3333-333333333333",
			Score: 1,
			Views: 1,
			Type:  posts.WithLink,
			Title: "TEST POST",
			URL:   "http://84.23.52.45:3000/createpost",
			Author: jwt.TokenPayload{
				Login: "admin",
				ID:    "ffffffff-ffff-ffff-ffff-ffffffffffff",
			},
			Category: posts.Programming,
			Text:     "",
			Votes: posts.Votes{
				"ffffffff-ffff-ffff-ffff-ffffffffffff": &posts.PostVote{
					UserID: "ffffffff-ffff-ffff-ffff-ffffffffffff",
					Vote:   1,
				},
			},
			Comments:         []*posts.PostComment{},
			Created:          "1984-02-20T10:21:04.716Z",
			UpvotePercentage: 100,
		},
	}
	postPayload = posts.PostPayload{
		Type:     posts.WithLink,
		Title:    "TEST POST",
		URL:      "http://84.23.52.45:3000/createpost",
		Category: posts.Programming,
		Text:     "",
	}
	tokenPayloadAdmin = &jwt.TokenPayload{
		Login: "admin",
		ID:    "ffffffff-ffff-ffff-ffff-ffffffffffff",
	}
	tokenPayloadUser = &jwt.TokenPayload{
		Login: "user",
		ID:    "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
	}
	findInternalErr = "Find internal error"
	errSimulatedErr = errors.New("simulated error")
)

func toBSON(data any) bson.D {
	doc := bson.D{}
	bytes, _ := bson.Marshal(data) //nolint:errcheck
	bson.Unmarshal(bytes, &doc)    //nolint:errcheck
	return doc
}

func deepCopyPost(src *posts.Post) *posts.Post {
	cpy := *src
	comms := make([]*posts.PostComment, 0, len(src.Comments))
	for _, comm := range src.Comments {
		comms = append(comms, &posts.PostComment{
			Created: comm.Created,
			Author: jwt.TokenPayload{
				Login: comm.Author.Login,
				ID:    comm.Author.ID,
			},
			Body: comm.Body,
			ID:   comm.ID,
		})
	}
	cpy.Comments = comms

	votes := make(posts.Votes, len(src.Votes))
	for key, val := range src.Votes {
		votes[key] = &posts.PostVote{
			UserID: val.UserID,
			Vote:   val.Vote,
		}
	}
	cpy.Votes = votes
	return &cpy
}

func TestGetAllPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		responses := make([]bson.D, 0, len(expectedPosts)+1)

		responses = append(responses,
			mtest.CreateCursorResponse(1, "db.test", mtest.FirstBatch, toBSON(expectedPosts[0])),
		)
		for i := 1; i < len(expectedPosts); i++ {
			responses = append(responses,
				mtest.CreateCursorResponse(1, "db.test", mtest.NextBatch, toBSON(expectedPosts[i])),
			)
		}
		responses = append(responses, mtest.CreateCursorResponse(1, "db.test", mtest.NextBatch))
		mt.AddMockResponses(responses...)

		posts, err := postRepo.GetAllPosts(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, len(expectedPosts), len(posts))
		assert.Equal(t, expectedPosts, posts)
	})

	mt.Run(t.Name()+"_find_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
			Message: findInternalErr,
		}))

		posts, err := postRepo.GetAllPosts(context.Background())
		assert.Error(t, err)
		assert.Nil(t, posts)
		assert.Contains(t, err.Error(), findInternalErr)
	})

	mt.Run(t.Name()+"_cursor_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		badRecord := mtest.CreateCursorResponse(1, "db.test", mtest.FirstBatch, toBSON(nil))
		mt.AddMockResponses(badRecord)

		posts, err := postRepo.GetAllPosts(context.Background())
		assert.Error(t, err)
		assert.Nil(t, posts)
		assert.Contains(t, err.Error(), "no responses remaining")
	})
}

func TestGetPostsByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		responses := make([]bson.D, 0, len(expectedPosts)+1)
		responses = append(responses,
			mtest.CreateCursorResponse(1, "db.test", mtest.FirstBatch, toBSON(expectedPosts[0])),
		)
		responses = append(responses, mtest.CreateCursorResponse(1, "db.test", mtest.NextBatch))
		mt.AddMockResponses(responses...)

		expected := []*posts.Post{expectedPosts[0]}
		posts, err := postRepo.GetPostsByCategory(context.Background(), posts.Music)

		assert.NoError(t, err)
		assert.Equal(t, len(expected), len(posts))
		assert.Equal(t, expected, posts)
	})

	mt.Run(t.Name()+"_find_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
			Message: findInternalErr,
		}))

		posts, err := postRepo.GetPostsByCategory(context.Background(), posts.Music)
		assert.Error(t, err)
		assert.Nil(t, posts)
		assert.Contains(t, err.Error(), findInternalErr)
	})

	mt.Run(t.Name()+"_cursor_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		badRecord := mtest.CreateCursorResponse(1, "db.test", mtest.FirstBatch, toBSON(nil))
		mt.AddMockResponses(badRecord)

		posts, err := postRepo.GetPostsByCategory(context.Background(), posts.Music)
		assert.Error(t, err)
		assert.Nil(t, posts)
		assert.Contains(t, err.Error(), "no responses remaining")

	})
}

func TestGetPostsByUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		responses := make([]bson.D, 0, len(expectedPosts)+1)
		responses = append(responses,
			mtest.CreateCursorResponse(1, "db.test", mtest.FirstBatch, toBSON(expectedPosts[0])),
			mtest.CreateCursorResponse(1, "db.test", mtest.NextBatch, toBSON(expectedPosts[1])),
		)
		responses = append(responses, mtest.CreateCursorResponse(1, "db.test", mtest.NextBatch))
		mt.AddMockResponses(responses...)

		expected := []*posts.Post{expectedPosts[0], expectedPosts[1]}
		posts, err := postRepo.GetPostsByUser(context.Background(), tokenPayloadAdmin.Login)

		assert.NoError(t, err)
		assert.Equal(t, len(expected), len(posts))
		assert.Equal(t, expected, posts)
	})

	mt.Run(t.Name()+"_find_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
			Message: findInternalErr,
		}))

		posts, err := postRepo.GetPostsByUser(context.Background(), tokenPayloadAdmin.Login)
		assert.Error(t, err)
		assert.Nil(t, posts)
		assert.Contains(t, err.Error(), findInternalErr)
	})

	mt.Run(t.Name()+"_cursor_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		badRecord := mtest.CreateCursorResponse(1, "db.test", mtest.FirstBatch, toBSON(nil))
		mt.AddMockResponses(badRecord)

		posts, err := postRepo.GetPostsByUser(context.Background(), tokenPayloadAdmin.Login)
		assert.Error(t, err)
		assert.Nil(t, posts)
		assert.Contains(t, err.Error(), "no responses remaining")

	})
}

func TestGetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abstractCollection := mocks.NewMockAbstractCollection(ctrl)
	singleResult := mocks.NewMockAbstractSingleResult(ctrl)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		filter := bson.M{"uuid": expected.ID}
		abstractCollection.EXPECT().FindOne(context.Background(), filter).Return(singleResult)
		singleResult.EXPECT().Err().Return(nil)
		singleResult.EXPECT().Decode(gomock.Any()).SetArg(0, *expected).Return(nil)

		post, err := postRepo.GetPostByID(context.Background(), expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, post)
	})

	mt.Run(t.Name()+"_post_not_found", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)

		abstractCollection.EXPECT().FindOne(context.Background(), gomock.Any()).Return(singleResult)
		singleResult.EXPECT().Err().Return(mongo.ErrNoDocuments)

		post, err := postRepo.GetPostByID(context.Background(), expectedPosts[0].ID)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errs.ErrPostNotFound)
	})

	mt.Run(t.Name()+"_decode_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)

		abstractCollection.EXPECT().FindOne(context.Background(), gomock.Any()).Return(singleResult)
		singleResult.EXPECT().Err().Return(nil)
		singleResult.EXPECT().Decode(gomock.Any()).Return(errSimulatedErr)

		post, err := postRepo.GetPostByID(context.Background(), expectedPosts[0].ID)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errSimulatedErr)
	})
}

func TestCreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		expected := expectedPosts[1]
		ctx := context.WithValue(context.Background(), jwt.Payload, tokenPayloadAdmin)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		post, err := postRepo.CreatePost(ctx, postPayload)
		assert.NoError(t, err)
		post.ID, post.Created = expected.ID, expected.Created
		assert.Equal(t, expected, post)
	})

	mt.Run(t.Name()+"_bad_payload", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		ctx := context.WithValue(context.Background(), jwt.Payload, "bad payload")

		post, err := postRepo.CreatePost(ctx, postPayload)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errs.ErrBadPayload)
	})

	mt.Run(t.Name()+"_insert_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		ctx := context.WithValue(context.Background(), jwt.Payload, tokenPayloadAdmin)

		mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
			Message: findInternalErr,
		}))

		post, err := postRepo.CreatePost(ctx, postPayload)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.Contains(t, err.Error(), findInternalErr)
	})
}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abstractCollection := mocks.NewMockAbstractCollection(ctrl)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)

		filter := bson.M{"uuid": expectedPosts[0].ID}
		abstractCollection.EXPECT().DeleteOne(context.Background(), filter).Return(int64(1), nil)

		err := postRepo.DeletePost(context.Background(), expectedPosts[0].ID)
		assert.NoError(t, err)
	})

	mt.Run(t.Name()+"_delete_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)

		abstractCollection.EXPECT().DeleteOne(context.Background(), gomock.Any()).Return(int64(0), errSimulatedErr)

		err := postRepo.DeletePost(context.Background(), expectedPosts[0].ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errSimulatedErr)
	})

	mt.Run(t.Name()+"_post_not_found", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)

		abstractCollection.EXPECT().DeleteOne(context.Background(), gomock.Any()).Return(int64(0), nil)

		err := postRepo.DeletePost(context.Background(), expectedPosts[0].ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrPostNotFound)
	})
}

func TestAddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abstractCollection := mocks.NewMockAbstractCollection(ctrl)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	commentBody := "comment body"
	ctx := context.WithValue(context.Background(), jwt.Payload, tokenPayloadAdmin)

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)
		filter := bson.M{"uuid": expected.ID}

		abstractCollection.EXPECT().UpdateOne(ctx, filter, gomock.Any()).Return(int64(1), nil)
		updatedPost.AddComment(*tokenPayloadAdmin, commentBody)

		post, err := postRepo.AddComment(ctx, expected, posts.Comment{Body: commentBody})
		assert.NoError(t, err)
		assert.Equal(t, updatedPost.Comments[len(updatedPost.Comments)-1].Body, post.Comments[len(post.Comments)-1].Body)
		assert.Equal(t, updatedPost.Comments[len(updatedPost.Comments)-1].Author, post.Comments[len(post.Comments)-1].Author)
	})

	mt.Run(t.Name()+"_bad_payload", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		badCTX := context.WithValue(context.Background(), jwt.Payload, "bad payload")

		post, err := postRepo.AddComment(badCTX, expectedPosts[0], posts.Comment{Body: commentBody})
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errs.ErrBadPayload)
	})

	mt.Run(t.Name()+"_update_comments_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		filter := bson.M{"uuid": expectedPosts[0].ID}

		abstractCollection.EXPECT().UpdateOne(ctx, filter, gomock.Any()).Return(int64(0), errSimulatedErr)

		post, err := postRepo.AddComment(ctx, expectedPosts[0], posts.Comment{Body: commentBody})
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errSimulatedErr)
	})
}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abstractCollection := mocks.NewMockAbstractCollection(ctrl)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.WithValue(context.Background(), jwt.Payload, tokenPayloadAdmin)

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)
		filter := bson.M{"uuid": expected.ID}
		update := bson.M{"$pull": bson.M{"comments": bson.M{"uuid": expected.Comments[0].ID}}}

		abstractCollection.EXPECT().UpdateOne(ctx, filter, update).Return(int64(1), nil)
		updatedPost.DeleteComment(updatedPost.Comments[0].ID) //nolint:errcheck

		post, err := postRepo.DeleteComment(ctx, expected, expected.Comments[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, len(updatedPost.Comments), len(post.Comments))
		assert.Equal(t, updatedPost.Comments, post.Comments)
	})

	mt.Run(t.Name()+"_comment_not_found", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])

		post, err := postRepo.DeleteComment(ctx, expected, expected.ID)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errs.ErrCommentNotFound)
	})

	mt.Run(t.Name()+"_update_comments_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		filter := bson.M{"uuid": expected.ID}
		update := bson.M{"$pull": bson.M{"comments": bson.M{"uuid": expected.Comments[0].ID}}}

		abstractCollection.EXPECT().UpdateOne(ctx, filter, update).Return(int64(0), errSimulatedErr)

		post, err := postRepo.DeleteComment(ctx, expected, expected.Comments[0].ID)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errSimulatedErr)
	})
}

func TestUpvote(t *testing.T) { //nolint:funlen
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abstractCollection := mocks.NewMockAbstractCollection(ctrl)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctxAdmin := context.WithValue(context.Background(), jwt.Payload, tokenPayloadAdmin)
	ctxUser := context.WithValue(context.Background(), jwt.Payload, tokenPayloadUser)

	mt.Run(t.Name()+"_success_create", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)
		newVote, _ := updatedPost.Upvote(tokenPayloadUser.ID)
		filter := bson.M{"uuid": expected.ID}
		update := bson.M{
			"$push": bson.M{
				"votes": newVote,
			},
			"$set": bson.M{
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxUser, filter, update).Return(int64(1), nil)

		post, err := postRepo.Upvote(ctxUser, expected)
		assert.NoError(t, err)
		assert.Equal(t, updatedPost, post)
	})

	mt.Run(t.Name()+"_create_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)
		newVote, _ := updatedPost.Upvote(tokenPayloadUser.ID)
		filter := bson.M{"uuid": expected.ID}
		update := bson.M{
			"$push": bson.M{
				"votes": newVote,
			},
			"$set": bson.M{
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxUser, filter, update).Return(int64(0), errSimulatedErr)

		post, err := postRepo.Upvote(ctxUser, expected)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errSimulatedErr)
	})

	mt.Run(t.Name()+"_bad_payload", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		ctx := context.WithValue(context.Background(), jwt.Payload, "bad payload")

		post, err := postRepo.Upvote(ctx, expectedPosts[0])
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errs.ErrBadPayload)
	})

	mt.Run(t.Name()+"_success_update", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)
		expected.Votes[tokenPayloadAdmin.ID].Vote = -1
		updatedPost.Votes[tokenPayloadAdmin.ID].Vote = -1
		expected.Score, updatedPost.Score = -1, -1

		newVote, _ := updatedPost.Upvote(tokenPayloadAdmin.ID)
		filter := bson.M{
			"uuid":       updatedPost.ID,
			"votes.user": newVote.UserID,
		}
		update := bson.M{
			"$set": bson.M{
				"votes.$.vote":     newVote.Vote,
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxAdmin, filter, update).Return(int64(1), nil)

		post, err := postRepo.Upvote(ctxAdmin, expected)
		assert.NoError(t, err)
		assert.Equal(t, updatedPost, post)
	})

	mt.Run(t.Name()+"_update_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)
		expected.Votes[tokenPayloadAdmin.ID].Vote = -1
		updatedPost.Votes[tokenPayloadAdmin.ID].Vote = -1
		expected.Score, updatedPost.Score = -1, -1

		newVote, _ := updatedPost.Upvote(tokenPayloadAdmin.ID)
		filter := bson.M{
			"uuid":       updatedPost.ID,
			"votes.user": newVote.UserID,
		}
		update := bson.M{
			"$set": bson.M{
				"votes.$.vote":     newVote.Vote,
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxAdmin, filter, update).Return(int64(0), errSimulatedErr)

		post, err := postRepo.Upvote(ctxAdmin, expected)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errSimulatedErr)
	})
}

func TestDownVote(t *testing.T) { //nolint:funlen
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abstractCollection := mocks.NewMockAbstractCollection(ctrl)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctxAdmin := context.WithValue(context.Background(), jwt.Payload, tokenPayloadAdmin)
	ctxUser := context.WithValue(context.Background(), jwt.Payload, tokenPayloadUser)

	mt.Run(t.Name()+"_success_create", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)
		newVote, _ := updatedPost.Downvote(tokenPayloadUser.ID)
		filter := bson.M{"uuid": expected.ID}
		update := bson.M{
			"$push": bson.M{
				"votes": newVote,
			},
			"$set": bson.M{
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxUser, filter, update).Return(int64(1), nil)

		post, err := postRepo.Downvote(ctxUser, expected)
		assert.NoError(t, err)
		assert.Equal(t, updatedPost, post)
	})

	mt.Run(t.Name()+"_create_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)
		newVote, _ := updatedPost.Downvote(tokenPayloadUser.ID)
		filter := bson.M{"uuid": expected.ID}
		update := bson.M{
			"$push": bson.M{
				"votes": newVote,
			},
			"$set": bson.M{
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxUser, filter, update).Return(int64(0), errSimulatedErr)

		post, err := postRepo.Downvote(ctxUser, expected)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errSimulatedErr)
	})

	mt.Run(t.Name()+"_bad_payload", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		ctx := context.WithValue(context.Background(), jwt.Payload, "bad payload")

		post, err := postRepo.Downvote(ctx, expectedPosts[0])
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errs.ErrBadPayload)
	})

	mt.Run(t.Name()+"_success_update", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)

		newVote, _ := updatedPost.Downvote(tokenPayloadAdmin.ID)
		filter := bson.M{
			"uuid":       updatedPost.ID,
			"votes.user": newVote.UserID,
		}
		update := bson.M{
			"$set": bson.M{
				"votes.$.vote":     newVote.Vote,
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxAdmin, filter, update).Return(int64(1), nil)

		post, err := postRepo.Downvote(ctxAdmin, expected)
		assert.NoError(t, err)
		assert.Equal(t, updatedPost, post)
	})

	mt.Run(t.Name()+"_update_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)

		newVote, _ := updatedPost.Downvote(tokenPayloadAdmin.ID)
		filter := bson.M{
			"uuid":       updatedPost.ID,
			"votes.user": newVote.UserID,
		}
		update := bson.M{
			"$set": bson.M{
				"votes.$.vote":     newVote.Vote,
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxAdmin, filter, update).Return(int64(0), errSimulatedErr)

		post, err := postRepo.Downvote(ctxAdmin, expected)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errSimulatedErr)
	})
}

func TestUnvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abstractCollection := mocks.NewMockAbstractCollection(ctrl)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctxAdmin := context.WithValue(context.Background(), jwt.Payload, tokenPayloadAdmin)
	ctxUser := context.WithValue(context.Background(), jwt.Payload, tokenPayloadUser)

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)

		updatedPost.Unvote(tokenPayloadAdmin.ID) //nolint:errcheck
		filter := bson.M{"uuid": updatedPost.ID}
		update := bson.M{
			"$pull": bson.M{
				"votes": bson.M{
					"user": tokenPayloadAdmin.ID,
				},
			},
			"$set": bson.M{
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxAdmin, filter, update).Return(int64(1), nil)

		post, err := postRepo.Unvote(ctxAdmin, expected)
		assert.NoError(t, err)
		assert.Equal(t, updatedPost, post)
	})

	mt.Run(t.Name()+"_bad_payload", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(storage.NewMongoCollection(mt.Coll))
		ctx := context.WithValue(context.Background(), jwt.Payload, "bad payload")

		post, err := postRepo.Unvote(ctx, expectedPosts[0])
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errs.ErrBadPayload)
	})

	mt.Run(t.Name()+"_vote_not_found", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])

		post, err := postRepo.Unvote(ctxUser, expected)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errs.ErrVoteNotFound)
	})

	mt.Run(t.Name()+"_update_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		expected := deepCopyPost(expectedPosts[0])
		updatedPost := deepCopyPost(expected)

		updatedPost.Unvote(tokenPayloadAdmin.ID) //nolint:errcheck
		filter := bson.M{"uuid": updatedPost.ID}
		update := bson.M{
			"$pull": bson.M{
				"votes": bson.M{
					"user": tokenPayloadAdmin.ID,
				},
			},
			"$set": bson.M{
				"upvotePercentage": updatedPost.UpvotePercentage,
				"score":            updatedPost.Score,
			},
		}

		abstractCollection.EXPECT().UpdateOne(ctxAdmin, filter, update).Return(int64(0), errSimulatedErr)

		post, err := postRepo.Unvote(ctxAdmin, expected)
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.ErrorIs(t, err, errSimulatedErr)
	})
}

func TestUpdateViews(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abstractCollection := mocks.NewMockAbstractCollection(ctrl)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run(t.Name()+"_success", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		filter := bson.M{"uuid": expectedPosts[0].ID}
		update := bson.M{"$inc": bson.M{"views": 1}}

		abstractCollection.EXPECT().UpdateOne(context.Background(), filter, update).Return(int64(1), nil)

		err := postRepo.UpdateViews(context.Background(), expectedPosts[0].ID)
		assert.NoError(t, err)
	})

	mt.Run(t.Name()+"_update_error", func(mt *mtest.T) {
		postRepo := storage.NewPostRepoMongoDB(abstractCollection)
		filter := bson.M{"uuid": expectedPosts[0].ID}
		update := bson.M{"$inc": bson.M{"views": 1}}

		abstractCollection.EXPECT().UpdateOne(context.Background(), filter, update).Return(int64(0), errSimulatedErr)

		err := postRepo.UpdateViews(context.Background(), expectedPosts[0].ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errSimulatedErr)
	})
}
