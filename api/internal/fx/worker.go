package fx

import (
	"api/internal/repository"
	internalWorkers "api/internal/workers"

	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

type WorkerParams struct {
	fx.In
	Repository repository.Repository
}

func ProvideWorker(params WorkerParams) *river.Workers {
	workers := river.NewWorkers()
	river.AddWorker(workers, &internalWorkers.SyncParameterWorker{
		Repository: params.Repository,
	})
	return workers
}

var WorkerModule = fx.Module("worker", fx.Provide(ProvideWorker))
