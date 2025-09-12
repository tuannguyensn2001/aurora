package fx

import (
	"api/internal/endpoint"
	"api/internal/handler"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

// EndpointParams holds the parameters needed for endpoints
type EndpointParams struct {
	fx.In
	Handler *handler.Handler
	Logger  zerolog.Logger
}

// ProvideEndpoints provides the endpoints
func ProvideEndpoints(params EndpointParams) endpoint.Endpoints {
	return endpoint.MakeEndpoints(params.Handler, params.Logger)
}

// EndpointModule provides the endpoint module
var EndpointModule = fx.Module("endpoint",
	fx.Provide(ProvideEndpoints),
)
