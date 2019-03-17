package config

import (
	"time"
)

// WebConfig stores the configuration for the web server
type WebConfig struct {
	Address         string
	GracefulTimeout time.Duration
}
