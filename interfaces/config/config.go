package config

// Config stores the child configurations
type Config struct {
	API       APIConfig
	CLI       CLIConfig
	Converter ConverterConfig
	Storage   StorageConfig
	Log       LogConfig
	Queue     QueueConfig
	Store     StoreConfig
	Web       WebConfig
	Worker    WorkerConfig
}
