package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/queue"
	"github.com/johnwyles/vrddt-droplets/interfaces/rest"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/graceful"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/pkg/middlewares"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	"github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

var (
	// BuildTimestamp is the build date
	BuildTimestamp string

	// GitHash is the git build hash
	GitHash string

	// Version is the version of the software
	Version string

	// loggerHandle is the current logger facility
	loggerHandle logger.Logger
)

func main() {
	// Setup some sensible defaults for the vrddt configuration - this is
	// somewhat ugly but necessary if we want to override configuration file
	// options with the arguments on the command line
	cfg := &config.Config{
		API: config.APIConfig{
			Address:         ":9090",
			CertFile:        "config/ssl/server.crt",
			GracefulTimeout: 30,
			KeyFile:         "config/ssl/server.key",
		},
		Log: config.LogConfig{
			Format: "text",
			Level:  "debug",
		},
		Queue: config.QueueConfig{
			Memory: config.QueueMemoryConfig{
				MaxSize: 100000,
			},
			RabbitMQ: config.QueueRabbitMQConfig{
				BindingKeyName: "vrddt-bindingkey-converter",
				ExchangeName:   "vrddt-exchange-converter",
				QueueName:      "vrddt-queue-converter",
				URI:            "amqp://admin:password@localhost:5672",
			},
			Type: config.QueueConfigRabbitMQ,
		},
		Store: config.StoreConfig{
			Memory: config.StoreMemoryConfig{
				MaxSize: 100000,
			},
			Mongo: config.StoreMongoConfig{
				RedditVideosCollectionName: "reddit_videos",
				Timeout:                    60,
				URI:                        "mongodb://admin:password@localhost:27017/vrddt",
				VrddtVideosCollectionName:  "vrddt_videos",
			},
			Type: config.StoreConfigMongo,
		},
	}

	// Loading of all the configuration from environment variables, toml
	// configuration file, or command-line flags
	flags := []cli.Flag{
		&cli.StringFlag{
			Aliases: []string{"c"},
			EnvVars: []string{"VRDDT_CONFIG"},
			Name:    "config",
			Usage:   "vrddt-admin TOML configuration file",
			Value:   "",
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Aliases:     []string{"a"},
				Destination: &cfg.API.Address,
				EnvVars:     []string{"VRDDT_API_ADDRESS"},
				Name:        "API.Address",
				Usage:       "API listening address",
				Value:       cfg.API.Address,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Aliases:     []string{"f"},
				Destination: &cfg.API.CertFile,
				EnvVars:     []string{"VRDDT_API_CERT_FILE"},
				Name:        "API.CertFile",
				Usage:       "API SSL certificate file",
				Value:       cfg.API.CertFile,
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Aliases:     []string{"t"},
				Destination: &cfg.API.GracefulTimeout,
				EnvVars:     []string{"VRDDT_API_GRACEFUL_TIMEOUT"},
				Name:        "API.GracefulTimeout",
				Usage:       "API graceful timeout (in seconds)",
				Value:       cfg.API.GracefulTimeout,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Aliases:     []string{"k"},
				Destination: &cfg.API.KeyFile,
				EnvVars:     []string{"VRDDT_API_KEY_FILE"},
				Name:        "API.KeyFile",
				Usage:       "API SSL key file",
				Value:       cfg.API.KeyFile,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Aliases:     []string{"lf"},
				Destination: &cfg.Log.Format,
				EnvVars:     []string{"VRDDT_LOG_FORMAT"},
				Name:        "Log.Format",
				Usage:       "Logging format",
				Value:       cfg.Log.Format,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Aliases:     []string{"ll"},
				Destination: &cfg.Log.Level,
				EnvVars:     []string{"VRDDT_LOG_LEVEL"},
				Name:        "Log.Level",
				Usage:       "Set logging level",
				Value:       cfg.Log.Level,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Queue.RabbitMQ.BindingKeyName,
				EnvVars:     []string{"VRDDT_RABBITMQ_BINDING_KEY_NAME"},
				Name:        "Queue.RabbitMQ.BindingKeyName",
				Usage:       "RabbitMQ binding key name",
				Value:       cfg.Queue.RabbitMQ.BindingKeyName,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Queue.RabbitMQ.ExchangeName,
				EnvVars:     []string{"VRDDT_RABBITMQ_EXCHANGE_NAME"},
				Name:        "Queue.RabbitMQ.ExchangeName",
				Usage:       "RabbitMQ exchange name",
				Value:       cfg.Queue.RabbitMQ.ExchangeName,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Queue.RabbitMQ.URI,
				EnvVars:     []string{"VRDDT_RABBITMQ_URI"},
				Name:        "Queue.RabbitMQ.URI",
				Usage:       "RabbitMQ connection string",
				Value:       cfg.Queue.RabbitMQ.URI,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Store.Mongo.RedditVideosCollectionName,
				EnvVars:     []string{"VRDDT_STORE_MONGO_REDDIT_VIDEOS_COLLECTION_NAME"},
				Name:        "Store.Mongo.RedditVideosCollectionName",
				Usage:       "Collection name where we will store information about the Reddit videos",
				Value:       cfg.Store.Mongo.RedditVideosCollectionName,
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Destination: &cfg.Store.Mongo.Timeout,
				EnvVars:     []string{"VRDDT_STORE_MONGO_TIMEOUT"},
				Name:        "Store.Mongo.Timeout",
				Usage:       "Connection timeout",
				Value:       cfg.Store.Mongo.Timeout,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Store.Mongo.URI,
				EnvVars:     []string{"VRDDT_STORE_MONGO_URI"},
				Name:        "Store.Mongo.URI",
				Usage:       "MongoDB connection string",
				Value:       cfg.Store.Mongo.URI,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Store.Mongo.VrddtVideosCollectionName,
				EnvVars:     []string{"VRDDT_STORE_MONGO_VRDDT_VIDEOS_COLLECTION_NAME"},
				Name:        "Store.Mongo.ReddtiVideosCollectionName",
				Usage:       "Collection name where we will store information about the vrddt videos",
				Value:       cfg.Store.Mongo.VrddtVideosCollectionName,
			},
		),
	}

	timeStamp, err := strconv.ParseInt(BuildTimestamp, 10, 64)
	if err != nil {
		now := time.Now()
		timeStamp = now.Unix()
	}

	app := &cli.App{
		Action: rootAction(cfg),
		After:  afterResources(cfg),
		Authors: []*cli.Author{
			{
				Name:  "John Wyles",
				Email: "john@johnwyles.com",
			},
		},
		Before: altsrc.InitInputSourceWithContext(
			flags,
			func(context *cli.Context) (altsrc.InputSourceContext, error) {
				if context.IsSet("config") {
					return altsrc.NewTomlSourceFromFile(context.String("config"))
				}

				return &altsrc.MapInputSource{}, nil
			},
		),
		Commands: allCommands(cfg),
		Compiled: time.Now(),
		Flags:    flags,
		Name:     "vrddt-api",
		Prepare:  prepareResources(cfg),
		Usage:    "vrddt API service",
		Version:  fmt.Sprintf("%s [Build Date: %s, Git Hash: %s]", Version, time.Unix(timeStamp, 0), GitHash),
	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "Print the help",
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Print the current version",
	}

	if err := app.Run(os.Args); err != nil {
		loggerHandle.Fatalf("An error occured running the application: %s", err)
		os.Exit(1)

		return
	}

	return
}

