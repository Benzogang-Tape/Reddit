package main

import (
	"database/sql"
	"github.com/Benzogang-Tape/Reddit/internal/service"
	"github.com/Benzogang-Tape/Reddit/internal/storage"
	"github.com/Benzogang-Tape/Reddit/internal/transport/rest"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	dsn := "root:pass@tcp(localhost:3306)/reddit?"
	dsn += "charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)

	//db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// ---

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Logger init error")
	}
	defer zapLogger.Sync() //nolint:errcheck
	logger := zapLogger.Sugar()

	userStorage := storage.NewUserRepo(db)
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
