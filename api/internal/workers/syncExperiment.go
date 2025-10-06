package workers

import (
	"api/config"
	"api/internal/dto"
	"api/internal/repository"
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
)

type SyncExperimentWorker struct {
	river.WorkerDefaults[dto.SyncExperimentArgs]
	Repository repository.Repository
	Cfg        config.Config
	S3         *s3.Client
}

func (w *SyncExperimentWorker) Work(ctx context.Context, job *river.Job[dto.SyncExperimentArgs]) error {
	logger := log.Ctx(ctx).With().Str("worker", "sync-experiment").Logger()
	logger.Info().Msg("Syncing experiment")

	return nil
}
