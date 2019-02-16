package main

import (
	"context"
	"flag"
	"net/url"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/johnwyles/vrddt-droplets/interfaces/mongo"
	"github.com/johnwyles/vrddt-droplets/interfaces/rabbitmq"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	// "github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

func main() {
	// using standard library "flag" package
	flag.String("REDDIT_URL", "", "Reddit URL for content")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	cfg := loadConfig()
	lg := logger.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)
	lg.Infof("config: %#v", cfg)

	_, err := url.ParseRequestURI(cfg.RedditURL)
	if err != nil {
		lg.Fatalf("You did not supply a valid URL: %s", cfg.RedditURL)
	}

	db, closeMongoSession, err := mongo.Connect(cfg.MongoURI, true)
	if err != nil {
		lg.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer closeMongoSession()

	redditVideoStore := mongo.NewRedditVideoStore(db)
	vrddtVideoStore := mongo.NewVrddtVideoStore(db)

	q, closeRabbitMQSession, err := rabbitmq.Connect(cfg.RabbitMQURI)
	if err != nil {
		lg.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer closeRabbitMQSession()

	workQueue := rabbitmq.NewWorkQueue(q)

	redditVideoConstructor := redditvideos.NewConstructor(lg, workQueue, redditVideoStore)
	redditVideoRetriever := redditvideos.NewRetriever(lg, redditVideoStore, vrddtVideoStore)

	redditVideo, err := redditVideoRetriever.GetByURL(context.TODO(), cfg.RedditURL)
	if err != nil {
		switch errors.Type(err) {
		case errors.TypeUnknown:
			lg.Warnf("error getting URL from db: %s", err)
			return
		case errors.TypeResourceNotFound:
		}
	}

	if redditVideo != nil {
		vrddtVideo, errVrddt := redditVideoRetriever.GetVrddtVideoByID(context.TODO(), redditVideo.VrddtVideoID)
		if err != nil {
			switch errors.Type(errVrddt) {
			case errors.TypeResourceNotFound:
				lg.Fatalf("Reddit Video found (ID: %s) but vrddt Video (ID: %s) was not", redditVideo.ID.Hex(), redditVideo.VrddtVideoID.Hex())
			default:
				lg.Fatalf("Something went wrong: %s", errVrddt)
			}
		}

		lg.Infof("reddit video already in db with id '%s' and a vrddt URL of: ", redditVideo.ID, vrddtVideo.URL)
		return
	}

	if err := redditVideoConstructor.Push(context.TODO(), redditVideo); err != nil {
		lg.Fatalf("failed to create reddit video: %s", err)
	}

	lg.Infof("unique reddit video URL queued with id '%s' and URL of: ", redditVideo.ID, redditVideo.URL)

	timeoutTime := 5
	pollTime := 500

	// Wait a pre-determined amount of time for the worker to fetch, convert,
	// store in the database, and store in storage the video
	timeout := time.After(time.Duration(time.Duration(timeoutTime) * time.Second))
	tick := time.Tick(time.Duration(pollTime) * time.Millisecond)
	for {
		select {
		case <-timeout:
			lg.Fatalf("Operation timed out")
		case <-tick:
			// If the Reddit URL is not found in the database yet keep checking
			temporaryRedditVideo, err := redditVideoRetriever.GetByURL(context.TODO(), redditVideo.URL)
			if err != nil {
				switch errors.Type(err) {
				case errors.TypeUnknown:
					lg.Fatalf("Something went wrong: %s", err)
				case errors.TypeResourceNotFound:
					continue
				}
			}

			vrddtVideo, errVrddt := redditVideoRetriever.GetVrddtVideoByID(context.TODO(), redditVideo.VrddtVideoID)
			if errVrddt != nil {
				switch errors.Type(errVrddt) {
				case errors.TypeResourceNotFound:
					lg.Fatalf("Reddit Video found (ID: %s) but associated vrddt Video (ID: %s) was not", temporaryRedditVideo.ID.Hex(), temporaryRedditVideo.VrddtVideoID.Hex())
				default:
					lg.Fatalf("Something went wrong: %s", errVrddt)
				}
			}

			lg.Infof("vrddt Video Created: %#v", vrddtVideo)
			return
		}
	}
}

type config struct {
	LogLevel    string
	LogFormat   string
	MongoURI    string
	RabbitMQURI string
	RedditURL   string
}

func loadConfig() config {
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_FORMAT", "text")
	viper.SetDefault("MONGO_URI", "mongodb://admin:password@localhost:27017/vrddt")
	viper.SetDefault("RABBITMQ_URI", "amqp://admin:password@localhost:5672")
	viper.SetDefault("REDDIT_URL", "")

	viper.ReadInConfig()
	viper.AutomaticEnv()

	return config{
		// application configuration
		LogLevel:    viper.GetString("LOG_LEVEL"),
		LogFormat:   viper.GetString("LOG_FORMAT"),
		MongoURI:    viper.GetString("MONGO_URI"),
		RabbitMQURI: viper.GetString("RABBITMQ_URI"),
		RedditURL:   viper.GetString("REDDIT_URL"),
	}
}
