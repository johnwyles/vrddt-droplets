package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// We tried and failed so many times...
//
// https://stackoverflow.com/questions/35583735/unmarshaling-into-an-interface-and-then-performing-type-assertion
// https://stackoverflow.com/questions/28254102/how-to-unmarshal-json-into-interface-in-go
// http://nesv.github.io/golang/2014/02/25/worker-queues-in-go.html
// https://programming.guide/go/type-assertion-switch.html

// rabbitmqConnection contains all the information about a RabbitMQ connection
type rabbitmqConnection struct {
	bindingKeyName string
	consumerID     string
	channel        *amqp.Channel
	connection     *amqp.Connection
	connectionType ConnectionType
	delivery       <-chan amqp.Delivery
	exchangeName   string
	log            logger.Logger
	queue          amqp.Queue
	queueName      string
	uri            string
}

// RabbitMQ initializes a RabbitMQ AMQP connection
func RabbitMQ(cfg *config.QueueRabbitMQConfig, loggerHandle logger.Logger) (queue Queue, err error) {
	loggerHandle.Debugf("RabbitMQ(cfg): %#v", cfg)

	queue = &rabbitmqConnection{
		bindingKeyName: cfg.BindingKeyName,
		connectionType: Client,
		exchangeName:   cfg.ExchangeName,
		log:            loggerHandle,
		queueName:      cfg.QueueName,
		uri:            cfg.URI,
	}

	return
}

// Cleanup will close the channel and connection
func (r *rabbitmqConnection) Cleanup(ctx context.Context) (err error) {
	if r.connection == nil || r.channel == nil {
		return fmt.Errorf("channel and connection have not be initialized")
	}

	if err = r.channel.Close(); err != nil {
		r.log.Errorf("Error closing channel: %s", err)
		return r.connection.Close()
	}
	r.log.Infof("Channel closed")

	if err = r.connection.Close(); err != nil {
		return fmt.Errorf("Error closing connection")
	}
	r.log.Infof("Connection closed")

	return
}

func (r *rabbitmqConnection) Init(ctx context.Context) (err error) {
	r.connection, err = amqp.Dial(r.uri)
	if err != nil {
		return
	}
	r.log.Infof("Opened connection")

	r.channel, err = r.connection.Channel()
	if err != nil {
		r.log.Errorf("Unable to open channel; %s", err)

		if err = r.connection.Close(); err != nil {
			return r.connection.Close()
		}

		return
	}
	r.log.Infof("Opened channel")

	if err = r.channel.ExchangeDeclare(
		r.exchangeName, // name of the exchange
		"direct",       // type
		true,           // durable
		false,          // delete when complete
		false,          // internal
		false,          // noWait
		nil,            // arguments
	); err != nil {
		r.log.Errorf("Error declaring exchange %s: %s", r.exchangeName, err)
		return r.Cleanup(ctx)
	}
	r.log.Infof("Exchanged declared: %s", r.exchangeName)

	r.queue, err = r.channel.QueueDeclare(
		r.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		r.log.Errorf("Error declaring queue %s: %s", r.queueName, err)
		return r.Cleanup(ctx)
	}
	r.log.Infof("Queue declared: %s", r.queueName)

	err = r.channel.QueueBind(
		r.queueName,      // name of the queue
		r.bindingKeyName, // bindingKey
		r.exchangeName,   // sourceExchange
		false,            // noWait
		nil,              // arguments
	)
	if err != nil {
		r.log.Errorf("Error binding queue '%s' to exchange '%s' using binding key name '%s': %s", r.queueName, r.exchangeName, r.bindingKeyName, err)
		return r.Cleanup(ctx)
	}
	r.log.Infof("Queue '%s' bound to exchange '%s' using binding key name '%s'", r.queueName, r.exchangeName, r.bindingKeyName)

	return
}

// MakeClient is implemented but does nothing as there is no additional steps
// required by RabbitMQ to make the connection a client vs a consumer
func (r *rabbitmqConnection) MakeClient(ctx context.Context) (err error) {
	r.connectionType = Client
	return
}

// MakeConsumer will setup whatever is necessary to pop messages and
// will set the Delivery channel
func (r *rabbitmqConnection) MakeConsumer(ctx context.Context) (err error) {
	if r.consumerID != "" && r.connectionType == Consumer {
		r.log.Debugf("Channel was already made a consumer: %s", r.consumerID)
		return
	}

	// Setup a new AMQP consumer UUID
	uuid, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("Error generating new random UUID: %s", err)
	}
	r.consumerID = uuid.String()

	if err = r.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("Error setting QoS level: %s", err)
	}

	r.delivery, err = r.channel.Consume(
		r.queueName,  // queue
		r.consumerID, // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return
	}

	r.connectionType = Consumer
	r.log.Debugf("Channel consumer created: %s", r.consumerID)

	return
}

// Pop will pull off a Reddit video struct from the queue
func (r *rabbitmqConnection) Pop(ctx context.Context) (msg interface{}, err error) {
	if r.connectionType != Consumer {
		return nil, fmt.Errorf("Connection type must be '%s' but it is '%s' instead", Consumer, r.connectionType)
	}

	data := <-r.delivery
	data.Ack(false)
	msg = data.Body
	r.log.Infof("Popped message: %#v", string(data.Body))

	return
}

// Push will put a Reddit video struct onto the queue
func (r *rabbitmqConnection) Push(ctx context.Context, msg interface{}) (err error) {
	if r.connectionType != Client {
		return fmt.Errorf("Connection type must be '%s' but it is '%s' instead", Client, r.connectionType)
	}

	if _, ok := msg.([]byte); !ok {
		return fmt.Errorf("RabbitMQ Push(msg), msg must be of type []byte")
	}

	byteMsg := msg.([]byte)
	err = r.channel.Publish(
		"",          // exchange
		r.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			Body:         byteMsg,
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Headers:      amqp.Table{},
			Priority:     0,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("Error publishing message: %#v", string(byteMsg))
	}

	r.log.Infof("Pushed message: %#v", string(byteMsg))

	return
}
