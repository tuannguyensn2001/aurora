package fx

import (
	"api/config"
	"api/internal/handler"
	"api/internal/router"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

// ServerParams holds the parameters needed for server
type ServerParams struct {
	fx.In
	Handler *handler.Handler
	Logger  zerolog.Logger
	Config  *config.Config
}

// ProvideHTTPServer provides the HTTP server
func ProvideHTTPServer(lc fx.Lifecycle, params ServerParams) *http.Server {
	// Set gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	// Create gin engine
	engine := gin.New()

	// Setup CORS middleware
	engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
		c.Header("Access-Control-Allow-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Create router and setup routes
	r := router.New(params.Handler, params.Logger, params.Config)
	r.SetupRoutes(engine)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", params.Config.Service.Port),
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info().Msgf("Starting server on :%d", params.Config.Service.Port)
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
