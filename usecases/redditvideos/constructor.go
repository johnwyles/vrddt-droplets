package redditvideos

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/queue"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewConstructor initializes a Creation service object.
func NewConstructor(lg logger.Logger, queue queue.Queue, store store.Store) *Constructor {
	return &Constructor{
		Logger: lg,
		queue:  queue,
		store:  store,
	}
}

// Constructor provides functions for reddit video creation operations.
type Constructor struct {
	logger.Logger

	queue queue.Queue
	store store.Store
}

// Create creates a new reddit video in the system using the supplied
// RedditVideo object
func (cons *Constructor) Create(ctx context.Context, redditVideo *domain.RedditVideo) (err error) {
	if err := redditVideo.Validate(); err != nil {
		return err
	}

	if err := redditVideo.SetFinalURL(); err != nil {
		return err
	}

	redditVideo, err = cons.store.GetRedditVideo(
		ctx, store.Selector{
			"_id": redditVideo.ID,
		},
	)
	if err != nil {
		return errors.Conflict("ID", redditVideo.ID.Hex())
	}

	return cons.store.CreateRedditVideo(ctx, redditVideo)
}

// Push pops a reddit video from the queue.
func (cons *Constructor) Push(ctx context.Context, redditVideo *domain.RedditVideo) error {
	if err := redditVideo.Validate(); err != nil {
		return err
	}

	if err := redditVideo.SetFinalURL(); err != nil {
		return err
	}

	cons.queue.MakeClient(ctx)
	if err := cons.queue.Push(ctx, redditVideo); err != nil {
		cons.Debugf("failed to pop reddit video: %v", err)
		return err
	}

	return nil
}
