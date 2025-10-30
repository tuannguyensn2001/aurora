package model

import (
	"time"
)

// EventType represents the type of event being tracked
type EventType string

const (
	EventTypeParameterEvaluation  EventType = "parameter_evaluation"
	EventTypeExperimentEvaluation EventType = "experiment_evaluation"
)

// EvaluationEvent represents an evaluation event stored in the database
type EvaluationEvent struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID        string    `gorm:"uniqueIndex;not null;size:255" json:"eventId"`
	ServiceName    string    `gorm:"not null;size:255" json:"serviceName"`
	EventType      string    `gorm:"not null;size:50" json:"eventType"`
	ParameterName  string    `gorm:"not null;size:255" json:"parameterName"`
	Source         string    `gorm:"not null;size:50" json:"source"`
	UserAttributes string    `gorm:"type:jsonb" json:"userAttributes"` // JSON string
	RolloutValue   *string   `gorm:"type:text" json:"rolloutValue,omitempty"`
	Error          *string   `gorm:"type:text" json:"error,omitempty"`
	Timestamp      time.Time `gorm:"not null" json:"timestamp"`
	ExperimentID   *int      `json:"experimentId,omitempty"`
	ExperimentUUID *string   `gorm:"size:255" json:"experimentUuid,omitempty"`
	VariantID      *int      `json:"variantId,omitempty"`
	VariantName    *string   `gorm:"size:255" json:"variantName,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
