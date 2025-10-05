package sdk

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

type engine struct {
	logger *slog.Logger
}

func newEngine(logger *slog.Logger) *engine {
	return &engine{
		logger: logger,
	}
}

func (e *engine) evaluateParameter(parameter *Parameter, attribute *attribute) string {

	if len(parameter.Rules) == 0 {
		e.logger.Info("no rules found for parameter", "parameter", parameter)
		return parameter.DefaultRolloutValue
	}

	for _, rule := range parameter.Rules {
		count := 0
		if rule.Type == RuleTypeAttribute {
			for _, condition := range rule.Conditions {
				if e.evaluateRuleCondition(&condition, attribute) {
					count++
				}
			}
		} else if rule.Type == RuleTypeSegment {
			if rule.MatchType == ConditionMatchTypeMatch && e.evaluateSegmentRule(&rule, attribute) {
				return rule.RolloutValue
			} else if rule.MatchType == ConditionMatchTypeNotMatch && !e.evaluateSegmentRule(&rule, attribute) {
				return rule.RolloutValue
			}
			continue
		}
		if count == len(rule.Conditions) {
			return rule.RolloutValue
		}
	}

	return parameter.DefaultRolloutValue

}

func (e *engine) evaluateRuleCondition(condition *ParameterRuleCondition, attribute *attribute) bool {
	e.logger.Debug("evaluating rule condition", "dataType", condition.AttributeDataType, "condition", condition, "attribute", attribute)
	switch condition.AttributeDataType {
	case "string":
		return e.evaluateRuleConditionString(condition, attribute)
	case "number":
		return e.evaluateRuleConditionNumber(condition, attribute)
	case "boolean":
		return e.evaluateRuleConditionBoolean(condition, attribute)
	case "enum":
		return e.evaluateRuleConditionEnum(condition, attribute)
	}

	return false
}

func (e *engine) evaluateRuleConditionEnum(condition *ParameterRuleCondition, attribute *attribute) bool {
	e.logger.Debug("evaluating rule condition enum", "condition", condition, "attribute", attribute)
	value, ok := attribute.Get(condition.AttributeName).(string)
	if !ok {
		return false
	}
	if !slices.Contains(condition.EnumOptions, value) {
		return false
	}
	return value == condition.Value
}

func (e *engine) evaluateRuleConditionString(condition *ParameterRuleCondition, attribute *attribute) bool {
	e.logger.Debug("evaluating rule condition string", "condition", condition, "attribute", attribute)
	value, ok := attribute.Get(condition.AttributeName).(string)
	if !ok {
		return false
	}
	switch condition.Operator {
	case ConditionOperatorEquals:
		return value == condition.Value
	case ConditionOperatorNotEquals:
		return value != condition.Value
	case ConditionOperatorContains:
		return strings.Contains(value, condition.Value)
	case ConditionOperatorNotContains:
		return !strings.Contains(value, condition.Value)
	case ConditionOperatorIn:
		values := strings.Split(condition.Value, ",")
		return slices.Contains(values, value)
	case ConditionOperatorNotIn:
		values := strings.Split(condition.Value, ",")
		return !slices.Contains(values, value)
	}
	return false
}

func (e *engine) evaluateRuleConditionNumber(condition *ParameterRuleCondition, attribute *attribute) bool {
	value, ok := attribute.Get(condition.AttributeName).(float64)
	if !ok {
		return false
	}
	expectedValue, err := strconv.ParseFloat(condition.Value, 64)
	if err != nil {
		return false
	}
	switch condition.Operator {
	case ConditionOperatorEquals:
		return value == expectedValue
	case ConditionOperatorNotEquals:
		return value != expectedValue
	case ConditionOperatorGreaterThan:
		return value > expectedValue
	case ConditionOperatorLessThan:
		return value < expectedValue
	case ConditionOperatorGreaterThanOrEqual:
		return value >= expectedValue
	case ConditionOperatorLessThanOrEqual:
		return value <= expectedValue
	case ConditionOperatorIn:
		values := strings.Split(condition.Value, ",")
		return slices.Contains(values, strconv.FormatFloat(value, 'f', -1, 64))
	case ConditionOperatorNotIn:
		values := strings.Split(condition.Value, ",")
		return !slices.Contains(values, strconv.FormatFloat(value, 'f', -1, 64))
	}
	return false
}

func (e *engine) evaluateRuleConditionBoolean(condition *ParameterRuleCondition, attribute *attribute) bool {
	value, ok := attribute.Get(condition.AttributeName).(bool)
	if !ok {
		return false
	}
	expectedValue, err := strconv.ParseBool(condition.Value)
	if err != nil {
		return false
	}
	switch condition.Operator {
	case ConditionOperatorEquals:
		return value == expectedValue
	case ConditionOperatorNotEquals:
		return value != expectedValue
	}
	return false
}

