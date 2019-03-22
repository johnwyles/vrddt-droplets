package vrddtvideos

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewConstructor initializes the vrddt  usecase.
func NewConstructor(lg logger.Logger, store store.Store) *Constructor {
	return &Constructor{
		Logger: lg,

		store: store,
	}
}

// Constructor implements the publishing usecases.
type Constructor struct {
	logger.Logger

	store store.Store
}

// Create validates and persists the vrddt video into the store.
func (c *Constructor) Create(ctx context.Context, vrddtVideo *domain.VrddtVideo) (resultVideo *domain.VrddtVideo, err error) {
	if err = vrddtVideo.Validate(); err != nil {
		return nil, err
	}

	resultVideo, err = c.store.GetVrddtVideo(ctx,
		store.Selector{
			"_id": vrddtVideo.ID,
		},
	)
	if err != nil {
		return nil, errors.Conflict("VrddtVideo", vrddtVideo.ID.Hex())
	}

	err = c.store.CreateVrddtVideo(ctx, vrddtVideo)

	return
}
