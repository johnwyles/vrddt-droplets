package main

import (
	"context"
	"time"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/mongo"
	"github.com/johnwyles/vrddt-droplets/interfaces/rabbitmq"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	"github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

// Converter will process a reddit URL into a vrddt video using our internal
// services
func Converter(cfg *config.Config) {
	errs := make(chan error, 0)
	successes := make(chan bool, 0)
	errorCount := 0
	successCount := 0
	messageCount := 0

	db, closeMongoSession, err := mongo.Connect(cfg.Store.Mongo.URI, true)
	if err != nil {
		loggerHandle.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer closeMongoSession()

	redditVideoStore := mongo.NewRedditVideoStore(db)
	vrddtVideoStore := mongo.NewVrddtVideoStore(db)

	q, closeRabbitMQSession, err := rabbitmq.Connect(cfg.Queue.RabbitMQ.URI)
	if err != nil {
		loggerHandle.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer closeRabbitMQSession()

	redditVideoWorkQueue := rabbitmq.NewRedditVideoWorkQueue(q)

	vrddtVideoConstructor := vrddtvideos.NewConstructor(loggerHandle, vrddtVideoStore)
	vrddtVideoDestructor := vrddtvideos.NewDestructor(loggerHandle, vrddtVideoStore)
	vrddtVideoRetriever := vrddtvideos.NewRetriever(loggerHandle, vrddtVideoStore)

	redditVideoConstructor := redditvideos.NewConstructor(loggerHandle, redditVideoWorkQueue, redditVideoStore)
	redditVideoDestructor := redditvideos.NewDestructor(loggerHandle, redditVideoWorkQueue, redditVideoStore)
	redditVideoRetriever := redditvideos.NewRetriever(loggerHandle, redditVideoStore, vrddtVideoStore)

	conv, err := ffmpeg.NewConverter(
		GlobalServices.Converter,
		q,
		db,
		GlobalServices.Storage,
	)

	for {
		messageCount++

		go func() {
			// TODO: How do we use context? Study
			// ctx := context.Background()
			ctx := context.TODO()

			if goErr := conv.GetWork(&ctx); goErr != nil {
				loggerHandle.Errorf("Error getting element of work: %s", err)
				errs <- goErr
				successes <- false
				return
			}
			loggerHandle.Infof("Received new request (%d) for work", messageCount)

			if goErr := conv.DoWork(&ctx); goErr != nil {
				loggerHandle.Errorf("Unable to perform work: %s", err)
				errs <- goErr
				successes <- false
				return
			}

			if goErr := conv.CompleteWork(&ctx); goErr != nil {
				loggerHandle.Errorf("Unable to complete work: %s", err)
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
				loggerHandle.Warnf("Warning while processing media: %s", err)
			default:
				errorCount++
				loggerHandle.Warnf("Error (#%d of %d allowed) while processing media: %s",
					errorCount,
					cliContext.Int("max-error"),
					err,
				)
			}
		case success := <-successes:
			successCount++
			loggerHandle.Infof("Success (#%d): %#v", successCount, success)
		}

		// Let's take a break
		time.Sleep(time.Duration(cliContext.Int64("WorkerConverter.Sleep")) * time.Millisecond)

		// We have exceeded the max-error count so break out of the loop
		// This will exit the infinite for-loop because we exceeded "max-error"
		if errorCount >= cliContext.Int("WorkerConverter.MaxErrors") {
			loggerHandle.Errorf("Maximum errors (#%d) while processing videos", errorCount)
			break
		}
	}

	close(errs)
	close(successes)

	return
}
