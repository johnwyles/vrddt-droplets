package redditvideos

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
)

// Store implementation is responsible for managing persistence of
// reddit videos.
type Store interface {
	Exists(ctx context.Context, name string) bool
	Save(ctx context.Context, user domain.RedditVideo) (*domain.RedditVideo, error)
	FindByName(ctx context.Context, name string) (*domain.RedditVideo, error)
	FindAll(ctx context.Context, tags []string, limit int) ([]domain.RedditVideo, error)
}
