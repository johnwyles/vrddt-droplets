package redditvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// Query represents parameters for executing a search. Zero valued fields
// in the query will be ignored.
type Query struct {
	ID           bson.ObjectId `json:"id,omitempty"`
	URL          string        `json:"url,omitempty"`
	VrddtVideoID bson.ObjectId `json:"vrddt_video_id,omitempty"`
}

// Retriever provides functions for retrieving user and user info.
type Retriever struct {
	logger.Logger

	store store.Store
}

// NewRetriever initializes an instance of Retriever with given store.
func NewRetriever(lg logger.Logger, store store.Store) *Retriever {
	return &Retriever{
		Logger: lg,

		store: store,
	}
}

// GetByID finds a reddit video by id.
func (ret *Retriever) GetByID(ctx context.Context, id bson.ObjectId) (redditVideo *domain.RedditVideo, err error) {
	redditVideo, err = ret.store.GetRedditVideo(
		ctx, store.Selector{
			"_id": id,
		},
	)
	if err != nil {
		ret.Debugf("Failed to find Reddit video with ID '%s': %v", id.Hex(), err)
		return nil, err
	}

	return
}

// GetByURL finds a reddit video by url.
func (ret *Retriever) GetByURL(ctx context.Context, url string) (redditVideo *domain.RedditVideo, err error) {
	// TODO: If there is a way to do this entirely client-side we can save some time
	finalURL, err := domain.GetFinalURL(url)
	if err != nil {
		return nil, err
	}

	redditVideo, err = ret.store.GetRedditVideo(
		ctx,
		store.Selector{
			"url": finalURL,
		},
	)
	if err != nil {
		ret.Debugf("Failed to find Reddit video with URL '%s': %v", url, err)
		return nil, err
	}

	return
}

// GetVrddtVideoByID will return the vrddt video by it's ID in the store.
func (ret *Retriever) GetVrddtVideoByID(ctx context.Context, id bson.ObjectId) (vrddtVideo *domain.VrddtVideo, err error) {
	vrddtVideo, err = ret.store.GetVrddtVideo(
		ctx,
		store.Selector{
			"_id": id,
		},
	)
	if err != nil {
		ret.Debugf("Failed to find vrddt video with ID '%s': %v", id.Hex(), err)
		return nil, err
	}

	return vrddtVideo, nil
}

// Search finds all the vrddt videos matching the parameters in the query.
// TODO: This is incomplete
func (ret *Retriever) Search(ctx context.Context, selector store.Selector, limit int) ([]*domain.RedditVideo, error) {
	redditVideos, err := ret.store.GetRedditVideos(ctx, selector, limit)
	if err != nil {
		return nil, err
	}

	return redditVideos, nil
}
