package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/Benzogang-Tape/Reddit/internal/config"
	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/service"
	"github.com/Benzogang-Tape/Reddit/internal/storage"
	"github.com/Benzogang-Tape/Reddit/internal/transport/rest"
)

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
	v, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	if err = jwt.SetJWTSecret(v.GetString("jwt.secret")); err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?%s",
		v.GetString("mysql.user"),
		v.GetString("mysql.password"),
		v.GetString("mysql.host"),
		v.GetString("mysql.port"),
		v.GetString("mysql.database"),
		v.GetString("mysql.params"),
	)

	fmt.Println(dsn)
	usersDB, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = usersDB.Ping()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	sess, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(
		"%s%s",
		v.GetString("mongo.uri"),
		v.GetString("mongo.host"),
	)))

	if err != nil {
		panic(err)
	}

	postsDB := sess.Database(v.GetString("mongo.initdb.database")).Collection(v.GetString("mongo.collection.posts"))

	sessionDB := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", v.GetString("redis.host"), v.GetString("redis.port")),
		Password: v.GetString("redis.password"),
		DB:       0,
	})

	if _, err = sessionDB.Ping().Result(); err != nil {
		panic(err)
	}

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Logger init error")
	}
	defer zapLogger.Sync() //nolint:errcheck
	logger := zapLogger.Sugar()

	sessionStorage := storage.NewSessionRepoRedis(sessionDB)
	sessionHandler := service.NewSessionHandler(sessionStorage)

	userStorage := storage.NewUserRepoMySQL(usersDB)
	userHandler := service.NewUserHandler(userStorage)
	u := rest.NewUserHandler(userHandler, sessionHandler, logger)

	mongoAbstraction := storage.NewMongoCollection(postsDB)
	postStorage := storage.NewPostRepoMongoDB(mongoAbstraction)
	postHandler := service.NewPostHandler(postStorage, postStorage)
	p := rest.NewPostHandler(postHandler, logger)

	router := rest.NewAppRouter(u, p).InitRouter(logger)

	addr := fmt.Sprintf(":%s", v.GetString("app.port"))
	logger.Infow(fmt.Sprintf("Starting server on %s", addr))
	log.Panic(http.ListenAndServe(addr, router))
}
