package fx

import (
	"api/internal/repository"
	"api/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

// ServiceParams holds the parameters needed for service
type ServiceParams struct {
	fx.In
	Repository  repository.Repository
	RiverClient *river.Client[pgx.Tx]
}

// ProvideService provides the service instance
func ProvideService(params ServiceParams) service.Service {
	return service.New(params.Repository, params.RiverClient)
}

// ServiceModule provides the service module
var ServiceModule = fx.Module("service",
	fx.Provide(ProvideService),
)
