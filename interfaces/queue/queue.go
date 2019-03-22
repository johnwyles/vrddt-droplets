package queue

import (
	"context"
)

const (
	// Client will set the queue type to be able to push data
	Client ConnectionType = iota

	// Consumer will set the queue type to be able to pop data
	Consumer ConnectionType = iota
)

// ConnectionType holds whether the connection is a Client or Consumer
type ConnectionType int

// Queue is the generic interface for a queue
type Queue interface {
	Cleanup(ctx context.Context) (err error)
	Init(ctx context.Context) (err error)
	MakeClient(ctx context.Context) (err error)
	MakeConsumer(ctx context.Context) (err error)
	Push(ctx context.Context, msg interface{}) (err error)
	Pop(ctx context.Context) (msg interface{}, err error)
}

func (c ConnectionType) String() string {
	return [...]string{"client", "consumer"}[c]
}
