package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/johnwyles/vrddt-droplets/interfaces/mongo"
	"github.com/johnwyles/vrddt-droplets/interfaces/rest"
	"github.com/johnwyles/vrddt-droplets/interfaces/web"
	"github.com/johnwyles/vrddt-droplets/pkg/graceful"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/pkg/middlewares"
	"github.com/johnwyles/vrddt-droplets/usecases/redditvideos"
	"github.com/johnwyles/vrddt-droplets/usecases/vrddtvideos"
	"github.com/spf13/viper"
)

func main() {
	cfg := loadConfig()
	lg := logger.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)

	db, closeSession, err := mongo.Connect(cfg.MongoURI, true)
	if err != nil {
		lg.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer closeSession()

	lg.Debugf("setting up rest api service")
	redditVideoStore := mongo.NewRedditVideoStore(db)
	vrddtVideoStore := mongo.NewVrddtVideoStore(db)

	userRegistration := redditvideos.NewRegistrar(lg, redditVideoStore)
	userRetriever := redditvideos.NewRetriever(lg, redditVideoStore)

	vrddtVideoPub := vrddtvideos.NewPublication(lg, vrddtVideoStore, redditVideoStore)
	vrddtVideoRet := vrddtvideos.NewRetriever(lg, vrddtVideoStore)

	restHandler := rest.New(lg, userRegistration, userRetriever, vrddtVideoRet, vrddtVideoPub)
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
	rest = middlewares.WithBasicAuth(lg, rest,
		middlewares.UserVerifierFunc(func(ctx context.Context, name, secret string) bool {
			return secret == "secret@123"
		}),
	)

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
}

func loadConfig() config {
	viper.SetDefault("MONGO_URI", "mongodb://localhost/droplets")
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_FORMAT", "text")
	viper.SetDefault("ADDR", ":8080")
	viper.SetDefault("STATIC_DIR", "./web/static/")
	viper.SetDefault("TEMPLATE_DIR", "./web/templates/")
	viper.SetDefault("GRACEFUL_TIMEOUT", 20*time.Second)

	viper.ReadInConfig()
	viper.AutomaticEnv()

	return config{
		// application configuration
		Addr:            viper.GetString("ADDR"),
		StaticDir:       viper.GetString("STATIC_DIR"),
		TemplateDir:     viper.GetString("TEMPLATE_DIR"),
		LogLevel:        viper.GetString("LOG_LEVEL"),
		LogFormat:       viper.GetString("LOG_FORMAT"),
		GracefulTimeout: viper.GetDuration("GRACEFUL_TIMEOUT"),

		// store configuration
		MongoURI: viper.GetString("MONGO_URI"),
	}
}
