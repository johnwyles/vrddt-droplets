package config

// Config stores the child configurations
type Config struct {
	Converter ConverterConfig
	Storage   StorageConfig
	Log       LogConfig
	Queue     QueueConfig
	Store     StoreConfig
	Worker    WorkerConfig
}
