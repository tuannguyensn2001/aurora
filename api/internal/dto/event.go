package dto

import "time"

// EventType represents the type of event being tracked
type EventType string

const (
	EventTypeParameterEvaluation  EventType = "parameter_evaluation"
	EventTypeExperimentEvaluation EventType = "experiment_evaluation"
)

// TrackEventRequest represents the request to track an evaluation event
type TrackEventRequest struct {
	ID             string                 `json:"id" binding:"required"`
	ServiceName    string                 `json:"serviceName" binding:"required"`
	EventType      EventType              `json:"eventType" binding:"required"`
	ParameterName  string                 `json:"parameterName" binding:"required"`
	Source         string                 `json:"source" binding:"required"`
	UserAttributes map[string]interface{} `json:"userAttributes" binding:"required"`
	RolloutValue   *string                `json:"rolloutValue,omitempty"`
	Error          *string                `json:"error,omitempty"`
	Timestamp      time.Time              `json:"timestamp" binding:"required"`
	ExperimentID   *int                   `json:"experimentId,omitempty"`
	ExperimentUUID *string                `json:"experimentUuid,omitempty"`
	VariantID      *int                   `json:"variantId,omitempty"`
	VariantName    *string                `json:"variantName,omitempty"`
}

// TrackEventResponse represents the response after tracking an event
type TrackEventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
