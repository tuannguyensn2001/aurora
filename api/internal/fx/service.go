package fx

import (
	"api/internal/repository"
	"api/internal/service"

	"go.uber.org/fx"
)

// ServiceParams holds the parameters needed for service
type ServiceParams struct {
	fx.In
	Repository repository.Repository
}

// ProvideService provides the service instance
func ProvideService(params ServiceParams) service.Service {
	return service.New(params.Repository)
}

// ServiceModule provides the service module
var ServiceModule = fx.Module("service",
	fx.Provide(ProvideService),
)