// evaluateSegmentRule evaluates whether an attribute matches a segment rule
func (e *engine) evaluateSegmentRule(rule *ParameterRule, attribute *attribute) bool {
	e.logger.Debug("evaluating segment rule", "rule", rule, "segmentID", rule.SegmentID)

	// Check if segment is nil or has no rules
	if rule.Segment == nil {
		e.logger.Debug("segment is nil for rule", "ruleID", rule.ID)
		return false
	}

	if len(rule.Segment.Rules) == 0 {
		e.logger.Debug("no segment rules found", "segmentID", rule.SegmentID)
		return false
	}

	// Evaluate each segment rule - if any rule matches, the segment matches
	for _, segmentRule := range rule.Segment.Rules {
		if e.evaluateSegmentRuleConditions(&segmentRule, attribute) {
			e.logger.Debug("segment rule matched", "segmentRuleID", segmentRule.ID, "segmentID", rule.SegmentID)
			return true
		}
	}

	e.logger.Debug("no segment rules matched", "segmentID", rule.SegmentID)
	return false
}

// evaluateSegmentRuleConditions evaluates all conditions in a segment rule
func (e *engine) evaluateSegmentRuleConditions(segmentRule *SegmentRule, attribute *attribute) bool {
	if len(segmentRule.Conditions) == 0 {
		e.logger.Debug("no conditions found for segment rule", "segmentRuleID", segmentRule.ID)
		return false
	}

	// All conditions must match for the segment rule to match
	matchedConditions := 0
	for _, condition := range segmentRule.Conditions {
		if e.evaluateSegmentRuleCondition(&condition, attribute) {
			matchedConditions++
		}
	}

	return matchedConditions == len(segmentRule.Conditions)
}

// evaluateSegmentRuleCondition evaluates a single segment rule condition
func (e *engine) evaluateSegmentRuleCondition(condition *SegmentRuleCondition, attribute *attribute) bool {
	e.logger.Debug("evaluating segment rule condition", "condition", condition, "attributeDataType", condition.AttributeDataType)

	switch condition.AttributeDataType {
	case "string":
		return e.evaluateSegmentRuleConditionString(condition, attribute)
	case "number":
		return e.evaluateSegmentRuleConditionNumber(condition, attribute)
	case "boolean":
		return e.evaluateSegmentRuleConditionBoolean(condition, attribute)
	case "enum":
		return e.evaluateSegmentRuleConditionEnum(condition, attribute)
	}

	return false
}

// evaluateSegmentRuleConditionString evaluates string-type segment rule conditions
func (e *engine) evaluateSegmentRuleConditionString(condition *SegmentRuleCondition, attribute *attribute) bool {
	e.logger.Debug("evaluating segment rule condition string", "condition", condition, "attribute", attribute)
	value, ok := attribute.Get(condition.AttributeName).(string)
	if !ok {
		return false
	}
	switch condition.Operator {
	case ConditionOperatorEquals:
		return value == condition.Value
	case ConditionOperatorNotEquals:
		return value != condition.Value
	case ConditionOperatorContains:
		return strings.Contains(value, condition.Value)
	case ConditionOperatorNotContains:
		return !strings.Contains(value, condition.Value)
	case ConditionOperatorIn:
		values := strings.Split(condition.Value, ",")
		return slices.Contains(values, value)
	case ConditionOperatorNotIn:
		values := strings.Split(condition.Value, ",")
		return !slices.Contains(values, value)
	}
	return false
}

// evaluateSegmentRuleConditionNumber evaluates number-type segment rule conditions
func (e *engine) evaluateSegmentRuleConditionNumber(condition *SegmentRuleCondition, attribute *attribute) bool {
	value, ok := attribute.Get(condition.AttributeName).(float64)
	if !ok {
		return false
	}
	expectedValue, err := strconv.ParseFloat(condition.Value, 64)
	if err != nil {
		return false
	}
	switch condition.Operator {
	case ConditionOperatorEquals:
		return value == expectedValue
	case ConditionOperatorNotEquals:
		return value != expectedValue
	case ConditionOperatorGreaterThan:
		return value > expectedValue
	case ConditionOperatorLessThan:
		return value < expectedValue
	case ConditionOperatorGreaterThanOrEqual:
		return value >= expectedValue
	case ConditionOperatorLessThanOrEqual:
		return value <= expectedValue
	case ConditionOperatorIn:
		values := strings.Split(condition.Value, ",")
		return slices.Contains(values, strconv.FormatFloat(value, 'f', -1, 64))
	case ConditionOperatorNotIn:
		values := strings.Split(condition.Value, ",")
		return !slices.Contains(values, strconv.FormatFloat(value, 'f', -1, 64))
	}
	return false
}

// evaluateSegmentRuleConditionBoolean evaluates boolean-type segment rule conditions
func (e *engine) evaluateSegmentRuleConditionBoolean(condition *SegmentRuleCondition, attribute *attribute) bool {
	value, ok := attribute.Get(condition.AttributeName).(bool)
	if !ok {
		return false
	}
	expectedValue, err := strconv.ParseBool(condition.Value)
	if err != nil {
		return false
	}
	switch condition.Operator {
	case ConditionOperatorEquals:
		return value == expectedValue
	case ConditionOperatorNotEquals:
		return value != expectedValue
	}
	return false
}

// evaluateSegmentRuleConditionEnum evaluates enum-type segment rule conditions
func (e *engine) evaluateSegmentRuleConditionEnum(condition *SegmentRuleCondition, attribute *attribute) bool {
	e.logger.Debug("evaluating segment rule condition enum", "condition", condition, "attribute", attribute)
	value, ok := attribute.Get(condition.AttributeName).(string)
	if !ok {
		return false
	}
	if !slices.Contains(condition.EnumOptions, value) {
		return false
	}
	return value == condition.Value
}
