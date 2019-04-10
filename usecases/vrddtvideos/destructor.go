package vrddtvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewDestructor initializes the vrddt  usecase.
func NewDestructor(lg logger.Logger, store store.Store) (dstr *Destructor) {
	return &Destructor{
		Logger: lg,

		store: store,
	}
}

// Destructor implements the publishing usecases.
type Destructor struct {
	logger.Logger

	store store.Store
}

// Delete removes the vrddt video from the store.
func (d *Destructor) Delete(ctx context.Context, id bson.ObjectId) (err error) {
	err = d.store.DeleteVrddtVideo(
		ctx,
		store.Selector{
			"_id": id,
		},
	)

	return
}
