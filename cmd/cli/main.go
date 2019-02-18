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
		Use:   "process-with-api",
		Short: "Use the public API to process a reddit video",
		Long:  "Use the public API to process a reddit video",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ProcessWithAPI(cfg, lg)
		},
	}

	cmdProcessWithAPI.Flags().StringVarP(
		&cfg.RedditURL,
		"api-uri",
		"a",
		"",
		"vrddt API URI",
	)
	cmdProcessWithAPI.MarkFlagRequired("api-uri")
	viper.BindPFlag("VRDDT_API_URI", cmdProcessWithAPI.PersistentFlags().Lookup("api-uri"))

	cmdProcessWithAPI.Flags().StringVarP(
		&cfg.RedditURL,
		"reddit-url",
		"r",
		"",
		"Reddit URL",
	)
	cmdProcessWithAPI.MarkFlagRequired("reddit-url")
	viper.BindPFlag("VRDDT_REDDIT_URL", cmdProcessWithAPI.PersistentFlags().Lookup("reddit-url"))

	cmdProcessWithAPI.Flags().IntVarP(
		&cfg.Timeout,
		"timeout",
		"t",
		60,
		"Timeout (in seconds) before giving up on the operation",
	)
	viper.BindPFlag("VRDDT_TIMEOUT", cmdProcessWithAPI.PersistentFlags().Lookup("timeout"))

	cmdProcessWithAPI.Flags().IntVarP(
		&cfg.PollTime,
		"poll-time",
		"p",
		500,
		"Poll time frequency (in milliseconds) to poll the API to see if the vrddt video is ready",
	)
	viper.BindPFlag("VRDDT_POLL_TIME", cmdProcessWithAPI.PersistentFlags().Lookup("poll-time"))

	var rootCmd = &cobra.Command{Use: "vrddt-admin"}
	rootCmd.AddCommand(cmdProcessWithAPI)
	rootCmd.Execute()
}

type config struct {
	LogLevel    string
	LogFormat   string
	MongoURI    string
	PollTime    int
	RabbitMQURI string
	RedditURL   string
	Timeout     int
	VrddtAPIURI string
}

func loadConfig() config {
	viper.SetDefault("VRDDT_LOG_LEVEL", "debug")
	viper.SetDefault("VRDDT_LOG_FORMAT", "text")
	viper.SetDefault("VRDDT_API_URI", "http://localhost:8080/api")

	viper.ReadInConfig()
	viper.AutomaticEnv()

	return config{
		// application configuration
		LogLevel:    viper.GetString("VRDDT_LOG_LEVEL"),
		LogFormat:   viper.GetString("VRDDT_LOG_FORMAT"),
		VrddtAPIURI: viper.GetString("VRDDT_API_URI"),
	}
}
