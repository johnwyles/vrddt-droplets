package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	cli "gopkg.in/urfave/cli.v2"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
)

// Download will process a Reddit URL with the vrddt API and download the
// resulting video locally
func Download(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Action: download,
		Before: beforeDownload,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Aliases: []string{"r"},
				EnvVars: []string{"VRDDT_CLI_DOWNLOAD_REDDIT_URL"},
				Name:    "reddit-url",
				Usage:   "Specifies the Reddit URL to pull the video from",
			},
			&cli.IntFlag{
				Aliases: []string{"t"},
				EnvVars: []string{"VRDDT_CLI_DOWNLOAD_TIMEOUT"},
				Name:    "timeout",
				Usage:   "Specifies the amount of time (in seconds: 1 to 600) to wait for the download",
				Value:   60,
			},
			&cli.IntFlag{
				Aliases: []string{"p"},
				EnvVars: []string{"VRDDT_CLI_DOWNLOAD_POLL"},
				Name:    "poll",
				Usage:   "Specifies the amount of time (in milliseconds: 10 to 5000) to wait between polling the database for process completion",
				Value:   500,
			},
		},
		Name:  "download",
		Usage: "Download a Reddit video from a given Reddit URL using the vrddt service",
	}
}

// beforeDownloads will validate that we have set a Reddit URL
func beforeDownload(cliContext *cli.Context) (err error) {
	if !cliContext.IsSet("reddit-url") {
		cli.ShowCommandHelp(cliContext, cliContext.Command.Name)
		err = fmt.Errorf("A Reddit URL was not given")
	}

	_, err = url.ParseRequestURI(cliContext.String("reddit-url"))
	if err != nil {
		loggerHandle.Fatalf("You did not supply a valid Reddit URL: %s", cliContext.String("reddit-url"))
	}

	return
}

// download will process a Reddit URL using the vrddt API and download the
// resulting video locally
func download(cliContext *cli.Context) (err error) {
	apiURL := cliContext.String("CLI.APIURI") + "/reddit_videos/"
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		loggerHandle.Fatalf("NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	q := req.URL.Query()
	q.Add("url", cliContext.String("reddit-url"))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		loggerHandle.Fatalf("Do: %s", err)
	}
	defer resp.Body.Close()

	vrddtVideo := domain.NewVrddtVideo()
	if err = json.NewDecoder(resp.Body).Decode(&vrddtVideo); err != nil {
		loggerHandle.Fatalf("An error was encountered: %s", err)
		return
	}

	loggerHandle.Infof("vrddt video URL: %s", vrddtVideo.URL)
	return
}
