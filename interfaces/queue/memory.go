package queue

import (
	"context"
	"fmt"

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
		return fmt.Errorf("connection type must be '%s' but is is '%s' instead", Client, m.connectionType)
	}

	select {
	case m.queue <- msg:
		return
	default:
		return fmt.Errorf("memory queue is full of %d maximum items trying to publish: %#v", m.maxSize, msg)
	}
}

func (m *memory) Pop(ctx context.Context) (msg interface{}, err error) {
	if m.connectionType != Consumer {
		return nil, fmt.Errorf("connection type must be '%s' but it is '%s' instead", Consumer, m.connectionType)
	}

	select {
	case item := <-m.queue:
		return item, nil
	default:
		return nil, fmt.Errorf("memory queue is empty of items")
	}
}
