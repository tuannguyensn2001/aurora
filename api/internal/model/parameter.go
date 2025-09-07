package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ParameterDataType represents the enum for parameter data types
type ParameterDataType string

const (
	ParameterDataTypeBoolean ParameterDataType = "boolean"
	ParameterDataTypeString  ParameterDataType = "string"
	ParameterDataTypeNumber  ParameterDataType = "number"
)

// ConditionMatchType represents the enum for condition match types
type ConditionMatchType string

const (
	ConditionMatchTypeMatch    ConditionMatchType = "match"
	ConditionMatchTypeNotMatch ConditionMatchType = "not_match"
)

// RuleType represents the enum for rule types
type RuleType string

const (
	RuleTypeSegment   RuleType = "segment"
	RuleTypeAttribute RuleType = "attribute"
)

// RolloutValue represents a flexible value that can be string, number, or boolean
type RolloutValue struct {
	Data interface{} `json:"value"`
}

// Scan implements the sql.Scanner interface
func (rv *RolloutValue) Scan(value interface{}) error {
	if value == nil {
		rv.Data = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &rv.Data)
	case string:
		return json.Unmarshal([]byte(v), &rv.Data)
	default:
		return errors.New("cannot scan RolloutValue")
	}
}

// Value implements the driver.Valuer interface
func (rv RolloutValue) Value() (driver.Value, error) {
	return json.Marshal(rv.Data)
}

