package mapper

import (
	"api/internal/model"
	"errors"
	"sdk"
)

// SegmentToSDK converts a model.Segment to sdk.Segment
func SegmentToSDK(segment *model.Segment) (sdk.Segment, error) {
	if segment == nil {
		return sdk.Segment{}, errors.New("segment is nil")
	}

	// Map rules
	sdkRules, err := segmentRulesToSDK(segment.Rules)
	if err != nil {
		return sdk.Segment{}, err
	}

	return sdk.Segment{
		ID:          segment.ID,
		Name:        segment.Name,
		Description: segment.Description,
		CreatedAt:   segment.CreatedAt,
		UpdatedAt:   segment.UpdatedAt,
		Rules:       sdkRules,
	}, nil
}

// SegmentsToSDK converts a slice of model.Segment to slice of sdk.Segment
func SegmentsToSDK(segments []*model.Segment) ([]sdk.Segment, error) {
	sdkSegments := make([]sdk.Segment, len(segments))
	for i, segment := range segments {
		sdkSegment, err := SegmentToSDK(segment)
		if err != nil {
			return nil, err
		}
		sdkSegments[i] = sdkSegment
	}
	return sdkSegments, nil
}

// segmentRulesToSDK converts model segment rules to SDK segment rules
func segmentRulesToSDK(rules []model.SegmentRule) ([]sdk.SegmentRule, error) {
	sdkRules := make([]sdk.SegmentRule, len(rules))

	for i, rule := range rules {
		// Map conditions
		sdkConditions, err := segmentRuleConditionsToSDK(rule.Conditions)
		if err != nil {
			return nil, err
		}

		sdkRules[i] = sdk.SegmentRule{
			ID:          rule.ID,
			Name:        rule.Name,
			Description: rule.Description,
			SegmentID:   rule.SegmentID,
			Conditions:  sdkConditions,
		}
	}

	return sdkRules, nil
}

// segmentRuleConditionsToSDK converts model segment rule conditions to SDK rule conditions
func segmentRuleConditionsToSDK(conditions []model.SegmentRuleCondition) ([]sdk.RuleCondition, error) {
	sdkConditions := make([]sdk.RuleCondition, len(conditions))

	for i, condition := range conditions {
		// Get attribute information if available
		var attributeName, attributeDataType string
		var enumOptions []string
		if condition.Attribute != nil {
			attributeName = condition.Attribute.Name
			attributeDataType = string(condition.Attribute.DataType)
			if condition.Attribute.DataType == model.DataTypeEnum {
				enumOptions = condition.Attribute.EnumOptions
			}
		}

		sdkConditions[i] = sdk.RuleCondition{
			ID:                condition.ID,
			AttributeID:       condition.AttributeID,
			Operator:          sdk.ConditionOperator(condition.Operator),
			Value:             condition.Value,
			AttributeName:     attributeName,
			AttributeDataType: attributeDataType,
			EnumOptions:       enumOptions,
		}
	}

	return sdkConditions, nil
}
