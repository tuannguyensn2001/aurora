package fx

import (
	"api/internal/handler"
	"api/internal/service"

	"go.uber.org/fx"
)

// HandlerParams holds the parameters needed for handler
type HandlerParams struct {
	fx.In
	Service service.Service
}

// ProvideHandler provides the handler instance
func ProvideHandler(params HandlerParams) *handler.Handler {
	return handler.New(params.Service)
}

// HandlerModule provides the handler module
var HandlerModule = fx.Module("handler",
	fx.Provide(ProvideHandler),
)
