package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	// HTTPHeaders are the default headers we will set before making each
	// HTTP request
	HTTPHeaders = map[string]string{
		"User-Agent": "My User Agent 1.0",
		"From":       "testyouremail@domain.com",
	}

	// RedirectMax will set the maximum ollowable redirects for discovering
	// the final URL
	RedirectMax = 10

	// TemporaryDirectoryPrefix is the prefix we will associate with the
	// directory where we will temporarily download content
	TemporaryDirectoryPrefix = "vrddt-download"
)

// DownloadToTemporaryFile is a helper function to download a given URL to a temporary
// file with a specified prefix
func DownloadToTemporaryFile(originalURL string, filePrefix string) (outputFile *os.File, err error) {
	// Attempt to create a tenporary directory
	temporaryDirectory, err := ioutil.TempDir(os.TempDir(), TemporaryDirectoryPrefix)
	if err != nil {
		return
	}

	// Get the content
	httpResponse, err := http.Get(originalURL)
	if err != nil {
		return
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != 200 {
		err = errors.New("HTTP response for the URL to download is not a 200 response code")
		return
	}

	// Attempt to create a temporary file with a specified prefix
	outputFile, err = ioutil.TempFile(temporaryDirectory, filePrefix)
	if err != nil {
		return
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, httpResponse.Body)

	return
}

// GetFinalURL will get the final URL after redirects for a supplied URL
func GetFinalURL(originalURL string) (finalURL string, err error) {
	// Check this is a valid URL
	_, err = url.Parse(originalURL)
	if err != nil {
		return
	}

	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	nextURL := originalURL
	for i := 0; i < RedirectMax; i++ {
		_, err = url.Parse(nextURL)
		if err != nil {
			return
		}

		// Make a header-only call so this is lightweight
		var httpRequest *http.Request
		httpRequest, err = http.NewRequest("HEAD", nextURL, nil)
		if err != nil {
			break
		}

		// Add User-Agent so that reddit doesn't throw us a 429:
		// Too Many Requests
		for key, value := range HTTPHeaders {
			httpRequest.Header.Add(key, value)
		}

		var httpResponse *http.Response
		httpResponse, err = httpClient.Do(httpRequest)
		if err != nil {
			return
		}

		if httpResponse.StatusCode == 200 {
			url := httpResponse.Request.URL
			finalURL = fmt.Sprintf("%s://%s/%s",
				url.Scheme,
				url.Host,
				strings.TrimPrefix(url.Path, "/"),
			)
			break
		} else {
			nextURL = httpResponse.Header.Get("Location")
		}
	}

	return
}

// GetJSONData will return the JSON structured data from a URL
func GetJSONData(url string) (jsonData interface{}, err error) {
	rawData, err := GetRawDataFromURL(url)
	if err != nil {
		return
	}

	jsonData, err = GetJSONDataFromRawData(rawData)

	return
}

// GetJSONDataFromRawData will return a JSON representation of the raw data
// that is passed in
func GetJSONDataFromRawData(data []byte) (jsonData interface{}, err error) {
	if !json.Valid(data) {
		err = errors.New("Invalid JSON: " + string(data[:100]))
		return
	}

	json.Unmarshal(data, &jsonData)

	return
}

// GetRawDataFromURL returns the raw data for a given URL
func GetRawDataFromURL(originalURL string) (data []byte, err error) {
	var httpClient http.Client

	httpRequest, err := http.NewRequest(http.MethodGet, originalURL, nil)
	if err != nil {
		return
	}

	// Add User-Agent so that reddit doesn't throw us a 429: Too Many Requests
	for key, value := range HTTPHeaders {
		httpRequest.Header.Add(key, value)
	}

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return
	}

	data, err = ioutil.ReadAll(httpResponse.Body)

	return
}
