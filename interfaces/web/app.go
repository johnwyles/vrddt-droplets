package web

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// URL holds the association between a Reddit URL and vrddt URL for rendering
type URL struct {
	logger.Logger

	RedditURL       string
	VrddtAPIAddress string
	VrddtURL        string
}

type app struct {
	logger.Logger

	render          func(wr http.ResponseWriter, tpl string, data interface{})
	tpl             template.Template
	vrddtAPIAddress string
}

func (app app) indexHandler(wr http.ResponseWriter, req *http.Request) {
	app.render(wr, "index.tpl", URL{VrddtAPIAddress: app.vrddtAPIAddress})
}

// uriHandler will get the vrddt video by the URI path to Reddit
// without the URL (this will catch "reddit.photos" substitution for the URL
// instead of "reddit.com")
func (app app) uriHandler(wr http.ResponseWriter, req *http.Request) {
	if uri, ok := mux.Vars(req)["uri"]; ok {
		url := fmt.Sprintf("https://%s/%s", domain.RedditDomain, uri)

		apiURL := fmt.Sprintf("https://%s/vrddt_videos/", app.vrddtAPIAddress)
		req, err := http.NewRequest(http.MethodGet, apiURL, nil)
		if err != nil {
			app.Fatalf("NewRequest: %s", err)
		}
		req.Header.Set("Content-Type", "application/json")

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		client := &http.Client{Transport: tr}

		q := req.URL.Query()
		q.Add("url", url)
		req.URL.RawQuery = q.Encode()

		resp, err := client.Do(req)
		if err != nil {
			app.Fatalf("Do: %s", err)
		}
		defer resp.Body.Close()

		vrddtVideo := domain.NewVrddtVideo()
		if err = json.NewDecoder(resp.Body).Decode(&vrddtVideo); err != nil {
			app.Fatalf("An error was encountered: %s", err)
			return
		}

		urls := URL{
			RedditURL:       url,
			VrddtURL:        vrddtVideo.URL,
			VrddtAPIAddress: app.vrddtAPIAddress,
		}

		app.Infof("vrddt video URL: %s", urls.VrddtURL)

		app.render(wr, "index.tpl", urls)
	}

}
