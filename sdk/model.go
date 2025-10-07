package sdk

import (
	"errors"
	"time"
)

// ParameterDataType represents the data type of a parameter
type ParameterDataType string

const (
	ParameterDataTypeBoolean ParameterDataType = "boolean"
	ParameterDataTypeString  ParameterDataType = "string"
	ParameterDataTypeNumber  ParameterDataType = "number"
)

// Parameter represents an experiment parameter with rules and conditions
type Parameter struct {
	Name                string            `json:"name"`
	DataType            ParameterDataType `json:"dataType"`
	DefaultRolloutValue string            `json:"defaultRolloutValue"`
	Rules               []ParameterRule   `json:"rules"`
}

// ConditionMatchType represents how conditions should be matched
type ConditionMatchType string

const (
	ConditionMatchTypeMatch    ConditionMatchType = "match"
	ConditionMatchTypeNotMatch ConditionMatchType = "not_match"
)

// RuleType represents the type of rule (segment or attribute)
type RuleType string

const (
	RuleTypeSegment   RuleType = "segment"
	RuleTypeAttribute RuleType = "attribute"
)

// ConditionOperator represents operators used in rule conditions
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

// ParameterRule represents a rule for parameter evaluation
type ParameterRule struct {
	ID           uint               `json:"id"`
	Name         string             `json:"name"`
	Type         RuleType           `json:"type"`
	RolloutValue string             `json:"rolloutValue"`
	SegmentID    int64              `json:"segmentId"`
	MatchType    ConditionMatchType `json:"matchType"`
	Conditions   []RuleCondition    `json:"conditions"`
	Segment      *Segment           `json:"segment,omitempty"`
}

// RuleCondition represents a unified condition structure for both parameter and segment rules
type RuleCondition struct {
	ID                uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	AttributeID       uint              `gorm:"not null" json:"attributeId"`
	Operator          ConditionOperator `gorm:"type:condition_operator;not null" json:"operator"`
	Value             string            `gorm:"type:text;not null" json:"value"`
	AttributeName     string            `json:"attributeName"`
	AttributeDataType string            `json:"attributeDataType"`
	EnumOptions       []string          `json:"enumOptions"`
}

// Implement the Condition interface for RuleCondition
func (rc *RuleCondition) GetAttributeName() string {
	return rc.AttributeName
}
func (rc *RuleCondition) GetAttributeDataType() string {
	return rc.AttributeDataType
}
func (rc *RuleCondition) GetOperator() ConditionOperator {
	return rc.Operator
}
func (rc *RuleCondition) GetValue() string {
	return rc.Value
}
func (rc *RuleCondition) GetEnumOptions() []string {
	return rc.EnumOptions
}

// SegmentRule represents a rule within a segment
type SegmentRule struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string          `gorm:"not null;size:255" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	SegmentID   uint            `gorm:"not null" json:"segmentId"`
	Conditions  []RuleCondition `gorm:"foreignKey:RuleID" json:"conditions"`
}

// Segment represents a segment with its associated rules
type Segment struct {
	ID          uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string        `gorm:"uniqueIndex;not null;size:255" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`
	Rules       []SegmentRule `gorm:"foreignKey:SegmentID" json:"rules"`
}

type Experiment struct {
	ID                int                 `json:"id"`
	Name              string              `json:"name"`
	Uuid              string              `json:"uuid"`
	StartDate         int64               `json:"startDate"`
	EndDate           int64               `json:"endDate"`
	HashAttributeID   int                 `json:"hashAttributeId"`
	PopulationSize    int                 `json:"populationSize"`
	Strategy          string              `json:"strategy"`
	Status            string              `json:"status"`
	SegmentID         int                 `json:"segmentId"`
	Segment           *Segment            `json:"segment,omitempty"`
	Variants          []ExperimentVariant `json:"variants"`
	HashAttributeName string              `json:"hashAttributeName"`
}

type ExperimentVariant struct {
	ID                int                          `json:"id"`
	ExperimentID      int                          `json:"experimentId"`
	Name              string                       `json:"name"`
	Description       string                       `json:"description"`
	TrafficAllocation int                          `json:"trafficAllocation"`
	Parameters        []ExperimentVariantParameter `json:"parameters"`
}

type ExperimentVariantParameter struct {
	ID                int               `json:"id"`
	ParameterDataType ParameterDataType `json:"parameterDataType"`
	ParameterID       int               `json:"parameterId"`
	ParameterName     string            `json:"parameterName"`
	RolloutValue      string            `json:"rolloutValue"`
}

func (e *Experiment) isValid() error {
	if e.HashAttributeName == "" {
		return errors.New("hash attribute name is empty")
	}
	now := time.Now().Unix()
	if now < e.StartDate || now > e.EndDate {
		return errors.New("experiment is not active")
	}
	if e.StartDate >= e.EndDate {
		return errors.New("start date must be before end date")
	}
	if len(e.Uuid) == 0 {
		return errors.New("uuid is empty")
	}
	if e.PopulationSize <= 0 || e.PopulationSize > 100 {
		return errors.New("population size is invalid")
	}
	if e.Strategy != "percentage_split" {
		return errors.New("strategy is invalid")
	}
	if (e.Status != ExperimentStatusSchedule) && (e.Status != ExperimentStatusRunning) {
		return errors.New("status is invalid")
	}

	if len(e.Variants) == 0 {
		return errors.New("variants are empty")
	}

	totalTrafficAllocation := 0
	for _, variant := range e.Variants {
		totalTrafficAllocation += variant.TrafficAllocation
	}
	if totalTrafficAllocation != 100 {
		return errors.New("total traffic allocation must be 100")
	}

	for _, variant := range e.Variants {
		if variant.TrafficAllocation <= 0 || variant.TrafficAllocation > 100 {
			return errors.New("traffic allocation is invalid")
		}
		if len(variant.Parameters) == 0 {
			return errors.New("parameters are empty")
		}

		for _, parameter := range variant.Parameters {
			if len(parameter.ParameterName) == 0 {
				return errors.New("parameter name is empty")
			}
			if parameter.ParameterDataType != ParameterDataTypeString && parameter.ParameterDataType != ParameterDataTypeNumber && parameter.ParameterDataType != ParameterDataTypeBoolean {
				return errors.New("parameter data type is invalid")
			}
		}

	}
	return nil

}
