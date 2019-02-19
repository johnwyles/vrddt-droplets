package vrddtvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewCreator initializes the vrddt  usecase.
func NewDestructor(lg logger.Logger, store Store) *Destructor {
	return &Destructor{
		Logger: lg,

		store: store,
	}
}

// Creator implements the publishing usecases.
type Destructor struct {
	logger.Logger

	store Store
}

// Delete removes the vrddt video from the store.
func (d *Destructor) Delete(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error) {
	return d.store.Delete(ctx, id)
}
