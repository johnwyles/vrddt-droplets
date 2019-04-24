package main

import (
	"context"
	"net/url"
	"os"

	cli "gopkg.in/urfave/cli.v2"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
)

// DownloadLocally will process a Reddit URL using only local resources (i.e. http download and ffmpeg for conversion)
func DownloadLocally(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Action: downloadLocally,
		Before: beforeDownloadLocally,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Aliases: []string{"r"},
				EnvVars: []string{"VRDDT_CLI_DOWNLOAD_LOCALLY_REDDIT_URL"},
				Name:    "reddit-url",
				Usage:   "Specifies the Reddit URL to pull the video from",
			},
			&cli.StringFlag{
				Aliases: []string{"o"},
				EnvVars: []string{"VRDDT_CLI_OUTPUT_FILE"},
				Name:    "output-file",
				Usage:   "Specifies the output file where the final processed video will reside",
				Value:   "vrddt-output.mp4",
			},
		},
		Name:  "download-locally",
		Usage: "Download a Reddit video from a given Reddit URL using only local resouces (i.e. http download and ffmpeg for conversion)",
	}
}

// beforeDownloadLocally will validate that we have set a Reddit URL
func beforeDownloadLocally(cliContext *cli.Context) (err error) {
	if !cliContext.IsSet("reddit-url") {
		cli.ShowCommandHelp(cliContext, cliContext.Command.Name)
		loggerHandle.Fatalf("A Reddit URL was not given")
		os.Exit(1)

		return
	}

	_, err = url.ParseRequestURI(cliContext.String("reddit-url"))
	if err != nil {
		loggerHandle.Fatalf("You did not supply a valid Reddit URL: %s", cliContext.String("reddit-url"))
		os.Exit(1)

		return
	}

	if !cliContext.IsSet("output-file") {
		cli.ShowCommandHelp(cliContext, cliContext.Command.Name)
		loggerHandle.Fatalf("You have not specified an output file path")
		os.Exit(1)

		return
	}

	// TODO: Context
	ctx := context.TODO()

	// Initialize converter
	if err = services.Converter.Init(ctx); err != nil {
		return
	}

	return
}

// downloadLocally will process a Reddit URL using the vrddt API and download the
// resulting video locally
func downloadLocally(cliContext *cli.Context) (err error) {
	outputFile := cliContext.String("output-file")

	// Setup a new Reddit video with all the video information
	redditVideo := domain.NewRedditVideo()
	redditVideo.URL = cliContext.String("reddit-url")
	err = redditVideo.SetFinalURL()
	if err != nil {
		return
	}

	loggerHandle.Infof("Getting video for Reddit URL: %s", redditVideo.URL)

	// We only really need this because we want to check that the URL contains
	// valid JSON and is a video link and not just any other Reddit URL
	err = redditVideo.SetMetadata()
	if err != nil {
		return
	}

	if err = redditVideo.Download(); err != nil {
		return
	}

	loggerHandle.Infof("Downloaded Reddit video: %#v", redditVideo)

	// We don't care if the Audio file fails to download as there are
	// plenty of videos on Reddit that do not have audio
	if redditVideo.RedditAudio != nil && redditVideo.RedditAudio.FileHandle != nil && redditVideo.RedditAudio.FilePath != "" {
		defer redditVideo.RedditAudio.FileHandle.Close()
		defer os.Remove(redditVideo.RedditAudio.FilePath)
	} else {
		redditVideo.RedditAudio = &domain.RedditAudio{
			FileHandle: nil,
			FilePath:   "",
		}
	}

	defer redditVideo.FileHandle.Close()
	defer os.Remove(redditVideo.FilePath)

	loggerHandle.Infof("Converting media for Reddit URL: %s", redditVideo.URL)

	ctx := context.TODO()
	if err = services.Converter.Convert(
		ctx,
		redditVideo.FilePath,
		redditVideo.RedditAudio.FilePath,
		outputFile,
	); err != nil {
		return
	}

	loggerHandle.Infof("Completed downloading video for Reddit URL [%s]: %s",
		redditVideo.URL,
		outputFile,
	)

	return
}
