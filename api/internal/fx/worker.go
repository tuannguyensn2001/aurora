package fx

import (
	"api/config"
	"api/internal/repository"
	internalWorkers "api/internal/workers"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

type WorkerParams struct {
	fx.In
	Repository repository.Repository
	Cfg        *config.Config
	S3         *s3.Client
}

func ProvideWorker(params WorkerParams) *river.Workers {
	workers := river.NewWorkers()
	river.AddWorker(workers, &internalWorkers.SyncParameterWorker{
		Repository: params.Repository,
		Cfg:        *params.Cfg,
		S3:         params.S3,
	})
	river.AddWorker(workers, &internalWorkers.SyncExperimentWorker{
		Repository: params.Repository,
		Cfg:        *params.Cfg,
		S3:         params.S3,
	})
	return workers
}

var WorkerModule = fx.Module("worker", fx.Provide(ProvideWorker))
