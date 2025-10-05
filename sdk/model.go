package sdk

import "time"

type ParameterDataType string

const (
	ParameterDataTypeBoolean ParameterDataType = "boolean"
	ParameterDataTypeString  ParameterDataType = "string"
	ParameterDataTypeNumber  ParameterDataType = "number"
)

type Parameter struct {
	Name                string            `json:"name"`
	DataType            ParameterDataType `json:"dataType"`
	DefaultRolloutValue string            `json:"defaultRolloutValue"`
	Rules               []ParameterRule   `json:"rules"`
}

type ConditionMatchType string

const (
	ConditionMatchTypeMatch    ConditionMatchType = "match"
	ConditionMatchTypeNotMatch ConditionMatchType = "not_match"
)

type RuleType string

const (
	RuleTypeSegment   RuleType = "segment"
	RuleTypeAttribute RuleType = "attribute"
)

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

type ParameterRule struct {
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	Type         RuleType                 `json:"type"`
	RolloutValue string                   `json:"rolloutValue"`
	SegmentID    int64                    `json:"segmentId"`
	MatchType    ConditionMatchType       `json:"matchType"`
	Conditions   []ParameterRuleCondition `json:"conditions"`
	Segment      *Segment                 `json:"segment,omitempty"`
}

type SegmentRuleCondition struct {
	ID                uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	AttributeID       uint              `gorm:"not null" json:"attributeId"`
	Operator          ConditionOperator `gorm:"type:condition_operator;not null" json:"operator"`
	Value             string            `gorm:"type:text;not null" json:"value"`
	AttributeName     string            `json:"attributeName"`
	AttributeDataType string            `json:"attributeDataType"`
	EnumOptions       []string          `json:"enumOptions"`
}

type SegmentRule struct {
	ID          uint                   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string                 `gorm:"not null;size:255" json:"name"`
	Description string                 `gorm:"type:text" json:"description"`
	SegmentID   uint                   `gorm:"not null" json:"segmentId"`
	Conditions  []SegmentRuleCondition `gorm:"foreignKey:RuleID" json:"conditions"`
}

type Segment struct {
	ID          uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string        `gorm:"uniqueIndex;not null;size:255" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`
	Rules       []SegmentRule `gorm:"foreignKey:SegmentID" json:"rules"`
}

type ParameterRuleCondition struct {
	ID                uint              `json:"id"`
	AttributeID       uint              `json:"attributeId"`
	Operator          ConditionOperator `json:"operator"`
	Value             string            `json:"value"`
	AttributeName     string            `json:"attributeName"`
	AttributeDataType string            `json:"attributeDataType"`
	EnumOptions       []string          `json:"enumOptions"`
}
