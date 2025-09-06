package server

import (
	"api/config"
	"api/internal/database"
	"api/internal/endpoint"
	"api/internal/handler"
	"api/internal/repository"
	"api/internal/service"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

func Serve(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := zerolog.New(getWriter(cfg)).Level(getZerologLevel(cfg.Logging.Level)).With().Timestamp().Logger()

	logger.Info().Msg("Starting server")

	// Initialize database connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repository, service, and handler
	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)

	// Create endpoints
	endpoints := endpoint.MakeEndpoints(h)

	// Create HTTP handler
	httpHandler := endpoint.MakeHTTPHandler(endpoints)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
	})

	httpHandler = c.Handler(httpHandler)

	logger.Info().Msg("Server listening on :8080")
	if err := http.ListenAndServe(":8080", httpHandler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
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
		return zerolog.PanicLevel
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
