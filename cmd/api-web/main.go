package main

import (
	// "context"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/johnwyles/vrddt-droplets/interfaces/mongo"
	"github.com/johnwyles/vrddt-droplets/interfaces/rabbitmq"
	"github.com/johnwyles/vrddt-droplets/interfaces/rest"
	"github.com/johnwyles/vrddt-droplets/interfaces/web"
	"github.com/johnwyles/vrddt-droplets/pkg/graceful"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/pkg/middlewares"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	"github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
)

func main() {
	cfg := loadConfig()
	lg := logger.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)

	lg.Debugf("setting up rest api service")

	db, closeMongoSession, err := mongo.Connect(cfg.MongoURI, true)
	if err != nil {
		lg.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer closeMongoSession()

	redditVideoStore := mongo.NewRedditVideoStore(db)
	vrddtVideoStore := mongo.NewVrddtVideoStore(db)

	q, closeRabbitMQSession, err := rabbitmq.Connect(cfg.RabbitMQURI)
	if err != nil {
		lg.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer closeRabbitMQSession()

	redditVideoWorkQueue := rabbitmq.NewRedditVideoWorkQueue(q)

	vrddtVideoConstructor := vrddtvideos.NewConstructor(lg, vrddtVideoStore)
	vrddtVideoDestructor := vrddtvideos.NewDestructor(lg, vrddtVideoStore)
	vrddtVideoRetriever := vrddtvideos.NewRetriever(lg, vrddtVideoStore)

	redditVideoConstructor := redditvideos.NewConstructor(lg, redditVideoWorkQueue, redditVideoStore)
	redditVideoDestructor := redditvideos.NewDestructor(lg, redditVideoWorkQueue, redditVideoStore)
	redditVideoRetriever := redditvideos.NewRetriever(lg, redditVideoStore, vrddtVideoStore)

	restHandler := rest.New(
		lg,
		redditVideoConstructor,
		redditVideoDestructor,
		redditVideoRetriever,
		vrddtVideoConstructor,
		vrddtVideoDestructor,
		vrddtVideoRetriever,
	)
	webHandler, err := web.New(lg, web.Config{
		TemplateDir: cfg.TemplateDir,
		StaticDir:   cfg.StaticDir,
	})
	if err != nil {
		lg.Fatalf("failed to setup web handler: %v", err)
	}

	srv := setupServer(cfg, lg, webHandler, restHandler)
	lg.Infof("listening for requests on :8080...")
	if err := srv.ListenAndServe(); err != nil {
		lg.Fatalf("http server exited: %s", err)
	}
}

func setupServer(cfg config, lg logger.Logger, web http.Handler, rest http.Handler) *graceful.Server {
	router := mux.NewRouter()
	router.PathPrefix("/api").Handler(http.StripPrefix("/api", rest))
	router.PathPrefix("/").Handler(web)

	handler := middlewares.WithRequestLogging(lg, router)
	handler = middlewares.WithRecovery(lg, handler)

	srv := graceful.NewServer(handler, cfg.GracefulTimeout, os.Interrupt)
	srv.Log = lg.Errorf
	srv.Addr = cfg.Addr
	return srv
}

type config struct {
	Addr            string
	LogLevel        string
	LogFormat       string
	StaticDir       string
	TemplateDir     string
	GracefulTimeout time.Duration
	MongoURI        string
	RabbitMQURI     string
}

func loadConfig() config {
	viper.SetDefault("ADDR", ":8080")
	viper.SetDefault("GRACEFUL_TIMEOUT", 20*time.Second)
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_FORMAT", "text")
	viper.SetDefault("VRDDT_MONGO_URI", "mongodb://admin:password@localhost:27017/vrddt")
	viper.SetDefault("VRDDT_RABBITMQ_URI", "amqp://admin:password@localhost:5672")
	viper.SetDefault("STATIC_DIR", "../../web/static/")
	viper.SetDefault("TEMPLATE_DIR", "../../web/templates/")

	viper.ReadInConfig()
	viper.AutomaticEnv()

	return config{
		// application configuration
		Addr:            viper.GetString("ADDR"),
		GracefulTimeout: viper.GetDuration("GRACEFUL_TIMEOUT"),
		LogLevel:        viper.GetString("LOG_LEVEL"),
		LogFormat:       viper.GetString("LOG_FORMAT"),
		MongoURI:        viper.GetString("VRDDT_MONGO_URI"),
		RabbitMQURI:     viper.GetString("VRDDT_RABBITMQ_URI"),
		StaticDir:       viper.GetString("STATIC_DIR"),
		TemplateDir:     viper.GetString("TEMPLATE_DIR"),
	}
}