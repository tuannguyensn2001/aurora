package mapper

import (
	"api/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"sdk/types"
)

// ExperimentToSDK converts a model.Experiment to sdk.Experiment with its variants
func ExperimentToSDK(experiment *model.Experiment) (types.Experiment, error) {
	if experiment == nil {
		return types.Experiment{}, errors.New("experiment is nil")
	}

	var segment *types.Segment
	if experiment.Segment != nil {
		sdkSegment, err := SegmentToSDK(experiment.Segment)
		if err != nil {
			return types.Experiment{}, err
		}
		segment = &sdkSegment
	}

	// Convert variants with their parameters
	var sdkVariants []types.ExperimentVariant
	if experiment.Variants != nil {
		sdkVariants = make([]types.ExperimentVariant, len(experiment.Variants))
		for i, variant := range experiment.Variants {
			sdkVariant, err := ExperimentVariantToSDK(&variant)
			if err != nil {
				return types.Experiment{}, err
			}
			sdkVariants[i] = sdkVariant
		}
	}

	return types.Experiment{
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
func ExperimentsToSDK(experiments []*model.Experiment) ([]types.Experiment, error) {
	sdkExperiments := make([]types.Experiment, len(experiments))
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
func ExperimentVariantToSDK(variant *model.ExperimentVariant) (types.ExperimentVariant, error) {
	if variant == nil {
		return types.ExperimentVariant{}, errors.New("experiment variant is nil")
	}

	// Convert parameters
	var sdkParameters []types.ExperimentVariantParameter
	if variant.Parameters != nil {
		sdkParameters = make([]types.ExperimentVariantParameter, len(variant.Parameters))
		for i, param := range variant.Parameters {
			sdkParam, err := ExperimentVariantParameterToSDK(&param)
			if err != nil {
				return types.ExperimentVariant{}, err
			}
			sdkParameters[i] = sdkParam
		}
	}

	return types.ExperimentVariant{
		ID:                variant.ID,
		ExperimentID:      variant.ExperimentID,
		Name:              variant.Name,
		Description:       variant.Description,
		TrafficAllocation: variant.TrafficAllocation,
		Parameters:        sdkParameters,
	}, nil
}

// ExperimentVariantsToSDK converts a slice of model.ExperimentVariant to slice of sdk.ExperimentVariant
func ExperimentVariantsToSDK(variants []*model.ExperimentVariant) ([]types.ExperimentVariant, error) {
	sdkVariants := make([]types.ExperimentVariant, len(variants))
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
func ExperimentVariantParameterToSDK(parameter *model.ExperimentVariantParameter) (types.ExperimentVariantParameter, error) {
	if parameter == nil {
		return types.ExperimentVariantParameter{}, errors.New("experiment variant parameter is nil")
	}

	return types.ExperimentVariantParameter{
		ID:                parameter.ID,
		ParameterDataType: types.ParameterDataType(parameter.ParameterDataType),
		ParameterID:       parameter.ParameterID,
		ParameterName:     parameter.ParameterName,
		RolloutValue:      parameter.RolloutValue,
	}, nil
}

// ExperimentVariantParametersToSDK converts a slice of model.ExperimentVariantParameter to slice of sdk.ExperimentVariantParameter
func ExperimentVariantParametersToSDK(parameters []*model.ExperimentVariantParameter) ([]types.ExperimentVariantParameter, error) {
	sdkParameters := make([]types.ExperimentVariantParameter, len(parameters))
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
func ExperimentWithVariantsToSDK(experiment *model.Experiment, variants []*model.ExperimentVariant, variantParameters map[int][]*model.ExperimentVariantParameter) (types.Experiment, error) {
	if experiment == nil {
		return types.Experiment{}, errors.New("experiment is nil")
	}

	var segment *types.Segment
	if experiment.Segment != nil {
		sdkSegment, err := SegmentToSDK(experiment.Segment)
		if err != nil {
			return types.Experiment{}, err
		}
		segment = &sdkSegment
	}

	// Convert variants with their parameters from the separate collections
	var sdkVariants []types.ExperimentVariant
	if variants != nil {
		sdkVariants = make([]types.ExperimentVariant, len(variants))
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
				return types.Experiment{}, err
			}
			sdkVariants[i] = sdkVariant
		}
	}

	return types.Experiment{
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
func ExperimentCompleteToSDK(experiment *model.Experiment) (types.Experiment, error) {
	if experiment == nil {
		return types.Experiment{}, errors.New("experiment is nil")
	}

	// This function assumes the experiment model already has all nested data loaded
	return ExperimentToSDK(experiment)
}

// ExperimentBatchToSDK converts multiple experiments with their variants to SDK format
func ExperimentBatchToSDK(experiments []*model.Experiment, experimentsVariants map[int][]*model.ExperimentVariant, variantParameters map[int][]*model.ExperimentVariantParameter) ([]types.Experiment, error) {
	if experiments == nil {
		return []types.Experiment{}, nil
	}

	sdkExperiments := make([]types.Experiment, len(experiments))
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

// ExperimentsToSDKFromRawValue converts experiments with raw values to SDK format
// This function uses the raw_value field from the database for efficient conversion
func ExperimentsToSDKFromRawValue(experiments []model.Experiment) ([]types.Experiment, error) {
	if experiments == nil {
		return []types.Experiment{}, nil
	}

	sdkExperiments := make([]types.Experiment, len(experiments))
	for i, experiment := range experiments {
		sdkExp, err := ExperimentToSDKFromRawValue(&experiment)
		if err != nil {
			return nil, err
		}
		sdkExperiments[i] = sdkExp
	}

	return sdkExperiments, nil
}

// ExperimentToSDKFromRawValue converts a single experiment with raw value to SDK format
// This function uses the raw_value field for efficient conversion
func ExperimentToSDKFromRawValue(experiment *model.Experiment) (types.Experiment, error) {
	if experiment == nil {
		return types.Experiment{}, errors.New("experiment is nil")
	}

	// If raw_value is available, use it for conversion
	if len(experiment.RawValue) > 0 {
		var rawData map[string]interface{}
		if err := json.Unmarshal(experiment.RawValue, &rawData); err != nil {
			return types.Experiment{}, fmt.Errorf("failed to unmarshal raw value: %w", err)
		}

		// Convert raw data to SDK format
		sdkExp, err := convertRawDataToSDK(rawData)
		if err != nil {
			return types.Experiment{}, err
		}
		return sdkExp, nil
	}

	// Fallback to regular conversion if raw_value is not available
	return ExperimentToSDK(experiment)
}

// convertRawDataToSDK converts raw JSON data to SDK experiment format
func convertRawDataToSDK(rawData map[string]interface{}) (types.Experiment, error) {
	// Extract basic experiment fields
	id, ok := rawData["id"].(float64)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment id")
	}

	name, ok := rawData["name"].(string)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment name")
	}

	uuid, ok := rawData["uuid"].(string)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment uuid")
	}

	startDate, ok := rawData["startDate"].(float64)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment startDate")
	}

	endDate, ok := rawData["endDate"].(float64)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment endDate")
	}

	hashAttributeID, ok := rawData["hashAttributeId"].(float64)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment hashAttributeId")
	}

	populationSize, ok := rawData["populationSize"].(float64)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment populationSize")
	}

	strategy, ok := rawData["strategy"].(string)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment strategy")
	}

	status, ok := rawData["status"].(string)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment status")
	}

	segmentID, ok := rawData["segmentId"].(float64)
	if !ok {
		return types.Experiment{}, errors.New("invalid experiment segmentId")
	}

	// Extract hash attribute name
	hashAttributeName := ""
	if hashAttribute, ok := rawData["hashAttribute"].(map[string]interface{}); ok {
		if name, ok := hashAttribute["name"].(string); ok {
			hashAttributeName = name
		}
	}

	// Extract segment
	var segment *types.Segment
	if segmentData, ok := rawData["segment"].(map[string]interface{}); ok {
		sdkSegment, err := convertRawSegmentToSDK(segmentData)
		if err != nil {
			return types.Experiment{}, err
		}
		segment = &sdkSegment
	}

	// Extract variants
	var variants []types.ExperimentVariant
	if variantsData, ok := rawData["variants"].([]interface{}); ok {
		variants = make([]types.ExperimentVariant, len(variantsData))
		for i, variantData := range variantsData {
			if variantMap, ok := variantData.(map[string]interface{}); ok {
				sdkVariant, err := convertRawVariantToSDK(variantMap)
				if err != nil {
					return types.Experiment{}, err
				}
				variants[i] = sdkVariant
			}
		}
	}

	return types.Experiment{
		ID:                int(id),
		Name:              name,
		Uuid:              uuid,
		StartDate:         int64(startDate),
		EndDate:           int64(endDate),
		HashAttributeID:   int(hashAttributeID),
		PopulationSize:    int(populationSize),
		Strategy:          strategy,
		Status:            status,
		SegmentID:         int(segmentID),
		Segment:           segment,
		Variants:          variants,
		HashAttributeName: hashAttributeName,
	}, nil
}

