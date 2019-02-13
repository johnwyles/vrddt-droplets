package vrddtvideos

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/domain"
)

// Store implementation is responsible for managing persistance of vrddt videos.
type Store interface {
	Get(ctx context.Context, name string) (*domain.VrddtVideo, error)
	Exists(ctx context.Context, name string) bool
	Save(ctx context.Context, post domain.VrddtVideo) (*domain.VrddtVideo, error)
	Delete(ctx context.Context, name string) (*domain.VrddtVideo, error)
}

// userVerifier is responsible for verifying existence of a user.
type userVerifier interface {
	Exists(ctx context.Context, name string) bool
}
