package repository

import (
	"api/internal/model"
	"context"

	"gorm.io/gorm"
)

// EventRepository handles database operations for evaluation events
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// CreateEvent creates a new evaluation event
func (r *EventRepository) CreateEvent(ctx context.Context, event *model.EvaluationEvent) error {
	// Convert user attributes to JSON string
	if event.UserAttributes != "" {
		// Already a JSON string, no conversion needed
	} else {
		// This shouldn't happen if called from service layer, but handle gracefully
		event.UserAttributes = "{}"
	}

	return r.db.WithContext(ctx).Create(event).Error
}

// GetEventsByServiceName retrieves events for a specific service
func (r *EventRepository) GetEventsByServiceName(ctx context.Context, serviceName string, limit, offset int) ([]model.EvaluationEvent, error) {
	var events []model.EvaluationEvent
	err := r.db.WithContext(ctx).
		Where("service_name = ?", serviceName).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

// GetEventsByParameterName retrieves events for a specific parameter
func (r *EventRepository) GetEventsByParameterName(ctx context.Context, parameterName string, limit, offset int) ([]model.EvaluationEvent, error) {
	var events []model.EvaluationEvent
	err := r.db.WithContext(ctx).
		Where("parameter_name = ?", parameterName).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

// GetEventsByExperimentID retrieves events for a specific experiment
func (r *EventRepository) GetEventsByExperimentID(ctx context.Context, experimentID int, limit, offset int) ([]model.EvaluationEvent, error) {
	var events []model.EvaluationEvent
	err := r.db.WithContext(ctx).
		Where("experiment_id = ?", experimentID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

// GetEventStats retrieves statistics about events
func (r *EventRepository) GetEventStats(ctx context.Context, serviceName string) (map[string]interface{}, error) {
	var stats map[string]interface{}

	// Get total events count
	var totalEvents int64
	err := r.db.WithContext(ctx).
		Model(&model.EvaluationEvent{}).
		Where("service_name = ?", serviceName).
		Count(&totalEvents).Error
	if err != nil {
		return nil, err
	}

	// Get events by type
	var parameterEvents, experimentEvents int64
	err = r.db.WithContext(ctx).
		Model(&model.EvaluationEvent{}).
		Where("service_name = ? AND event_type = ?", serviceName, string(model.EventTypeParameterEvaluation)).
		Count(&parameterEvents).Error
	if err != nil {
		return nil, err
	}

	err = r.db.WithContext(ctx).
		Model(&model.EvaluationEvent{}).
		Where("service_name = ? AND event_type = ?", serviceName, string(model.EventTypeExperimentEvaluation)).
		Count(&experimentEvents).Error
	if err != nil {
		return nil, err
	}

	stats = map[string]interface{}{
		"totalEvents":      totalEvents,
		"parameterEvents":  parameterEvents,
		"experimentEvents": experimentEvents,
	}

	return stats, nil
}
