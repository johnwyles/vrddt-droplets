package redditvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
)

// Store implementation is responsible for managing persistence of reddit
// videos.
type Store interface {
	Delete(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error)
	Exists(ctx context.Context, id bson.ObjectId) bool
	FindAll(ctx context.Context, limit int) ([]domain.RedditVideo, error)
	FindByID(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error)
	FindByURL(ctx context.Context, url string) (*domain.RedditVideo, error)
	Save(ctx context.Context, rv *domain.RedditVideo) (*domain.RedditVideo, error)
}
