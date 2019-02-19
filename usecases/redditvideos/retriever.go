package redditvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

// NewRetriever initializes an instance of Retriever with given store.
func NewRetriever(lg logger.Logger, store Store, vrddtStore vrddtvideos.Store) *Retriever {
	return &Retriever{
		Logger: lg,

		store:      store,
		vrddtStore: vrddtStore,
	}
}

// Retriever provides functions for retrieving user and user info.
type Retriever struct {
	logger.Logger

	store      Store
	vrddtStore vrddtvideos.Store
}

// GetByID finds a reddit video by id.
func (ret *Retriever) GetByID(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error) {
	redditVideo, err := ret.store.FindByID(ctx, id)
	if err != nil {
		ret.Debugf("failed to find reddit video with id '%s': %v", id.Hex(), err)
		return nil, err
	}

	return redditVideo, nil
}

// GetByURL finds a reddit video by url.
func (ret *Retriever) GetByURL(ctx context.Context, url string) (*domain.RedditVideo, error) {
	// TODO: If there is a way to do this entirely client-side we can save some time
	finalURL, err := domain.GetFinalURL(url)
	if err != nil {
		return nil, err
	}

	redditVideo, err := ret.store.FindByURL(ctx, finalURL)
	if err != nil {
		ret.Debugf("failed to find reddit video with url '%s': %v", url, err)
		return nil, err
	}

	return redditVideo, nil
}

// GetVrddtVideoByID will return the vrddt video by it's ID in the store.
func (ret *Retriever) GetVrddtVideoByID(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error) {
	vrddtVideo, err := ret.vrddtStore.FindByID(ctx, id)
	if err != nil {
		ret.Debugf("failed to find vrddt video with id '%s': %v", id.Hex(), err)
		return nil, err
	}

	return vrddtVideo, nil
}

// Search finds all the vrddt videos matching the parameters in the query.
// TODO: This is incomplete
func (ret *Retriever) Search(ctx context.Context, q []string, limit int) ([]domain.RedditVideo, error) {
	redditVideos, err := ret.store.FindAll(ctx, limit)
	if err != nil {
		return nil, err
	}

	return redditVideos, nil
}

// Query represents parameters for executing a search. Zero valued fields
// in the query will be ignored.
type Query struct {
	ID           bson.ObjectId `json:"id,omitempty"`
	URL          string        `json:"url,omitempty"`
	VrddtVideoID bson.ObjectId `json:"vrddt_video_id,omitempty"`
}
