package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ParameterChangeRequestStatus represents the status of a parameter change request
type ParameterChangeRequestStatus string

const (
	ChangeRequestStatusPending   ParameterChangeRequestStatus = "pending"
	ChangeRequestStatusApproved  ParameterChangeRequestStatus = "approved"
	ChangeRequestStatusRejected  ParameterChangeRequestStatus = "rejected"
	ChangeRequestStatusCancelled ParameterChangeRequestStatus = "cancelled"
)

// ParameterChangeData holds the proposed changes to a parameter
type ParameterChangeData struct {
	Name                *string                `json:"name,omitempty"`
	Description         *string                `json:"description,omitempty"`
	DataType            *ParameterDataType     `json:"dataType,omitempty"`
	DefaultRolloutValue interface{}            `json:"defaultRolloutValue,omitempty"`
	Rules               []ParameterRuleRequest `json:"rules,omitempty"`
}

// ParameterCurrentConfig holds the current configuration of a parameter at the time of change request
type ParameterCurrentConfig struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	DataType            ParameterDataType      `json:"dataType"`
	DefaultRolloutValue interface{}            `json:"defaultRolloutValue"`
	Rules               []ParameterRuleRequest `json:"rules,omitempty"`
}

// ParameterRuleRequest represents a rule in the change request
type ParameterRuleRequest struct {
	Name         string                          `json:"name"`
	Description  string                          `json:"description,omitempty"`
	Type         RuleType                        `json:"type"`
	RolloutValue interface{}                     `json:"rolloutValue"`
	SegmentID    *uint                           `json:"segmentId,omitempty"`
	MatchType    *ConditionMatchType             `json:"matchType,omitempty"`
	Conditions   []ParameterRuleConditionRequest `json:"conditions,omitempty"`
}

// ParameterRuleConditionRequest represents a condition in a rule
type ParameterRuleConditionRequest struct {
	AttributeID uint              `json:"attributeId"`
	Operator    ConditionOperator `json:"operator"`
	Value       string            `json:"value"`
}

// Scan implements the sql.Scanner interface for ParameterChangeData
func (pcd *ParameterChangeData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, pcd)
	case string:
		return json.Unmarshal([]byte(v), pcd)
	default:
		return errors.New("cannot scan ParameterChangeData")
	}
}

// Value implements the driver.Valuer interface for ParameterChangeData
func (pcd ParameterChangeData) Value() (driver.Value, error) {
	return json.Marshal(pcd)
}

// Scan implements the sql.Scanner interface for ParameterCurrentConfig
func (pcc *ParameterCurrentConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, pcc)
	case string:
		return json.Unmarshal([]byte(v), pcc)
	default:
		return errors.New("cannot scan ParameterCurrentConfig")
	}
}

// Value implements the driver.Valuer interface for ParameterCurrentConfig
func (pcc ParameterCurrentConfig) Value() (driver.Value, error) {
	return json.Marshal(pcc)
}

// ParameterChangeRequest represents a request to change a parameter
type ParameterChangeRequest struct {
	ID                uint                         `gorm:"primaryKey;autoIncrement" json:"id"`
	ParameterID       uint                         `gorm:"not null" json:"parameterId"`
	RequestedByUserID uint                         `gorm:"not null" json:"requestedByUserId"`
	Status            ParameterChangeRequestStatus `gorm:"type:parameter_change_request_status;not null;default:'pending'" json:"status"`
	Description       string                       `gorm:"type:text" json:"description"`
	ChangeData        ParameterChangeData          `gorm:"type:jsonb;not null;column:change_data" json:"changeData"`
	CurrentConfig     ParameterCurrentConfig       `gorm:"type:jsonb;not null;column:current_config" json:"currentConfig"`
	ReviewedByUserID  *uint                        `json:"reviewedByUserId,omitempty"`
	ReviewedAt        *time.Time                   `json:"reviewedAt,omitempty"`
	CreatedAt         time.Time                    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time                    `gorm:"autoUpdateTime" json:"updatedAt"`
	Parameter         *Parameter                   `gorm:"foreignKey:ParameterID" json:"parameter,omitempty"`
	RequestedByUser   *User                        `gorm:"foreignKey:RequestedByUserID" json:"requestedByUser,omitempty"`
	ReviewedByUser    *User                        `gorm:"foreignKey:ReviewedByUserID" json:"reviewedByUser,omitempty"`
}

// TableName specifies the table name for GORM
func (ParameterChangeRequest) TableName() string {
	return "parameter_change_requests"
}

// BeforeCreate hook to validate parameter change request
func (pcr *ParameterChangeRequest) BeforeCreate(tx *gorm.DB) error {
	return pcr.validate()
}

// BeforeUpdate hook to validate parameter change request
func (pcr *ParameterChangeRequest) BeforeUpdate(tx *gorm.DB) error {
	return pcr.validate()
}

// validate performs validation on the parameter change request
func (pcr *ParameterChangeRequest) validate() error {
	// Validate status
	switch pcr.Status {
	case ChangeRequestStatusPending, ChangeRequestStatusApproved, ChangeRequestStatusRejected, ChangeRequestStatusCancelled:
		return nil
	default:
		return gorm.ErrInvalidData
	}
}
