package main

import (
	"fmt"
	"os"

	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/converter"
	"github.com/johnwyles/vrddt-droplets/interfaces/queue"
	"github.com/johnwyles/vrddt-droplets/interfaces/storage"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// Services holds all of various services to the subcommands for use
type Services struct {
	Converter converter.Converter
	Queue     queue.Queue
	Storage   storage.Storage
	Store     store.Store
}

// loggerHandle is the current logger facility
var loggerHandle logger.Logger

// services will be a refer to our global services avaiable to the subcommands
var services = &Services{}

// allCommands are all of the commands we are able to run
func allCommands(cfg *config.Config) []*cli.Command {
	return []*cli.Command{
		Converter(cfg),
	}
}

func main() {
	// Setup some sensible defaults for the vrddt worker converter
	// configuration - this is somewhat ugly but necessary if we want to
	// override configuration file options with the arguments on the command
	// line
	cfg := &config.Config{
		Converter: config.ConverterConfig{
			FFmpeg: config.ConverterFFmpegConfig{
				Path: "/usr/local/bin/ffmpeg",
			},
		},
		Log: config.LogConfig{
			Format: "text",
			Level:  "warn",
		},
		Queue: config.QueueConfig{
			RabbitMQ: config.QueueRabbitMQConfig{
				BindingKeyName: "vrddt-bindingkey-converter",
				ExchangeName:   "vrddt-exchange-converter",
				QueueName:      "vrddt-queue-converter",
				URI:            "amqp://admin:password@localhost:5672",
			},
			Type: config.QueueConfigRabbitMQ,
		},
		Storage: config.StorageConfig{
			GCS: config.StorageGCSConfig{
				CredentialsJSON: "",
				Bucket:          "vrddt-media",
			},
			Type: config.StorageConfigGCS,
		},
		Store: config.StoreConfig{
			Mongo: config.StoreMongoConfig{
				RedditVideosCollectionName: "reddit_videos",
				URI:                        "mongodb://admin:password@localhost:27017/vrddt",
				VrddtVideosCollectionName:  "vrddt_videos",
			},
			Type: config.StoreConfigMongo,
		},
		Worker: config.WorkerConfig{
			Converter: config.WorkerConverterConfig{
				MaxErrors: 10,
				Sleep:     1000,
			},
		},
	}

	// Loading of all the configuration from environment variables, toml
	// configuration file, or command-line flags
	flags := []cli.Flag{
		&cli.StringFlag{
			Aliases: []string{"c"},
			EnvVars: []string{"VRDDT_CONFIG"},
			Name:    "config",
			Usage:   "vrddt TOML configuration file (see: config/config.workerconverter.example.toml)",
			Value:   "",
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Aliases:     []string{"lf"},
				Destination: &cfg.Log.Format,
				EnvVars:     []string{"VRDDT_LOG_FORMAT"},
				Name:        "Log.Format",
				Usage:       "Format for logging (e.g. text or json)",
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
				Destination: &cfg.Converter.FFmpeg.Path,
				EnvVars:     []string{"VRDDT_CONVERTER_FFMPEG_PATH"},
				Name:        "Converter.FFmpeg.ExecutableName",
				Usage:       "Enable colored logging",
				Value:       cfg.Converter.FFmpeg.ExecutableName,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Queue.RabbitMQ.BindingKeyName,
				EnvVars:     []string{"VRDDT_QUEUE_RABBITMQ_BINDING_KEY_NAME"},
				Name:        "Queue.RabbitMQ.BindingKeyName",
				Usage:       "RabbitMQ binding key name",
				Value:       cfg.Queue.RabbitMQ.BindingKeyName,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Queue.RabbitMQ.ExchangeName,
				EnvVars:     []string{"VRDDT_QUEUE_RABBITMQ_EXCHANGE_NAME"},
				Name:        "Queue.RabbitMQ.ExchangeName",
				Usage:       "RabbitMQ exchange name",
				Value:       cfg.Queue.RabbitMQ.ExchangeName,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Queue.RabbitMQ.URI,
				EnvVars:     []string{"VRDDT_QUEUE_RABBITMQ_URI"},
				Name:        "Queue.RabbitMQ.URI",
				Usage:       "RabbitMQ connection string",
				Value:       cfg.Queue.RabbitMQ.URI,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Storage.GCS.CredentialsJSON,
				EnvVars:     []string{"GOOGLE_APPLICATION_CREDENTIALS"},
				Name:        "Storage.GCS.CredentialsJSON",
				Usage:       "Set the path to the GCP JSON credentials file for the storage user",
				Value:       cfg.Storage.GCS.CredentialsJSON,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Storage.GCS.Bucket,
				EnvVars:     []string{"VRDDT_STROAGE_GCS_BUCKET"},
				Name:        "Storage.GCS.Bucket",
				Usage:       "GCS bucket for vrddt media",
				Value:       cfg.Storage.GCS.Bucket,
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
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Destination: &cfg.Worker.Converter.MaxErrors,
				EnvVars:     []string{"VRDDT_WORKER_CONVERTER_MAX_ERRORS"},
				Name:        "WorkerConverter.MaxErrors",
				Usage:       "Maximum number of errors tolerated before the worker dies completely",
				Value:       cfg.Worker.Converter.MaxErrors,
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Destination: &cfg.Worker.Converter.Sleep,
				EnvVars:     []string{"VRDDT_WORKER_CONVERTER_SLEEP"},
				Name:        "WorkerConverter.Sleep",
				Usage:       "Number of milliseconds to sleep after processing each request",
				Value:       cfg.Worker.Converter.Sleep,
			},
		),
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
			func(cliContext *cli.Context) (altsrc.InputSourceContext, error) {
				if cliContext.IsSet("config") {
					return altsrc.NewTomlSourceFromFile(cliContext.String("config"))
				}

				return &altsrc.MapInputSource{}, nil
			},
		),
		Flags:   flags,
		Name:    "vrddt-worker",
		Prepare: prepareResources(cfg),
		Usage:   "vrddt Long-lived worker processes",
		Version: "v0.0.1",
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
	}
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
		// Initalize connections
		loggerHandle = logger.New(os.Stderr, cfg.Log.Level, cfg.Log.Format)

		services.Queue, err = queue.RabbitMQ(&cfg.Queue.RabbitMQ, loggerHandle)
		if err != nil {
			return
		}

		// Setup the store
		services.Store, err = store.Mongo(&cfg.Store.Mongo, loggerHandle)
		if err != nil {
			return
		}

		services.Converter, err = converter.FFmpeg(&cfg.Converter.FFmpeg, loggerHandle)
		if err != nil {
			return
		}

		return nil
	}
}

func rootAction(cfg *config.Config) cli.ActionFunc {
	return func(cliContext *cli.Context) (err error) {
		cli.ShowAppHelp(cliContext)
		return fmt.Errorf("No sub-command specified")
	}
}
