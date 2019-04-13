package config

// WebConfig stores the configuration for the web server
type WebConfig struct {
	Address         string
	GracefulTimeout int
	PathPrefix      string
}
