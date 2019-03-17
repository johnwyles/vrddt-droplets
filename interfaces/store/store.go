package store

import (
	"github.com/johnwyles/vrddt-droplets/domain"
)

// Selector is the map for selecting content from the data  store
type Selector map[string]interface{}

// Store is the generic interface for a persistence store
type Store interface {
	Cleanup() (err error)
	CreateRedditVideo(redditVideo *domain.RedditVideo) (err error)
	CreateVrddtVideo(vrddtVideo *domain.VrddtVideo) (err error)
	DeleteRedditVideo(selector Selector) (err error)
	DeleteRedditVideos(selector Selector) (err error)
	DeleteVrddtVideo(selector Selector) (err error)
	DeleteVrddtVideos(selector Selector) (err error)
	GetRedditVideo(selector Selector) (redditVideo *domain.RedditVideo, err error)
	GetRedditVideos(selector Selector) (redditVideo []*domain.RedditVideo, err error)
	GetVrddtVideo(selector Selector) (vrddtVideo *domain.VrddtVideo, err error)
	GetVrddtVideos(selector Selector) (vrddtVideo []*domain.VrddtVideo, err error)
	Init() (err error)
}
