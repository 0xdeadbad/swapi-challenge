package api

import (
	"context"
	httpserver "swapi-challenge/api/server"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func Start(ctx context.Context, cancel context.CancelCauseFunc, redisClient *redis.Client, mongoClient *mongo.Client, options ...httpserver.HTTPServerOption) error {

	apiServer, err := httpserver.NewAPIServer(ctx, redisClient, mongoClient, options...)
	if err != nil {
		return err
	}

	go func() {
		if err := apiServer.HttpServer.ListenAndServe(); err != nil {
			cancel(err)
		}
	}()

	<-ctx.Done()

	return apiServer.HttpServer.Shutdown(context.Background())
}
