package config

// StoreType is the type of store
type StoreType int

const (
	// StoreConfigeMemory is the type reserved for a memory store
	StoreConfigeMemory StoreType = iota

	// StoreConfigMongo is the type reserved for a Mongo store
	StoreConfigMongo StoreType = iota
)

// StoreConfig holds all the different implementations for a persistence store service
type StoreConfig struct {
	Memory StoreMemoryConfig
	Mongo  StoreMongoConfig
	Type   StoreType
}

// String will return the string representation of the iota
func (s StoreType) String() string {
	return [...]string{"memory", "mongo"}[s]
}
