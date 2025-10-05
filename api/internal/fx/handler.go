package fx

import (
	"api/config"
	"api/internal/handler"
	"api/internal/service"

	"go.uber.org/fx"
)

// HandlerParams holds the parameters needed for handler
type HandlerParams struct {
	fx.In
	Service service.Service
	Config  *config.Config
}

// ProvideHandler provides the handler instance
func ProvideHandler(params HandlerParams) *handler.Handler {
	return handler.New(params.Service, params.Config)
}

// HandlerModule provides the handler module
var HandlerModule = fx.Module("handler",
	fx.Provide(ProvideHandler),
)
