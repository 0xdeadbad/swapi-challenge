package main

import (
	"context"

	"log"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"swapi-challenge/api"
	httpserver "swapi-challenge/api/server"

	"github.com/pkg/profile"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

func main() {

	err := optsFlags.Parse()
	if err != nil {
		log.Fatalln(err)
	}

	bindAddr, err := netip.ParseAddrPort(optsFlags.Bind)
	if err != nil {
		log.Fatalln(err)
	}

	for _, prof := range optsFlags.Profile {
		switch strings.ToUpper(prof) {
		case "MEM":
			defer profile.Start(profile.MemProfile, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		case "CPU":
			defer profile.Start(profile.CPUProfile, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		case "TRAC":
			defer profile.Start(profile.TraceProfile, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		case "GORO":
			defer profile.Start(profile.GoroutineProfile, profile.NoShutdownHook, profile.ProfilePath(".")).Stop()

		default:
			log.Fatalf("invalid %s profiling parameter", prof)
		}
	}

	ctx, cancel := context.WithCancelCause(context.Background())

	sigChMain := make(chan os.Signal, 1)
	signal.Notify(sigChMain, syscall.SIGINT, syscall.SIGTERM)

	go func() { <-sigChMain; cancel(SWAPIQuitSignal) }()

	defer signal.Stop(sigChMain)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	defer close(sigChMain)

main_loop:
	for {
		g, gctx := errgroup.WithContext(ctx)
		goroutineCtx, goroutineCancel := context.WithCancelCause(ctx)

		g.Go(func() error {
			sigCh := make(chan os.Signal, 1)

			signal.Notify(sigCh, syscall.SIGHUP)
			defer signal.Stop(sigCh)
			defer signal.Reset(syscall.SIGHUP)
			defer close(sigCh)

			go func() {
				if sig, ok := <-sigCh; ok {
					if sig != syscall.SIGHUP {
						return
					}

					goroutineCancel(SWAPIRestartSignal)
				}
			}()

			redisClient := redis.NewClient(&redis.Options{
				Addr:    optsFlags.RedisURI,
				Network: optsFlags.RedisNetwork,
			})
			_, err := redisClient.Ping(ctx).Result()
			if err != nil {
				return err
			}
			defer redisClient.Close()

			mongoClient, err := mongo.Connect(ctx, mongoOptions.Client().ApplyURI(optsFlags.MongoURI))
			if err != nil {
				return err
			}
			defer mongoClient.Disconnect(context.Background())

			return api.Start(goroutineCtx, goroutineCancel,
				redisClient,
				mongoClient,
				httpserver.WithAddress(&bindAddr),
				httpserver.WithMaxHeaderBytes(4096),
				httpserver.WithHandler(api.Router(redisClient, mongoClient)),
			)
		})

		select {
		case <-gctx.Done():
			cause := context.Cause(goroutineCtx)

			cancel(cause)

			if err := g.Wait(); err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)
				log.SetOutput(os.Stdout)
			}

			break main_loop

		case <-goroutineCtx.Done():
			cause := context.Cause(goroutineCtx)

			if cause == SWAPIRestartSignal {
				log.Println(SWAPIRestartSignal)
				goroutineCancel(SWAPIRestartSignal)

				continue main_loop
			}

			log.Println(SWAPIQuitSignal)
			cancel(cause)

			if err := g.Wait(); err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)
				log.SetOutput(os.Stdout)
			}

			break main_loop

		case <-ctx.Done():

			if err := g.Wait(); err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)
				log.SetOutput(os.Stdout)
			}

			break main_loop
		}

	}

	if ctx.Err() != nil && ctx.Err() != context.Canceled {
		log.Fatalln(ctx.Err())
	}

	defer log.Println("Program has reached its end")
	defer log.Println(context.Cause(ctx))
}

type SWAPIError string

const (
	SWAPIQuitSignal    SWAPIError = "signal to end process received"
	SWAPIRestartSignal SWAPIError = "signal to restart process received"
	// SWAPIUnknownSignal SWAPIError = "unknown signal received by process"
)

func (e SWAPIError) String() string {
	return string(e)
}

func (e SWAPIError) Error() string {
	return e.String()
}

type ReqKey string
type ReqValue any
