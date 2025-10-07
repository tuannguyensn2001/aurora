package mapper

import (
	"api/internal/model"
	"errors"
	"sdk"
)

// ExperimentToSDK converts a model.Experiment to sdk.Experiment with its variants
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

	// Convert variants with their parameters
	var sdkVariants []sdk.ExperimentVariant
	if experiment.Variants != nil {
		sdkVariants = make([]sdk.ExperimentVariant, len(experiment.Variants))
		for i, variant := range experiment.Variants {
			sdkVariant, err := ExperimentVariantToSDK(&variant)
			if err != nil {
				return sdk.Experiment{}, err
			}
			sdkVariants[i] = sdkVariant
		}
	}

	return sdk.Experiment{
		ID:                experiment.ID,
		Name:              experiment.Name,
		Uuid:              experiment.Uuid,
		StartDate:         experiment.StartDate,
		EndDate:           experiment.EndDate,
		HashAttributeID:   experiment.HashAttributeID,
		PopulationSize:    experiment.PopulationSize,
		Strategy:          experiment.Strategy,
		Status:            experiment.Status,
		SegmentID:         experiment.SegmentID,
		Segment:           segment,
		Variants:          sdkVariants,
		HashAttributeName: experiment.HashAttribute.Name,
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

// ExperimentVariantToSDK converts a model.ExperimentVariant to sdk.ExperimentVariant with its parameters
func ExperimentVariantToSDK(variant *model.ExperimentVariant) (sdk.ExperimentVariant, error) {
	if variant == nil {
		return sdk.ExperimentVariant{}, errors.New("experiment variant is nil")
	}

	// Convert parameters
	var sdkParameters []sdk.ExperimentVariantParameter
	if variant.Parameters != nil {
		sdkParameters = make([]sdk.ExperimentVariantParameter, len(variant.Parameters))
		for i, param := range variant.Parameters {
			sdkParam, err := ExperimentVariantParameterToSDK(&param)
			if err != nil {
				return sdk.ExperimentVariant{}, err
			}
			sdkParameters[i] = sdkParam
		}
	}

	return sdk.ExperimentVariant{
		ID:                variant.ID,
		ExperimentID:      variant.ExperimentID,
		Name:              variant.Name,
		Description:       variant.Description,
		TrafficAllocation: variant.TrafficAllocation,
		Parameters:        sdkParameters,
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
		ParameterDataType: sdk.ParameterDataType(parameter.ParameterDataType),
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
// This function handles the case where variants and parameters are loaded separately
func ExperimentWithVariantsToSDK(experiment *model.Experiment, variants []*model.ExperimentVariant, variantParameters map[int][]*model.ExperimentVariantParameter) (sdk.Experiment, error) {
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

	// Convert variants with their parameters from the separate collections
	var sdkVariants []sdk.ExperimentVariant
	if variants != nil {
		sdkVariants = make([]sdk.ExperimentVariant, len(variants))
		for i, variant := range variants {
			// Get parameters for this variant
			var variantParams []model.ExperimentVariantParameter
			if params, exists := variantParameters[variant.ID]; exists {
				variantParams = make([]model.ExperimentVariantParameter, len(params))
				for j, param := range params {
					variantParams[j] = *param
				}
			}

			// Create a temporary variant with parameters for conversion
			tempVariant := *variant
			tempVariant.Parameters = variantParams

			sdkVariant, err := ExperimentVariantToSDK(&tempVariant)
			if err != nil {
				return sdk.Experiment{}, err
			}
			sdkVariants[i] = sdkVariant
		}
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
		Variants:        sdkVariants,
	}, nil
}

// ExperimentCompleteToSDK converts a complete experiment structure to SDK format
// This is the most comprehensive function that handles all nested relationships
func ExperimentCompleteToSDK(experiment *model.Experiment) (sdk.Experiment, error) {
	if experiment == nil {
		return sdk.Experiment{}, errors.New("experiment is nil")
	}

	// This function assumes the experiment model already has all nested data loaded
	return ExperimentToSDK(experiment)
}

// ExperimentBatchToSDK converts multiple experiments with their variants to SDK format
func ExperimentBatchToSDK(experiments []*model.Experiment, experimentsVariants map[int][]*model.ExperimentVariant, variantParameters map[int][]*model.ExperimentVariantParameter) ([]sdk.Experiment, error) {
	if experiments == nil {
		return []sdk.Experiment{}, nil
	}

	sdkExperiments := make([]sdk.Experiment, len(experiments))
	for i, experiment := range experiments {
		variants := experimentsVariants[experiment.ID]
		sdkExp, err := ExperimentWithVariantsToSDK(experiment, variants, variantParameters)
		if err != nil {
			return nil, err
		}
		sdkExperiments[i] = sdkExp
	}

	return sdkExperiments, nil
}
