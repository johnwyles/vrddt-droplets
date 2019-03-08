package main

import (
	"context"
	"time"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/mongo"
	"github.com/johnwyles/vrddt-droplets/interfaces/rabbitmq"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
)

// Converter will process a reddit URL into a vrddt video using our internal
// services
func Converter(cfg *config.Config, lg logger.Logger) {
	errs := make(chan error, 0)
	successes := make(chan bool, 0)
	errorCount := 0
	successCount := 0
	messageCount := 0

	// // Initialize the queue
	// q, closeRabbitMQSession, err := rabbitmq.Connect(cfg.RabbitMQ.URI)
	// if err != nil {
	// 	lg.Fatalf("failed to connect to rabbitmq: %v", err)
	// }
	// defer closeRabbitMQSession()
	// redditVideoWorkQueue := rabbitmq.NewRedditVideoWorkQueue(q)
	//
	// // Initialze the store
	// db, closeMongoSession, err := mongo.Connect(cfg.Mongo.URI, true)
	// if err != nil {
	// 	lg.Fatalf("failed to connect to mongodb: %v", err)
	// }
	// defer closeMongoSession()
	// redditVideoStore := mongo.NewRedditVideoStore(db)
	// vrddtVideoStore := mongo.NewVrddtVideoStore(db)
	//
	// // TODO: Initialize the storage
	// vrddtVideoStorage, err = gcs.NewVrddtVideoStorage(&cfg.GCS)
	// if err != nil {
	// 	return
	// }
	//
	// redditVideoConstructor := redditvideos.NewConstructor(lg, redditVideoWorkQueue, redditVideoStore)
	// redditVideoRetriever := redditvideos.NewRetriever(lg, redditVideoStore, vrddtVideoStore)
	//
	// workerConverter := workers.NewConverter(lg, redditVideoWorkQueue, redditVideoStore, vrddtVideoStore)
	//
	// for {
	// 	messageCount++
	//
	// 	go func() {
	// 		// ctx := context.Background()
	// 		ctx := context.TODO()
	//
	// 		if goErr := conv.GetWork(&ctx); goErr != nil {
	// 			lg.Errorf("error getting element of work: %s", err)
	// 			errs <- goErr
	// 			successes <- false
	// 			return
	// 		}
	// 		lg.Infof("Received new request (%d) for work", messageCount)
	//
	// 		if goErr := conv.DoWork(&ctx); goErr != nil {
	// 			lg.Warnf("unable to perform work: %s", err)
	// 			errs <- goErr
	// 			successes <- false
	// 			return
	// 		}
	//
	// 		if goErr := conv.CompleteWork(&ctx); goErr != nil {
	// 			lg.Warnf("unable to complete work: %s", err)
	// 			errs <- goErr
	// 			successes <- false
	// 			return
	// 		}
	//
	// 		successes <- true
	// 		// ctx.Done()
	// 		return
	// 	}()
	//
	// 	select {
	// 	case err := <-errs:
	// 		switch err {
	// 		case reddit.JSONTitleErr, reddit.JSONVideoURLErr, reddit.NotDASHErr:
	// 			lg.Warnf("Warning while processing media: %s", err)
	// 		default:
	// 			errorCount++
	// 			lg.Errorf("Error (#%d of %d allowed) while processing media: %s",
	// 				errorCount,
	// 				cliContext.Int("max-error"),
	// 				err,
	// 			)
	// 		}
	// 	case success := <-successes:
	// 		successCount++
	// 		lg.Infof("Success (#%d): %#v", successCount, success)
	// 	}
	//
	// 	// Let's take a break
	// 	time.Sleep(time.Duration(cliContext.Int64("WorkerConverter.Sleep")) * time.Millisecond)
	//
	// 	// We have exceeded the max-error count so break out of the loop
	// 	// This will exit the infinite for-loop because we exceeded "max-error"
	// 	if errorCount >= cliContext.Int("WorkerConverter.MaxErrors") {
	// 		lg.Errorf("Maximum errors (#%d) while processing videos", errorCount)
	// 		break
	// 	}
	// }
	//
	// close(errs)
	// close(successes)
	//
	// return
}
