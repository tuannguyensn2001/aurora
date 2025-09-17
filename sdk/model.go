package sdk

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
}

type ParameterRuleCondition struct {
	ID                uint              `json:"id"`
	AttributeID       uint              `json:"attributeId"`
	Operator          ConditionOperator `json:"operator"`
	Value             string            `json:"value"`
	AttributeName     string            `json:"attributeName"`
	AttributeDataType string            `json:"attributeDataType"`
}
