package server

import (
	"api/config"
	"api/internal/endpoint"
	"api/internal/handler"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

func Serve(configPath string) {

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := zerolog.New(getWriter(cfg)).Level(getZerologLevel(cfg.Logging.Level)).With().Timestamp().Logger()

	logger.Info().Msg("Starting server")

	handler := handler.New()

	endpointList := endpoint.Get(handler, logger)

	r := mux.NewRouter()

	endpoint.Routes(r, endpointList)

	http.ListenAndServe(":8080", r)
}

func getZerologLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
	}

	return zerolog.InfoLevel
}

func getWriter(cfg *config.Config) io.Writer {
	if !cfg.IsDevelopment() {
		return os.Stdout
	}

	return zerolog.ConsoleWriter{
		Out: os.Stderr,
	}
}
