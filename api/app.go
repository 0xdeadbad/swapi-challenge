package api

import (
	"context"
	"log"
	"net/http"
	httpserver "swapi-challenge/api/server"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func Start(ctx context.Context, cancel context.CancelCauseFunc, redisClient *redis.Client, mongoClient *mongo.Client, options ...httpserver.HTTPServerOption) error {

	apiServer, err := httpserver.NewAPIServer(ctx, redisClient, mongoClient, options...)
	if err != nil {
		return err
	}

	httpServer := apiServer.HttpServer

	retCh := make(chan error)

	go func() {
		log.Println("Server has started")
		retCh <- httpServer.ListenAndServe()
		log.Println("Server has ended")
		close(retCh)
	}()

	select {
	case <-ctx.Done():

	case err := <-retCh:
		if err != http.ErrServerClosed {
			cancel(err)
		}
	}

	wctx, wcancel := context.WithTimeout(context.Background(), 30)
	defer wcancel()

	return httpServer.Shutdown(wctx)
}
