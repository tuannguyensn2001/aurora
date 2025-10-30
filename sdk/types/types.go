package types

import "time"

// ParameterDataType represents the data type of a parameter
type ParameterDataType string

const (
	ParameterDataTypeBoolean ParameterDataType = "boolean"
	ParameterDataTypeString  ParameterDataType = "string"
	ParameterDataTypeNumber  ParameterDataType = "number"
)

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

// EventType represents the type of event being tracked
type EventType string

const (
	EventTypeParameterEvaluation  EventType = "parameter_evaluation"
	EventTypeExperimentEvaluation EventType = "experiment_evaluation"
)

// ExperimentStatus represents the status of an experiment
const (
	ExperimentStatusDraft    = "draft"
	ExperimentStatusSchedule = "schedule"
	ExperimentStatusRunning  = "running"
	ExperimentStatusFinish   = "finish"
	ExperimentStatusCancel   = "cancel"
	ExperimentStatusAbort    = "abort"
)

// Parameter represents an experiment parameter with rules and conditions
type Parameter struct {
	Name                string            `json:"name"`
	DataType            ParameterDataType `json:"dataType"`
	DefaultRolloutValue string            `json:"defaultRolloutValue"`
	Rules               []ParameterRule   `json:"rules"`
}

// ParameterRule represents a rule for parameter evaluation
type ParameterRule struct {
	ID           uint               `json:"id"`
	Type         RuleType           `json:"type"`
	MatchType    ConditionMatchType `json:"matchType"`
	RolloutValue string             `json:"rolloutValue"`
	SegmentID    *uint              `json:"segmentId,omitempty"`
	Segment      *Segment           `json:"segment,omitempty"`
	Conditions   []RuleCondition    `gorm:"foreignKey:RuleID" json:"conditions"`
}

// RuleCondition represents a condition within a rule
type RuleCondition struct {
	ID                uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleID            uint              `gorm:"not null" json:"ruleId"`
	AttributeName     string            `gorm:"not null;size:255" json:"attributeName"`
	AttributeDataType string            `gorm:"not null;size:50" json:"attributeDataType"`
	Operator          ConditionOperator `gorm:"not null;size:50" json:"operator"`
	Value             string            `gorm:"type:text" json:"value"`
	EnumOptions       []string          `gorm:"type:text" json:"enumOptions"`
}

// GetAttributeName returns the attribute name
func (rc *RuleCondition) GetAttributeName() string {
	return rc.AttributeName
}

// GetAttributeDataType returns the attribute data type
func (rc *RuleCondition) GetAttributeDataType() string {
	return rc.AttributeDataType
}

// GetOperator returns the operator
func (rc *RuleCondition) GetOperator() ConditionOperator {
	return rc.Operator
}

// GetValue returns the value
func (rc *RuleCondition) GetValue() string {
	return rc.Value
}

// GetEnumOptions returns the enum options
func (rc *RuleCondition) GetEnumOptions() []string {
	return rc.EnumOptions
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

// SegmentRule represents a rule within a segment
type SegmentRule struct {
	ID         uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	SegmentID  uint            `gorm:"not null" json:"segmentId"`
	Conditions []RuleCondition `gorm:"foreignKey:RuleID" json:"conditions"`
}

// Experiment represents an A/B test experiment
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

// ExperimentVariant represents a variant within an experiment
type ExperimentVariant struct {
	ID                int                          `json:"id"`
	ExperimentID      int                          `json:"experimentId"`
	Name              string                       `json:"name"`
	Description       string                       `json:"description"`
	TrafficAllocation int                          `json:"trafficAllocation"`
	Parameters        []ExperimentVariantParameter `json:"parameters"`
}

// ExperimentVariantParameter represents a parameter within an experiment variant
type ExperimentVariantParameter struct {
	ID                int               `json:"id"`
	ExperimentID      int               `json:"experimentId"`
	VariantID         int               `json:"variantId"`
	ParameterID       int               `json:"parameterId"`
	ParameterName     string            `json:"parameterName"`
	ParameterDataType ParameterDataType `json:"parameterDataType"`
	RolloutValue      string            `json:"rolloutValue"`
}

// EvaluationEvent represents an event that occurred during parameter or experiment evaluation
type EvaluationEvent struct {
	ID             string                 `json:"id"`
	ServiceName    string                 `json:"serviceName"`
	EventType      EventType              `json:"eventType"`
	ParameterName  string                 `json:"parameterName"`
	Source         string                 `json:"source"` // "parameter" or "experiment"
	UserAttributes map[string]interface{} `json:"userAttributes"`
	RolloutValue   *string                `json:"rolloutValue,omitempty"`
	Error          *string                `json:"error,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	ExperimentID   *int                   `json:"experimentId,omitempty"`
	ExperimentUUID *string                `json:"experimentUuid,omitempty"`
	VariantID      *int                   `json:"variantId,omitempty"`
	VariantName    *string                `json:"variantName,omitempty"`
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

// MetadataResponse represents the response from the metadata API
type MetadataResponse struct {
	EnableS3 bool `json:"enableS3"`
}

// UpstreamParametersResponse represents the response from the upstream parameters API
type UpstreamParametersResponse struct {
	Parameters []Parameter `json:"parameters"`
}

// UpstreamExperimentsResponse represents the response from the upstream experiments API
type UpstreamExperimentsResponse struct {
	Experiments []Experiment `json:"experiments"`
}

// IsValid validates that an experiment is valid for evaluation
func (e *Experiment) IsValid() error {
	// Add validation logic here
	// For now, just return nil (valid)
	return nil
}

// BatchConfig holds configuration for event batching
type BatchConfig struct {
	MaxSize     int           // Maximum number of events per batch
	MaxBytes    int           // Maximum bytes per batch
	MaxWaitTime time.Duration // Maximum wait time before flushing batch
	FlushSize   int           // Size at which to flush batch immediately
	FlushBytes  int           // Bytes at which to flush batch immediately
}

// ClientOptions holds the required configuration options for the client
type ClientOptions struct {
	S3BucketName string
	EndpointURL  string
	ServiceName  string
}
