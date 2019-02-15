package redditvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewRetriever initializes an instance of Retriever with given store.
func NewRetriever(lg logger.Logger, store Store) *Retriever {
	return &Retriever{
		Logger: lg,
		store:  store,
	}
}

// Retriever provides functions for retrieving user and user info.
type Retriever struct {
	logger.Logger

	store Store
	queue Queue
}

// Get finds a reddit video by id.
func (ret *Retriever) GetByID(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error) {
	redditVideo, err := ret.store.FindByID(ctx, id)
	if err != nil {
		ret.Debugf("failed to find reddit video with id '%s': %v", id.Hex(), err)
		return nil, err
	}

	return redditVideo, nil
}

// Get finds a reddit video by url.
func (ret *Retriever) GetByURL(ctx context.Context, url string) (*domain.RedditVideo, error) {
	redditVideo, err := ret.store.FindByURL(ctx, url)
	if err != nil {
		ret.Debugf("failed to find reddit video with url '%s': %v", url, err)
		return nil, err
	}

	return redditVideo, nil
}

// TODO
// Search finds all the vrddt videos matching the parameters in the query.
func (ret *Retriever) Search(ctx context.Context, limit int) ([]domain.RedditVideo, error) {
	redditVideos, err := ret.store.FindAll(ctx, limit)
	if err != nil {
		return nil, err
	}

	return redditVideos, nil
}

// Query represents parameters for executing a search. Zero valued fields
// in the query will be ignored.
type Query struct {
	ID  bson.ObjectId `json:"id,omitempty"`
	URL string        `json:"url,omitempty"`
}
