package redditvideos

import (
	"context"
	"fmt"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/queue"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// Destructor implements the publishing usecases.
type Destructor struct {
	logger.Logger

	queue queue.Queue
	store store.Store
}

// NewDestructor initializes the vrddt  usecase.
func NewDestructor(loggerHandle logger.Logger, queue queue.Queue, store store.Store) *Destructor {
	return &Destructor{
		Logger: loggerHandle,

		queue: queue,
		store: store,
	}
}

// Delete removes the vrddt video from the store.
func (d *Destructor) Delete(ctx context.Context, id bson.ObjectId) (err error) {
	return d.store.DeleteVrddtVideo(
		ctx,
		store.Selector{
			"_id": id,
		},
	)
}

// Pop pops a reddit video off of the queue.
func (d *Destructor) Pop(ctx context.Context) (redditVideo *domain.RedditVideo, err error) {
	d.queue.MakeConsumer(ctx)
	result, err := d.queue.Pop(ctx)
	if err != nil {
		d.Debugf("failed to pop reddit video: %v", err)
		return nil, err
	}

	if redditVideo, ok := result.(*domain.RedditVideo); ok {
		return redditVideo, nil
	}

	return nil, errors.ResourceUnknown("result", fmt.Sprintf("#%v", result))
}
