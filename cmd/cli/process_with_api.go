package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

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
	query := map[string]string{"url": cfg.RedditURL}

	jsonData, err := json.Marshal(query)
	if err != nil {
		lg.Fatalf("JSON Marshal: %s", err)
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		lg.Fatalf("NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		lg.Fatalf("Do: %s", err)
	}
	defer resp.Body.Close()

	vrddtVideo := &domain.VrddtVideo{}

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&vrddtVideo); err != nil {
		lg.Warnf("Unable to decode response into a vrddt video: %s", err)
	}

	var pollTime int
	pollTime = cfg.PollTime
	if pollTime > 5000 {
		pollTime = 5000
	}
	if pollTime < 10 {
		pollTime = 10
	}

	timeoutTime := cfg.Timeout
	if timeoutTime > 600 {
		pollTime = 600
	}
	if timeoutTime < 1 {
		timeoutTime = 1
	}

	// Wait a pre-determined amount of time for the worker to fetch, convert,
	// store in the database, and store in storage the video
	timeout := time.After(
		time.Duration(
			time.Duration(timeoutTime) * time.Second,
		),
	)
	tick := time.Tick(time.Duration(pollTime) * time.Millisecond)
	for {
		select {
		case <-timeout:
			lg.Fatalf("Operation timed out at after '%d' seconds.", timeout)
		case <-tick:
			return
		}
	}
}
