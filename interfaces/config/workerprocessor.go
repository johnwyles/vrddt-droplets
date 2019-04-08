package config

// WorkerProcessorConfig holds all the configuration for the worker that performs
// video conversion
type WorkerProcessorConfig struct {
	MaxErrors int
	Sleep     int
}