// ConvertRawSegmentToSDK converts raw segment data to SDK format (public function)
func ConvertRawSegmentToSDK(segmentData map[string]interface{}) (types.Segment, error) {
	return convertRawSegmentToSDK(segmentData)
}

// convertRawSegmentToSDK converts raw segment data to SDK format
func convertRawSegmentToSDK(segmentData map[string]interface{}) (types.Segment, error) {
	id, ok := segmentData["id"].(float64)
	if !ok {
		return types.Segment{}, errors.New("invalid segment id")
	}

	name, ok := segmentData["name"].(string)
	if !ok {
		return types.Segment{}, errors.New("invalid segment name")
	}

	description, ok := segmentData["description"].(string)
	if !ok {
		description = ""
	}

	// Extract rules
	var rules []types.SegmentRule
	if rulesData, ok := segmentData["rules"].([]interface{}); ok {
		rules = make([]types.SegmentRule, len(rulesData))
		for i, ruleData := range rulesData {
			if ruleMap, ok := ruleData.(map[string]interface{}); ok {
				sdkRule, err := convertRawSegmentRuleToSDK(ruleMap)
				if err != nil {
					return types.Segment{}, err
				}
				rules[i] = sdkRule
			}
		}
	}

	return types.Segment{
		ID:          uint(id),
		Name:        name,
		Description: description,
		Rules:       rules,
	}, nil
}

