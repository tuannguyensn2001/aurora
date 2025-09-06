package fx

import (
	"api/config"
	"io"
	"os"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

// LoggerParams holds the parameters needed for logger
type LoggerParams struct {
	fx.In
	Config *config.Config
}

// ProvideLogger provides the logger instance
func ProvideLogger(params LoggerParams) zerolog.Logger {
	return zerolog.New(getWriter(params.Config)).
		Level(getZerologLevel(params.Config.Logging.Level)).
		With().
		Timestamp().
		Logger()
}

// LoggerModule provides the logger module
var LoggerModule = fx.Module("logger",
	fx.Provide(ProvideLogger),
)

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
