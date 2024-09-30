package main

import (
	"github.com/Benzogang-Tape/Reddit/internal/service"
	"github.com/Benzogang-Tape/Reddit/internal/storage"
	"github.com/Benzogang-Tape/Reddit/internal/transport/rest"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Logger init error")
	}
	defer zapLogger.Sync() //nolint:errcheck
	logger := zapLogger.Sugar()

	userStorage := storage.NewUserRepo()
	userHandler := service.NewUserHandler(userStorage)
	u := rest.NewUserHandler(userHandler, logger)

	postStorage := storage.NewPostRepo()
	postHandler := service.NewPostHandler(postStorage, postStorage)
	p := rest.NewPostHandler(postHandler, logger)

	router := rest.NewAppRouter(u, p).InitRouter(logger)

	addr := ":8080"
	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Panicf("RUNTIME ERROR")
	}
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
