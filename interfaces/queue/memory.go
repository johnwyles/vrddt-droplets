package queue

import (
	"context"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

type memory struct {
	maxSize        int
	queue          chan interface{}
	log            logger.Logger
	connectionType ConnectionType
}

// Memory is the contructor for a new memory based queue
func Memory(cfg *config.QueueMemoryConfig, loggerHandle logger.Logger) (queue Queue, err error) {
	loggerHandle.Debugf("Memory(cfg): %#v", cfg)

	queue = &memory{
		log:     loggerHandle,
		maxSize: cfg.MaxSize,
	}

	return
}

func (m *memory) Cleanup(ctx context.Context) (err error) {
	close(m.queue)
	return
}

func (m *memory) Init(ctx context.Context) (err error) {
	m.queue = make(chan interface{}, m.maxSize)
	return
}

func (m *memory) MakeClient(ctx context.Context) (err error) {
	m.connectionType = Client
	return
}

func (m *memory) MakeConsumer(ctx context.Context) (err error) {
	m.connectionType = Consumer
	return
}

func (m *memory) Push(ctx context.Context, msg interface{}) (err error) {
	if m.connectionType != Client {
		return errors.Conflict("Connection type", m.connectionType.String())
	}

	select {
	case m.queue <- msg:
		return
	default:
		return errors.ResourceLimit("memory", m.maxSize)
	}
}

func (m *memory) Pop(ctx context.Context) (msg interface{}, err error) {
	if m.connectionType != Consumer {
		return nil, errors.ConnectionFailure("memory", m.connectionType.String())
	}

	select {
	case item := <-m.queue:
		return item, nil
	default:
		return nil, errors.ResourceLimit("memory", 0)
	}
}
