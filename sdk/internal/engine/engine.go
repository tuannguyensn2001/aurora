package engine

import (
	"sdk/types"
)

// Engine interface defines the evaluation operations
type Engine interface {
	// Parameter evaluation
	EvaluateParameter(parameter *types.Parameter, attribute Attribute) string

	// Experiment evaluation
	EvaluateExperiment(experiment *types.Experiment, attribute Attribute, parameterName string) (string, types.ParameterDataType, bool)
	EvaluateExperimentDetailed(experiment *types.Experiment, attribute Attribute, parameterName string) *types.ExperimentEvaluationResult
}

// Attribute interface for dependency injection
type Attribute interface {
	Get(key string) interface{}
	ToMap() map[string]interface{}
}

// Condition interface for unified condition evaluation
type Condition interface {
	GetAttributeName() string
	GetAttributeDataType() string
	GetOperator() types.ConditionOperator
	GetValue() string
	GetEnumOptions() []string
}
