package config

// APIConfig stores the configuration for the API server
type APIConfig struct {
	Address         string
	GracefulTimeout int
}
