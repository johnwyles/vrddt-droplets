package redditvideos

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// NewConstructor initializes a Creation service object.
func NewConstructor(lg logger.Logger, queue Queue, store Store) *Constructor {
	return &Constructor{
		Logger: lg,
		queue:  queue,
		store:  store,
	}
}

// Constructor provides functions for reddit video creation operations.
type Constructor struct {
	logger.Logger

	queue Queue
	store Store
}

// Create creates a new reddit video in the system using the supplied
// RedditVideo object
func (cons *Constructor) Create(ctx context.Context, redditVideo *domain.RedditVideo) (*domain.RedditVideo, error) {
	if err := redditVideo.Validate(); err != nil {
		return nil, err
	}

	if err := redditVideo.SetFinalURL(); err != nil {
		return nil, err
	}

	if cons.store.Exists(ctx, redditVideo.ID) {
		return nil, errors.Conflict("ID", redditVideo.ID.Hex())
	}

	saved, err := cons.store.Save(ctx, redditVideo)
	if err != nil {
		cons.Logger.Warnf("failed to save user object: %v", err)
		return nil, err
	}

	return saved, nil
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
