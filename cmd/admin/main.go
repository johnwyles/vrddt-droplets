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
	lg.Infof("config: %#v", cfg)

	// cmdProcessWithInternalServices will setup the
	// "process-with-internal-services" sub-command
	var cmdProcessWithInternalServices = &cobra.Command{
		Use:   "process-with-internal-services",
		Short: "Use the internal store and queue to process a reddit video",
		Long: `Use the internal store and queue to process a reddit video
 rather than hitting the public API`,
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ProcessWithInternalServices(cfg, lg)
		},
	}
	cmdProcessWithInternalServices.Flags().StringVarP(
		&cfg.RedditURL,
		"reddit-url",
		"r",
		"",
		"Reddit URL",
	)
	cmdProcessWithInternalServices.MarkFlagRequired("reddit-url")
	viper.BindPFlag("VRDDT_REDDIT_URL", cmdProcessWithInternalServices.PersistentFlags().Lookup("reddit-url"))

	var rootCmd = &cobra.Command{Use: "vrddt-admin"}
	rootCmd.Flags().StringVarP(
		&cfg.MongoURI,
		"mongodb-uri",
		"m",
		"mongodb://admin:password@localhost:27017/vrddt",
		"MongoDB URI with credentials, host, port, and database",
	)
	cmdProcessWithInternalServices.MarkFlagRequired("mongo-uri")
	viper.BindPFlag("VRDDT_MONGODB_URI", rootCmd.PersistentFlags().Lookup("mongodb-uri"))

	rootCmd.Flags().StringVarP(
		&cfg.RabbitMQURI,
		"rabbitmq-uri",
		"m",
		"amqp://admin:password@localhost:5672t",
		"RabbitMQ URI with credentials, host, port",
	)
	cmdProcessWithInternalServices.MarkFlagRequired("rabbitmq-uri")
	viper.BindPFlag("VRDDT_RABBITMQ_URI", rootCmd.PersistentFlags().Lookup("rabbitmq-uri"))

	rootCmd.AddCommand(cmdProcessWithInternalServices)
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
}

func loadConfig() config {
	viper.SetDefault("VRDDT_LOG_LEVEL", "debug")
	viper.SetDefault("VRDDT_LOG_FORMAT", "text")
	viper.SetDefault("VRDDT_MONGO_URI", "mongodb://admin:password@localhost:27017/vrddt")
	viper.SetDefault("VRDDT_RABBITMQ_URI", "amqp://admin:password@localhost:5672")

	viper.ReadInConfig()
	viper.AutomaticEnv()

	return config{
		// application configuration
		LogLevel:    viper.GetString("VRDDT_LOG_LEVEL"),
		LogFormat:   viper.GetString("VRDDT_LOG_FORMAT"),
		MongoURI:    viper.GetString("VRDDT_MONGO_URI"),
		RabbitMQURI: viper.GetString("VRDDT_RABBITMQ_URI"),
	}
}
