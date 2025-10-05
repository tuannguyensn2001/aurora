package mapper

import (
	"api/internal/model"
	"encoding/json"
	"errors"
	"sdk"
)

// ParameterToSDK converts a model.Parameter to sdk.Parameter
func ParameterToSDK(parameter *model.Parameter) (sdk.Parameter, error) {
	if parameter == nil {
		return sdk.Parameter{}, errors.New("parameter is nil")
	}

	// Convert default rollout value to string
	defaultRolloutValueStr, err := rolloutValueToString(parameter.DefaultRolloutValue)
	if err != nil {
		return sdk.Parameter{}, err
	}

	// Map rules
	sdkRules, err := parameterRulesToSDK(parameter.Rules)
	if err != nil {
		return sdk.Parameter{}, err
	}

	return sdk.Parameter{
		Name:                parameter.Name,
		DataType:            sdk.ParameterDataType(parameter.DataType),
		DefaultRolloutValue: defaultRolloutValueStr,
		Rules:               sdkRules,
	}, nil
}

// ParametersToSDK converts a slice of model.Parameter to slice of sdk.Parameter
func ParametersToSDK(parameters []*model.Parameter) ([]sdk.Parameter, error) {
	sdkParameters := make([]sdk.Parameter, len(parameters))
	for i, parameter := range parameters {
		sdkParameter, err := ParameterToSDK(parameter)
		if err != nil {
			return nil, err
		}
		sdkParameters[i] = sdkParameter
	}
	return sdkParameters, nil
}

// rolloutValueToString converts a RolloutValue to its string representation
func rolloutValueToString(rolloutValue model.RolloutValue) (string, error) {
	if rolloutValue.Data == nil {
		return "", nil
	}

	// Convert to JSON string representation
	jsonBytes, err := json.Marshal(rolloutValue.Data)
	if err != nil {
		return "", err
	}

	// Remove quotes if it's a simple string value
	jsonStr := string(jsonBytes)
	if len(jsonStr) >= 2 && jsonStr[0] == '"' && jsonStr[len(jsonStr)-1] == '"' {
		return jsonStr[1 : len(jsonStr)-1], nil
	}

	return jsonStr, nil
}

// parameterRulesToSDK converts model parameter rules to SDK parameter rules
func parameterRulesToSDK(rules []model.ParameterRule) ([]sdk.ParameterRule, error) {
	sdkRules := make([]sdk.ParameterRule, len(rules))

	for i, rule := range rules {
		// Convert rollout value to string
		rolloutValueStr, err := rolloutValueToString(rule.RolloutValue)
		if err != nil {
			return nil, err
		}

		// Handle segment ID conversion (pointer uint to int64)
		var segmentID int64
		if rule.SegmentID != nil {
			segmentID = int64(*rule.SegmentID)
		}

		// Handle match type conversion (pointer to value)
		var matchType sdk.ConditionMatchType
		if rule.MatchType != nil {
			matchType = sdk.ConditionMatchType(*rule.MatchType)
		}

		// Map conditions
		sdkConditions, err := parameterRuleConditionsToSDK(rule.Conditions)
		if err != nil {
			return nil, err
		}

		sdkRules[i] = sdk.ParameterRule{
			ID:           rule.ID,
			Name:         rule.Name,
			Type:         sdk.RuleType(rule.Type),
			RolloutValue: rolloutValueStr,
			SegmentID:    segmentID,
			MatchType:    matchType,
			Conditions:   sdkConditions,
		}
	}

	return sdkRules, nil
}

// parameterRuleConditionsToSDK converts model parameter rule conditions to SDK parameter rule conditions
func parameterRuleConditionsToSDK(conditions []model.ParameterRuleCondition) ([]sdk.ParameterRuleCondition, error) {
	sdkConditions := make([]sdk.ParameterRuleCondition, len(conditions))

	for i, condition := range conditions {
		// Get attribute information if available
		var attributeName, attributeDataType string
		if condition.Attribute != nil {
			attributeName = condition.Attribute.Name
			attributeDataType = string(condition.Attribute.DataType)
		}

		sdkConditions[i] = sdk.ParameterRuleCondition{
			ID:                condition.ID,
			AttributeID:       condition.AttributeID,
			Operator:          sdk.ConditionOperator(condition.Operator),
			Value:             condition.Value,
			AttributeName:     attributeName,
			AttributeDataType: attributeDataType,
		}
		if condition.Attribute != nil && condition.Attribute.DataType == model.DataTypeEnum {
			sdkConditions[i].EnumOptions = condition.Attribute.EnumOptions
		}
	}

	return sdkConditions, nil
}
