package rabbitmq

import (
	"github.com/streadway/amqp"
)

// Connect to a RabbitMQ instance located by rabbitmq-uri using the `amqp` package
func Connect(uri string) (*amqp.Connection, func() error, error) {
	_, err := amqp.ParseURI(uri)
	if err != nil {
		return nil, doNothing, err
	}

	session, err := amqp.Dial(uri)
	if err != nil {
		return nil, doNothing, err
	}

	return session, session.Close, nil
}

func doNothing() error {
	return nil
}
