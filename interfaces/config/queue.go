package config

// QueueType is the type of queue
type QueueType int

const (
	// QueueConfigMemory is the type reserved for a Memory queue type
	QueueConfigMemory QueueType = iota

	// QueueConfigRabbitMQ is the type reserved for a RabbitMQ queue type
	QueueConfigRabbitMQ QueueType = iota
)

// QueueConfig holds all the different implentations for a queue service
type QueueConfig struct {
	Memory   QueueMemoryConfig
	RabbitMQ QueueRabbitMQConfig
	Type     QueueType
}

// String will return the string representation of the iota
func (q QueueType) String() string {
	return [...]string{"memory", "rabbitmq"}[q]
}
