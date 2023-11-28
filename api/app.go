package api

import (
	"context"
	"swapi-challenge/server"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func Start(ctx context.Context, cancel context.CancelCauseFunc, redisClient *redis.Client, mongoClient *mongo.Client, options ...server.HTTPServerOption) error {

	apiServer, err := server.NewAPIServer(ctx, redisClient, mongoClient, options...)
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
