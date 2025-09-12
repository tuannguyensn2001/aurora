package workers

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
)

type SyncParameterArgs struct {
	ParameterID int
}

func (SyncParameterArgs) Kind() string {
	return "sync_parameter"
}

type SyncParameterWorker struct {
	river.WorkerDefaults[SyncParameterArgs]
}

func (w *SyncParameterWorker) Work(ctx context.Context, job *river.Job[SyncParameterArgs]) error {
	logger := log.Ctx(ctx).With().Str("worker", "sync-parameter").Logger()
	logger.Info().Msg("Syncing parameter")

	return nil
}
