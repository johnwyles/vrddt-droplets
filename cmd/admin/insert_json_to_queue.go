package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	cli "gopkg.in/urfave/cli.v2"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/mongo"
	"github.com/johnwyles/vrddt-droplets/interfaces/rabbitmq"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
)

// InsertJSONToQueueCommand will take whatever garbage or valid URLs you throw in a
// JSON file formatted with unmarshaled Reddit Video structs and insert them
// to the Queue
func InsertJSONToQueueCommand(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Action: insertJSONToQueue,
		Before: beforeInsertJSONToQueue,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Aliases: []string{"f"},
				EnvVars: []string{"VRDDT_ADMIN_INSERT_JSON_TO_QUEUE_FILE"},
				Name:    "json-file",
				Usage:   "Specifies the JSON file to load Reddit URLs from",
				Value:   "",
			},
		},
		Name:  "insert-json-to-queue",
		Usage: "Blindly instert a JSON file of Reddit data to the Queue",
	}
}

// beforeInsertJSONToQueue will validate that we have set a JSON file
func beforeInsertJSONToQueue(cliContext *cli.Context) (err error) {
	if !cliContext.IsSet("json-file") {
		cli.ShowCommandHelp(cliContext, cliContext.Command.Name)
		err = fmt.Errorf("A JSON file was not supplied")
	}

	return
}

// insertJSONToQueue will throw whatever "json-file" argument as unmarshaled
// RedditVideo structs into the Queue
func insertJSONToQueue(cliContext *cli.Context) (err error) {
	data, err := ioutil.ReadFile(cliContext.String("json-file"))
	if err != nil {
		return
	}

	var redditVideos []domain.RedditVideo
	if err = json.Unmarshal(data, &redditVideos); err != nil {
		return
	}

	// Initialize the queue
	q, closeRabbitMQSession, err := rabbitmq.Connect(cliContext.String("Queue.RabbitMQ.URI"))
	if err != nil {
		loggerHandle.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer closeRabbitMQSession()
	redditVideoWorkQueue := rabbitmq.NewRedditVideoWorkQueue(q)

	// // Initialze the store
	db, closeMongoSession, err := mongo.Connect(cliContext.String("Store.Mongo.URI"), true)
	if err != nil {
		loggerHandle.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer closeMongoSession()
	redditVideoStore := mongo.NewRedditVideoStore(db)

	redditVideoConstructor := redditvideos.NewConstructor(loggerHandle, redditVideoWorkQueue, redditVideoStore)

	// NOTE: This does NOT check the DB at all before inserting the video into
	// the Queue so we can test if the API and Web tiers are doing their job as
	// this should never occur
	for _, redditVideo := range redditVideos {
		// message, err := json.Marshal(redditVideo)
		if err != nil {
			loggerHandle.Errorf("Problem marshaling to JSON %#v: %s", redditVideo, err)
			continue
		}
		loggerHandle.Debugf("Enqueuing video: %#v\n", redditVideo)

		if err = redditVideoConstructor.Push(context.TODO(), &redditVideo); err != nil {
			loggerHandle.Errorf("Error pushing JSON to queue: %s", err)
		}
	}

	return
}
