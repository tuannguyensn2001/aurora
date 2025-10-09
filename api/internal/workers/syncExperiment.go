package workers

import (
	"api/config"
	"api/internal/dto"
	"api/internal/mapper"
	"api/internal/repository"
	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	err := w.ProcessSyncExperiment(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to sync experiment")
		return err
	}

	return nil
}

func (w *SyncExperimentWorker) ProcessSyncExperiment(ctx context.Context) error {
	logger := log.Ctx(ctx).With().Str("worker", "sync-experiment").Logger()
	logger.Info().Msg("Processing sync experiment")

	if !w.Cfg.S3.Enable {
		logger.Info().Msg("S3 is not enabled, skipping sync experiment")
		return nil
	}

	experiments, err := w.Repository.GetExperimentsActive(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get active experiments")
		return err
	}

	// Convert to SDK format using mapper
	sdkExperiments, err := mapper.ExperimentsToSDKFromRawValue(experiments)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to map experiments to SDK")
		return err
	}

	logger.Info().Int("experiments_count", len(sdkExperiments)).Msg("Found experiments to sync")
	if len(sdkExperiments) > 0 {
		logger.Debug().Interface("experiments", sdkExperiments[0]).Msg("Experiments to sync")
	}

	jsonExperiments, err := json.Marshal(sdkExperiments)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal experiments")
		return err
	}

	_, err = w.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(w.Cfg.S3.BucketName),
		Key:    aws.String("experiments.json"),
		Body:   bytes.NewReader(jsonExperiments),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to put experiments to S3")
		return err
	}

	return nil
}
