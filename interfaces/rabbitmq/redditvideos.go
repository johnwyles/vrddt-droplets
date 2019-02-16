package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"

	"github.com/johnwyles/vrddt-droplets/domain"
)

const (
	// Client is telling all queries that we are a client of RabbitMQ
	Client ConnectionType = iota

	// Consumer is telling all queries that we are a consumer of RabbitMQ
	Consumer ConnectionType = iota
)

// ConnectionType is the representation in iota of whether we are a client
// or a consumer process
type ConnectionType int

// TODO: Don't hardcode the below in NewWorkQueue()

// NewWorkQueue initializes a work queue with the given queue handle.
func NewWorkQueue(c *amqp.Connection) *WorkQueue {
	return &WorkQueue{
		bindingKeyName: "vrddt-bindingkey-converter",
		connection:     c,
		exchangeName:   "vrddt-exchange-converter",
		queueName:      "vrddt-queue-converter",
	}
}

// WorkQueue provides functions for queueing work entities in RabbitMQ.
type WorkQueue struct {
	delivery       <-chan amqp.Delivery
	channel        *amqp.Channel
	connection     *amqp.Connection
	connectionType ConnectionType
	consmerID      string
	bindingKeyName string
	exchangeName   string
	queue          amqp.Queue
	queueName      string
}

// MakeClient is implemented but does nothing as there is no additional steps
// required by RabbitMQ to make the connection a client vs a consumer
func (w *WorkQueue) MakeClient(ctx context.Context) error {
	w.connectionType = Client
	err := w.init(ctx)

	return err
}

// MakeConsumer is implemented but does nothing as there is no additional steps
// required by RabbitMQ to make the connection a client vs a consumer
func (w *WorkQueue) MakeConsumer(ctx context.Context) error {
	if err := w.init(ctx); err != nil {
		return err
	}

	// Setup a new AMQP consumer UUID
	uuid, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("error generating new random uuid")
	}
	w.consmerID = uuid.String()

	if err = w.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("error setting QoS level")
	}

	w.delivery, err = w.channel.Consume(
		w.queueName, // queue
		w.consmerID, // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return err
	}

	w.connectionType = Consumer

	return nil
}

// Pop will pull off a Reddit video struct from the queue
func (w *WorkQueue) Pop(ctx context.Context) (*domain.RedditVideo, error) {
	if w.connectionType != Consumer {
		return nil, fmt.Errorf("connection type must be '%s' but it is '%s' instead", Consumer, w.connectionType)
	}

	data := <-w.delivery
	data.Ack(false)
	msg := data.Body

	rv := &domain.RedditVideo{}
	if err := json.Unmarshal(msg, rv); err != nil {
		return nil, err
	}

	return rv, nil
}

// Push will put a Reddit video struct onto the queue
func (w *WorkQueue) Push(ctx context.Context, rv *domain.RedditVideo) (err error) {
	if w.connectionType != Client {
		return fmt.Errorf("connection type must be '%s' but it is '%s' instead", Client, w.connectionType)
	}

	rvJSON, err := json.Marshal(rv)
	if err != nil {
		return fmt.Errorf("rabbitmq Push(msg), msg must be of type []byte")
	}

	err = w.channel.Publish(
		"",          // exchange
		w.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			Body:         rvJSON,
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Headers:      amqp.Table{},
			Priority:     0,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("error publishing message: %#v", string(rvJSON))
	}

	return
}

// init initializes the connection by setting up all the particular details
// about a RabbitMQ queue
func (w *WorkQueue) init(ctx context.Context) error {
	var err error

	w.channel, err = w.connection.Channel()
	if err != nil {
		return err
	}

	if err = w.channel.ExchangeDeclare(
		w.exchangeName, // name of the exchange
		"direct",       // type
		true,           // durable
		false,          // delete when complete
		false,          // internal
		false,          // noWait
		nil,            // arguments
	); err != nil {
		return err
	}

	w.queue, err = w.channel.QueueDeclare(
		w.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return err
	}

	err = w.channel.QueueBind(
		w.queueName,      // name of the queue
		w.bindingKeyName, // bindingKey
		w.exchangeName,   // sourceExchange
		false,            // noWait
		nil,              // arguments
	)
	if err != nil {
		return err
	}

	return nil
}

func (c ConnectionType) String() string {
	return [...]string{"client", "consumer"}[c]
}
