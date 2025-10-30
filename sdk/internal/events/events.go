package events

import (
	"context"
	"fmt"
	"math/rand"
	"sdk/pkg/logger"
	"sdk/types"
	"time"
)

// EventTracker interface defines event tracking operations
type EventTracker interface {
	TrackEvent(ctx context.Context, event types.EvaluationEvent)
	CreateParameterEvaluationEvent(parameterName string, attribute Attribute, rolloutValue *string, err error) types.EvaluationEvent
	CreateExperimentEvaluationEvent(parameterName string, attribute Attribute, rolloutValue *string, err error, experimentID *int, experimentUUID *string, variantID *int, variantName *string) types.EvaluationEvent
	Stop(ctx context.Context)
}

// Attribute interface for dependency injection
type Attribute interface {
	Get(key string) interface{}
	ToMap() map[string]interface{}
}

// EventSender interface for sending events
type EventSender interface {
	SendEvents(ctx context.Context, events []types.EvaluationEvent) error
}

// BatchEventTracker implements EventTracker with batching
type BatchEventTracker struct {
	endpointURL string
	serviceName string
	logger      logger.Logger
	batchConfig types.BatchConfig
	eventBatch  []types.EvaluationEvent
	lastFlush   time.Time
	flushTimer  *time.Timer
	flushChan   chan struct{}
	sender      EventSender
}

// NewBatchEventTracker creates a new batch event tracker
func NewBatchEventTracker(endpointURL, serviceName string, logger logger.Logger, batchConfig types.BatchConfig, sender EventSender) EventTracker {
	return &BatchEventTracker{
		endpointURL: endpointURL,
		serviceName: serviceName,
		logger:      logger,
		batchConfig: batchConfig,
		eventBatch:  make([]types.EvaluationEvent, 0, batchConfig.MaxSize),
		lastFlush:   time.Now(),
		flushChan:   make(chan struct{}, 1),
		sender:      sender,
	}
}

// TrackEvent adds an event to the batch
func (t *BatchEventTracker) TrackEvent(ctx context.Context, event types.EvaluationEvent) {
	t.eventBatch = append(t.eventBatch, event)

	// Check if we should flush immediately
	if len(t.eventBatch) >= t.batchConfig.FlushSize {
		t.flushEvents(ctx)
		return
	}

	// Start timer if not already running
	if t.flushTimer == nil {
		t.flushTimer = time.AfterFunc(t.batchConfig.MaxWaitTime, func() {
			select {
			case t.flushChan <- struct{}{}:
			default:
			}
		})
	}
}

// CreateParameterEvaluationEvent creates a parameter evaluation event
func (t *BatchEventTracker) CreateParameterEvaluationEvent(parameterName string, attribute Attribute, rolloutValue *string, err error) types.EvaluationEvent {
	event := types.EvaluationEvent{
		ID:             generateEventID(),
		ServiceName:    t.serviceName,
		EventType:      types.EventTypeParameterEvaluation,
		ParameterName:  parameterName,
		Source:         "parameter",
		UserAttributes: attribute.ToMap(),
		RolloutValue:   rolloutValue,
		Timestamp:      time.Now(),
	}

	if err != nil {
		errStr := err.Error()
		event.Error = &errStr
	}

	return event
}

// CreateExperimentEvaluationEvent creates an experiment evaluation event
func (t *BatchEventTracker) CreateExperimentEvaluationEvent(parameterName string, attribute Attribute, rolloutValue *string, err error, experimentID *int, experimentUUID *string, variantID *int, variantName *string) types.EvaluationEvent {
	event := types.EvaluationEvent{
		ID:             generateEventID(),
		ServiceName:    t.serviceName,
		EventType:      types.EventTypeExperimentEvaluation,
		ParameterName:  parameterName,
		Source:         "experiment",
		UserAttributes: attribute.ToMap(),
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

// Stop stops the event tracker and flushes any pending events
func (t *BatchEventTracker) Stop(ctx context.Context) {
	if t.flushTimer != nil {
		t.flushTimer.Stop()
	}
	t.flushEvents(ctx)
	close(t.flushChan)
}

// flushEvents sends the current batch of events
func (t *BatchEventTracker) flushEvents(ctx context.Context) {
	if len(t.eventBatch) == 0 {
		return
	}

	events := make([]types.EvaluationEvent, len(t.eventBatch))
	copy(events, t.eventBatch)
	t.eventBatch = t.eventBatch[:0]

	if t.flushTimer != nil {
		t.flushTimer.Stop()
		t.flushTimer = nil
	}

	if err := t.sender.SendEvents(ctx, events); err != nil {
		t.logger.Error("failed to send events", "error", err)
	}
	
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("event_%d_%d", time.Now().UnixNano(), rand.Int63())
}
