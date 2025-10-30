package client

import (
	"context"
	"sdk/types"
)

// Client interface defines the main SDK operations
type Client interface {
	Start(ctx context.Context) error
	Stop()
	EvaluateParameter(ctx context.Context, parameterName string, attribute Attribute) RolloutValue
	GetMetadata(ctx context.Context) (*types.MetadataResponse, error)
}

// Attribute interface for dependency injection
type Attribute interface {
	Get(key string) interface{}
	ToMap() map[string]interface{}
}

// RolloutValue interface for evaluation results
type RolloutValue interface {
	HasError() bool
	Error() error
	AsString(defaultValue string) string
	AsNumber(defaultValue float64) float64
	AsInt(defaultValue int) int
	AsBool(defaultValue bool) bool
	Raw() *string
}

// DataFetcher interface for fetching data from various sources
type DataFetcher interface {
	GetParameters(ctx context.Context) ([]types.Parameter, error)
	GetExperiments(ctx context.Context) ([]types.Experiment, error)
}

// Storage interface for data persistence
type Storage interface {
	PersistParameters(ctx context.Context, parameters []types.Parameter) error
	GetParameterByName(ctx context.Context, name string) (types.Parameter, error)
	PersistExperiments(ctx context.Context, experiments []types.Experiment) error
	GetExperimentsByParameterName(ctx context.Context, parameterName string) ([]types.Experiment, error)
	Close(ctx context.Context) error
}

// Engine interface for evaluation logic
type Engine interface {
	EvaluateParameter(parameter *types.Parameter, attribute Attribute) string
	EvaluateExperiment(experiment *types.Experiment, attribute Attribute, parameterName string) (string, types.ParameterDataType, bool)
	EvaluateExperimentDetailed(experiment *types.Experiment, attribute Attribute, parameterName string) *types.ExperimentEvaluationResult
}

// EventTracker interface for event tracking
type EventTracker interface {
	TrackEvent(ctx context.Context, event types.EvaluationEvent)
	CreateParameterEvaluationEvent(parameterName string, attribute Attribute, rolloutValue *string, err error) types.EvaluationEvent
	CreateExperimentEvaluationEvent(parameterName string, attribute Attribute, rolloutValue *string, err error, experimentID *int, experimentUUID *string, variantID *int, variantName *string) types.EvaluationEvent
	Stop(ctx context.Context)
}