// convertRawSegmentRuleToSDK converts raw segment rule data to SDK format
func convertRawSegmentRuleToSDK(ruleData map[string]interface{}) (types.SegmentRule, error) {
	id, ok := ruleData["id"].(float64)
	if !ok {
		return types.SegmentRule{}, errors.New("invalid segment rule id")
	}

	segmentID, ok := ruleData["segmentId"].(float64)
	if !ok {
		return types.SegmentRule{}, errors.New("invalid segment rule segmentId")
	}

	// Extract conditions
	var conditions []types.RuleCondition
	if conditionsData, ok := ruleData["conditions"].([]interface{}); ok {
		conditions = make([]types.RuleCondition, len(conditionsData))
		for i, conditionData := range conditionsData {
			if conditionMap, ok := conditionData.(map[string]interface{}); ok {
				sdkCondition, err := convertRawConditionToSDK(conditionMap)
				if err != nil {
					return types.SegmentRule{}, err
				}
				conditions[i] = sdkCondition
			}
		}
	}

	return types.SegmentRule{
		ID:         uint(id),
		SegmentID:  uint(segmentID),
		Conditions: conditions,
	}, nil
}

// convertRawConditionToSDK converts raw condition data to SDK format
func convertRawConditionToSDK(conditionData map[string]interface{}) (types.RuleCondition, error) {
	id, ok := conditionData["id"].(float64)
	if !ok {
		return types.RuleCondition{}, errors.New("invalid condition id")
	}

	operator, ok := conditionData["operator"].(string)
	if !ok {
		return types.RuleCondition{}, errors.New("invalid condition operator")
	}

	value, ok := conditionData["value"].(string)
	if !ok {
		return types.RuleCondition{}, errors.New("invalid condition value")
	}

	// Extract attribute information from nested attribute object
	var attributeName, attributeDataType string
	if attribute, ok := conditionData["attribute"].(map[string]interface{}); ok {
		if name, ok := attribute["name"].(string); ok {
			attributeName = name
		}
		if dataType, ok := attribute["dataType"].(string); ok {
			attributeDataType = dataType
		}
	} else {
		// Fallback to direct fields if nested attribute is not available
		if name, ok := conditionData["attributeName"].(string); ok {
			attributeName = name
		}
		if dataType, ok := conditionData["attributeDataType"].(string); ok {
			attributeDataType = dataType
		}
	}

	// Extract enum options from nested attribute object
	var enumOptions []string
	if attribute, ok := conditionData["attribute"].(map[string]interface{}); ok {
		if enumData, ok := attribute["enumOptions"].([]interface{}); ok {
			enumOptions = make([]string, len(enumData))
			for i, option := range enumData {
				if optionStr, ok := option.(string); ok {
					enumOptions[i] = optionStr
				}
			}
		}
	} else {
		// Fallback to direct field if nested attribute is not available
		if enumData, ok := conditionData["enumOptions"].([]interface{}); ok {
			enumOptions = make([]string, len(enumData))
			for i, option := range enumData {
				if optionStr, ok := option.(string); ok {
					enumOptions[i] = optionStr
				}
			}
		}
	}

	return types.RuleCondition{
		ID:                uint(id),
		Operator:          types.ConditionOperator(operator),
		Value:             value,
		AttributeName:     attributeName,
		AttributeDataType: attributeDataType,
		EnumOptions:       enumOptions,
	}, nil
}

