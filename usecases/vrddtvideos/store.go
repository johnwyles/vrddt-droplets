package vrddtvideos

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
)

// Store implementation is responsible for managing persistance of vrddt videos.
type Store interface {
	Delete(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error)
	Exists(ctx context.Context, id bson.ObjectId) bool
	FindAll(ctx context.Context, limit int) ([]domain.VrddtVideo, error)
	FindByID(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error)
	FindByMD5(ctx context.Context, md5 string) (*domain.VrddtVideo, error)
	Save(ctx context.Context, vrddtVideo *domain.VrddtVideo) (*domain.VrddtVideo, error)
}
