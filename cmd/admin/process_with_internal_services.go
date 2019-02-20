package main

import (
	"context"
	"net/url"
	"time"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/mongo"
	"github.com/johnwyles/vrddt-droplets/interfaces/rabbitmq"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	// "github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

// ProcessWithInternalServices will process the Reddit URL using the internal
// services directly instead of calling the public API
func ProcessWithInternalServices(cfg config, lg logger.Logger) {
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

	redditVideoWorkQueue := rabbitmq.NewRedditVideoWorkQueue(q)

	redditVideoConstructor := redditvideos.NewConstructor(lg, redditVideoWorkQueue, redditVideoStore)
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
	} else {
		redditVideo = &domain.RedditVideo{
			URL: cfg.RedditURL,
		}
	}

	if err := redditVideoConstructor.Push(context.TODO(), redditVideo); err != nil {
		lg.Fatalf("failed to create reddit video: %s", err)
	}

	lg.Infof("unique reddit video URL queued with URL of: ", redditVideo.URL)

	var pollTime int
	pollTime = cfg.PollTime
	if pollTime > 5000 {
		pollTime = 5000
	}
	if pollTime < 10 {
		pollTime = 10
	}

	timeoutTime := cfg.Timeout
	if timeoutTime > 600 {
		pollTime = 600
	}
	if timeoutTime < 1 {
		timeoutTime = 1
	}

	// Wait a pre-determined amount of time for the worker to fetch, convert,
	// store in the database, and store in storage the video
	timeout := time.After(
		time.Duration(
			time.Duration(timeoutTime) * time.Second,
		),
	)
	tick := time.Tick(time.Duration(pollTime) * time.Millisecond)
	for {
		select {
		case <-timeout:
			lg.Fatalf("Operation timed out at after '%d' seconds.", timeout)
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
