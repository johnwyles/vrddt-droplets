package config

// QueueRabbitMQConfig stores the configuration for the RabbitMQ queue
type QueueRabbitMQConfig struct {
	BindingKeyName string
	ExchangeName   string
	QueueName      string
	URI            string
}
