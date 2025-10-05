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
