package service

import (
	"api/internal/dto"
	"api/internal/model"
	"api/internal/repository"
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
)

// EventService handles business logic for evaluation events
type EventService struct {
	eventRepo *repository.EventRepository
	logger    zerolog.Logger
}

// NewEventService creates a new event service
func NewEventService(eventRepo *repository.EventRepository, logger zerolog.Logger) *EventService {
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
