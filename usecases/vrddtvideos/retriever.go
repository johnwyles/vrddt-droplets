package vrddtvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewRetriever initializes the retrieval usecase with given store.
func NewRetriever(lg logger.Logger, store Store) *Retriever {
	return &Retriever{
		Logger: lg,
		store:  store,
	}
}

// Retriever provides retrieval related usecases.
type Retriever struct {
	logger.Logger

	store Store
}

// Get finds a vrddt video by its id.
func (ret *Retriever) GetByID(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error) {
	return ret.store.FindByID(ctx, id)
}

// Get finds a vrddt video by its md5 hash.
func (ret *Retriever) GetByMD5(ctx context.Context, md5 string) (*domain.VrddtVideo, error) {
	return ret.store.FindByMD5(ctx, md5)
}

// TODO
// Search finds all the vrddt videos matching the parameters in the query.
func (ret *Retriever) Search(ctx context.Context, limit int) ([]domain.VrddtVideo, error) {
	vrddtVideos, err := ret.store.FindAll(ctx, limit)
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
