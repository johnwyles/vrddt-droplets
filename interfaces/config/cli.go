package config

import (
	"time"
)

// CLIConfig stores the configuration for the client CLI
type CLIConfig struct {
	PollTime int
	Timeout  time.Duration
	APIURI   string
}
