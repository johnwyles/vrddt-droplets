package config

type QueueType int

const (
	QueueConfigRabbitMQ QueueType = iota
	QueueConfigMemory   QueueType = iota
)

// QueueConfig holds all the different implentations for a queue service
type QueueConfig struct {
	RabbitMQ QueueRabbitMQConfig
	Memory   QueueMemoryConfig
	Type     QueueType
}

// String will return the string representation of the iota
func (q QueueType) String() string {
	return [...]string{"memory", "rabbitmq"}[q]
}