// Parameter represents the parameters table
type Parameter struct {
	ID                  uint                 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                string               `gorm:"uniqueIndex;not null;size:255" json:"name"`
	Description         string               `gorm:"type:text;not null" json:"description"`
	DataType            ParameterDataType    `gorm:"type:parameter_data_type;not null;default:'string'" json:"dataType"`
	DefaultRolloutValue RolloutValue         `gorm:"type:jsonb;not null" json:"defaultRolloutValue"`
	UsageCount          int                  `gorm:"not null;default:0" json:"usageCount"`
	CreatedAt           time.Time            `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time            `gorm:"autoUpdateTime" json:"updatedAt"`
	Conditions          []ParameterCondition `gorm:"foreignKey:ParameterID" json:"conditions"`
	Rules               []ParameterRule      `gorm:"foreignKey:ParameterID" json:"rules"`
}

// TableName specifies the table name for GORM
func (Parameter) TableName() string {
	return "parameters"
}

// BeforeCreate hook to validate parameter
func (p *Parameter) BeforeCreate(tx *gorm.DB) error {
	return p.validate()
}

// BeforeUpdate hook to validate parameter
func (p *Parameter) BeforeUpdate(tx *gorm.DB) error {
	return p.validate()
}

// validate performs validation on the parameter
func (p *Parameter) validate() error {
	// Validate data type
	switch p.DataType {
	case ParameterDataTypeBoolean, ParameterDataTypeString, ParameterDataTypeNumber:
		return nil
	default:
		return gorm.ErrInvalidData
	}
}

// ParameterCondition represents the parameter_conditions table (legacy support)
type ParameterCondition struct {
	ID           uint               `gorm:"primaryKey;autoIncrement" json:"id"`
	ParameterID  uint               `gorm:"not null" json:"parameterId"`
	SegmentID    uint               `gorm:"not null" json:"segmentId"`
	MatchType    ConditionMatchType `gorm:"type:condition_match_type;not null" json:"matchType"`
	RolloutValue RolloutValue       `gorm:"type:jsonb;not null" json:"rolloutValue"`
	Parameter    *Parameter         `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`
	Segment      *Segment           `gorm:"foreignKey:SegmentID" json:"segment,omitempty"`
}

// TableName specifies the table name for GORM
func (ParameterCondition) TableName() string {
	return "parameter_conditions"
}

// BeforeCreate hook to validate parameter condition
func (pc *ParameterCondition) BeforeCreate(tx *gorm.DB) error {
	return pc.validate()
}

// BeforeUpdate hook to validate parameter condition
func (pc *ParameterCondition) BeforeUpdate(tx *gorm.DB) error {
	return pc.validate()
}

// validate performs validation on the parameter condition
func (pc *ParameterCondition) validate() error {
	// Validate match type
	switch pc.MatchType {
	case ConditionMatchTypeMatch, ConditionMatchTypeNotMatch:
		return nil
	default:
		return gorm.ErrInvalidData
	}
}

// ParameterRule represents the parameter_rules table
type ParameterRule struct {
	ID           uint                     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string                   `gorm:"not null;size:255" json:"name"`
	Description  string                   `gorm:"type:text" json:"description"`
	Type         RuleType                 `gorm:"type:rule_type;not null" json:"type"`
	RolloutValue RolloutValue             `gorm:"type:jsonb;not null" json:"rolloutValue"`
	ParameterID  uint                     `gorm:"not null" json:"parameterId"`
	SegmentID    *uint                    `gorm:"" json:"segmentId,omitempty"`
	MatchType    *ConditionMatchType      `gorm:"type:condition_match_type" json:"matchType,omitempty"`
	Parameter    *Parameter               `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`
	Segment      *Segment                 `gorm:"foreignKey:SegmentID" json:"segment,omitempty"`
	Conditions   []ParameterRuleCondition `gorm:"foreignKey:RuleID" json:"conditions"`
}

// TableName specifies the table name for GORM
func (ParameterRule) TableName() string {
	return "parameter_rules"
}

// BeforeCreate hook to validate parameter rule
func (pr *ParameterRule) BeforeCreate(tx *gorm.DB) error {
	return pr.validate()
}

// BeforeUpdate hook to validate parameter rule
func (pr *ParameterRule) BeforeUpdate(tx *gorm.DB) error {
	return pr.validate()
}

// validate performs validation on the parameter rule
func (pr *ParameterRule) validate() error {
	// Validate rule type
	switch pr.Type {
	case RuleTypeSegment, RuleTypeAttribute:
		// For segment rules, validate segment ID and match type are provided
		if pr.Type == RuleTypeSegment {
			if pr.SegmentID == nil || pr.MatchType == nil {
				return gorm.ErrInvalidData
			}
			// Validate match type
			switch *pr.MatchType {
			case ConditionMatchTypeMatch, ConditionMatchTypeNotMatch:
				// Valid
			default:
				return gorm.ErrInvalidData
			}
		}
		return nil
	default:
		return gorm.ErrInvalidData
	}
}

// ParameterRuleCondition represents the parameter_rule_conditions table
type ParameterRuleCondition struct {
	ID          uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	AttributeID uint              `gorm:"not null" json:"attributeId"`
	Operator    ConditionOperator `gorm:"type:condition_operator;not null" json:"operator"`
	Value       string            `gorm:"type:text;not null" json:"value"`
	RuleID      uint              `gorm:"not null" json:"ruleId"`
	Rule        *ParameterRule    `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	Attribute   *Attribute        `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
}

// TableName specifies the table name for GORM
func (ParameterRuleCondition) TableName() string {
	return "parameter_rule_conditions"
}

// BeforeCreate hook to validate parameter rule condition
func (prc *ParameterRuleCondition) BeforeCreate(tx *gorm.DB) error {
	return prc.validate()
}

// BeforeUpdate hook to validate parameter rule condition
func (prc *ParameterRuleCondition) BeforeUpdate(tx *gorm.DB) error {
	return prc.validate()
}

// validate performs validation on the parameter rule condition
func (prc *ParameterRuleCondition) validate() error {
	// Validate that operator is one of the allowed values
	switch prc.Operator {
	case ConditionOperatorEquals, ConditionOperatorNotEquals,
		ConditionOperatorContains, ConditionOperatorNotContains,
		ConditionOperatorGreaterThan, ConditionOperatorLessThan,
		ConditionOperatorGreaterThanOrEqual, ConditionOperatorLessThanOrEqual,
		ConditionOperatorIn, ConditionOperatorNotIn:
		return nil
	default:
		return gorm.ErrInvalidData
	}
}
