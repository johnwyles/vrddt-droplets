package storage

import (
	"context"
)

// Storage is the generic interface for a file store
type Storage interface {
	Attributes(ctx context.Context, remotePath string) (attrs interface{}, err error)
	Cleanup(ctx context.Context) (err error)
	Delete(ctx context.Context, remotePath string) (err error)
	Download(ctx context.Context, remotePath string, localPath string) (err error)
	Init(ctx context.Context) (err error)
	GetLocation(ctx context.Context, remotePath string) (url string, err error)
	List(ctx context.Context, remotePath string) (files []interface{}, err error)
	Upload(ctx context.Context, localPath string, remotePath string) (err error)
}
