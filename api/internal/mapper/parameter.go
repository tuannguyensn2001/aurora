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

// ParametersToSDKFromRawValue efficiently converts parameters to SDK format using raw_value field
// This avoids expensive preloading since raw_value already contains all necessary data
func ParametersToSDKFromRawValue(parameters []*model.Parameter) ([]sdk.Parameter, error) {
	sdkParameters := make([]sdk.Parameter, len(parameters))
	for i, parameter := range parameters {
		sdkParameter, err := ParameterToSDKFromRawValue(parameter)
		if err != nil {
			return nil, err
		}
		sdkParameters[i] = sdkParameter
	}
	return sdkParameters, nil
}

// ParameterToSDKFromRawValue converts a model.Parameter to sdk.Parameter using only the raw_value field
func ParameterToSDKFromRawValue(parameter *model.Parameter) (sdk.Parameter, error) {
	if parameter == nil {
		return sdk.Parameter{}, errors.New("parameter is nil")
	}

	// RawValue must be present since that's all we query
	if len(parameter.RawValue) == 0 {
		return sdk.Parameter{}, errors.New("raw_value is empty")
	}

	// Parse raw value to extract all data
	var rawData map[string]interface{}
	if err := json.Unmarshal(parameter.RawValue, &rawData); err != nil {
		return sdk.Parameter{}, err
	}

	// Extract name from raw data
	name, _ := rawData["name"].(string)

	// Extract data type from raw data
	dataType, _ := rawData["dataType"].(string)

	// Extract and convert default rollout value from raw data
	var defaultRolloutValueStr string
	if defaultRolloutData, ok := rawData["defaultRolloutValue"].(map[string]interface{}); ok {
		if value, exists := defaultRolloutData["value"]; exists {
			valueBytes, err := json.Marshal(value)
			if err != nil {
				return sdk.Parameter{}, err
			}
			valueStr := string(valueBytes)
			// Remove quotes if it's a simple string value
			if len(valueStr) >= 2 && valueStr[0] == '"' && valueStr[len(valueStr)-1] == '"' {
				defaultRolloutValueStr = valueStr[1 : len(valueStr)-1]
			} else {
				defaultRolloutValueStr = valueStr
			}
		}
	}

	// Extract and convert rules from raw data
	var sdkRules []sdk.ParameterRule
	if rulesData, ok := rawData["rules"].([]interface{}); ok {
		var err error
		sdkRules, err = extractRulesFromRawData(rulesData)
		if err != nil {
			return sdk.Parameter{}, err
		}
	}

	return sdk.Parameter{
		Name:                name,
		DataType:            sdk.ParameterDataType(dataType),
		DefaultRolloutValue: defaultRolloutValueStr,
		Rules:               sdkRules,
	}, nil
}

