package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type optionsFlags struct {
	// Show verbose logging
	// Verbose [3]bool `short:"v" long:"verbose" description:"Show verbose debug information"`

	// HTTP server's bind address to listen on
	Bind string `short:"b" long:"bind" description:"Bind address and port. <ip>:<port>" required:"true"`

	// Enable program profiling code file generation
	Profile []string `short:"P" long:"profile" description:"Generate profile files for perfomance analisys with go tool pprof"`

	// Profile path to write files
	ProfilePath string `short:"p" long:"profile-path" description:"Path to write profile files"`

	// MongoDB connection URI
	MongoURI string `short:"M" long:"mongo" description:"URI pointing to MongoDB" required:"true"`

	// Redis connection URI
	RedisURI string `short:"R" long:"redis" description:"URI pointing to Redis" required:"true"`

	// Redis connection type
	RedisNetwork string `long:"redis-network" description:"URI pointing to Redis" default:"tcp"`
}

var optsFlags = optionsFlags{}

func (o *optionsFlags) Parse() error {
	_, err := flags.ParseArgs(&optsFlags, os.Args)
	if err != nil {
		if optsFlags.Bind == "" {
			if env, ok := os.LookupEnv("BIND_ADDRESS"); ok {
				optsFlags.Bind = env
			} else {
				return err
			}
		}

		if optsFlags.MongoURI == "" {
			if env, ok := os.LookupEnv("MONGO_URI"); ok {
				optsFlags.MongoURI = env
			} else {
				return err
			}
		}

		if optsFlags.RedisURI == "" {
			if env, ok := os.LookupEnv("REDIS_URI"); ok {
				optsFlags.RedisURI = env
			} else {
				return err
			}
		}
	}

	if env, ok := os.LookupEnv("REDIS_NETWORK"); ok {
		optsFlags.RedisNetwork = env
	}

	return nil
}
