package service

import (
	"api/internal/dto"
	"api/internal/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
)

// EventRepositoryInterface defines the interface for event repository
type EventRepositoryInterface interface {
	CreateEvent(ctx context.Context, event *model.EvaluationEvent) error
	CreateEventsBatch(ctx context.Context, events []*model.EvaluationEvent) error
	GetEventsByServiceName(ctx context.Context, serviceName string, limit, offset int) ([]model.EvaluationEvent, error)
	GetEventsByParameterName(ctx context.Context, parameterName string, limit, offset int) ([]model.EvaluationEvent, error)
	GetEventsByExperimentID(ctx context.Context, experimentID int, limit, offset int) ([]model.EvaluationEvent, error)
	GetEventStats(ctx context.Context, serviceName string) (map[string]interface{}, error)
}

// EventService handles business logic for evaluation events
type EventService struct {
	eventRepo EventRepositoryInterface
	logger    zerolog.Logger
}

// NewEventService creates a new event service
func NewEventService(eventRepo EventRepositoryInterface, logger zerolog.Logger) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// TrackEvent tracks an evaluation event
func (s *EventService) TrackEvent(ctx context.Context, req *dto.TrackEventRequest) (*dto.TrackEventResponse, error) {
	// Convert user attributes to JSON string
	userAttributesJSON, err := json.Marshal(req.UserAttributes)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to marshal user attributes")
		return &dto.TrackEventResponse{
			Success: false,
			Message: "failed to process user attributes",
		}, err
	}

	// Create event model
	event := &model.EvaluationEvent{
		EventID:        req.ID,
		ServiceName:    req.ServiceName,
		EventType:      string(req.EventType),
		ParameterName:  req.ParameterName,
		Source:         req.Source,
		UserAttributes: string(userAttributesJSON),
		RolloutValue:   req.RolloutValue,
		Error:          req.Error,
		Timestamp:      req.Timestamp,
		ExperimentID:   req.ExperimentID,
		ExperimentUUID: req.ExperimentUUID,
		VariantID:      req.VariantID,
		VariantName:    req.VariantName,
	}

	// Save to database
	err = s.eventRepo.CreateEvent(ctx, event)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create event")
		return &dto.TrackEventResponse{
			Success: false,
			Message: "failed to save event",
		}, err
	}

	s.logger.Info().
		Str("eventId", req.ID).
		Str("serviceName", req.ServiceName).
		Str("eventType", string(req.EventType)).
		Str("parameterName", req.ParameterName).
		Msg("event tracked successfully")

	return &dto.TrackEventResponse{
		Success: true,
		Message: "event tracked successfully",
	}, nil
}

// GetEventsByServiceName retrieves events for a specific service
func (s *EventService) GetEventsByServiceName(ctx context.Context, serviceName string, limit, offset int) ([]model.EvaluationEvent, error) {
	return s.eventRepo.GetEventsByServiceName(ctx, serviceName, limit, offset)
}

// GetEventsByParameterName retrieves events for a specific parameter
func (s *EventService) GetEventsByParameterName(ctx context.Context, parameterName string, limit, offset int) ([]model.EvaluationEvent, error) {
	return s.eventRepo.GetEventsByParameterName(ctx, parameterName, limit, offset)
}

// GetEventsByExperimentID retrieves events for a specific experiment
func (s *EventService) GetEventsByExperimentID(ctx context.Context, experimentID int, limit, offset int) ([]model.EvaluationEvent, error) {
	return s.eventRepo.GetEventsByExperimentID(ctx, experimentID, limit, offset)
}

// GetEventStats retrieves statistics about events
func (s *EventService) GetEventStats(ctx context.Context, serviceName string) (map[string]interface{}, error) {
	return s.eventRepo.GetEventStats(ctx, serviceName)
}

// TrackBatchEvent tracks multiple evaluation events in batch
func (s *EventService) TrackBatchEvent(ctx context.Context, req *dto.TrackBatchEventRequest) (*dto.TrackBatchEventResponse, error) {
	if len(req.Events) == 0 {
		return &dto.TrackBatchEventResponse{
			Success: false,
			Message: "no events provided",
		}, nil
	}

	// Convert events to models
	events := make([]*model.EvaluationEvent, 0, len(req.Events))
	failedEvents := make([]int, 0)

	for i, eventReq := range req.Events {
		// Convert user attributes to JSON string
		userAttributesJSON, err := json.Marshal(eventReq.UserAttributes)
		if err != nil {
			s.logger.Error().Err(err).Int("eventIndex", i).Msg("failed to marshal user attributes")
			failedEvents = append(failedEvents, i)
			continue
		}

		// Create event model
		event := &model.EvaluationEvent{
			EventID:        eventReq.ID,
			ServiceName:    eventReq.ServiceName,
			EventType:      string(eventReq.EventType),
			ParameterName:  eventReq.ParameterName,
			Source:         eventReq.Source,
			UserAttributes: string(userAttributesJSON),
			RolloutValue:   eventReq.RolloutValue,
			Error:          eventReq.Error,
			Timestamp:      eventReq.Timestamp,
			ExperimentID:   eventReq.ExperimentID,
			ExperimentUUID: eventReq.ExperimentUUID,
			VariantID:      eventReq.VariantID,
			VariantName:    eventReq.VariantName,
		}
		events = append(events, event)
	}

	// Save batch to database
	err := s.eventRepo.CreateEventsBatch(ctx, events)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create events batch")
		return &dto.TrackBatchEventResponse{
			Success: false,
			Message: "failed to save events batch",
		}, err
	}

	processed := len(events)
	failed := len(failedEvents)

	s.logger.Info().
		Int("processed", processed).
		Int("failed", failed).
		Msg("batch events tracked successfully")

	return &dto.TrackBatchEventResponse{
		Success:      true,
		Message:      fmt.Sprintf("processed %d events, %d failed", processed, failed),
		Processed:    processed,
		Failed:       failed,
		FailedEvents: failedEvents,
	}, nil
}
