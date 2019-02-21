package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

func main() {
	cfg := loadConfig()
	lg := logger.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)

	// cmdProcessWithInternalServices will setup the
	// "process-with-internal-services" sub-command
	var cmdProcessWithAPI = &cobra.Command{
		Use:   "converter",
		Short: "Use the internal services to process a reddit URL to vrddt video",
		Long:  "Use the internal services to process a reddit URL to vrddt video",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			Converter(cfg, lg)
		},
	}

	var rootCmd = &cobra.Command{Use: "vrddt-admin"}
	rootCmd.AddCommand(cmdProcessWithAPI)
	rootCmd.Execute()
}

type config struct {
	LogLevel    string
	LogFormat   string
	MongoURI    string
	RabbitMQURI string
}

func loadConfig() config {
	viper.SetDefault("VRDDT_LOG_LEVEL", "debug")
	viper.SetDefault("VRDDT_LOG_FORMAT", "text")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	return config{
		// application configuration
		LogLevel:  viper.GetString("VRDDT_LOG_LEVEL"),
		LogFormat: viper.GetString("VRDDT_LOG_FORMAT"),
	}
}
