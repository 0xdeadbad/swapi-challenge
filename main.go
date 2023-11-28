package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"net/http"
	"net/netip"
	"os/signal"
	"strings"
	"syscall"

	"swapi-challenge/api"
	"swapi-challenge/server"

	"github.com/pkg/profile"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

var logger = slog.Default()

func fatalLogger(msg string) {
	logger.Error(msg)
	os.Exit(1)
}

func main() {

	err := optsFlags.Parse()
	if err != nil {
		fatalLogger(err.Error())
	}

	bindAddr, err := netip.ParseAddrPort(optsFlags.Bind)
	if err != nil {
		fatalLogger(err.Error())
	}

	for _, prof := range optsFlags.Profile {
		switch strings.ToUpper(prof) {
		case "MEM":
			defer profile.Start(profile.MemProfile, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		case "HEAP":
			defer profile.Start(profile.MemProfileHeap, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		case "CPU":
			defer profile.Start(profile.CPUProfile, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		case "TRACE":
			defer profile.Start(profile.TraceProfile, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		case "GOROUTINES":
			defer profile.Start(profile.GoroutineProfile, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		default:
			fatalLogger(fmt.Sprintf("invalid %s profiling parameter", prof))
		}
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisClient := redis.NewClient(&redis.Options{
		Addr:    optsFlags.RedisURI,
		Network: optsFlags.RedisNetwork,
	})
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		fatalLogger(err.Error())
	}
	defer redisClient.Close()

	mongoClient, err := mongo.Connect(ctx, mongoOptions.Client().ApplyURI(optsFlags.MongoURI))
	if err != nil {
		fatalLogger(err.Error())
	}
	defer mongoClient.Disconnect(context.Background())

	err = api.Start(ctx, cancel,
		redisClient,
		mongoClient,
		server.WithAddress(&bindAddr),
		server.WithMaxHeaderBytes(4096),
		server.WithHandler(api.Router(redisClient, mongoClient)),
	)

	if err != nil && err != context.Canceled && err != http.ErrServerClosed {
		fatalLogger(err.Error())
	}

	if ctx.Err() != nil && ctx.Err() != http.ErrServerClosed && ctx.Err() != context.Canceled {
		fatalLogger(ctx.Err().Error())
	}

	logger.Info("Program has reached its end")
}
