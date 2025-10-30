package storage

import (
	"context"
	"sdk/types"
)

// Storage interface defines the storage operations
type Storage interface {
	// Parameter operations
	PersistParameters(ctx context.Context, parameters []types.Parameter) error
	GetParameterByName(ctx context.Context, name string) (types.Parameter, error)

	// Experiment operations
	PersistExperiments(ctx context.Context, experiments []types.Experiment) error
	GetExperimentsByParameterName(ctx context.Context, parameterName string) ([]types.Experiment, error)

	// Lifecycle operations
	Close(ctx context.Context) error
}

// Factory function type for creating storage instances
type Factory func(config interface{}) (Storage, error)
