package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/models/posts"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
	"github.com/Benzogang-Tape/Reddit/internal/storage/mocks"
	"github.com/Benzogang-Tape/Reddit/internal/transport/rest"
)

var (
	validPostPayload = posts.PostPayload{
		Type:     posts.WithText,
		Title:    "TEST POST",
		URL:      "",
		Category: posts.Music,
		Text:     "I love music of Kartik.",
	}
	invalidPostPayload = posts.PostPayload{
		Type:     posts.WithLink,
		Title:    "Test invalid payload",
		URL:      "a:",
		Category: posts.Programming,
		Text:     "",
	}
	rawCommentPayload        = `{"comment":"New comment"}`
	rawInvalidCommentPayload = `{"comment":""}`
	commentPayload           = posts.Comment{
		Body: "New comment",
	}
	invalidCommentPayload          = posts.Comment{}
	fakeID                users.ID = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	postList                       = []*posts.Post{
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
	}
)

func TestGetAllPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	st.EXPECT().GetAllPosts(context.Background()).Return(postList, nil)

	r := httptest.NewRequest("GET", "/api/posts/", nil)
	w := httptest.NewRecorder()

	handler.GetAllPosts(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Unknown error
	st.EXPECT().GetAllPosts(context.Background()).Return(nil, errs.ErrUnknownError)

	r = httptest.NewRequest("GET", "/api/posts/", nil)
	w = httptest.NewRecorder()

	handler.GetAllPosts(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())

}

func TestCreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	st.EXPECT().CreatePost(context.Background(), validPostPayload).Return(postList[0], nil)
	rawValidPostPayload, _ := json.Marshal(validPostPayload) //nolint:errcheck
	r := httptest.NewRequest("POST", "/api/posts", bytes.NewReader(rawValidPostPayload))
	w := httptest.NewRecorder()

	handler.CreatePost(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList[0]) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Read body error
	r = httptest.NewRequest("POST", "/api/posts", bytes.NewReader(nil))
	r.Body = &fakeBody{data: rawCredentials}
	w = httptest.NewRecorder()
	handler.CreatePost(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Unmarshal body error
	r = httptest.NewRequest("POST", "/api/posts", bytes.NewReader(nil))
	w = httptest.NewRecorder()
	handler.CreatePost(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Unknown error
	st.EXPECT().CreatePost(context.Background(), validPostPayload).Return(nil, errs.ErrUnknownError)
	rawValidPostPayload, _ = json.Marshal(validPostPayload) //nolint:errcheck
	r = httptest.NewRequest("POST", "/api/posts", bytes.NewReader(rawValidPostPayload))
	w = httptest.NewRecorder()

	handler.CreatePost(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())

	// Invalid url
	st.EXPECT().CreatePost(context.Background(), invalidPostPayload).Return(nil, errs.ErrInvalidURL)
	rawInvalidPostPayload, _ := json.Marshal(invalidPostPayload) //nolint:errcheck
	r = httptest.NewRequest("POST", "/api/posts", bytes.NewReader(rawInvalidPostPayload))
	w = httptest.NewRecorder()

	handler.CreatePost(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Contains(t, string(body), `"msg":"is invalid"`)
}

func TestGetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w := httptest.NewRecorder()
	st.EXPECT().GetPostByID(r.Context(), postList[0].ID).Return(postList[0], nil)

	handler.GetPostByID(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList[0]) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Invalid post id
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": "1",
	})
	w = httptest.NewRecorder()

	handler.GetPostByID(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidPostID.Error())

	// Post not found
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(fakeID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().GetPostByID(r.Context(), fakeID).Return(nil, errs.ErrPostNotFound)

	handler.GetPostByID(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrPostNotFound.Error())

	// Unknown error
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().GetPostByID(r.Context(), postList[0].ID).Return(nil, errs.ErrUnknownError)

	handler.GetPostByID(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestGetPostsByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("GET", "/api/posts/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"CATEGORY_NAME": posts.Music.String(),
	})
	w := httptest.NewRecorder()
	st.EXPECT().GetPostsByCategory(r.Context(), posts.Music).Return(postList, nil)

	handler.GetPostsByCategory(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Invalid category
	r = httptest.NewRequest("GET", "/api/posts/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"CATEGORY_NAME": "ski",
	})
	w = httptest.NewRecorder()

	handler.GetPostsByCategory(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidCategory.Error())

	// Unknown Error
	r = httptest.NewRequest("GET", "/api/posts/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"CATEGORY_NAME": posts.Music.String(),
	})
	w = httptest.NewRecorder()
	st.EXPECT().GetPostsByCategory(r.Context(), posts.Music).Return(nil, errs.ErrUnknownError)

	handler.GetPostsByCategory(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestGetPostsByUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("GET", "/api/user/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"USER_LOGIN": string(postList[0].Author.Login),
	})
	w := httptest.NewRecorder()
	st.EXPECT().GetPostsByUser(r.Context(), postList[0].Author.Login).Return(postList, nil)

	handler.GetPostsByUser(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Unknown error
	r = httptest.NewRequest("GET", "/api/user/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"USER_LOGIN": string(postList[0].Author.Login),
	})
	w = httptest.NewRecorder()
	st.EXPECT().GetPostsByUser(r.Context(), postList[0].Author.Login).Return(nil, errs.ErrUnknownError)

	handler.GetPostsByUser(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w := httptest.NewRecorder()
	st.EXPECT().DeletePost(r.Context(), postList[0].ID).Return(nil)

	handler.DeletePost(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "success")

	// Invalid post id
	r = httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": "1",
	})
	w = httptest.NewRecorder()

	handler.DeletePost(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidPostID.Error())

	// Post not found
	r = httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(fakeID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().DeletePost(r.Context(), fakeID).Return(errs.ErrPostNotFound)

	handler.DeletePost(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrPostNotFound.Error())

	// Unknown error
	r = httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().DeletePost(r.Context(), postList[0].ID).Return(errs.ErrUnknownError)

	handler.DeletePost(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestUpvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w := httptest.NewRecorder()
	st.EXPECT().Upvote(r.Context(), postList[0].ID).Return(postList[0], nil)

	handler.Upvote(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList[0]) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Invalid post id
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": "1",
	})
	w = httptest.NewRecorder()

	handler.Upvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidPostID.Error())

	// Post not found
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(fakeID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().Upvote(r.Context(), fakeID).Return(nil, errs.ErrPostNotFound)

	handler.Upvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrPostNotFound.Error())

	// Unknown error
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().Upvote(r.Context(), postList[0].ID).Return(nil, errs.ErrUnknownError)

	handler.Upvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestDownvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w := httptest.NewRecorder()
	st.EXPECT().Downvote(r.Context(), postList[0].ID).Return(postList[0], nil)

	handler.Downvote(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList[0]) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Invalid post id
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": "1",
	})
	w = httptest.NewRecorder()

	handler.Downvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidPostID.Error())

	// Post not found
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(fakeID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().Downvote(r.Context(), fakeID).Return(nil, errs.ErrPostNotFound)

	handler.Downvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrPostNotFound.Error())

	// Unknown error
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().Downvote(r.Context(), postList[0].ID).Return(nil, errs.ErrUnknownError)

	handler.Downvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestUnvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w := httptest.NewRecorder()
	st.EXPECT().Unvote(r.Context(), postList[0].ID).Return(postList[0], nil)

	handler.Unvote(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList[0]) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Invalid post id
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": "1",
	})
	w = httptest.NewRecorder()

	handler.Unvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidPostID.Error())

	// Post not found
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(fakeID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().Unvote(r.Context(), fakeID).Return(nil, errs.ErrPostNotFound)

	handler.Unvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrPostNotFound.Error())

	// Unknown error
	r = httptest.NewRequest("GET", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().Unvote(r.Context(), postList[0].ID).Return(nil, errs.ErrUnknownError)

	handler.Unvote(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())
}

