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

	if resp.StatusCode == 404 {
		resp.Body.Close()

		lg.Debugf("reddit video does not exist yet for URL: %s", cfg.RedditURL)
		jsonData, err := json.Marshal(map[string]string{"url": cfg.RedditURL})
		if err != nil {
			lg.Fatalf("There was an issue marshalling JSON: %s", err)
		}

		req, err = http.NewRequest(http.MethodPost, apiURL+"queue", bytes.NewBuffer(jsonData))
		if err != nil {
			lg.Fatalf("An error occurred POST to API URL: %s Data: %s Reason: %s", apiURL, string(jsonData), err)
		}

		resp, err = client.Do(req)
		if err != nil {
			lg.Fatalf("Queueing error: %s", err)
		}
		defer resp.Body.Close()
	} else if resp.StatusCode < 200 || resp.StatusCode > 400 {
		lg.Fatalf("Received a status code < 200 or > 400: %d reason: %s", resp.StatusCode, resp.Status)
	} else {
		redditVideo := &domain.RedditVideo{}
		if err = json.NewDecoder(resp.Body).Decode(&redditVideo); err != nil {
			lg.Warnf("Unable to decode response into a reddit video: %s", err)
		}

		req, err = http.NewRequest(http.MethodGet, cfg.VrddtAPIURI+"/vrddt_videos/"+redditVideo.VrddtVideoID.Hex(), nil)
		if err != nil {
			lg.Fatalf("VrddtVideo: %s", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		if err != nil {
			lg.Fatalf("Queueing error: %s", err)
		}
		defer resp.Body.Close()

		vrddtVideo := map[string]string{"url": ""}
		if err = json.NewDecoder(resp.Body).Decode(&vrddtVideo); err != nil {
			lg.Fatalf("Unable to decode response into a vrddt video: %s", err)
		}

		lg.Infof("Got vrddt video URL: %#v", vrddtVideo["url"])
		return
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
			lg.Fatalf("Operation timed out at after '%d' seconds.", timeoutTime)
		case <-tick:
			req, err = http.NewRequest(http.MethodGet, apiURL, nil)
			if err != nil {
				lg.Fatalf("An error occurred GET to API URL: %s Reason: %s", apiURL, err)
			}
			q := req.URL.Query()
			q.Add("url", cfg.RedditURL)
			req.URL.RawQuery = q.Encode()

			resp, err = client.Do(req)
			if err != nil {
				lg.Fatalf("Do LOOP: %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 {
				lg.Debugf("Got status 404 so vrddt video does not exist yet")
				continue
			} else if resp.StatusCode < 200 || resp.StatusCode > 400 {
				lg.Fatalf("Received a status code < 200 or > 400: %d reason: %s", resp.StatusCode, resp.Status)
			}

			// Use json.Decode for reading streams of JSON data
			redditVideo := &domain.RedditVideo{}
			if err = json.NewDecoder(resp.Body).Decode(&redditVideo); err != nil {
				lg.Warnf("Unable to decode response into a reddit video: %s", err)
			}

			req, err = http.NewRequest(http.MethodGet, apiURL+"/"+redditVideo.VrddtVideoID.Hex(), nil)
			if err != nil {
				lg.Fatalf("VrddtVideo: %s", err)
			}
			req.Header.Set("Content-Type", "application/json")

			vrddtVideo := &domain.VrddtVideo{}
			if err = json.NewDecoder(resp.Body).Decode(&vrddtVideo); err != nil {
				lg.Warnf("Unable to decode response into a vrddt video: %s", err)
			}

			lg.Infof("Got vrddt video: %#v", vrddtVideo)

			return
		}
	}
}
