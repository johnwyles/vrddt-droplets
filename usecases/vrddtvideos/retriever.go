package vrddtvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewRetriever initializes the retrieval usecase with given store.
func NewRetriever(lg logger.Logger, store store.Store) *Retriever {
	return &Retriever{
		Logger: lg,

		store: store,
	}
}

// Retriever provides retrieval related usecases.
type Retriever struct {
	logger.Logger

	store store.Store
}

// GetByID finds a vrddt video by its id.
func (ret *Retriever) GetByID(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error) {
	return ret.store.GetVrddtVideo(
		ctx,
		store.Selector{
			"_id": id,
		},
	)
}

// GetByMD5 finds a vrddt video by its md5 hash.
func (ret *Retriever) GetByMD5(ctx context.Context, md5 string) (*domain.VrddtVideo, error) {
	return ret.store.GetVrddtVideo(
		ctx,
		store.Selector{
			"md5": md5,
		},
	)
}

// Search finds all the vrddt videos matching the parameters in the query.
func (ret *Retriever) Search(ctx context.Context, selector store.Selector, limit int) ([]*domain.VrddtVideo, error) {
	vrddtVideos, err := ret.store.GetVrddtVideos(
		ctx,
		selector,
		limit,
	)
	if err != nil {
		return nil, err
	}

	return vrddtVideos, nil
}

// Query represents parameters for executing a search. Zero valued fields
// in the query will be ignored.
type Query struct {
	ID  bson.ObjectId `json:"id,omitempty"`
	MD5 string        `json:"md5"`
}
