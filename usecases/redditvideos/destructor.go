package redditvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewDestructor initializes the vrddt  usecase.
func NewDestructor(lg logger.Logger, queue Queue, store Store) *Destructor {
	return &Destructor{
		Logger: lg,
		queue:  queue,
		store:  store,
	}
}

// Destructor implements the publishing usecases.
type Destructor struct {
	logger.Logger

	queue Queue
	store Store
}

// Delete removes the vrddt video from the store.
func (d *Destructor) Delete(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error) {
	return d.store.Delete(ctx, id)
}

// Pop pops a reddit video off of the queue.
func (d *Destructor) Pop(ctx context.Context) (*domain.RedditVideo, error) {
	d.queue.MakeConsumer(ctx)
	redditVideo, err := d.queue.Pop(ctx)
	if err != nil {
		d.Debugf("failed to pop reddit video: %v", err)
		return nil, err
	}

	return redditVideo, nil
}
