package sdk

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/spaolacci/murmur3"
)

// Condition represents a common interface for all condition types
type Condition interface {
	GetAttributeName() string
	GetAttributeDataType() string
	GetOperator() ConditionOperator
	GetValue() string
	GetEnumOptions() []string
}

const bucketSize = 10000

type engine struct {
	logger *slog.Logger
}

func newEngine(logger *slog.Logger) *engine {
	return &engine{
		logger: logger,
	}
}

// evaluateCondition is a unified method to evaluate any condition type
func (e *engine) evaluateCondition(condition Condition, attribute *Attribute) bool {
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
func (e *engine) evaluateStringCondition(condition Condition, attribute *Attribute) bool {
	e.logger.Debug("evaluating string condition", "condition", condition, "attribute", attribute)
	value, ok := attribute.Get(condition.GetAttributeName()).(string)
	if !ok {
		return false
	}
	switch condition.GetOperator() {
	case ConditionOperatorEquals:
		return value == condition.GetValue()
	case ConditionOperatorNotEquals:
		return value != condition.GetValue()
	case ConditionOperatorContains:
		return strings.Contains(value, condition.GetValue())
	case ConditionOperatorNotContains:
		return !strings.Contains(value, condition.GetValue())
	case ConditionOperatorIn:
		values := strings.Split(condition.GetValue(), ",")
		return slices.Contains(values, value)
	case ConditionOperatorNotIn:
		values := strings.Split(condition.GetValue(), ",")
		return !slices.Contains(values, value)
	}
	return false
}

// evaluateNumberCondition evaluates number-type conditions
func (e *engine) evaluateNumberCondition(condition Condition, attribute *Attribute) bool {
	value, ok := attribute.Get(condition.GetAttributeName()).(float64)
	if !ok {
		return false
	}
	expectedValue, err := strconv.ParseFloat(condition.GetValue(), 64)
	if err != nil {
		return false
	}
	switch condition.GetOperator() {
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
		values := strings.Split(condition.GetValue(), ",")
		return slices.Contains(values, strconv.FormatFloat(value, 'f', -1, 64))
	case ConditionOperatorNotIn:
		values := strings.Split(condition.GetValue(), ",")
		return !slices.Contains(values, strconv.FormatFloat(value, 'f', -1, 64))
	}
	return false
}

// evaluateBooleanCondition evaluates boolean-type conditions
func (e *engine) evaluateBooleanCondition(condition Condition, attribute *Attribute) bool {
	value, ok := attribute.Get(condition.GetAttributeName()).(bool)
	if !ok {
		return false
	}
	expectedValue, err := strconv.ParseBool(condition.GetValue())
	if err != nil {
		return false
	}
	switch condition.GetOperator() {
	case ConditionOperatorEquals:
		return value == expectedValue
	case ConditionOperatorNotEquals:
		return value != expectedValue
	}
	return false
}

// evaluateEnumCondition evaluates enum-type conditions
func (e *engine) evaluateEnumCondition(condition Condition, attribute *Attribute) bool {
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

// ExperimentEvaluationResult contains the result of experiment evaluation with metadata
type ExperimentEvaluationResult struct {
	Value          string
	DataType       ParameterDataType
	Success        bool
	ExperimentID   *int
	ExperimentUUID *string
	VariantID      *int
	VariantName    *string
}

func (e *engine) evaluateExperiment(experiment *Experiment, attribute *Attribute, parameterName string) (string, ParameterDataType, bool) {
	if err := experiment.isValid(); err != nil {
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

// evaluateExperimentDetailed returns detailed experiment evaluation result
func (e *engine) evaluateExperimentDetailed(experiment *Experiment, attribute *Attribute, parameterName string) *ExperimentEvaluationResult {
	result := &ExperimentEvaluationResult{
		ExperimentID:   &experiment.ID,
		ExperimentUUID: &experiment.Uuid,
		Success:        false,
	}

	if err := experiment.isValid(); err != nil {
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

func (e *engine) evaluateParameter(parameter *Parameter, attribute *Attribute) string {

	if len(parameter.Rules) == 0 {
		e.logger.Info("no rules found for parameter", "parameter", parameter)
		return parameter.DefaultRolloutValue
	}

	for _, rule := range parameter.Rules {
		count := 0
		if rule.Type == RuleTypeAttribute {
			for _, condition := range rule.Conditions {
				if e.evaluateCondition(&condition, attribute) {
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

// evaluateSegmentRule evaluates whether an attribute matches a segment rule
func (e *engine) evaluateSegmentRule(rule *ParameterRule, attribute *Attribute) bool {
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
func (e *engine) evaluateSegmentRuleConditions(segmentRule *SegmentRule, attribute *Attribute) bool {
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

func (e *engine) inPopulation(key string, start int, end int) bool {
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
