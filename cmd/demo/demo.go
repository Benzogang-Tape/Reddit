package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/service"
	"github.com/Benzogang-Tape/Reddit/internal/storage/inmem"
	"github.com/Benzogang-Tape/Reddit/internal/transport/rest"
)

var port = flag.Int("port", 8081, "HTTP port")

func init() {
	os.Setenv("JWT_SECRET", "super secret key")
}

//	@title			Reddit-Clone API
//	@version		1.0
//	@description	Basic restfull api for reddit-clone backend.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8081
//	@BasePath	/api

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	flag.Parse()

	if err := jwt.SetJWTSecret(os.Getenv("JWT_SECRET")); err != nil {
		panic(err)
	}

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Logger init error")
	}
	defer zapLogger.Sync() //nolint:errcheck
	logger := zapLogger.Sugar()

	sessionRepo := inmem.NewSessionRepo()
	sessionHandler := service.NewSessionHandler(sessionRepo)

	userStorage := inmem.NewUserRepo()
	userHandler := service.NewUserHandler(userStorage)
	u := rest.NewUserHandler(userHandler, sessionHandler, logger)

	postStorage := inmem.NewPostRepo()
	postHandler := service.NewPostHandler(postStorage, postStorage)
	p := rest.NewPostHandler(postHandler, logger)

	router := rest.NewAppRouter(u, p).InitRouter(logger)

	addr := fmt.Sprintf(":%d", *port)
	logger.Infow(fmt.Sprintf("Starting server on %s", addr))
	log.Panic(http.ListenAndServe(addr, router))
}
