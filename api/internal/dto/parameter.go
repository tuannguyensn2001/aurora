package dto

import (
	"api/internal/model"
	"time"
)

// CreateParameterRuleConditionRequest represents the request to create a parameter rule condition
type CreateParameterRuleConditionRequest struct {
	AttributeID uint                    `json:"attributeId" validate:"required"`
	Operator    model.ConditionOperator `json:"operator" validate:"required,oneof=equals not_equals contains not_contains greater_than less_than greater_than_or_equal less_than_or_equal in not_in"`
	Value       string                  `json:"value" validate:"required"`
}

// CreateParameterRuleRequest represents the request to create a parameter rule
type CreateParameterRuleRequest struct {
	Name         string                                `json:"name" validate:"required"`
	Description  string                                `json:"description,omitempty"`
	Type         model.RuleType                        `json:"type" validate:"required,oneof=segment attribute"`
	RolloutValue interface{}                           `json:"rolloutValue" validate:"required"`
	SegmentID    *uint                                 `json:"segmentId,omitempty"`
	MatchType    *model.ConditionMatchType             `json:"matchType,omitempty"`
	Conditions   []CreateParameterRuleConditionRequest `json:"conditions,omitempty" validate:"dive"`
}

// CreateParameterRequest represents the request to create a parameter
type CreateParameterRequest struct {
	Name                string                  `json:"name" validate:"required"`
	Description         string                  `json:"description" validate:"required"`
	DataType            model.ParameterDataType `json:"dataType" validate:"required,oneof=boolean string number"`
	DefaultRolloutValue interface{}             `json:"defaultRolloutValue" validate:"required"`
}

// UpdateParameterRuleConditionRequest represents the request to update a parameter rule condition
type UpdateParameterRuleConditionRequest struct {
	AttributeID uint                    `json:"attributeId" validate:"required"`
	Operator    model.ConditionOperator `json:"operator" validate:"required,oneof=equals not_equals contains not_contains greater_than less_than greater_than_or_equal less_than_or_equal in not_in"`
	Value       string                  `json:"value" validate:"required"`
}

// UpdateParameterRuleRequest represents the request to update a parameter rule
type UpdateParameterRuleRequest struct {
	Name         *string                               `json:"name,omitempty"`
	Description  *string                               `json:"description,omitempty"`
	Type         *model.RuleType                       `json:"type,omitempty"`
	RolloutValue interface{}                           `json:"rolloutValue,omitempty"`
	SegmentID    *uint                                 `json:"segmentId,omitempty"`
	MatchType    *model.ConditionMatchType             `json:"matchType,omitempty"`
	Conditions   []UpdateParameterRuleConditionRequest `json:"conditions,omitempty" validate:"dive"`
}

// UpdateParameterRequest represents the request to update a parameter
type UpdateParameterRequest struct {
	Name                *string                  `json:"name,omitempty"`
	Description         *string                  `json:"description,omitempty"`
	DataType            *model.ParameterDataType `json:"dataType,omitempty"`
	DefaultRolloutValue interface{}              `json:"defaultRolloutValue,omitempty"`
}

// UpdateParameterWithRulesRequest represents the comprehensive request to update a parameter with all its rules
type UpdateParameterWithRulesRequest struct {
	Name                *string                      `json:"name,omitempty"`
	Description         *string                      `json:"description,omitempty"`
	DataType            *model.ParameterDataType     `json:"dataType,omitempty"`
	DefaultRolloutValue interface{}                  `json:"defaultRolloutValue,omitempty"`
	Rules               []CreateParameterRuleRequest `json:"rules,omitempty" validate:"dive"`
}

// ParameterRuleConditionResponse represents the response for parameter rule condition operations
type ParameterRuleConditionResponse struct {
	ID          uint                    `json:"id"`
	AttributeID uint                    `json:"attributeId"`
	Operator    model.ConditionOperator `json:"operator"`
	Value       string                  `json:"value"`
	RuleID      uint                    `json:"ruleId"`
	Attribute   *AttributeResponse      `json:"attribute,omitempty"`
}

// ParameterRuleResponse represents the response for parameter rule operations
type ParameterRuleResponse struct {
	ID           uint                             `json:"id"`
	Name         string                           `json:"name"`
	Description  string                           `json:"description"`
	Type         model.RuleType                   `json:"type"`
	RolloutValue interface{}                      `json:"rolloutValue"`
	ParameterID  uint                             `json:"parameterId"`
	SegmentID    *uint                            `json:"segmentId,omitempty"`
	MatchType    *model.ConditionMatchType        `json:"matchType,omitempty"`
	Segment      *SegmentResponse                 `json:"segment,omitempty"`
	Conditions   []ParameterRuleConditionResponse `json:"conditions"`
}

// ParameterConditionResponse represents the response for parameter condition operations (legacy)
type ParameterConditionResponse struct {
	ID           uint                     `json:"id"`
	ParameterID  uint                     `json:"parameterId"`
	SegmentID    uint                     `json:"segmentId"`
	MatchType    model.ConditionMatchType `json:"matchType"`
	RolloutValue interface{}              `json:"rolloutValue"`
	Segment      *SegmentResponse         `json:"segment,omitempty"`
}