// allCommands are all of the commands we are able to run
func allCommands(cfg *config.Config) []*cli.Command {
	return []*cli.Command{}
}

// afterResources will execute after Action() to cleanup
func afterResources(cfg *config.Config) cli.AfterFunc {
	return func(cliContext *cli.Context) (err error) {
		return
	}
}

// prepareResources will setup some common shared resources amoung all of the
// commands and make them avaiable to use
func prepareResources(cfg *config.Config) cli.PrepareFunc {
	return func(cliContext *cli.Context) (err error) {
		// Initalize logger
		loggerHandle = logger.New(os.Stderr, cfg.Log.Level, cfg.Log.Format)

		return
	}
}

// rootAction is the what we execute if no commands are specified
func rootAction(cfg *config.Config) cli.ActionFunc {
	return func(cliContext *cli.Context) (err error) {
		// Initalize connections
		loggerHandle = logger.New(os.Stderr, cfg.Log.Level, cfg.Log.Format)

		// Setup the queue
		q, err := queue.RabbitMQ(&cfg.Queue.RabbitMQ, loggerHandle)
		if err != nil {
			return
		}

		// Initialize the queue
		if err = q.Init(context.TODO()); err != nil {
			return
		}

		// Setup the store
		str, err := store.Mongo(&cfg.Store.Mongo, loggerHandle)
		if err != nil {
			return
		}

		// Initialize the store
		if err = str.Init(context.TODO()); err != nil {
			return
		}

		// Get the REST controller
		router := mux.NewRouter()

		// Setup router with default handlers
		router.NotFoundHandler = http.HandlerFunc(rest.NotFoundHandler)
		router.MethodNotAllowedHandler = http.HandlerFunc(rest.MethodNotAllowedHandler)

		// Setup API endpoints for Reddit videos
		rvc := redditvideos.NewConstructor(loggerHandle, q, str)
		rvd := redditvideos.NewDestructor(loggerHandle, q, str)
		rvr := redditvideos.NewRetriever(loggerHandle, str)
		rest.AddRedditVideosAPI(loggerHandle, router, rvc, rvd, rvr)

		// Setup API endpoints for vrddt videos
		vvc := vrddtvideos.NewConstructor(loggerHandle, str)
		vvd := vrddtvideos.NewDestructor(loggerHandle, str)
		vvr := vrddtvideos.NewRetriever(loggerHandle, str)
		rest.AddVrddtVideosAPI(loggerHandle, router, vvc, vvd, vvr, rvc, rvr)

		// Setup API middleware
		handler := middlewares.WithRequestLogging(loggerHandle, router)
		handler = middlewares.WithRecovery(loggerHandle, handler)
		co := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET"},
		})
		handler = co.Handler(handler)

		// Setup HTTP server
		srv := graceful.NewServer(handler, time.Duration(cfg.API.GracefulTimeout)*time.Second, os.Interrupt)
		srv.Log = loggerHandle.Errorf
		srv.Addr = cfg.API.Address

		loggerHandle.Infof("API server is listening as: %s", cfg.API.Address)
		if err := srv.ListenAndServeTLS(cfg.API.CertFile, cfg.API.KeyFile); err != nil {
			loggerHandle.Fatalf("API server exited: %s", err)
		}

		return
	}
}
