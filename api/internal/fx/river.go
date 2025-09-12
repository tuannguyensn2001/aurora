package fx

import (
	"api/config"
	internalWorkers "api/internal/workers"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

type RiverParams struct {
	fx.In
	Config *config.Config
	Logger zerolog.Logger
}

func ProvideRiver(lc fx.Lifecycle, params RiverParams) *river.Client[pgx.Tx] {
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", params.Config.Database.User, params.Config.Database.Password, params.Config.Database.Host, params.Config.Database.Port, params.Config.Database.DBName)
	dbPool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		panic(err)
	}
	workers := river.NewWorkers()
	river.AddWorker(workers, &internalWorkers.SyncParameterWorker{})
	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 1},
		},
		Middleware: []rivertype.Middleware{
			&loggingMiddleware{
				logger: params.Logger,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return riverClient.Stop(ctx)

		},
		OnStart: func(ctx context.Context) error {
			return riverClient.Start(context.WithoutCancel(ctx))
		},
	})

	return riverClient
}

var RiverModule = fx.Module("river",
	fx.Provide(ProvideRiver),
)

type loggingMiddleware struct {
	river.MiddlewareDefaults
	logger zerolog.Logger
}

func (m *loggingMiddleware) Work(ctx context.Context, job *rivertype.JobRow, doInner func(ctx context.Context) error) error {

	ctx = m.logger.WithContext(ctx)
	return doInner(ctx)
}
