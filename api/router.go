package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIRouter struct {
	*mux.Router
}

func NewAPIRouter(redisClient *redis.Client, mongoClient *mongo.Client) http.Handler {
	ar := &APIRouter{
		Router: mux.NewRouter(),
	}

	mongoDatabase := mongoClient.Database("swapi")

	peopleApiRouter := ar.PathPrefix("/api/people").Subrouter()
	NewPeopleAPIEndpoints(peopleApiRouter, redisClient, mongoDatabase)

	return ar
}
