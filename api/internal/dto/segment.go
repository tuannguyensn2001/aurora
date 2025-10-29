package dto

import (
	"api/internal/model"
	"time"
)

// CreateSegmentRuleConditionRequest represents the request to create a segment rule condition
type CreateSegmentRuleConditionRequest struct {
	AttributeID uint                    `json:"attributeId" validate:"required"`
	Operator    model.ConditionOperator `json:"operator" validate:"required,oneof=equals not_equals contains not_contains greater_than less_than greater_than_or_equal less_than_or_equal in not_in"`
	Value       string                  `json:"value" validate:"required"`
}

// CreateSegmentRuleRequest represents the request to create a segment rule
type CreateSegmentRuleRequest struct {
	Name        string                              `json:"name" validate:"required"`
	Description string                              `json:"description,omitempty"`
	Conditions  []CreateSegmentRuleConditionRequest `json:"conditions" validate:"required,dive"`
}

// CreateSegmentRequest represents the request to create a segment
type CreateSegmentRequest struct {
	Name        string                     `json:"name" validate:"required"`
	Description string                     `json:"description,omitempty"`
	Rules       []CreateSegmentRuleRequest `json:"rules,omitempty" validate:"dive"`
}

// UpdateSegmentRuleConditionRequest represents the request to update a segment rule condition
type UpdateSegmentRuleConditionRequest struct {
	AttributeID uint                    `json:"attributeId" validate:"required"`
	Operator    model.ConditionOperator `json:"operator" validate:"required,oneof=equals not_equals contains not_contains greater_than less_than greater_than_or_equal less_than_or_equal in not_in"`
	Value       string                  `json:"value" validate:"required"`
}

// UpdateSegmentRuleRequest represents the request to update a segment rule
type UpdateSegmentRuleRequest struct {
	Name        string                              `json:"name" validate:"required"`
	Description string                              `json:"description,omitempty"`
	Conditions  []UpdateSegmentRuleConditionRequest `json:"conditions" validate:"required,dive"`
}

// UpdateSegmentRequest represents the request to update a segment
type UpdateSegmentRequest struct {
	Name        *string                    `json:"name,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Rules       []UpdateSegmentRuleRequest `json:"rules,omitempty" validate:"dive"`
}

// SegmentRuleConditionResponse represents the response for segment rule condition operations
type SegmentRuleConditionResponse struct {
	ID          uint                    `json:"id"`
	AttributeID uint                    `json:"attributeId"`
	Operator    model.ConditionOperator `json:"operator"`
	Value       string                  `json:"value"`
	RuleID      uint                    `json:"ruleId"`
	Attribute   *AttributeResponse      `json:"attribute,omitempty"`
}

// SegmentRuleResponse represents the response for segment rule operations
type SegmentRuleResponse struct {
	ID          uint                           `json:"id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	SegmentID   uint                           `json:"segmentId"`
	Conditions  []SegmentRuleConditionResponse `json:"conditions"`
}

// SegmentResponse represents the response for segment operations
type SegmentResponse struct {
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
	Rules       []SegmentRuleResponse `json:"rules"`
}

// SegmentListResponse represents the response for listing segments
type SegmentListResponse struct {
	Segments []SegmentResponse `json:"segments"`
}

// ToSegmentRuleConditionResponse converts model.SegmentRuleCondition to SegmentRuleConditionResponse
func ToSegmentRuleConditionResponse(condition *model.SegmentRuleCondition) SegmentRuleConditionResponse {
	response := SegmentRuleConditionResponse{
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

// ToSegmentRuleResponse converts model.SegmentRule to SegmentRuleResponse
func ToSegmentRuleResponse(rule *model.SegmentRule) SegmentRuleResponse {
	conditions := make([]SegmentRuleConditionResponse, len(rule.Conditions))
	for i, condition := range rule.Conditions {
		conditions[i] = ToSegmentRuleConditionResponse(&condition)
	}

	return SegmentRuleResponse{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		SegmentID:   rule.SegmentID,
		Conditions:  conditions,
	}
}

// ToSegmentResponse converts model.Segment to SegmentResponse
func ToSegmentResponse(segment *model.Segment) SegmentResponse {
	rules := make([]SegmentRuleResponse, len(segment.Rules))
	for i, rule := range segment.Rules {
		rules[i] = ToSegmentRuleResponse(&rule)
	}

	return SegmentResponse{
		ID:          segment.ID,
		Name:        segment.Name,
		Description: segment.Description,
		CreatedAt:   segment.CreatedAt,
		UpdatedAt:   segment.UpdatedAt,
		Rules:       rules,
	}
}

// ToSegmentListResponse converts slice of model.Segment to SegmentListResponse
func ToSegmentListResponse(segments []*model.Segment) SegmentListResponse {
	responses := make([]SegmentResponse, len(segments))
	for i, segment := range segments {
		responses[i] = ToSegmentResponse(segment)
	}
	return SegmentListResponse{
		Segments: responses,
	}
}

// CheckSegmentOverlapRequest represents the request to check segment overlap
type CheckSegmentOverlapRequest struct {
	SegmentIDs []uint `json:"segmentIds" validate:"required,min=2"`
}

// CheckSegmentOverlapResponse represents the response for segment overlap check
type CheckSegmentOverlapResponse struct {
	Overlap bool `json:"overlap"`
}
