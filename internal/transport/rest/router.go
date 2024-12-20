package rest

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "github.com/Benzogang-Tape/Reddit/docs"
	"github.com/Benzogang-Tape/Reddit/internal/transport/middleware"
	mdwr "github.com/Benzogang-Tape/Reddit/pkg/middleware"
)

type AppRouter struct {
	userHandler *UserHandler
	postHandler *PostHandler
}

func NewAppRouter(u *UserHandler, p *PostHandler) *AppRouter {
	return &AppRouter{
		userHandler: u,
		postHandler: p,
	}
}

func (rtr *AppRouter) InitRouter(logger *zap.SugaredLogger) http.Handler {
	templates := template.Must(template.ParseGlob("./static/*/*"))

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			http.Error(w, `Template error`, http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodGet)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.HandleFunc("/api/register", rtr.userHandler.RegisterUser).Methods(http.MethodPost)
	r.HandleFunc("/api/login", rtr.userHandler.LoginUser).Methods(http.MethodPost)
	r.HandleFunc("/api/posts/", rtr.postHandler.GetAllPosts).Methods(http.MethodGet)
	r.HandleFunc("/api/posts", rtr.postHandler.CreatePost).Methods(http.MethodPost)
	r.HandleFunc("/api/post/{POST_ID:[0-9a-fA-F-]+$}", rtr.postHandler.GetPostByID).Methods(http.MethodGet)
	r.HandleFunc("/api/posts/{CATEGORY_NAME:[0-9a-zA-Z_-]+$}", rtr.postHandler.GetPostsByCategory).Methods(http.MethodGet)
	r.HandleFunc("/api/user/{USER_LOGIN:[0-9a-zA-Z_-]+$}", rtr.postHandler.GetPostsByUser).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{POST_ID:[0-9a-fA-F-]+$}", rtr.postHandler.DeletePost).Methods(http.MethodDelete)
	r.HandleFunc("/api/post/{POST_ID:[0-9a-fA-F-]+}/upvote", rtr.postHandler.Upvote).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{POST_ID:[0-9a-fA-F-]+}/downvote", rtr.postHandler.Downvote).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{POST_ID:[0-9a-fA-F-]+}/unvote", rtr.postHandler.Unvote).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{POST_ID:[0-9a-fA-F-]+$}", rtr.postHandler.AddComment).Methods(http.MethodPost)
	r.HandleFunc("/api/post/{POST_ID:[0-9a-fA-F-]+}/{COMMENT_ID:[0-9a-fA-F-]+$}", rtr.postHandler.DeleteComment).Methods(http.MethodDelete)

	router := middleware.Auth(r, rtr.userHandler.sessMngr, logger)
	router = mdwr.AccessLog(logger, router)
	router = middleware.Panic(router, logger)

	return router
}
