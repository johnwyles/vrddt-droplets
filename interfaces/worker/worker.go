package worker

import (
	"context"
)

// Worker is the generic interface for a worker process store
type Worker interface {
	GetWork(ctx *context.Context) error
	DoWork(ctx *context.Context) error
	CompleteWork(ctx *context.Context) error
}
