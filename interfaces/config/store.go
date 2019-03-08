package config

type StoreType int

const (
	StoreConfigMongo   StoreType = iota
	StoreConfigeMemory StoreType = iota
)

// StoreConfig holds all the different implementations for a persistence store service
type StoreConfig struct {
	Mongo  StoreMongoConfig
	Memory StoreMemoryConfig
	Type   StoreType
}

// String will return the string representation of the iota
func (s StoreType) String() string {
	return [...]string{"memory", "rabbitmq"}[s]
}
