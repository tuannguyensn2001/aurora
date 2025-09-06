package handler

import (
	"context"

	"api/internal/dto"
	"api/internal/service"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	service service.Service
}

func New(svc service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) HealthCheck(ctx context.Context) (string, error) {
	logger := log.Ctx(ctx).With().Str("handler", "health-check").Logger()
	logger.Info().Msg("Health check")
	return "OK", nil
}

// CreateAttribute handles the business logic for creating an attribute
func (h *Handler) CreateAttribute(ctx context.Context, req *dto.CreateAttributeRequest) (*dto.AttributeResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "create-attribute").Logger()
	logger.Info().Msg("Creating attribute")

	attr, err := h.service.CreateAttribute(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create attribute")
		return nil, err
	}

	response := dto.ToAttributeResponse(attr)
	return &response, nil
}

// GetAllAttributes handles the business logic for getting all attributes
func (h *Handler) GetAllAttributes(ctx context.Context) ([]dto.AttributeResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-all-attributes").Logger()
	logger.Info().Msg("Getting all attributes")

	attrs, err := h.service.GetAllAttributes(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all attributes")
		return nil, err
	}

	responses := make([]dto.AttributeResponse, len(attrs))
	for i, attr := range attrs {
		responses[i] = dto.ToAttributeResponse(attr)
	}

	return responses, nil
}

// GetAttributeByID handles the business logic for getting an attribute by ID
func (h *Handler) GetAttributeByID(ctx context.Context, id uint) (*dto.AttributeResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-attribute-by-id").Uint("id", id).Logger()
	logger.Info().Msg("Getting attribute by ID")

	attr, err := h.service.GetAttributeByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to get attribute by ID")
		return nil, err
	}

	response := dto.ToAttributeResponse(attr)
	return &response, nil
}

// UpdateAttribute handles the business logic for updating an attribute
func (h *Handler) UpdateAttribute(ctx context.Context, id uint, req *dto.UpdateAttributeRequest) (*dto.AttributeResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "update-attribute").Uint("id", id).Logger()
	logger.Info().Msg("Updating attribute")

	attr, err := h.service.UpdateAttribute(ctx, id, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to update attribute")
		return nil, err
	}

	response := dto.ToAttributeResponse(attr)
	return &response, nil
}

// DeleteAttribute handles the business logic for deleting an attribute
func (h *Handler) DeleteAttribute(ctx context.Context, id uint) error {
	logger := log.Ctx(ctx).With().Str("handler", "delete-attribute").Uint("id", id).Logger()
	logger.Info().Msg("Deleting attribute")

	err := h.service.DeleteAttribute(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to delete attribute")
		return err
	}

	return nil
}

// IncrementAttributeUsageCount handles the business logic for incrementing attribute usage count
func (h *Handler) IncrementAttributeUsageCount(ctx context.Context, id uint) error {
	logger := log.Ctx(ctx).With().Str("handler", "increment-attribute-usage").Uint("id", id).Logger()
	logger.Info().Msg("Incrementing attribute usage count")

	err := h.service.IncrementAttributeUsageCount(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to increment attribute usage count")
		return err
	}

	return nil
}

// DecrementAttributeUsageCount handles the business logic for decrementing attribute usage count
func (h *Handler) DecrementAttributeUsageCount(ctx context.Context, id uint) error {
	logger := log.Ctx(ctx).With().Str("handler", "decrement-attribute-usage").Uint("id", id).Logger()
	logger.Info().Msg("Decrementing attribute usage count")

	err := h.service.DecrementAttributeUsageCount(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to decrement attribute usage count")
		return err
	}

	return nil
}
