package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// ProcessWithAPI will process a reddit URL into a vrddt video using the API
func ProcessWithAPI(cfg config, lg logger.Logger) {
	_, err := url.ParseRequestURI(cfg.VrddtAPIURI)
	if err != nil {
		lg.Fatalf("You did not supply a valid vrddt API URI: %s", cfg.VrddtAPIURI)
	}

	_, err = url.ParseRequestURI(cfg.RedditURL)
	if err != nil {
		lg.Fatalf("You did not supply a valid Reddit URL: %s", cfg.RedditURL)
	}

	apiURL := cfg.VrddtAPIURI + "/reddit_videos/"
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		lg.Fatalf("NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	q := req.URL.Query()
	q.Add("url", cfg.RedditURL)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		lg.Fatalf("Do: %s", err)
	}
	defer resp.Body.Close()

	vrddtVideo := domain.NewVrddtVideo()
	if err := json.NewDecoder(resp.Body).Decode(&vrddtVideo); err != nil {
		lg.Fatalf("An error was encountered: %s", err)
		return
	}

	lg.Infof("vrddt video URL: %s", vrddtVideo.URL)
}
