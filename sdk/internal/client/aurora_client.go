package client

import (
	"context"
	"sdk/internal/config"
	"sdk/pkg/errors"
	"sdk/pkg/logger"
	"sdk/types"
	"time"
)

// AuroraClient implements the Client interface
type AuroraClient struct {
	config       *config.Config
	logger       logger.Logger
	storage      Storage
	engine       Engine
	eventTracker EventTracker
	dataFetcher  DataFetcher
	quit         chan struct{}
}

// NewAuroraClient creates a new Aurora client
func NewAuroraClient(cfg *config.Config, storage Storage, engine Engine, eventTracker EventTracker, dataFetcher DataFetcher) Client {
	return &AuroraClient{
		config:       cfg,
		logger:       cfg.Logger,
		storage:      storage,
		engine:       engine,
		eventTracker: eventTracker,
		dataFetcher:  dataFetcher,
		quit:         make(chan struct{}),
	}
}

// Start initializes and starts the client
func (c *AuroraClient) Start(ctx context.Context) error {
	c.logger.Info("starting Aurora client")

	if ctx.Err() != nil {
		c.logger.ErrorContext(ctx, "context is done")
		return ctx.Err()
	}

	err := c.persist(ctx)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to persist parameters", "error", err)
	}

	go c.dispatch(ctx)
	return nil
}

// Stop shuts down the client gracefully
func (c *AuroraClient) Stop() {
	c.logger.Info("stopping Aurora client")

	// Flush any pending events before stopping
	if c.eventTracker != nil {
		c.eventTracker.Stop(context.Background())
	}

	// Close storage
	if c.storage != nil {
		c.storage.Close(context.Background())
	}

	close(c.quit)
}

// EvaluateParameter evaluates a parameter against the given attributes
func (c *AuroraClient) EvaluateParameter(ctx context.Context, parameterName string, attribute Attribute) RolloutValue {
	// Try experiments first
	experimentResult, resExperiments := c.resolveFromExperiments(ctx, parameterName, attribute)
	if !resExperiments.HasError() {
		if c.config.OnEvaluate != nil {
			c.config.OnEvaluate("experiment", parameterName, attribute, resExperiments.Raw(), resExperiments.Error())
		}

		// Track experiment evaluation event
		if c.eventTracker != nil {
			event := c.eventTracker.CreateExperimentEvaluationEvent(
				parameterName,
				attribute,
				resExperiments.Raw(),
				resExperiments.Error(),
				experimentResult.ExperimentID,
				experimentResult.ExperimentUUID,
				experimentResult.VariantID,
				experimentResult.VariantName,
			)
			c.eventTracker.TrackEvent(ctx, event)
		}

		return resExperiments
	}

	// Fall back to parameters
	res := c.resolveFromParameter(ctx, parameterName, attribute)

	if c.config.OnEvaluate != nil {
		c.config.OnEvaluate("parameter", parameterName, attribute, res.Raw(), res.Error())
	}

	// Track parameter evaluation event
	if c.eventTracker != nil {
		event := c.eventTracker.CreateParameterEvaluationEvent(
			parameterName,
			attribute,
			res.Raw(),
			res.Error(),
		)
		c.eventTracker.TrackEvent(ctx, event)
	}

	return res
}

// GetMetadata retrieves metadata from the upstream service
func (c *AuroraClient) GetMetadata(ctx context.Context) (*types.MetadataResponse, error) {
	// This would be implemented by the data fetcher
	// For now, return a default response
	return &types.MetadataResponse{
		EnableS3: c.config.EnableS3,
	}, nil
}

// dispatch runs the background refresh loop
func (c *AuroraClient) dispatch(ctx context.Context) {
	c.logger.Info("starting dispatch loop")
	ticker := time.NewTicker(c.config.RefreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.logger.Info("refreshing data")
			err := c.persist(ctx)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to refresh data", "error", err)
			}
		case <-ctx.Done():
			c.logger.Info("context done, stopping dispatch")
			return
		case <-c.quit:
			c.logger.Info("quit signal received, stopping dispatch")
			return
		}
	}
}

// persist fetches and stores the latest data
func (c *AuroraClient) persist(ctx context.Context) error {
	// Fetch and persist experiments
	experiments, err := c.dataFetcher.GetExperiments(ctx)
	if err != nil {
		return err
	}
	err = c.storage.PersistExperiments(ctx, experiments)
	if err != nil {
		return err
	}

	// Fetch and persist parameters
	parameters, err := c.dataFetcher.GetParameters(ctx)
	if err != nil {
		return err
	}
	err = c.storage.PersistParameters(ctx, parameters)
	if err != nil {
		return err
	}

	c.logger.Info("data persisted successfully", "parameters", len(parameters), "experiments", len(experiments))
	return nil
}

// resolveFromExperiments tries to resolve a parameter from experiments
func (c *AuroraClient) resolveFromExperiments(ctx context.Context, parameterName string, attribute Attribute) (*types.ExperimentEvaluationResult, RolloutValue) {
	experiments, err := c.storage.GetExperimentsByParameterName(ctx, parameterName)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get experiments by parameter name", "error", err)
		return nil, NewRolloutValueWithError(errors.NewParameterNotFoundError(parameterName))
	}
	if len(experiments) == 0 {
		return nil, NewRolloutValueWithError(errors.NewParameterNotFoundError(parameterName))
	}

	for _, experiment := range experiments {
		result := c.engine.EvaluateExperimentDetailed(&experiment, attribute, parameterName)
		if result.Success {
			return result, NewRolloutValue(&result.Value, result.DataType)
		}
	}
	return nil, NewRolloutValueWithError(errors.NewParameterNotFoundError(parameterName))
}

// resolveFromParameter resolves a parameter from the parameter store
func (c *AuroraClient) resolveFromParameter(ctx context.Context, parameterName string, attribute Attribute) RolloutValue {
	c.logger.InfoContext(ctx, "resolving parameter", "parameterName", parameterName)
	parameter, err := c.storage.GetParameterByName(ctx, parameterName)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get parameter", "error", err)
		return NewRolloutValueWithError(errors.NewParameterNotFoundError(parameterName))
	}

	rolloutValueStr := c.engine.EvaluateParameter(&parameter, attribute)
	c.logger.InfoContext(ctx, "resolved parameter", "parameterName", parameterName, "rolloutValue", rolloutValueStr, "dataType", parameter.DataType, "rules count", len(parameter.Rules))
	return NewRolloutValue(&rolloutValueStr, parameter.DataType)
}
