package main

import (
	"context"
	"os"
	"strings"
	"time"

	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/mongo"
	"github.com/johnwyles/vrddt-droplets/interfaces/rabbitmq"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	"github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

func main() {
	// Setup some sensible defaults for the vrddt worker converter
	// configuration - this is somewhat ugly but necessary if we want to
	// override configuration file options with the arguments on the command
	// line
	cfg := &config.Config{
		Converter: config.ConverterConfig{
			FFmpeg: config.ConverterFFmpegConfig{
				ExecutableName: "ffmpeg",
			},
		},
		Log: config.LogConfig{
			Colored: true,
			Level:   "warn",
			Pretty:  true,
		},
		Queue: config.QueueConfig{
			RabbitMQ: config.QueueRabbitMQConfig{
				BindingKeyName: "vrddt-bindingkey-converter",
				ExchangeName:   "vrddt-exchange-converter",
				QueueName:      "vrddt-queue-converter",
				URL:            "amqp://admin:password@localhost:5672",
			},
			Memory: config.QueueMemoryConfig{
				MaxSize: 100000,
			},
			Type: config.QueueConfigRabbitMQ,
		},
		Storage: config.StorageConfig{
			GCS: config.StorageGCSConfig{
				CredentialsJSON: "",
				Bucket:          "vrddt-media",
			},
			Local: config.StorageLocalConfig{
				Path: "./",
			},
			Type: config.StorageConfigGCS,
		},
		Store: config.StoreConfig{
			Mongo: config.StoreMongoConfig{
				RedditVideosCollectionName: "reddit_videos",
				URL:                        "mongodb://admin:password@localhost:27017/vrddt",
				VrddtVideosCollectionName:  "vrddt_videos",
			},
			Memory: config.StoreMemoryConfig{
				MaxSize: 100000,
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
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Aliases:     []string{"lc"},
				Destination: &cfg.Log.Colored,
				EnvVars:     []string{"VRDDT_LOG_COLORED"},
				Name:        "Log.Colored",
				Usage:       "Enable colored logging",
				Value:       cfg.Log.Colored,
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
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Aliases:     []string{"lp"},
				Destination: &cfg.Log.Pretty,
				EnvVars:     []string{"VRDDT_LOG_PRETTY"},
				Name:        "Log.Pretty",
				Usage:       "Enable pretty logging",
				Value:       cfg.Log.Pretty,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Destination: &cfg.Converter.FFmpeg.ExecutableName,
				EnvVars:     []string{"VRDDT_CONVERTER_FFMPEG_EXECUTABLE_NAME"},
				Name:        "Converter.FFmpeg.ExecutableName",
				Usage:       "Enable colored logging",
				Value:       cfg.Converter.FFmpeg.ExecutableName,
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Destination: &cfg.Queue.Memory.MaxSize,
				EnvVars:     []string{"VRDDT_QUEUE_MEMORY_MAX_SIZE"},
				Name:        "Queue.Memory.MaxSize",
				Usage:       "Maximum number of items for queue",
				Value:       cfg.Queue.Memory.MaxSize,
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
				Destination: &cfg.Queue.RabbitMQ.URL,
				EnvVars:     []string{"VRDDT_QUEUE_RABBITMQ_URL"},
				Name:        "Queue.RabbitMQ.URL",
				Usage:       "RabbitMQ connection string",
				Value:       cfg.Queue.RabbitMQ.URL,
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
				Destination: &cfg.Storage.Local.Path,
				EnvVars:     []string{"VRDDT_STROAGE_LOCAL_PATH"},
				Name:        "Storage.Local.Path",
				Usage:       "Path for vrddt media",
				Value:       cfg.Storage.Local.Path,
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Destination: &cfg.Store.Memory.MaxSize,
				EnvVars:     []string{"VRDDT_STORE_MEMORY_MAX_SIZE"},
				Name:        "Store.Memory.MaxSize",
				Usage:       "Maximum number of items for store",
				Value:       cfg.Store.Memory.MaxSize,
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
				Destination: &cfg.Store.Mongo.URL,
				EnvVars:     []string{"VRDDT_STORE_MONGO_URL"},
				Name:        "Store.Mongo.URL",
				Usage:       "MongoDB connection string",
				Value:       cfg.Store.Mongo.URL,
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
		Name:    "vrddt-worker-converter",
		Prepare: prepareResources(cfg),
		Usage:   "vrddt Worker Coverter long-lived process for converting media",
		Version: version.Version.String(),
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
		log.Fatal().
			Err(err).
			Msg("An error occured running the application")
		os.Exit(1) // This may be repetitive with above
	}
}

// afterResources will execute after Action() to cleanup
func afterResources(cfg *config.Config) cli.AfterFunc {
	return func(cliContext *cli.Context) (err error) {
		getLogger().Debug().Msg("afterResources()")

		// We don't care about any cleanup errors
		GlobalServices.Queue.Cleanup()
		GlobalServices.Store.Cleanup()
		GlobalServices.Storage.Cleanup()

		return
	}
}

// IsSet cannot be used below because the env variable may be used
// if cliContext.NArg() < 1 && (cliContext.String("config") == "") {
// 	cli.ShowAppHelp(cliContext)
// 	os.Exit(1)
// }
// prepareResources will setup some common shared resources amoung all of the
// commands and make them avaiable to use
func prepareResources(cfg *config.Config) cli.PrepareFunc {
	return func(cliContext *cli.Context) (err error) {
		switch strings.ToLower(cfg.Log.Level) {
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		case "fatal":
			zerolog.SetGlobalLevel(zerolog.FatalLevel)
		case "panic":
			zerolog.SetGlobalLevel(zerolog.PanicLevel)
		default:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		if cfg.Log.Pretty {
			log.Logger = log.Output(
				zerolog.ConsoleWriter{
					Out:        os.Stderr,
					NoColor:    !cfg.Log.Colored,
					TimeFormat: time.RFC3339,
				},
			)
		}

		getLogger().Debug().Msg("prepareResources()")
		getLogger().Debug().Msgf("Config: %#v", cfg)

		// Setup the file converter
		GlobalServices.Converter, err = converter.FFmpeg(&cfg.Converter.FFmpeg)
		if err != nil {
			return
		}

		// Setup the queue as a consumer
		GlobalServices.Queue, err = queue.RabbitMQ(&cfg.Queue.RabbitMQ)
		if err != nil {
			return
		}

		// Setup the store
		GlobalServices.Store, err = store.Mongo(&cfg.Store.Mongo)
		if err != nil {
			return
		}

		// Setup the storage
		GlobalServices.Storage, err = storage.GCS(&cfg.Storage.GCS)
		if err != nil {
			return
		}

		return nil
	}
}

func rootAction(cfg *config.Config) cli.ActionFunc {
	return func(cliContext *cli.Context) (err error) {
		getLogger().Debug().Msg("rootAction()")

		errs := make(chan error, 0)
		successes := make(chan bool, 0)
		errorCount := 0
		successCount := 0
		messageCount := 0

		// Initialize the queue
		if err = GlobalServices.Queue.Init(); err != nil {
			return
		}
		if err = GlobalServices.Queue.MakeConsumer(); err != nil {
			return
		}

		// Initialze the store
		if err = GlobalServices.Store.Init(); err != nil {
			return
		}

		// Initialize the storage
		if err = GlobalServices.Storage.Init(); err != nil {
			return
		}

		// TODO: How do we use context? Study

		conv, err := worker.Converter(
			GlobalServices.Converter,
			GlobalServices.Queue,
			GlobalServices.Store,
			GlobalServices.Storage,
		)

		for {
			messageCount++

			go func() {
				// ctx := context.Background()
				ctx := context.TODO()

				if goErr := conv.GetWork(&ctx); goErr != nil {
					getLogger().Warn().Err(err).Msg("error getting element of work")
					errs <- goErr
					successes <- false
					return
				}
				getLogger().Info().Msgf("Received new request (%d) for work", messageCount)

				if goErr := conv.DoWork(&ctx); goErr != nil {
					getLogger().Warn().Err(err).Msg("unable to perform work")
					errs <- goErr
					successes <- false
					return
				}

				if goErr := conv.CompleteWork(&ctx); goErr != nil {
					getLogger().Warn().Err(err).Msg("unable to complete work")
					errs <- goErr
					successes <- false
					return
				}

				successes <- true
				// ctx.Done()
				return
			}()

			select {
			case err := <-errs:
				switch err {
				case reddit.JSONTitleErr, reddit.JSONVideoURLErr, reddit.NotDASHErr:
					getLogger().Warn().Msgf("Warning while processing media: %s", err)
				default:
					errorCount++
					getLogger().Error().Msgf("Error (#%d of %d allowed) while processing media: %s",
						errorCount,
						cliContext.Int("max-error"),
						err,
					)
				}
			case success := <-successes:
				successCount++
				getLogger().Info().Msgf("Success (#%d): %#v", successCount, success)
			}

			// Let's take a break
			time.Sleep(time.Duration(cliContext.Int64("WorkerConverter.Sleep")) * time.Millisecond)

			// We have exceeded the max-error count so break out of the loop
			// This will exit the infinite for-loop because we exceeded "max-error"
			if errorCount >= cliContext.Int("WorkerConverter.MaxErrors") {
				getLogger().Error().Msgf("Maximum errors (#%d) while processing videos", errorCount)
				break
			}
		}

		close(errs)
		close(successes)

		return
	}
}
