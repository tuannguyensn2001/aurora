package fx

import (
	"api/internal/endpoint"
	"context"
	"net/http"

	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

// ServerParams holds the parameters needed for server
type ServerParams struct {
	fx.In
	Endpoints endpoint.Endpoints
	Logger    zerolog.Logger
}

// ProvideHTTPServer provides the HTTP server
func ProvideHTTPServer(lc fx.Lifecycle, params ServerParams) *http.Server {
	// Create HTTP handler
	httpHandler := endpoint.MakeHTTPHandler(params.Endpoints)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
	})

	httpHandler = c.Handler(httpHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: httpHandler,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info().Msg("Starting server on :8080")
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					params.Logger.Fatal().Err(err).Msg("Server failed to start")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info().Msg("Stopping server")
			return server.Shutdown(ctx)
		},
	})

	return server
}

// ServerModule provides the server module
var ServerModule = fx.Module("server",
	fx.Provide(ProvideHTTPServer),
)
