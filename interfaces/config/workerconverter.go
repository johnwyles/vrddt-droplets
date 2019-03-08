package config

// WorkerConverterConfig holds all the different implementations for workers
// that perform video conversion
type WorkerConverterConfig struct {
	MaxErrors int
	Sleep     int
}
