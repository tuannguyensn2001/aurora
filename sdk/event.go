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
	batchConfig BatchConfig
	eventBatch  []EvaluationEvent
	lastFlush   time.Time
	flushTimer  *time.Timer
	flushChan   chan struct{}
}

// NewEventTracker creates a new event tracker
func NewEventTracker(endpointURL, serviceName string, logger *slog.Logger, batchConfig BatchConfig) *EventTracker {
	et := &EventTracker{
		endpointURL: endpointURL,
		serviceName: serviceName,
		logger:      logger,
		batchConfig: batchConfig,
		eventBatch:  make([]EvaluationEvent, 0, batchConfig.MaxSize),
		flushChan:   make(chan struct{}, 1),
	}

	// Start the flush timer
	et.startFlushTimer()

	return et
}

// TrackEvent adds an event to the batch and flushes if necessary
func (et *EventTracker) TrackEvent(ctx context.Context, event *EvaluationEvent) {
	// Add event to batch
	et.eventBatch = append(et.eventBatch, *event)

	// Check if we should flush immediately
	if et.shouldFlush() {
		et.flushEvents(ctx)
	}
}

// shouldFlush determines if the batch should be flushed
func (et *EventTracker) shouldFlush() bool {
	// Check size limits
	if len(et.eventBatch) >= et.batchConfig.FlushSize || len(et.eventBatch) >= et.batchConfig.MaxSize {
		return true
	}

	// Check byte limits
	if et.getBatchSizeBytes() >= et.batchConfig.FlushBytes || et.getBatchSizeBytes() >= et.batchConfig.MaxBytes {
		return true
	}

	return false
}

// getBatchSizeBytes calculates the approximate size of the current batch in bytes
func (et *EventTracker) getBatchSizeBytes() int {
	// Simple approximation - in production, you might want to be more precise
	return len(et.eventBatch) * 200 // Approximate 200 bytes per event
}

// startFlushTimer starts the timer for periodic flushing
func (et *EventTracker) startFlushTimer() {
	et.flushTimer = time.AfterFunc(et.batchConfig.MaxWaitTime, func() {
		select {
		case et.flushChan <- struct{}{}:
		default:
		}
	})
}

// flushEvents sends the current batch to the server
func (et *EventTracker) flushEvents(ctx context.Context) {
	if len(et.eventBatch) == 0 {
		return
	}

	// Create batch request
	batchReq := BatchEventRequest{
		Events: make([]EvaluationEvent, len(et.eventBatch)),
	}
	copy(batchReq.Events, et.eventBatch)

	// Send batch to server
	client := resty.New()
	defer client.Close()

	var response BatchEventResponse
	_, err := client.R().
		SetContext(ctx).
		SetBody(batchReq).
		SetResult(&response).
		Post(et.endpointURL + "/api/v1/sdk/events")

	if err != nil {
		et.logger.ErrorContext(ctx, "failed to track batch events", "error", err, "batchSize", len(et.eventBatch))
	} else {
		et.logger.DebugContext(ctx, "batch events tracked successfully",
			"processed", response.Processed,
			"failed", response.Failed,
			"batchSize", len(et.eventBatch))
	}

	// Clear the batch
	et.eventBatch = et.eventBatch[:0]
	et.lastFlush = time.Now()

	// Reset the timer
	et.flushTimer.Reset(et.batchConfig.MaxWaitTime)
}

// FlushPendingEvents flushes any pending events in the batch
func (et *EventTracker) FlushPendingEvents(ctx context.Context) {
	et.flushEvents(ctx)
}

// Stop stops the event tracker and flushes any pending events
func (et *EventTracker) Stop(ctx context.Context) {
	if et.flushTimer != nil {
		et.flushTimer.Stop()
	}
	et.FlushPendingEvents(ctx)
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

// BatchEventRequest represents a batch of events to be sent to the server
type BatchEventRequest struct {
	Events []EvaluationEvent `json:"events"`
}

// BatchEventResponse represents the response from the server for batch events
type BatchEventResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	Processed    int    `json:"processed"`
	Failed       int    `json:"failed"`
	FailedEvents []int  `json:"failedEvents,omitempty"`
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
