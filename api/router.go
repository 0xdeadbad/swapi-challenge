package api

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIRouter struct {
	http.Handler
}

func NewAPIRouter(redisClient *redis.Client, mongoClient *mongo.Client) http.Handler {
	ar := &APIRouter{}

	return ar
}
