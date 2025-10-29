package fx

import (
	"api/config"
	"api/internal/external/solver"

	"go.uber.org/fx"
)

type SolverParams struct {
	fx.In
	Config *config.Config
}

func ProvideSolver(params SolverParams) solver.Solver {
	return solver.New(params.Config.Solver.EndpointURL)
}

var SolverModule = fx.Module("solver",
	fx.Provide(ProvideSolver),
)
