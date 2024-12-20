package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/httpresp"
	"github.com/Benzogang-Tape/Reddit/internal/models/posts"
	"github.com/Benzogang-Tape/Reddit/internal/models/users"
)

//go:generate mockgen -source=post.go -destination=../../storage/mocks/posts_repo_mongoDB_mock.go -package=mocks PostAPI
type PostAPI interface {
	GetAllPosts(ctx context.Context) ([]*posts.Post, error)
	GetPostsByCategory(ctx context.Context, postCategory posts.PostCategory) ([]*posts.Post, error)
	GetPostsByUser(ctx context.Context, userLogin users.Username) ([]*posts.Post, error)
	GetPostByID(ctx context.Context, postID users.ID) (*posts.Post, error)
	CreatePost(ctx context.Context, postPayload posts.PostPayload) (*posts.Post, error)
	DeletePost(ctx context.Context, postID users.ID) error
	AddComment(ctx context.Context, postID users.ID, comment posts.Comment) (*posts.Post, error)
	DeleteComment(ctx context.Context, postID, commentID users.ID) (*posts.Post, error)
	Upvote(ctx context.Context, postID users.ID) (*posts.Post, error)
	Downvote(ctx context.Context, postID users.ID) (*posts.Post, error)
	Unvote(ctx context.Context, postID users.ID) (*posts.Post, error)
}

type PostHandler struct {
	logger  *zap.SugaredLogger
	service PostAPI
}

func NewPostHandler(p PostAPI, logger *zap.SugaredLogger) *PostHandler {
	return &PostHandler{
		logger:  logger,
		service: p,
	}
}

func validateID(alias string, params map[string]string) (id users.ID, err error) {
	extractedID := params[alias]
	if utf8.RuneCountInString(extractedID) != posts.UUIDLength {
		return id, errs.ErrBadID
	}

	return users.ID(extractedID), nil
}

