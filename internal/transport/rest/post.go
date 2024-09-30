package rest

import (
	"encoding/json"
	"errors"
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"github.com/Benzogang-Tape/Reddit/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io"
	"net/http"
	"unicode/utf8"
)

type PostAPI interface {
	service.PostStorage
	service.PostActions
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

func (p *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	postList, err := p.service.GetAllPosts(r.Context())
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(postList)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	if _, err = w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	postPayload := models.PostPayload{}
	if err = json.Unmarshal(body, &postPayload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newPost, err := p.service.CreatePost(r.Context(), postPayload)
	if errors.Is(err, models.ErrInvalidURL) {
		jsonComplexErr(w, http.StatusUnprocessableEntity, models.NewComplexErr(models.ComplexErr{
			Location: "body",
			Param:    "url",
			Value:    postPayload.URL,
			Msg:      "is invalid",
		}))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(newPost)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	postID := models.ID(mux.Vars(r)["POST_ID"])
	if utf8.RuneCountInString(string(postID)) != models.UUIDLength {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.GetPostByID(r.Context(), postID)
	if errors.Is(err, models.ErrPostNotFound) {
		jsonSimpleErr(w, http.StatusNotFound, models.NewSimpleErr(models.ErrPostNotFound.Error()))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(post)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	if _, err = w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) GetPostsByCategory(w http.ResponseWriter, r *http.Request) {
	postCategory, err := models.StringToPostCategory(mux.Vars(r)["CATEGORY_NAME"])
	if err != nil {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidCategory.Error()))
		return
	}

	postList, err := p.service.GetPostsByCategory(r.Context(), postCategory)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(postList)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	if _, err := w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) GetPostsByUser(w http.ResponseWriter, r *http.Request) {
	userLogin := models.Username(mux.Vars(r)["USER_LOGIN"])
	postList, err := p.service.GetPostsByUser(r.Context(), userLogin)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(postList)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	if _, err := w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID := models.ID(mux.Vars(r)["POST_ID"])
	if utf8.RuneCountInString(string(postID)) != models.UUIDLength {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidPostID.Error()))
		return
	}

	err := p.service.DeletePost(r.Context(), postID)
	if errors.Is(err, models.ErrPostNotFound) {
		jsonSimpleErr(w, http.StatusNotFound, models.NewSimpleErr(models.ErrPostNotFound.Error()))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}
	jsonSimpleErr(w, http.StatusOK, models.NewSimpleErr("success"))
}

func (p *PostHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	postID := models.ID(mux.Vars(r)["POST_ID"])
	if utf8.RuneCountInString(string(postID)) != models.UUIDLength {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.Upvote(r.Context(), postID)
	if errors.Is(err, models.ErrPostNotFound) {
		jsonSimpleErr(w, http.StatusNotFound, models.NewSimpleErr(models.ErrPostNotFound.Error()))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(post)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	if _, err = w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	postID := models.ID(mux.Vars(r)["POST_ID"])
	if utf8.RuneCountInString(string(postID)) != models.UUIDLength {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.Downvote(r.Context(), postID)
	if errors.Is(err, models.ErrPostNotFound) {
		jsonSimpleErr(w, http.StatusNotFound, models.NewSimpleErr(models.ErrPostNotFound.Error()))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(post)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	if _, err = w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	postID := models.ID(mux.Vars(r)["POST_ID"])
	if utf8.RuneCountInString(string(postID)) != models.UUIDLength {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.Unvote(r.Context(), postID)
	if errors.Is(err, models.ErrPostNotFound) {
		jsonSimpleErr(w, http.StatusNotFound, models.NewSimpleErr(models.ErrPostNotFound.Error()))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(post)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	if _, err = w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	comment := models.Comment{}
	if err = json.Unmarshal(body, &comment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	postID := models.ID(mux.Vars(r)["POST_ID"])
	if utf8.RuneCountInString(string(postID)) != models.UUIDLength {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidPostID.Error()))
		return
	}

	post, err := p.service.AddComment(r.Context(), postID, comment)
	if errors.Is(err, models.ErrBadCommentBody) {
		jsonComplexErr(w, http.StatusUnprocessableEntity, models.NewComplexErr(models.ComplexErr{
			Location: "body",
			Param:    "comment",
			Msg:      "is required",
		}))
		return
	}
	if errors.Is(err, models.ErrPostNotFound) {
		jsonSimpleErr(w, http.StatusNotFound, models.NewSimpleErr(models.ErrPostNotFound.Error()))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(post)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}

func (p *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	postID := models.ID(mux.Vars(r)["POST_ID"])
	commentID := models.ID(mux.Vars(r)["COMMENT_ID"])
	if utf8.RuneCountInString(string(postID)) != models.UUIDLength {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidPostID.Error()))
		return
	}
	if utf8.RuneCountInString(string(commentID)) != models.UUIDLength {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrInvalidCommentID.Error()))
		return
	}

	post, err := p.service.DeleteComment(r.Context(), postID, commentID)
	if errors.Is(err, models.ErrPostNotFound) {
		jsonSimpleErr(w, http.StatusNotFound, models.NewSimpleErr(models.ErrPostNotFound.Error()))
		return
	}
	if errors.Is(err, models.ErrCommentNotFound) {
		jsonSimpleErr(w, http.StatusNotFound, models.NewSimpleErr(models.ErrCommentNotFound.Error()))
		return
	}
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrUnknownError.Error()))
		return
	}

	resp, err := json.Marshal(post)
	if err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
		return
	}
	if _, err = w.Write(resp); err != nil {
		jsonSimpleErr(w, http.StatusInternalServerError, models.NewSimpleErr(models.ErrResponseError.Error()))
	}
}
