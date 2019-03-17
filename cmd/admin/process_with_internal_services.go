package main

import (
	"context"
	"fmt"
	"time"

	mgo "gopkg.in/mgo.v2"
	cli "gopkg.in/urfave/cli.v2"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/queue"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	"github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

const (
	// MaxPollTime specifies the meximum amount of time to wait between polling the database for process completion
	MaxPollTime = 5000

	// MaxTimeout specifies the maximum amount of time to wait for the download
	MaxTimeout = 600
)

// ProcessWithInternalServicesCommand will process a Reddit video using the interal
// vrddt services not exposed by the API
func ProcessWithInternalServicesCommand(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Action: processWithInternalServices,
		Before: beforeProcessWithInternalServices,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Aliases: []string{"r"},
				EnvVars: []string{"VRDDT_ADMIN_PROCESS_WITH_INTERNAL_SERVICES_REDDIT_URL"},
				Name:    "reddit-url",
				Usage:   "Specifies the Reddit URL to pull the video from",
			},
			&cli.IntFlag{
				Aliases: []string{"t"},
				EnvVars: []string{"VRDDT_ADMIN_PROCESS_WITH_INTERNAL_SERVICES_TIMEOUT"},
				Name:    "timeout",
				Usage:   "Specifies the amount of time (in seconds: 1 to 600) to wait for the download",
				Value:   60,
			},
			&cli.IntFlag{
				Aliases: []string{"p"},
				EnvVars: []string{"VRDDT_ADMIN_PROCESS_WITH_INTERNAL_SERVICES_POLL"},
				Name:    "poll",
				Usage:   "Specifies the amount of time (in milliseconds: 10 to 5000) to wait between polling the database for process completion",
				Value:   500,
			},
		},
		Name:  "process-with-internal-services",
		Usage: "Download a Reddit video from a given Reddit URL using the vrddt service",
	}
}

// beforeProcessWithInternalServices will validate that we have set a Reddit URL
func beforeProcessWithInternalServices(cliContext *cli.Context) (err error) {
	if !cliContext.IsSet("reddit-url") {
		cli.ShowCommandHelp(cliContext, cliContext.Command.Name)
		err = fmt.Errorf("A Reddit URL was not given")
	}

	return
}

// processWithInternalServices is basically what the API will be doing without
// endpoints and transport components
func processWithInternalServices(cliContext *cli.Context) (err error) {
	var pollTime int
	pollTime = cliContext.Int("poll")
	if pollTime > 5000 {
		pollTime = 5000
	}
	if pollTime < 10 {
		pollTime = 10
	}

	timeoutTime := cliContext.Int("timeout")
	if timeoutTime > 600 {
		pollTime = 600
	}
	if timeoutTime < 1 {
		timeoutTime = 1
	}

	redditVideo := domain.NewRedditVideo()
	redditVideo.URL = cliContext.String("reddit-url")
	err = redditVideo.SetFinalURL()
	if err != nil {
		return
	}

	// We only really need this because we want to check that the URL contains
	// valid JSON and is a video link and not just any other Reddit URL
	err = redditVideo.SetMetadata()
	if err != nil {
		return
	}

	// Initialize the queue
	q, closeRabbitMQSession, err := queue.RabbitMQ(cliContext.String("Queue.RabbitMQ.URI"))
	if err != nil {
		loggerHandle.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer closeRabbitMQSession()
	redditVideoWorkQueue := queue.RabbitMQ(q)

	// // Initialze the store
	db, closeMongoSession, err := mongo.Connect(cliContext.String("Store.Mongo.URI"), true)
	if err != nil {
		loggerHandle.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer closeMongoSession()
	redditVideoStore := mongo.NewRedditVideoStore(db)
	vrddtVideoStore := mongo.NewVrddtVideoStore(db)

	redditVideoConstructor := redditvideos.NewConstructor(loggerHandle, redditVideoWorkQueue, redditVideoStore)
	redditVideoRetriever := redditvideos.NewRetriever(loggerHandle, redditVideoStore, vrddtVideoStore)
	vrddtVideoRetriever := vrddtvideos.NewRetriever(loggerHandle, vrddtVideoStore)

	// Check if this already exists in the database
	dbRedditVideo, err := redditVideoRetriever.GetByURL(context.TODO(), redditVideo.URL)
	switch err {
	case nil:
		loggerHandle.Debugf("dbRedditVideo: %#v", dbRedditVideo)
		dbVrddtVideo, vrddtErr := redditVideoRetriever.GetByID(context.TODO(), dbRedditVideo.VrddtVideoID)
		switch vrddtErr {
		case nil:
			msg := fmt.Sprintf("vrddt Video Found: %#v", dbVrddtVideo.URL)
			loggerHandle.Infof(msg)
			fmt.Printf("%s\n", msg)
			return nil
		case mgo.ErrNotFound:
			return fmt.Errorf("Reddit Video found (ID: %s) but vrddt Video (ID: %s) was not", redditVideo.ID.Hex(), dbRedditVideo.VrddtVideoID.Hex())
		default:
			return vrddtErr
		}
	case mgo.ErrNotFound:
	default:
		return fmt.Errorf("Something went wrong: %s", err)
	}

	time.Sleep(time.Millisecond * 100)

	// Push a message on to the queue
	err = redditVideoConstructor.Push(context.TODO(), redditVideo)
	if err != nil {
		return
	}

	// Wait a pre-determined amount of time for the worker to fetch, convert,
	// store in the database, and store in storage the video
	timeout := time.After(time.Duration(time.Duration(timeoutTime) * time.Second))
	tick := time.Tick(time.Duration(pollTime) * time.Millisecond)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Operation timed out")
		case <-tick:
			// If the Reddit URL is not found in the database yet keep checking
			temporaryRedditVideo, err := redditVideoRetriever.GetByURL(context.TODO(), redditVideo.URL)
			switch err {
			case nil:
				vrddtVideo, errVrddt := vrddtVideoRetriever.GetByID(context.TODO(), temporaryRedditVideo.VrddtVideoID)
				switch errVrddt {
				case nil:
					msg := fmt.Sprintf("vrddt Video Created: %#v", vrddtVideo.URL)
					loggerHandle.Infof(msg)
					fmt.Printf("%s\n", msg)
					return nil
				case mgo.ErrNotFound:
					return fmt.Errorf("Reddit Video found (ID: %s) but vrddt Video (ID: %s) was not", temporaryRedditVideo.ID.Hex(), temporaryRedditVideo.VrddtVideoID.Hex())
				default:
					return fmt.Errorf("Something went wrong: %s", errVrddt)
				}
			case mgo.ErrNotFound:
				continue
			default:
				return fmt.Errorf("Something went wrong: %s", err)
			}
		}
	}
}
