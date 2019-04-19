package main

import (
	"context"
	"time"

	cli "gopkg.in/urfave/cli.v2"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
)

// Processor will process a Reddit URL into a vrddt video using our internal
// services
func Processor(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Action: processor,
		After:  afterProcessor(cfg),
		Before: beforeProcessor(cfg),
		Flags: []cli.Flag{
			&cli.IntFlag{
				Aliases: []string{"e"},
				EnvVars: []string{"VRDDT_WORKER_CONVERTER_MAX_ERRORS"},
				Name:    "max-errors",
				Usage:   "Maximum number of errors to encounter before exitings process",
				Value:   10,
			},
			&cli.IntFlag{
				Aliases: []string{"s"},
				EnvVars: []string{"VRDDT_WORKER_CONVERTER_SLEEP"},
				Name:    "sleep",
				Usage:   "Amount of time to sleep (in milliseconds) between requests",
				Value:   5090,
			},
		},
		Name:  "processor",
		Usage: "Run the Processor which will process Reddit URLs from the queue turning them in to vrddt videos",
	}
}

// afterProcessor will execute after Action() to cleanup
func afterProcessor(cfg *config.Config) cli.AfterFunc {
	return func(cliContext *cli.Context) (err error) {
		// TODO: Context
		ctx := context.TODO()

		// We don't care about any cleanup errors
		services.Queue.Cleanup(ctx)
		services.Store.Cleanup(ctx)
		services.Storage.Cleanup(ctx)

		return
	}
}

// beforeConverter will validate before the command is run
func beforeProcessor(cfg *config.Config) cli.BeforeFunc {
	return func(cliContext *cli.Context) (err error) {
		// TODO: Context
		ctx := context.TODO()

		services.Converter.Init(ctx)
		services.Queue.Init(ctx)
		services.Storage.Init(ctx)
		services.Store.Init(ctx)
		services.Worker.Init(ctx)

		return
	}
}

// processor is the main function which will perform work: digesting Reddit URLs
// from the queue, downloading the video, converting the file to a vrddt video,
// storing the viceo in storage, and saving the results in our data store
func processor(cliContext *cli.Context) (err error) {
	errs := make(chan error, 0)
	successes := make(chan bool, 0)
	errorCount := 0
	successCount := 0
	messageCount := 0

	for {
		messageCount++

		go func() {
			// TODO: How do we use context? Study
			// ctx := context.Background()
			ctx := context.TODO()

			if err := services.Worker.GetWork(ctx); err != nil {
				loggerHandle.Errorf("Error getting element of work: %s", err)
				errs <- err
				successes <- false
				return
			}
			loggerHandle.Infof("Received new request (%d) for work", messageCount)

			if err = services.Worker.DoWork(ctx); err != nil {
				loggerHandle.Errorf("Unable to perform work: %s", err)
				errs <- err
				successes <- false
				return
			}
			loggerHandle.Infof("Performing work on request (%d)", messageCount)

			if err = services.Worker.CompleteWork(ctx); err != nil {
				loggerHandle.Errorf("Unable to complete work: %s", err)
				errs <- err
				successes <- false
				return
			}
			loggerHandle.Infof("Completed work for request (%d)", messageCount)

			successes <- true
			// ctx.Done()
			return
		}()

		select {
		case err := <-errs:
			switch err {
			case domain.ErrJSONTitle, domain.ErrJSONVideoURL, domain.ErrNotDASH:
				loggerHandle.Warnf("Warning while processing media: %s", err)
			default:
				errorCount++
				loggerHandle.Warnf("Error (#%d of %d allowed) while processing media: %s",
					errorCount,
					cliContext.Int("max-errors"),
					err,
				)
			}
		case success := <-successes:
			successCount++
			loggerHandle.Infof("Success (#%d): %#v", successCount, success)
		}

		// Let's take a break
		time.Sleep(time.Duration(cliContext.Int64("sleep")) * time.Millisecond)

		// We have exceeded the max-error count so break out of the loop
		// This will exit the infinite for-loop because we exceeded "max-error"
		if errorCount >= cliContext.Int("max-errors") {
			loggerHandle.Errorf("Maximum errors (#%d) while processing videos", errorCount)
			break
		}
	}

	close(errs)
	close(successes)

	return
}
