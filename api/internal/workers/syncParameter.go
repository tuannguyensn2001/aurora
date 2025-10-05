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

type SyncParameterWorker struct {
	river.WorkerDefaults[dto.SyncParameterArgs]
	Repository repository.Repository
	Cfg        config.Config
	S3         *s3.Client
}

func (w *SyncParameterWorker) Work(ctx context.Context, job *river.Job[dto.SyncParameterArgs]) error {
	logger := log.Ctx(ctx).With().Str("worker", "sync-parameter").Logger()
	logger.Info().Msg("Syncing parameter")
	err := w.ProcessSyncParameter(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to sync parameter")
		return err
	}

	return nil
}

func (w *SyncParameterWorker) ProcessSyncParameter(ctx context.Context) error {
	logger := log.Ctx(ctx).With().Str("worker", "sync-parameter").Logger()
	logger.Info().Msg("Processing sync parameter")

	if !w.Cfg.S3.Enable {
		logger.Info().Msg("S3 is not enabled, skipping sync parameter")
		return nil
	}

	parameters, err := w.Repository.GetAllParameters(ctx, 0, 0)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get parameter by ID")
		return err
	}

	// Convert to SDK format using mapper
	sdkParameters, err := mapper.ParametersToSDK(parameters)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to map parameters to SDK")
		return err
	}

	logger.Info().Int("parameters_count", len(sdkParameters)).Msg("Found parameters to sync")
	jsonParameters, err := json.Marshal(sdkParameters)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal parameters")
		return err
	}

	_, err = w.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(w.Cfg.S3.BucketName),
		Key:    aws.String("parameters.json"),
		Body:   bytes.NewReader(jsonParameters),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to put parameters to S3")
		return err
	}

	return nil
}
