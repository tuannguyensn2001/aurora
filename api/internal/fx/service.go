package fx

import (
	"api/internal/external/solver"
	"api/internal/repository"
	"api/internal/service"
	"sdk"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

// ServiceParams holds the parameters needed for service
type ServiceParams struct {
	fx.In
	Repository   repository.Repository
	RiverClient  *river.Client[pgx.Tx]
	AuroraClient sdk.Client
	Solver       solver.Solver
}

// ProvideService provides the service instance
func ProvideService(params ServiceParams) service.Service {
	return service.New(params.Repository, params.RiverClient, params.AuroraClient, params.Solver)
}

// ServiceModule provides the service module
var ServiceModule = fx.Module("service",
	fx.Provide(ProvideService),
)
