package config

// APIConfig stores the configuration for the API server
type APIConfig struct {
	Address         string
	CertFile        string
	GracefulTimeout int
	KeyFile         string
}