// extractRulesFromRawData extracts and converts rules from raw JSON data
func extractRulesFromRawData(rulesData []interface{}) ([]sdk.ParameterRule, error) {
	sdkRules := make([]sdk.ParameterRule, len(rulesData))

	for i, ruleData := range rulesData {
		ruleMap, ok := ruleData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract basic rule fields
		var ruleID uint
		var ruleName string
		var ruleType sdk.RuleType
		var rolloutValue string
		var segmentID int64
		var matchType sdk.ConditionMatchType
		var segment *sdk.Segment

		if id, ok := ruleMap["id"].(float64); ok {
			ruleID = uint(id)
		}
		if name, ok := ruleMap["name"].(string); ok {
			ruleName = name
		}
		if rType, ok := ruleMap["type"].(string); ok {
			ruleType = sdk.RuleType(rType)
		}

		// Convert rollout value
		if rolloutData, ok := ruleMap["rolloutValue"].(map[string]interface{}); ok {
			if value, exists := rolloutData["value"]; exists {
				valueBytes, err := json.Marshal(value)
				if err != nil {
					return nil, err
				}
				valueStr := string(valueBytes)
				// Remove quotes if it's a simple string value
				if len(valueStr) >= 2 && valueStr[0] == '"' && valueStr[len(valueStr)-1] == '"' {
					rolloutValue = valueStr[1 : len(valueStr)-1]
				} else {
					rolloutValue = valueStr
				}
			}
		}

		// Extract segment information
		if segmentData, ok := ruleMap["segment"].(map[string]interface{}); ok {
			if segID, exists := segmentData["id"].(float64); exists {
				segmentID = int64(segID)
			}
			// Convert segment data
			segmentInfo := sdk.Segment{}
			if name, ok := segmentData["name"].(string); ok {
				segmentInfo.Name = name
			}
			if desc, ok := segmentData["description"].(string); ok {
				segmentInfo.Description = desc
			}
			segment = &segmentInfo
		} else if segID, ok := ruleMap["segmentId"].(float64); ok {
			segmentID = int64(segID)
		}

		// Extract match type
		if mType, ok := ruleMap["matchType"].(string); ok {
			matchType = sdk.ConditionMatchType(mType)
		}

		// Extract conditions
		var conditions []sdk.RuleCondition
		if conditionsData, ok := ruleMap["conditions"].([]interface{}); ok {
			conditions = make([]sdk.RuleCondition, len(conditionsData))
			for j, condData := range conditionsData {
				if condMap, ok := condData.(map[string]interface{}); ok {
					condition := sdk.RuleCondition{}
					if id, ok := condMap["id"].(float64); ok {
						condition.ID = uint(id)
					}
					if attrID, ok := condMap["attributeId"].(float64); ok {
						condition.AttributeID = uint(attrID)
					}
					if op, ok := condMap["operator"].(string); ok {
						condition.Operator = sdk.ConditionOperator(op)
					}
					if val, ok := condMap["value"].(string); ok {
						condition.Value = val
					}
					// Extract attribute information if available
					if attrData, ok := condMap["attribute"].(map[string]interface{}); ok {
						if name, ok := attrData["name"].(string); ok {
							condition.AttributeName = name
						}
						if dataType, ok := attrData["dataType"].(string); ok {
							condition.AttributeDataType = dataType
						}
						if enumOpts, ok := attrData["enumOptions"].([]interface{}); ok {
							condition.EnumOptions = make([]string, len(enumOpts))
							for k, opt := range enumOpts {
								if optStr, ok := opt.(string); ok {
									condition.EnumOptions[k] = optStr
								}
							}
						}
					}
					conditions[j] = condition
				}
			}
		}

		sdkRules[i] = sdk.ParameterRule{
			ID:           ruleID,
			Name:         ruleName,
			Type:         ruleType,
			RolloutValue: rolloutValue,
			SegmentID:    segmentID,
			MatchType:    matchType,
			Conditions:   conditions,
			Segment:      segment,
		}
	}

	return sdkRules, nil
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
		var segmentID int64
		var segment *sdk.Segment
		if rule.SegmentID != nil {
			segmentID = int64(*rule.SegmentID)
			segmentRaw, err := SegmentToSDK(rule.Segment)
			if err != nil {
				return nil, err
			}
			segment = &segmentRaw

		}

		sdkRules[i] = sdk.ParameterRule{
			ID:           rule.ID,
			Name:         rule.Name,
			Type:         sdk.RuleType(rule.Type),
			RolloutValue: rolloutValueStr,
			SegmentID:    segmentID,
			MatchType:    matchType,
			Conditions:   sdkConditions,
			Segment:      segment,
		}
	}

	return sdkRules, nil
}

// parameterRuleConditionsToSDK converts model parameter rule conditions to SDK rule conditions
func parameterRuleConditionsToSDK(conditions []model.ParameterRuleCondition) ([]sdk.RuleCondition, error) {
	sdkConditions := make([]sdk.RuleCondition, len(conditions))

	for i, condition := range conditions {
		// Get attribute information if available
		var attributeName, attributeDataType string
		if condition.Attribute != nil {
			attributeName = condition.Attribute.Name
			attributeDataType = string(condition.Attribute.DataType)
		}

		sdkConditions[i] = sdk.RuleCondition{
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
