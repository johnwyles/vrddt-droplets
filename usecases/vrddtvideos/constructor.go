package vrddtvideos

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// Constructor implements the publishing usecases.
type Constructor struct {
	logger.Logger

	store store.Store
}

// NewConstructor initializes the vrddt  usecase.
func NewConstructor(loggerHandle logger.Logger, store store.Store) *Constructor {
	return &Constructor{
		Logger: loggerHandle,

		store: store,
	}
}

// Create validates and persists the vrddt video into the store.
func (c *Constructor) Create(ctx context.Context, vrddtVideo *domain.VrddtVideo) (err error) {
	if err = vrddtVideo.Validate(); err != nil {
		return
	}

	_, err = c.store.GetVrddtVideo(ctx,
		store.Selector{
			"_id": vrddtVideo.ID,
		},
	)
	if err != nil {
		return errors.Conflict("VrddtVideo", vrddtVideo.ID.Hex())
	}

	err = c.store.CreateVrddtVideo(ctx, vrddtVideo)

	return
}
