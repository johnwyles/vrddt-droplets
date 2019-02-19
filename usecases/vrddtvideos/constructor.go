package vrddtvideos

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewCreator initializes the vrddt  usecase.
func NewConstructor(lg logger.Logger, store Store) *Constructor {
	return &Constructor{
		Logger: lg,

		store: store,
	}
}

// Creator implements the publishing usecases.
type Constructor struct {
	logger.Logger

	store Store
}

// Create validates and persists the vrddt video into the store.
func (c *Constructor) Create(ctx context.Context, vrddtVideo *domain.VrddtVideo) (*domain.VrddtVideo, error) {
	if err := vrddtVideo.Validate(); err != nil {
		return nil, err
	}

	if c.store.Exists(ctx, vrddtVideo.ID) {
		return nil, errors.Conflict("VrddtVideo", vrddtVideo.ID.Hex())
	}

	saved, err := c.store.Save(ctx, vrddtVideo)
	if err != nil {
		c.Warnf("failed to save vrddt video to the store: %+v", err)
	}

	return saved, nil
}
