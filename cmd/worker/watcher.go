package main

import (
	"os"

	cli "gopkg.in/urfave/cli.v2"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
)

// Watcher will watch Reddit for new videos and pre-process them without
// requiring a submission from the user
func Watcher(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Action: watcher,
		After:  afterWatcher(cfg),
		Before: beforeWatcher(cfg),
		Flags:  []cli.Flag{},
		Name:   "watcher",
		Usage:  "Run the Watcher which will watch Reddit for new videos and pre-process them without requiring a submission from the user",
	}
}

// afterWatcher will execute after Action() to cleanup
func afterWatcher(cfg *config.Config) cli.AfterFunc {
	return func(cliContext *cli.Context) (err error) {
		return
	}
}

// beforeWatcher will validate before the command is run
func beforeWatcher(cfg *config.Config) cli.BeforeFunc {
	return func(cliContext *cli.Context) (err error) {
		return
	}
}

// watcher is the main function which will perform work: it will watch the
// stream of Reddit submission and pre-process them withour requiring a
// submission from users
func watcher(cliContext *cli.Context) (err error) {
	loggerHandle.Fatalf("Not implemented yet.")
	os.Exit(1)

	return
}
