package model

import (
	"time"

	"gorm.io/gorm"
)

// ConditionOperator represents the enum for segment rule condition operators
type ConditionOperator string

const (
	ConditionOperatorEquals             ConditionOperator = "equals"
	ConditionOperatorNotEquals          ConditionOperator = "not_equals"
	ConditionOperatorContains           ConditionOperator = "contains"
	ConditionOperatorNotContains        ConditionOperator = "not_contains"
	ConditionOperatorGreaterThan        ConditionOperator = "greater_than"
	ConditionOperatorLessThan           ConditionOperator = "less_than"
	ConditionOperatorGreaterThanOrEqual ConditionOperator = "greater_than_or_equal"
	ConditionOperatorLessThanOrEqual    ConditionOperator = "less_than_or_equal"
	ConditionOperatorIn                 ConditionOperator = "in"
	ConditionOperatorNotIn              ConditionOperator = "not_in"
)

// Segment represents the segments table
type Segment struct {
	ID          uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string        `gorm:"uniqueIndex;not null;size:255" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`
	Rules       []SegmentRule `gorm:"foreignKey:SegmentID" json:"rules"`
}

// TableName specifies the table name for GORM
func (Segment) TableName() string {
	return "segments"
}

// SegmentRule represents the segment_rules table
type SegmentRule struct {
	ID          uint                   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string                 `gorm:"not null;size:255" json:"name"`
	Description string                 `gorm:"type:text" json:"description"`
	SegmentID   uint                   `gorm:"not null" json:"segmentId"`
	Segment     *Segment               `gorm:"foreignKey:SegmentID" json:"segment,omitempty"`
	Conditions  []SegmentRuleCondition `gorm:"foreignKey:RuleID" json:"conditions"`
}

// TableName specifies the table name for GORM
func (SegmentRule) TableName() string {
	return "segment_rules"
}

// SegmentRuleCondition represents the segment_rule_conditions table
type SegmentRuleCondition struct {
	ID          uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	AttributeID uint              `gorm:"not null" json:"attributeId"`
	Operator    ConditionOperator `gorm:"type:condition_operator;not null" json:"operator"`
	Value       string            `gorm:"type:text;not null" json:"value"`
	RuleID      uint              `gorm:"not null" json:"ruleId"`
	Rule        *SegmentRule      `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	Attribute   *Attribute        `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
}

// TableName specifies the table name for GORM
func (SegmentRuleCondition) TableName() string {
	return "segment_rule_conditions"
}

// BeforeCreate hook to validate segment rule condition
func (src *SegmentRuleCondition) BeforeCreate(tx *gorm.DB) error {
	return src.validate()
}

// BeforeUpdate hook to validate segment rule condition
func (src *SegmentRuleCondition) BeforeUpdate(tx *gorm.DB) error {
	return src.validate()
}

// validate performs validation on the segment rule condition
func (src *SegmentRuleCondition) validate() error {
	// Validate that operator is one of the allowed values
	switch src.Operator {
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
