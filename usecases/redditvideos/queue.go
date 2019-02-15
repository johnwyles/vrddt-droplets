package redditvideos

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
)

// Queue is the generic interface for a reddit video queue
type Queue interface {
	MakeClient(ctx context.Context) (err error)
	MakeConsumer(ctx context.Context) (err error)
	Push(ctx context.Context, rv *domain.RedditVideo) (err error)
	Pop(ctx context.Context) (rv *domain.RedditVideo, err error)
}
