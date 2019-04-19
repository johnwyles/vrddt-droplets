package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/converter"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// Services holds all of various services to the subcommands for use
type Services struct {
	Converter converter.Converter
}

var (
	// BuildTimestamp is the build date
	BuildTimestamp string

	// GitHash is the git build hash
	GitHash string

	// Version is the version of the software
	Version string

	// loggerHandle is the current logger facility
	loggerHandle logger.Logger

	// services will be a refer to our global services avaiable to the subcommands
	services = &Services{}
)

func main() {
	// Setup some sensible defaults for the vrddt configuration - this is
	// somewhat ugly but necessary if we want to override configuration file
	// options with the arguments on the command line
	cfg := &config.Config{
		CLI: config.CLIConfig{
			APIURI:  "http://localhost:8080",
			Timeout: 20 * time.Second,
		},
		Converter: config.ConverterConfig{
			FFmpeg: config.ConverterFFmpegConfig{
				Path: "/usr/local/bin/ffmpeg",
			},
		},
		Log: config.LogConfig{
			Format: "text",
			Level:  "info",
		},
	}

	// Loading of all the configuration from environment variables, toml
	// configuration file, or command-line flags
	flags := []cli.Flag{
		&cli.StringFlag{
			EnvVars: []string{"VRDDT_CONFIG"},
			Name:    "config",
			Usage:   "vrddt-cli TOML configuration file",
			Value:   "",
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Aliases:     []string{"a"},
				Destination: &cfg.CLI.APIURI,
				EnvVars:     []string{"VRDDT_CLI_API_URI"},
				Name:        "CLI.APIURI",
				Usage:       "vrddt API URI",
				Value:       cfg.CLI.APIURI,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Aliases:     []string{"lf"},
				Destination: &cfg.Log.Format,
				EnvVars:     []string{"VRDDT_LOG_FORMAT"},
				Name:        "Log.Format",
				Usage:       "Logging format",
				Value:       cfg.Log.Format,
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
	}

	timeStamp, err := strconv.ParseInt(BuildTimestamp, 10, 64)
	if err != nil {
		now := time.Now()
		timeStamp = now.Unix()
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
			func(context *cli.Context) (altsrc.InputSourceContext, error) {
				if context.IsSet("config") {
					return altsrc.NewTomlSourceFromFile(context.String("config"))
				}

				return &altsrc.MapInputSource{}, nil
			},
		),
		Commands: allCommands(cfg),
		Compiled: time.Now(),
		Flags:    flags,
		Name:     "vrddt-cli",
		Prepare:  prepareResources(cfg),
		Usage:    "vrddt standalone CLI tool",
		Version:  fmt.Sprintf("%s [Build Date: %s, Git Hash: %s]", Version, time.Unix(timeStamp, 0), GitHash),
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

	if err = app.Run(os.Args); err != nil {
		loggerHandle.Fatalf("An error occured running the application: %s", err)
		os.Exit(1)

		return
	}

	return
}

// allCommands are all of the commands we are able to run
func allCommands(cfg *config.Config) []*cli.Command {
	return []*cli.Command{
		DownloadWithAPI(cfg),
		DownloadLocally(cfg),
		// GetMetadata(cfg),
	}
}

// afterResources will execute after Action() to cleanup
func afterResources(cfg *config.Config) cli.AfterFunc {
	return func(cliContext *cli.Context) (err error) {
		// Cleanup
		return
	}
}

// prepareResources will setup some common shared resources amoung all of the
// commands and make them avaiable to use
func prepareResources(cfg *config.Config) cli.PrepareFunc {
	return func(cliContext *cli.Context) (err error) {
		// Initalize logger
		loggerHandle = logger.New(os.Stderr, cfg.Log.Level, cfg.Log.Format)

		_, err = url.ParseRequestURI(cfg.CLI.APIURI)
		if err != nil {
			loggerHandle.Fatalf("You did not supply a valid vrddt API URI: %s", cfg.CLI.APIURI)
		}

		// Setup converter
		services.Converter, err = converter.FFmpeg(&cfg.Converter.FFmpeg, loggerHandle)
		if err != nil {
			return
		}

		return nil
	}
}

// rootAction is the what we execute if no commands are specified
func rootAction(cfg *config.Config) cli.ActionFunc {
	return func(cliContext *cli.Context) (err error) {
		cli.ShowAppHelp(cliContext)
		loggerHandle.Fatalf("No sub-command specified")
		os.Exit(1)

		return
	}
}
