package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var logger = slog.Default()

func apiRoutesHandler(r *mux.Router, redisClient *redis.Client, mongoClient *mongo.Database) {

	peopleApiRouter := r.PathPrefix("/people").Subrouter()
	newPeopleApiHandler(peopleApiRouter, mongoClient.Collection("people"), redisClient)
}

func Router(redisClient *redis.Client, mongoClient *mongo.Client) http.Handler {
	mongoDatabase := mongoClient.Database("swapi")

	r := mux.NewRouter()

	r.Use(loggingMiddleware)

	apiRouter := r.PathPrefix("/api").Subrouter()

	apiRoutesHandler(apiRouter, redisClient, mongoDatabase)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "This is: %s", "root [/]")
	})

	return r
}

func loggingMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log(r.Context(), slog.LevelInfo, fmt.Sprintf("[%s : %s] -- ", r.RequestURI, r.Method))
		next.ServeHTTP(w, r)
	})
}
