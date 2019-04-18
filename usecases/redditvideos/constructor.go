package redditvideos

import (
	"context"
	"encoding/json"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/queue"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// Constructor provides functions for reddit video creation operations.
type Constructor struct {
	logger.Logger

	queue queue.Queue
	store store.Store
}

// NewConstructor initializes a Creation service object.
func NewConstructor(loggerHandle logger.Logger, queue queue.Queue, store store.Store) *Constructor {
	return &Constructor{
		Logger: loggerHandle,

		queue: queue,
		store: store,
	}
}

// Create creates a new reddit video in the system using the supplied
// RedditVideo object
func (cons *Constructor) Create(ctx context.Context, redditVideo *domain.RedditVideo) (err error) {
	if err = redditVideo.Validate(); err != nil {
		return
	}

	if err = redditVideo.SetFinalURL(); err != nil {
		return
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
func (cons *Constructor) Push(ctx context.Context, redditVideo *domain.RedditVideo) (err error) {
	if err = redditVideo.Validate(); err != nil {
		return
	}

	if err = redditVideo.SetFinalURL(); err != nil {
		return
	}

	cons.queue.MakeClient(ctx)

	message, err := json.Marshal(redditVideo)
	if err != nil {
		return
	}

	if err = cons.queue.Push(ctx, message); err != nil {
		cons.Errorf("Failed to push reddit video: %v", err)
		return
	}

	return
}