// GetAllPosts godoc
//
//	@Summary		Get all posts
//	@Description	Get a list of posts of all users and threads
//	@Tags			getting-posts
//	@ID				get-all-posts
//	@Produce		json
//	@Success		200	{array}		posts.Post		"Posts successfully received"
//	@Failure		500	{object}	errs.SimpleErr	"Internal server error"
//	@Router			/posts/ [get]
func (p *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	postList, err := p.service.GetAllPosts(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(postList, w)
}

// CreatePost godoc
//
//	@Summary		Create a post
//	@Description	Create a post of a specific type, category, and content
//	@Security		ApiKeyAuth
//	@Tags			managing-posts
//	@ID				create-post
//	@Accept			json
//	@Produce		json
//	@Param			post_payload	body		posts.PostPayload	true	"Post data"	validate(required)
//	@Success		201				{object}	posts.Post			"Post successfully created"
//	@Failure		400				"Bad payload"
//	@Failure		422				{object}	errs.ComplexErrArr	"Bad content"
//	@Failure		500				{object}	errs.SimpleErr		"Internal server error"
//	@Router			/posts [post]
func (p *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	postPayload := posts.PostPayload{}
	if err = json.Unmarshal(body, &postPayload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newPost, err := p.service.CreatePost(r.Context(), postPayload)
	switch {
	case errors.Is(err, errs.ErrInvalidURL):
		sendErrorResponse(w, http.StatusUnprocessableEntity, errs.NewComplexErrArr(errs.ComplexErr{
			Location: "body",
			Param:    "url",
			Value:    postPayload.URL,
			Msg:      "is invalid",
		}))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(newPost, w, httpresp.WithStatusCode(http.StatusCreated))
}

// GetPostByID godoc
//
//	@Summary		Get a certain post
//	@Description	Get information on a specific post by id
//	@Tags			getting-posts
//	@ID				get-post-by-id
//	@Produce		json
//	@Param			POST_ID	path		string			true	"Post uuid"	minlength(36)	maxlength(36)
//	@Success		200		{object}	posts.Post		"Post successfully received"
//	@Failure		400		{object}	errs.SimpleErr	"Bad post id"
//	@Failure		404		{object}	errs.SimpleErr	"No posts with the provided id were found"
//	@Failure		500		{object}	errs.SimpleErr	"Internal server error"
//	@Router			/post/{POST_ID} [get]
func (p *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	postID, err := validateID("POST_ID", mux.Vars(r))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.GetPostByID(r.Context(), postID)
	switch {
	case errors.Is(err, errs.ErrPostNotFound):
		sendErrorResponse(w, http.StatusNotFound, errs.NewSimpleErr(errs.ErrPostNotFound.Error()))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(post, w)
}

// GetPostsByCategory godoc
//
//	@Summary		Get posts by category
//	@Description	Get all posts belonging to a certain category
//	@Tags			getting-posts
//	@ID				get-posts-by-category
//	@Produce		json
//	@Param			CATEGORY_NAME	path		string			true	"Category name"
//	@Success		200				{array}		posts.Post		"Posts successfully received"
//	@Failure		400				{object}	errs.SimpleErr	"Bad category(doesn't exist)"
//	@Failure		500				{object}	errs.SimpleErr	"Internal server error"
//	@Router			/posts/{CATEGORY_NAME} [get]
func (p *PostHandler) GetPostsByCategory(w http.ResponseWriter, r *http.Request) {
	postCategory, err := posts.StringToPostCategory(mux.Vars(r)["CATEGORY_NAME"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidCategory.Error()))
		return
	}

	postList, err := p.service.GetPostsByCategory(r.Context(), postCategory)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(postList, w)
}

// GetPostsByUser godoc
//
//	@Summary		Get posts by user
//	@Description	Get all posts of a certain user by his/her username
//	@Tags			getting-posts
//	@ID				get-posts-by-user
//	@Produce		json
//	@Param			USER_LOGIN	path		string			true	"Username of user"
//	@Success		200			{array}		posts.Post		"Posts successfully received"
//	@Failure		400			{object}	errs.SimpleErr	"Bad username(doesn't exist)"
//	@Failure		500			{object}	errs.SimpleErr	"Internal server error"
//	@Router			/user/{USER_LOGIN} [get]
func (p *PostHandler) GetPostsByUser(w http.ResponseWriter, r *http.Request) {
	userLogin := users.Username(mux.Vars(r)["USER_LOGIN"])
	postList, err := p.service.GetPostsByUser(r.Context(), userLogin)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(postList, w)
}

// DeletePost godoc
//
//	@Summary		Delete a post
//	@Description	Delete a specific post by its id
//	@Security		ApiKeyAuth
//	@Tags			managing-posts
//	@ID				delete-post
//	@Param			POST_ID	path		string			true	"Post uuid"	minlength(36)	maxlength(36)
//	@Success		200		{object}	errs.SimpleErr	"Post successfully deleted"
//	@Failure		400		{object}	errs.SimpleErr	"Bad post id"
//	@Failure		404		{object}	errs.SimpleErr	"No posts with the provided id were found"
//	@Failure		500		{object}	errs.SimpleErr	"Internal server error"
//	@Router			/post/{POST_ID} [delete]
func (p *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := validateID("POST_ID", mux.Vars(r))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidPostID.Error()))
		return
	}

	err = p.service.DeletePost(r.Context(), postID)
	switch {
	case errors.Is(err, errs.ErrPostNotFound):
		sendErrorResponse(w, http.StatusNotFound, errs.NewSimpleErr(errs.ErrPostNotFound.Error()))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendErrorResponse(w, http.StatusOK, errs.NewSimpleErr("success"))
}

// Upvote godoc
//
//	@Summary		Vote up on a post
//	@Description	Increase post rating by 1 vote
//	@Security		ApiKeyAuth
//	@Tags			voting-posts
//	@ID				upvote-post
//	@Produce		json
//	@Param			POST_ID	path		string			true	"Post uuid"	minlength(36)	maxlength(36)
//	@Success		200		{object}	posts.Post		"Successfully upvoted"
//	@Failure		400		{object}	errs.SimpleErr	"Bad post id"
//	@Failure		404		{object}	errs.SimpleErr	"No posts with the provided id were found"
//	@Failure		500		{object}	errs.SimpleErr	"Internal server error"
//	@Router			/post/{POST_ID}/upvote [get]
func (p *PostHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	postID, err := validateID("POST_ID", mux.Vars(r))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.Upvote(r.Context(), postID)
	switch {
	case errors.Is(err, errs.ErrPostNotFound):
		sendErrorResponse(w, http.StatusNotFound, errs.NewSimpleErr(errs.ErrPostNotFound.Error()))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(post, w)
}

// Downvote godoc
//
//	@Summary		Vote down on a post
//	@Description	Decrease post rating by 1 vote
//	@Security		ApiKeyAuth
//	@Tags			voting-posts
//	@ID				downvote-post
//	@Produce		json
//	@Param			POST_ID	path		string			true	"Post uuid"	minlength(36)	maxlength(36)
//	@Success		200		{object}	posts.Post		"Successfully downvoted"
//	@Failure		400		{object}	errs.SimpleErr	"Bad post id"
//	@Failure		404		{object}	errs.SimpleErr	"No posts with the provided id were found"
//	@Failure		500		{object}	errs.SimpleErr	"Internal server error"
//	@Router			/post/{POST_ID}/downvote [get]
func (p *PostHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	postID, err := validateID("POST_ID", mux.Vars(r))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.Downvote(r.Context(), postID)
	switch {
	case errors.Is(err, errs.ErrPostNotFound):
		sendErrorResponse(w, http.StatusNotFound, errs.NewSimpleErr(errs.ErrPostNotFound.Error()))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(post, w)
}

// Unvote godoc
//
//	@Summary		Cancel your vote
//	@Description	Withdraw your vote from the post
//	@Security		ApiKeyAuth
//	@Tags			voting-posts
//	@ID				unvote-post
//	@Produce		json
//	@Param			POST_ID	path		string			true	"Post uuid"	minlength(36)	maxlength(36)
//	@Success		200		{object}	posts.Post		"Successfully unvoted"
//	@Failure		400		{object}	errs.SimpleErr	"Bad post id"
//	@Failure		404		{object}	errs.SimpleErr	"No posts with the provided id were found"
//	@Failure		500		{object}	errs.SimpleErr	"Internal server error"
//	@Router			/post/{POST_ID}/unvote [get]
func (p *PostHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	postID, err := validateID("POST_ID", mux.Vars(r))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.Unvote(r.Context(), postID)
	switch {
	case errors.Is(err, errs.ErrPostNotFound):
		sendErrorResponse(w, http.StatusNotFound, errs.NewSimpleErr(errs.ErrPostNotFound.Error()))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(post, w)
}

// AddComment godoc
//
//	@Summary		Comment on the post
//	@Description	Leave a comment under a certain post
//	@Security		ApiKeyAuth
//	@Tags			commenting-posts
//	@ID				add-comment
//	@Accept			json
//	@Produce		json
//	@Param			comment_payload	body		posts.Comment		true	"Comment data"	validate(required)
//	@Param			POST_ID			path		string				true	"Post uuid"		minlength(36)	maxlength(36)
//	@Success		201				{object}	posts.Post			"Comment successfully left"
//	@Failure		400				{object}	errs.SimpleErr		"Bad payload"
//	@Failure		404				{object}	errs.SimpleErr		"No posts with the provided id were found"
//	@Failure		422				{object}	errs.ComplexErrArr	"Bad content"
//	@Failure		500				{object}	errs.SimpleErr		"Internal server error"
//	@Router			/posts/{POST_ID} [post]
func (p *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	comment := posts.Comment{}
	if err = json.Unmarshal(body, &comment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	postID, err := validateID("POST_ID", mux.Vars(r))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.AddComment(r.Context(), postID, comment)
	switch {
	case errors.Is(err, errs.ErrBadCommentBody):
		sendErrorResponse(w, http.StatusUnprocessableEntity, errs.NewComplexErrArr(errs.ComplexErr{
			Location: "body",
			Param:    "comment",
			Msg:      "is required",
		}))
		return
	case errors.Is(err, errs.ErrPostNotFound):
		sendErrorResponse(w, http.StatusNotFound, errs.NewSimpleErr(errs.ErrPostNotFound.Error()))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(post, w, httpresp.WithStatusCode(http.StatusCreated))
}

// DeleteComment godoc
//
//	@Summary		Delete comment
//	@Description	Delete a certain comment on a certain post
//	@Security		ApiKeyAuth
//	@Tags			commenting-posts
//	@ID				delete-comment
//	@Accept			json
//	@Produce		json
//	@Param			POST_ID		path		string			true	"Post uuid"		minlength(36)	maxlength(36)
//	@Param			COMMENT_ID	path		string			true	"Comment uuid"	minlength(36)	maxlength(36)
//	@Success		200			{object}	posts.Post		"Comment successfully deleted"
//	@Failure		400			{object}	errs.SimpleErr	"Bad uuid"
//	@Failure		404			{object}	errs.SimpleErr	"No posts or comment with the provided id were found"
//	@Failure		500			{object}	errs.SimpleErr	"Internal server error"
//	@Router			/posts/{POST_ID}/{COMMENT_ID} [delete]
func (p *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	postID, err := validateID("POST_ID", params)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidPostID.Error()))
		return
	}
	commentID, err := validateID("COMMENT_ID", params)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.NewSimpleErr(errs.ErrInvalidCommentID.Error()))
		return
	}

	post, err := p.service.DeleteComment(r.Context(), postID, commentID)
	switch {
	case errors.Is(err, errs.ErrPostNotFound):
		sendErrorResponse(w, http.StatusNotFound, errs.NewSimpleErr(errs.ErrPostNotFound.Error()))
		return
	case errors.Is(err, errs.ErrCommentNotFound):
		sendErrorResponse(w, http.StatusNotFound, errs.NewSimpleErr(errs.ErrCommentNotFound.Error()))
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrUnknownError.Error()))
		return
	}

	sendResponse(post, w)

}
