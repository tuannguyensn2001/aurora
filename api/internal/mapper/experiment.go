package mapper

import (
	"api/internal/model"
	"errors"
	"sdk"
)

// ExperimentToSDK converts a model.Experiment to sdk.Experiment
func ExperimentToSDK(experiment *model.Experiment) (sdk.Experiment, error) {
	if experiment == nil {
		return sdk.Experiment{}, errors.New("experiment is nil")
	}

	var segment *sdk.Segment
	if experiment.Segment != nil {
		sdkSegment, err := SegmentToSDK(experiment.Segment)
		if err != nil {
			return sdk.Experiment{}, err
		}
		segment = &sdkSegment
	}

	return sdk.Experiment{
		ID:              experiment.ID,
		Name:            experiment.Name,
		Uuid:            experiment.Uuid,
		StartDate:       experiment.StartDate,
		EndDate:         experiment.EndDate,
		HashAttributeID: experiment.HashAttributeID,
		PopulationSize:  experiment.PopulationSize,
		Strategy:        experiment.Strategy,
		Status:          experiment.Status,
		SegmentID:       experiment.SegmentID,
		Segment:         segment,
		Variants:        []sdk.ExperimentVariant{}, // Will be populated separately
	}, nil
}

// ExperimentsToSDK converts a slice of model.Experiment to slice of sdk.Experiment
func ExperimentsToSDK(experiments []*model.Experiment) ([]sdk.Experiment, error) {
	sdkExperiments := make([]sdk.Experiment, len(experiments))
	for i, experiment := range experiments {
		sdkExperiment, err := ExperimentToSDK(experiment)
		if err != nil {
			return nil, err
		}
		sdkExperiments[i] = sdkExperiment
	}
	return sdkExperiments, nil
}

// ExperimentVariantToSDK converts a model.ExperimentVariant to sdk.ExperimentVariant
func ExperimentVariantToSDK(variant *model.ExperimentVariant) (sdk.ExperimentVariant, error) {
	if variant == nil {
		return sdk.ExperimentVariant{}, errors.New("experiment variant is nil")
	}

	return sdk.ExperimentVariant{
		ID:                variant.ID,
		ExperimentID:      variant.ExperimentID,
		Name:              variant.Name,
		Description:       variant.Description,
		TrafficAllocation: variant.TrafficAllocation,
		Parameters:        []sdk.ExperimentVariantParameter{}, // Will be populated separately
	}, nil
}

// ExperimentVariantsToSDK converts a slice of model.ExperimentVariant to slice of sdk.ExperimentVariant
func ExperimentVariantsToSDK(variants []*model.ExperimentVariant) ([]sdk.ExperimentVariant, error) {
	sdkVariants := make([]sdk.ExperimentVariant, len(variants))
	for i, variant := range variants {
		sdkVariant, err := ExperimentVariantToSDK(variant)
		if err != nil {
			return nil, err
		}
		sdkVariants[i] = sdkVariant
	}
	return sdkVariants, nil
}

// ExperimentVariantParameterToSDK converts a model.ExperimentVariantParameter to sdk.ExperimentVariantParameter
func ExperimentVariantParameterToSDK(parameter *model.ExperimentVariantParameter) (sdk.ExperimentVariantParameter, error) {
	if parameter == nil {
		return sdk.ExperimentVariantParameter{}, errors.New("experiment variant parameter is nil")
	}

	return sdk.ExperimentVariantParameter{
		ID:                parameter.ID,
		ParameterDataType: parameter.ParameterDataType,
		ParameterID:       parameter.ParameterID,
		ParameterName:     parameter.ParameterName,
		RolloutValue:      parameter.RolloutValue,
	}, nil
}

// ExperimentVariantParametersToSDK converts a slice of model.ExperimentVariantParameter to slice of sdk.ExperimentVariantParameter
func ExperimentVariantParametersToSDK(parameters []*model.ExperimentVariantParameter) ([]sdk.ExperimentVariantParameter, error) {
	sdkParameters := make([]sdk.ExperimentVariantParameter, len(parameters))
	for i, parameter := range parameters {
		sdkParameter, err := ExperimentVariantParameterToSDK(parameter)
		if err != nil {
			return nil, err
		}
		sdkParameters[i] = sdkParameter
	}
	return sdkParameters, nil
}

// ExperimentWithVariantsToSDK converts experiment with its variants to SDK format
func ExperimentWithVariantsToSDK(experiment *model.Experiment, variants []*model.ExperimentVariant, variantParameters map[int][]*model.ExperimentVariantParameter) (sdk.Experiment, error) {
	if experiment == nil {
		return sdk.Experiment{}, errors.New("experiment is nil")
	}

	// Convert base experiment
	sdkExperiment, err := ExperimentToSDK(experiment)
	if err != nil {
		return sdk.Experiment{}, err
	}

	// Convert variants with their parameters
	if variants != nil {
		sdkVariants := make([]sdk.ExperimentVariant, len(variants))
		for i, variant := range variants {
			sdkVariant, err := ExperimentVariantToSDK(variant)
			if err != nil {
				return sdk.Experiment{}, err
			}

			// Add parameters for this variant if they exist
			if params, exists := variantParameters[variant.ID]; exists {
				sdkParams, err := ExperimentVariantParametersToSDK(params)
				if err != nil {
					return sdk.Experiment{}, err
				}
				sdkVariant.Parameters = sdkParams
			}

			sdkVariants[i] = sdkVariant
		}
		sdkExperiment.Variants = sdkVariants
	}

	return sdkExperiment, nil
}