// ParameterResponse represents the response for parameter operations
type ParameterResponse struct {
	ID                  uint                         `json:"id"`
	Name                string                       `json:"name"`
	Description         string                       `json:"description"`
	DataType            model.ParameterDataType      `json:"dataType"`
	DefaultRolloutValue interface{}                  `json:"defaultRolloutValue"`
	UsageCount          int                          `json:"usageCount"`
	CreatedAt           time.Time                    `json:"createdAt"`
	UpdatedAt           time.Time                    `json:"updatedAt"`
	Conditions          []ParameterConditionResponse `json:"conditions"`
	Rules               []ParameterRuleResponse      `json:"rules"`
}

// ParameterListResponse represents the response for listing parameters
type ParameterListResponse struct {
	Parameters []ParameterResponse `json:"parameters"`
}

// ToParameterRuleConditionResponse converts model.ParameterRuleCondition to ParameterRuleConditionResponse
func ToParameterRuleConditionResponse(condition *model.ParameterRuleCondition) ParameterRuleConditionResponse {
	response := ParameterRuleConditionResponse{
		ID:          condition.ID,
		AttributeID: condition.AttributeID,
		Operator:    condition.Operator,
		Value:       condition.Value,
		RuleID:      condition.RuleID,
	}

	if condition.Attribute != nil {
		attr := ToAttributeResponse(condition.Attribute)
		response.Attribute = &attr
	}

	return response
}

// ToParameterRuleResponse converts model.ParameterRule to ParameterRuleResponse
func ToParameterRuleResponse(rule *model.ParameterRule) ParameterRuleResponse {
	conditions := make([]ParameterRuleConditionResponse, len(rule.Conditions))
	for i, condition := range rule.Conditions {
		conditions[i] = ToParameterRuleConditionResponse(&condition)
	}

	response := ParameterRuleResponse{
		ID:           rule.ID,
		Name:         rule.Name,
		Description:  rule.Description,
		Type:         rule.Type,
		RolloutValue: rule.RolloutValue.Data,
		ParameterID:  rule.ParameterID,
		SegmentID:    rule.SegmentID,
		MatchType:    rule.MatchType,
		Conditions:   conditions,
	}

	if rule.Segment != nil {
		segment := ToSegmentResponse(rule.Segment)
		response.Segment = &segment
	}

	return response
}

// ToParameterConditionResponse converts model.ParameterCondition to ParameterConditionResponse
func ToParameterConditionResponse(condition *model.ParameterCondition) ParameterConditionResponse {
	response := ParameterConditionResponse{
		ID:           condition.ID,
		ParameterID:  condition.ParameterID,
		SegmentID:    condition.SegmentID,
		MatchType:    condition.MatchType,
		RolloutValue: condition.RolloutValue.Data,
	}

	if condition.Segment != nil {
		segment := ToSegmentResponse(condition.Segment)
		response.Segment = &segment
	}

	return response
}

// ToParameterResponse converts model.Parameter to ParameterResponse
func ToParameterResponse(parameter *model.Parameter) ParameterResponse {
	conditions := make([]ParameterConditionResponse, len(parameter.Conditions))
	for i, condition := range parameter.Conditions {
		conditions[i] = ToParameterConditionResponse(&condition)
	}

	rules := make([]ParameterRuleResponse, len(parameter.Rules))
	for i, rule := range parameter.Rules {
		rules[i] = ToParameterRuleResponse(&rule)
	}

	return ParameterResponse{
		ID:                  parameter.ID,
		Name:                parameter.Name,
		Description:         parameter.Description,
		DataType:            parameter.DataType,
		DefaultRolloutValue: parameter.DefaultRolloutValue.Data,
		UsageCount:          parameter.UsageCount,
		CreatedAt:           parameter.CreatedAt,
		UpdatedAt:           parameter.UpdatedAt,
		Conditions:          conditions,
		Rules:               rules,
	}
}

// ToParameterListResponse converts slice of model.Parameter to ParameterListResponse
func ToParameterListResponse(parameters []*model.Parameter) ParameterListResponse {
	responses := make([]ParameterResponse, len(parameters))
	for i, parameter := range parameters {
		responses[i] = ToParameterResponse(parameter)
	}
	return ParameterListResponse{
		Parameters: responses,
	}
}

type SimulateParameterRequest struct {
	ParameterName string                     `json:"parameterName" validate:"required"`
	ParameterType model.ParameterDataType    `json:"parameterType" validate:"required,oneof=boolean string number"`
	Attributes    []SimulateAttributeRequest `json:"attributes" validate:"required"`
}

type SimulateAttributeRequest struct {
	DataType model.DataType `json:"dataType" validate:"required,oneof=boolean string number"`
	Value    string         `json:"value" validate:"required"`
	Name     string         `json:"name" validate:"required"`
}

type SimulateParameterResponse struct {
	Value interface{} `json:"value"`
}
