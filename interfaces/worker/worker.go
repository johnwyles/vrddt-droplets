package worker

import (
	"context"
)

// Worker is the generic interface for a worker process store
type Worker interface {
	CompleteWork(ctx context.Context) (err error)
	DoWork(ctx context.Context) (err error)
	GetWork(ctx context.Context) (err error)
	Init(ctx context.Context) (err error)
}
