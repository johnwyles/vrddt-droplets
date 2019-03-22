package store

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
)

// Selector is the map for selecting content from the data  store
type Selector map[string]interface{}

// Store is the generic interface for a persistence store
type Store interface {
	Cleanup(ctx context.Context) (err error)

	CreateRedditVideo(ctx context.Context, redditVideo *domain.RedditVideo) (err error)
	DeleteRedditVideo(ctx context.Context, selector Selector) (err error)
	DeleteRedditVideos(ctx context.Context, selector Selector) (err error)
	GetRedditVideo(ctx context.Context, selector Selector) (redditVideo *domain.RedditVideo, err error)
	GetRedditVideos(ctx context.Context, selector Selector, limit int) (redditVideo []*domain.RedditVideo, err error)

	CreateVrddtVideo(ctx context.Context, vrddtVideo *domain.VrddtVideo) (err error)
	DeleteVrddtVideo(ctx context.Context, selector Selector) (err error)
	DeleteVrddtVideos(ctx context.Context, selector Selector) (err error)
	GetVrddtVideo(ctx context.Context, selector Selector) (vrddtVideo *domain.VrddtVideo, err error)
	GetVrddtVideos(ctx context.Context, selector Selector, limit int) (vrddtVideo []*domain.VrddtVideo, err error)

	Init(ctx context.Context) (err error)
}
