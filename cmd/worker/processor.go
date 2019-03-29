package main

import (
	"context"
	"fmt"
	"time"

	cli "gopkg.in/urfave/cli.v2"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/usecases/converter"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	"github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

// Processor will process a reddit URL into a vrddt video using our internal
// services
func Processor(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Action: processor,
		Before: beforeProcessor,
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

// beforeConverter will validate that we have set a JSON file
func beforeProcessor(cliContext *cli.Context) (err error) {
	if !cliContext.IsSet("json-file") {
		cli.ShowCommandHelp(cliContext, cliContext.Command.Name)
		err = fmt.Errorf("A JSON file was not supplied")
	}

	return
}

func processor(cliContext *cli.Context) (err error) {
	errs := make(chan error, 0)
	successes := make(chan bool, 0)
	errorCount := 0
	successCount := 0
	messageCount := 0

	vrddtVideoConstructor := vrddtvideos.NewConstructor(loggerHandle, services.Store)
	vrddtVideoDestructor := vrddtvideos.NewDestructor(loggerHandle, services.Store)
	vrddtVideoRetriever := vrddtvideos.NewRetriever(loggerHandle, services.Store)

	redditVideoConstructor := redditvideos.NewConstructor(loggerHandle, services.Queue, services.Store)
	redditVideoDestructor := redditvideos.NewDestructor(loggerHandle, services.Queue, services.Store)
	redditVideoRetriever := redditvideos.NewRetriever(loggerHandle, services.Store)

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
}