// convertRawVariantToSDK converts raw variant data to SDK format
func convertRawVariantToSDK(variantData map[string]interface{}) (types.ExperimentVariant, error) {
	id, ok := variantData["id"].(float64)
	if !ok {
		return types.ExperimentVariant{}, errors.New("invalid variant id")
	}

	experimentID, ok := variantData["experimentId"].(float64)
	if !ok {
		return types.ExperimentVariant{}, errors.New("invalid variant experimentId")
	}

	name, ok := variantData["name"].(string)
	if !ok {
		return types.ExperimentVariant{}, errors.New("invalid variant name")
	}

	description, ok := variantData["description"].(string)
	if !ok {
		description = ""
	}

	trafficAllocation, ok := variantData["trafficAllocation"].(float64)
	if !ok {
		return types.ExperimentVariant{}, errors.New("invalid variant trafficAllocation")
	}

	// Extract parameters
	var parameters []types.ExperimentVariantParameter
	if parametersData, ok := variantData["parameters"].([]interface{}); ok {
		parameters = make([]types.ExperimentVariantParameter, len(parametersData))
		for i, paramData := range parametersData {
			if paramMap, ok := paramData.(map[string]interface{}); ok {
				sdkParam, err := convertRawVariantParameterToSDK(paramMap)
				if err != nil {
					return types.ExperimentVariant{}, err
				}
				parameters[i] = sdkParam
			}
		}
	}

	return types.ExperimentVariant{
		ID:                int(id),
		ExperimentID:      int(experimentID),
		Name:              name,
		Description:       description,
		TrafficAllocation: int(trafficAllocation),
		Parameters:        parameters,
	}, nil
}

// convertRawVariantParameterToSDK converts raw variant parameter data to SDK format
func convertRawVariantParameterToSDK(paramData map[string]interface{}) (types.ExperimentVariantParameter, error) {
	id, ok := paramData["id"].(float64)
	if !ok {
		return types.ExperimentVariantParameter{}, errors.New("invalid parameter id")
	}

	parameterDataType, ok := paramData["parameterDataType"].(string)
	if !ok {
		return types.ExperimentVariantParameter{}, errors.New("invalid parameter data type")
	}

	parameterID, ok := paramData["parameterId"].(float64)
	if !ok {
		return types.ExperimentVariantParameter{}, errors.New("invalid parameter id")
	}

	parameterName, ok := paramData["parameterName"].(string)
	if !ok {
		return types.ExperimentVariantParameter{}, errors.New("invalid parameter name")
	}

	rolloutValue, ok := paramData["rolloutValue"].(string)
	if !ok {
		return types.ExperimentVariantParameter{}, errors.New("invalid parameter rollout value")
	}

	return types.ExperimentVariantParameter{
		ID:                int(id),
		ParameterDataType: types.ParameterDataType(parameterDataType),
		ParameterID:       int(parameterID),
		ParameterName:     parameterName,
		RolloutValue:      rolloutValue,
	}, nil
}
