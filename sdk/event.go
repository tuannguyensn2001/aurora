package sdk

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"resty.dev/v3"
)

// EventType represents the type of event being tracked
type EventType string

const (
	EventTypeParameterEvaluation  EventType = "parameter_evaluation"
	EventTypeExperimentEvaluation EventType = "experiment_evaluation"
)

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

// EventTracker handles sending events to the server
type EventTracker struct {
	endpointURL string
	serviceName string
	logger      *slog.Logger
}

// NewEventTracker creates a new event tracker
func NewEventTracker(endpointURL, serviceName string, logger *slog.Logger) *EventTracker {
	return &EventTracker{
		endpointURL: endpointURL,
		serviceName: serviceName,
		logger:      logger,
	}
}

// TrackEvent sends an event to the server
func (et *EventTracker) TrackEvent(ctx context.Context, event *EvaluationEvent) {
	// Send event asynchronously to avoid blocking the main evaluation flow
	// go func() {
	client := resty.New()
	defer client.Close()

	_, err := client.R().
		SetContext(ctx).
		SetBody(event).
		Post(et.endpointURL + "/api/v1/sdk/events")

	if err != nil {
		et.logger.ErrorContext(ctx, "failed to track event", "error", err, "event", event)
	} else {
		et.logger.DebugContext(ctx, "event tracked successfully", "event", event)
	}
	// }()
}

// CreateParameterEvaluationEvent creates an event for parameter evaluation
func (et *EventTracker) CreateParameterEvaluationEvent(
	parameterName string,
	attributes *Attribute,
	rolloutValue *string,
	err error,
) *EvaluationEvent {
	event := &EvaluationEvent{
		ID:             generateEventID(),
		ServiceName:    et.serviceName,
		EventType:      EventTypeParameterEvaluation,
		ParameterName:  parameterName,
		Source:         "parameter",
		UserAttributes: attributes.ToMap(),
		RolloutValue:   rolloutValue,
		Timestamp:      time.Now(),
	}

	if err != nil {
		errStr := err.Error()
		event.Error = &errStr
	}

	return event
}

// CreateExperimentEvaluationEvent creates an event for experiment evaluation
func (et *EventTracker) CreateExperimentEvaluationEvent(
	parameterName string,
	attributes *Attribute,
	rolloutValue *string,
	err error,
	experimentID *int,
	experimentUUID *string,
	variantID *int,
	variantName *string,
) *EvaluationEvent {
	event := &EvaluationEvent{
		ID:             generateEventID(),
		ServiceName:    et.serviceName,
		EventType:      EventTypeExperimentEvaluation,
		ParameterName:  parameterName,
		Source:         "experiment",
		UserAttributes: attributes.ToMap(),
		RolloutValue:   rolloutValue,
		Timestamp:      time.Now(),
		ExperimentID:   experimentID,
		ExperimentUUID: experimentUUID,
		VariantID:      variantID,
		VariantName:    variantName,
	}

	if err != nil {
		errStr := err.Error()
		event.Error = &errStr
	}

	return event
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