func TestAddComment(t *testing.T) { //nolint:funlen
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("POST", "/api/post/", strings.NewReader(rawCommentPayload))
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w := httptest.NewRecorder()
	st.EXPECT().AddComment(r.Context(), postList[0].ID, commentPayload).Return(postList[0], nil)

	handler.AddComment(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList[0]) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Read body error
	r = httptest.NewRequest("POST", "/api/post/", bytes.NewReader(nil))
	r.Body = &fakeBody{data: rawCredentials}
	w = httptest.NewRecorder()
	handler.AddComment(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Unmarshal body error
	r = httptest.NewRequest("POST", "/api/post/", bytes.NewReader(nil))
	w = httptest.NewRecorder()
	handler.AddComment(w, r)
	resp = w.Result() //nolint:bodyclose

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Invalid post id
	r = httptest.NewRequest("POST", "/api/post/", strings.NewReader(rawCommentPayload))
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": "1",
	})
	w = httptest.NewRecorder()

	handler.AddComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidPostID.Error())

	// Invalid comment body
	r = httptest.NewRequest("POST", "/api/post/", strings.NewReader(rawInvalidCommentPayload))
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().AddComment(r.Context(), postList[0].ID, invalidCommentPayload).Return(nil, errs.ErrBadCommentBody)

	handler.AddComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Contains(t, string(body), "is required")

	// Post not found
	r = httptest.NewRequest("POST", "/api/post/", strings.NewReader(rawCommentPayload))
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(fakeID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().AddComment(r.Context(), fakeID, commentPayload).Return(nil, errs.ErrPostNotFound)

	handler.AddComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrPostNotFound.Error())

	// Unknown error
	r = httptest.NewRequest("POST", "/api/post/", strings.NewReader(rawCommentPayload))
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": string(postList[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().AddComment(r.Context(), postList[0].ID, commentPayload).Return(nil, errs.ErrUnknownError)

	handler.AddComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())

}

func TestDeleteComment(t *testing.T) { //nolint:funlen
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := mocks.NewMockPostAPI(ctrl)
	handler := rest.NewPostHandler(st, zap.NewNop().Sugar())

	// Success
	r := httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID":    string(postList[0].ID),
		"COMMENT_ID": string(postList[0].Comments[0].ID),
	})
	w := httptest.NewRecorder()
	st.EXPECT().DeleteComment(r.Context(), postList[0].ID, postList[0].Comments[0].ID).Return(postList[0], nil)

	handler.DeleteComment(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	expectedData, _ := json.Marshal(postList[0]) //nolint:errcheck
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expectedData, body)

	// Invalid post id
	r = httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID": "1",
	})
	w = httptest.NewRecorder()

	handler.DeleteComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidPostID.Error())

	// Invalid comment id
	r = httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID":    string(postList[0].ID),
		"COMMENT_ID": "1",
	})
	w = httptest.NewRecorder()

	handler.DeleteComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrInvalidCommentID.Error())

	// Post not found
	r = httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID":    string(fakeID),
		"COMMENT_ID": string(postList[0].Comments[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().DeleteComment(r.Context(), fakeID, postList[0].Comments[0].ID).Return(nil, errs.ErrPostNotFound)

	handler.DeleteComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrPostNotFound.Error())

	// Comment not found
	r = httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID":    string(postList[0].ID),
		"COMMENT_ID": string(fakeID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().DeleteComment(r.Context(), postList[0].ID, fakeID).Return(nil, errs.ErrCommentNotFound)

	handler.DeleteComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrCommentNotFound.Error())

	// Unknown error
	r = httptest.NewRequest("DELETE", "/api/post/", nil)
	r = mux.SetURLVars(r, map[string]string{
		"POST_ID":    string(postList[0].ID),
		"COMMENT_ID": string(postList[0].Comments[0].ID),
	})
	w = httptest.NewRecorder()
	st.EXPECT().DeleteComment(r.Context(), postList[0].ID, postList[0].Comments[0].ID).Return(nil, errs.ErrUnknownError)

	handler.DeleteComment(w, r)
	resp = w.Result()
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), errs.ErrUnknownError.Error())

}
