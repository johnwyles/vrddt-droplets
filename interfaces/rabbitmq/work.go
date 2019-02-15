package rabbitmq

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

const (
	Client   ConnectionType = iota
	Consumer ConnectionType = iota
)

type ConnectionType int

// NewWorkQueue initializes a work queue with the given queue handle.
func NewWorkQueue(c *amqp.Connection) *WorkQueue {

	return &WorkQueue{
		connection: c,
	}
}

// WorkQueue provides functions for queueing work entities in RabbitMQ.
type WorkQueue struct {
	connection     *amqp.Connection
	connectionType ConnectionType
}

// TODO
// MakeClient is implemented but does nothing as there is no additional steps
// required by RabbitMQ to make the connection a client vs a consumer
func (w *WorkQueue) MakeClient() {
	w.connectionType = Client
	return
}

// TODO
// MakeConsumer is implemented but does nothing as there is no additional steps
// required by RabbitMQ to make the connection a client vs a consumer
func (w *WorkQueue) MakeConsumer() (err error) {
	// Setup a new AMQP consumer UUID
	uuid, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("error generating new random uuid")
	}
	consmerID := uuid.String()

	if err = w.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("error setting QoS level")
	}

	w.delivery, err = w.channel.Consume(
		w.queueName, // queue
		consmerID,   // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return
	}

	w.connectionType = Consumer

	return
}

func (c ConnectionType) String() string {
	return [...]string{"client", "consumer"}[c]
}
