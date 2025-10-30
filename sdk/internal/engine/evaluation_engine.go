package engine

import (
	"fmt"
	"sdk/pkg/logger"
	"sdk/types"
	"slices"
	"strconv"
	"strings"

	"github.com/spaolacci/murmur3"
)

const bucketSize = 10000

// EvaluationEngine implements the Engine interface
type EvaluationEngine struct {
	logger logger.Logger
}

// NewEvaluationEngine creates a new evaluation engine
func NewEvaluationEngine(logger logger.Logger) Engine {
	return &EvaluationEngine{
		logger: logger,
	}
}

// EvaluateParameter evaluates a parameter against the given attributes
func (e *EvaluationEngine) EvaluateParameter(parameter *types.Parameter, attribute Attribute) string {
	if len(parameter.Rules) == 0 {
		e.logger.Info("no rules found for parameter", "parameter", parameter)
		return parameter.DefaultRolloutValue
	}

	for _, rule := range parameter.Rules {
		count := 0
		if rule.Type == types.RuleTypeAttribute {
			for _, condition := range rule.Conditions {
				if e.evaluateCondition(&condition, attribute) {
					count++
				}
			}
		} else if rule.Type == types.RuleTypeSegment {
			if rule.MatchType == types.ConditionMatchTypeMatch && e.evaluateSegmentRule(&rule, attribute) {
				return rule.RolloutValue
			} else if rule.MatchType == types.ConditionMatchTypeNotMatch && !e.evaluateSegmentRule(&rule, attribute) {
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

// EvaluateExperiment evaluates an experiment and returns the result
func (e *EvaluationEngine) EvaluateExperiment(experiment *types.Experiment, attribute Attribute, parameterName string) (string, types.ParameterDataType, bool) {
	if err := experiment.IsValid(); err != nil {
		e.logger.Debug("experiment is invalid", "experiment", experiment, "error", err)
		return "", "", false
	}

	if experiment.Segment != nil {
		rules := experiment.Segment.Rules
		passRule := 0
		for _, rule := range rules {
			if e.evaluateSegmentRuleConditions(&rule, attribute) {
				passRule++
				break
			}
		}
		if passRule == 0 {
			e.logger.Debug("no rules passed", "experiment", experiment)
			return "", "", false
		}
	}

	valuePopulation := fmt.Sprintf("%v", attribute.Get(experiment.HashAttributeName))
	keyPopulation := fmt.Sprintf("experiment:population:%s:%s", experiment.Uuid, valuePopulation)
	inPopulation := e.inPopulation(keyPopulation, 0, experiment.PopulationSize)
	if !inPopulation {
		e.logger.Debug("not in population", "experiment", experiment)
		return "", "", false
	}

	trafficAllocation := make([]int, 0)
	for _, variant := range experiment.Variants {
		if len(trafficAllocation) == 0 {
			trafficAllocation = append(trafficAllocation, variant.TrafficAllocation)
		} else {
			trafficAllocation = append(trafficAllocation, trafficAllocation[len(trafficAllocation)-1]+variant.TrafficAllocation)
		}
	}
	index := -1
	valueHash := fmt.Sprintf("experiment:hash:%s:%s", experiment.Uuid, valuePopulation)
	for i, allocation := range trafficAllocation {
		check := false
		if i == 0 {
			check = e.inPopulation(valueHash, 0, allocation)
		} else if i == len(trafficAllocation)-1 {
			check = e.inPopulation(valueHash, trafficAllocation[i-1], 100)
		} else {
			check = e.inPopulation(valueHash, trafficAllocation[i-1], allocation)
		}
		if check {
			index = i
			break
		}
	}
	if index == -1 {
		e.logger.Debug("not in traffic allocation", "experiment", experiment)
		return "", "", false
	}
	variant := experiment.Variants[index]
	for _, parameter := range variant.Parameters {
		if parameter.ParameterName == parameterName {
			return parameter.RolloutValue, parameter.ParameterDataType, true
		}
	}

	return "", "", false
}

// EvaluateExperimentDetailed returns detailed experiment evaluation result
func (e *EvaluationEngine) EvaluateExperimentDetailed(experiment *types.Experiment, attribute Attribute, parameterName string) *types.ExperimentEvaluationResult {
	result := &types.ExperimentEvaluationResult{
		ExperimentID:   &experiment.ID,
		ExperimentUUID: &experiment.Uuid,
		Success:        false,
	}

	if err := experiment.IsValid(); err != nil {
		e.logger.Debug("experiment is invalid", "experiment", experiment, "error", err)
		return result
	}

	if experiment.Segment != nil {
		rules := experiment.Segment.Rules
		passRule := 0
		for _, rule := range rules {
			if e.evaluateSegmentRuleConditions(&rule, attribute) {
				passRule++
				break
			}
		}
		if passRule == 0 {
			e.logger.Debug("no rules passed", "experiment", experiment)
			return result
		}
	}

	valuePopulation := fmt.Sprintf("%v", attribute.Get(experiment.HashAttributeName))
	keyPopulation := fmt.Sprintf("experiment:population:%s:%s", experiment.Uuid, valuePopulation)
	inPopulation := e.inPopulation(keyPopulation, 0, experiment.PopulationSize)
	if !inPopulation {
		e.logger.Debug("not in population", "experiment", experiment)
		return result
	}

	trafficAllocation := make([]int, 0)
	for _, variant := range experiment.Variants {
		if len(trafficAllocation) == 0 {
			trafficAllocation = append(trafficAllocation, variant.TrafficAllocation)
		} else {
			trafficAllocation = append(trafficAllocation, trafficAllocation[len(trafficAllocation)-1]+variant.TrafficAllocation)
		}
	}
	index := -1
	valueHash := fmt.Sprintf("experiment:hash:%s:%s", experiment.Uuid, valuePopulation)
	for i, allocation := range trafficAllocation {
		check := false
		if i == 0 {
			check = e.inPopulation(valueHash, 0, allocation)
		} else if i == len(trafficAllocation)-1 {
			check = e.inPopulation(valueHash, trafficAllocation[i-1], 100)
		} else {
			check = e.inPopulation(valueHash, trafficAllocation[i-1], allocation)
		}
		if check {
			index = i
			break
		}
	}
	if index == -1 {
		e.logger.Debug("not in traffic allocation", "experiment", experiment)
		return result
	}

	variant := experiment.Variants[index]
	result.VariantID = &variant.ID
	result.VariantName = &variant.Name

	for _, parameter := range variant.Parameters {
		if parameter.ParameterName == parameterName {
			result.Value = parameter.RolloutValue
			result.DataType = parameter.ParameterDataType
			result.Success = true
			return result
		}
	}

	return result
}

// evaluateCondition is a unified method to evaluate any condition type
func (e *EvaluationEngine) evaluateCondition(condition Condition, attribute Attribute) bool {
	e.logger.Debug("evaluating condition", "dataType", condition.GetAttributeDataType(), "condition", condition, "attribute", attribute)
	switch condition.GetAttributeDataType() {
	case "string":
		return e.evaluateStringCondition(condition, attribute)
	case "number":
		return e.evaluateNumberCondition(condition, attribute)
	case "boolean":
		return e.evaluateBooleanCondition(condition, attribute)
	case "enum":
		return e.evaluateEnumCondition(condition, attribute)
	}
	return false
}

// evaluateStringCondition evaluates string-type conditions
func (e *EvaluationEngine) evaluateStringCondition(condition Condition, attribute Attribute) bool {
	e.logger.Debug("evaluating string condition", "condition", condition, "attribute", attribute)
	value, ok := attribute.Get(condition.GetAttributeName()).(string)
	if !ok {
		return false
	}
	switch condition.GetOperator() {
	case types.ConditionOperatorEquals:
		return value == condition.GetValue()
	case types.ConditionOperatorNotEquals:
		return value != condition.GetValue()
	case types.ConditionOperatorContains:
		return strings.Contains(value, condition.GetValue())
	case types.ConditionOperatorNotContains:
		return !strings.Contains(value, condition.GetValue())
	case types.ConditionOperatorIn:
		values := strings.Split(condition.GetValue(), ",")
		return slices.Contains(values, value)
	case types.ConditionOperatorNotIn:
		values := strings.Split(condition.GetValue(), ",")
		return !slices.Contains(values, value)
	}
	return false
}

// evaluateNumberCondition evaluates number-type conditions
func (e *EvaluationEngine) evaluateNumberCondition(condition Condition, attribute Attribute) bool {
	value, ok := attribute.Get(condition.GetAttributeName()).(float64)
	if !ok {
		return false
	}
	expectedValue, err := strconv.ParseFloat(condition.GetValue(), 64)
	if err != nil {
		return false
	}
	switch condition.GetOperator() {
	case types.ConditionOperatorEquals:
		return value == expectedValue
	case types.ConditionOperatorNotEquals:
		return value != expectedValue
	case types.ConditionOperatorGreaterThan:
		return value > expectedValue
	case types.ConditionOperatorLessThan:
		return value < expectedValue
	case types.ConditionOperatorGreaterThanOrEqual:
		return value >= expectedValue
	case types.ConditionOperatorLessThanOrEqual:
		return value <= expectedValue
	case types.ConditionOperatorIn:
		values := strings.Split(condition.GetValue(), ",")
		return slices.Contains(values, strconv.FormatFloat(value, 'f', -1, 64))
	case types.ConditionOperatorNotIn:
		values := strings.Split(condition.GetValue(), ",")
		return !slices.Contains(values, strconv.FormatFloat(value, 'f', -1, 64))
	}
	return false
}

// evaluateBooleanCondition evaluates boolean-type conditions
func (e *EvaluationEngine) evaluateBooleanCondition(condition Condition, attribute Attribute) bool {
	value, ok := attribute.Get(condition.GetAttributeName()).(bool)
	if !ok {
		return false
	}
	expectedValue, err := strconv.ParseBool(condition.GetValue())
	if err != nil {
		return false
	}
	switch condition.GetOperator() {
	case types.ConditionOperatorEquals:
		return value == expectedValue
	case types.ConditionOperatorNotEquals:
		return value != expectedValue
	}
	return false
}

// evaluateEnumCondition evaluates enum-type conditions
func (e *EvaluationEngine) evaluateEnumCondition(condition Condition, attribute Attribute) bool {
	e.logger.Debug("evaluating enum condition", "condition", condition, "attribute", attribute)
	value, ok := attribute.Get(condition.GetAttributeName()).(string)
	if !ok {
		return false
	}
	if !slices.Contains(condition.GetEnumOptions(), value) {
		return false
	}
	values := strings.Split(condition.GetValue(), ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	return slices.Contains(values, value)
}

// evaluateSegmentRule evaluates whether an attribute matches a segment rule
func (e *EvaluationEngine) evaluateSegmentRule(rule *types.ParameterRule, attribute Attribute) bool {
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
func (e *EvaluationEngine) evaluateSegmentRuleConditions(segmentRule *types.SegmentRule, attribute Attribute) bool {
	if len(segmentRule.Conditions) == 0 {
		e.logger.Debug("no conditions found for segment rule", "segmentRuleID", segmentRule.ID)
		return false
	}

	// All conditions must match for the segment rule to match
	matchedConditions := 0
	for _, condition := range segmentRule.Conditions {
		if e.evaluateCondition(&condition, attribute) {
			matchedConditions++
		}
	}

	return matchedConditions == len(segmentRule.Conditions)
}

// inPopulation checks if a key falls within the specified population range
func (e *EvaluationEngine) inPopulation(key string, start int, end int) bool {
	if start > end {
		return false
	}
	if start < 0 || end > 100 {
		return false
	}

	hashValue := murmur3.Sum64([]byte(key))
	bucket := hashValue % bucketSize
	thresholdLeft := uint64(start * (bucketSize / 100))
	thresholdRight := uint64(end * (bucketSize / 100))
	return bucket < thresholdRight && bucket >= thresholdLeft
}
