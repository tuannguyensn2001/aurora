package handler

import (
	"api/internal/dto"
	"api/internal/model"
	"context"

	"github.com/rs/zerolog/log"
)

// CreateParameterChangeRequest handles creating a parameter change request
func (h *Handler) CreateParameterChangeRequest(ctx context.Context, userID uint, req *dto.CreateParameterChangeRequestRequest) (*dto.ParameterChangeRequestResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "create-parameter-change-request").Logger()
	logger.Info().Msg("Creating parameter change request")

	changeRequest, err := h.service.CreateParameterChangeRequest(ctx, userID, req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create parameter change request")
		return nil, err
	}

	response := dto.ToParameterChangeRequestResponse(changeRequest)
	return &response, nil
}

// GetParameterChangeRequestByID handles getting a parameter change request by ID
func (h *Handler) GetParameterChangeRequestByID(ctx context.Context, id uint) (*dto.ParameterChangeRequestResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-parameter-change-request-by-id").Uint("id", id).Logger()
	logger.Info().Msg("Getting parameter change request by ID")

	changeRequest, err := h.service.GetParameterChangeRequestByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get parameter change request")
		return nil, err
	}

	response := dto.ToParameterChangeRequestResponse(changeRequest)
	return &response, nil
}

// GetPendingParameterChangeRequestByParameterID handles getting pending change request for a parameter
func (h *Handler) GetPendingParameterChangeRequestByParameterID(ctx context.Context, parameterID uint) (*dto.ParameterChangeRequestResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-pending-parameter-change-request").Uint("parameterId", parameterID).Logger()
	logger.Info().Msg("Getting pending parameter change request")

	changeRequest, err := h.service.GetPendingParameterChangeRequestByParameterID(ctx, parameterID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get pending parameter change request")
		return nil, err
	}

	if changeRequest == nil {
		return nil, nil
	}

	response := dto.ToParameterChangeRequestResponse(changeRequest)
	return &response, nil
}

// GetParameterChangeRequestsByParameterID handles getting all change requests for a parameter
func (h *Handler) GetParameterChangeRequestsByParameterID(ctx context.Context, parameterID uint) ([]dto.ParameterChangeRequestResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-parameter-change-requests-by-parameter-id").Uint("parameterId", parameterID).Logger()
	logger.Info().Msg("Getting parameter change requests")

	changeRequests, err := h.service.GetParameterChangeRequestsByParameterID(ctx, parameterID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get parameter change requests")
		return nil, err
	}

	responses := make([]dto.ParameterChangeRequestResponse, len(changeRequests))
	for i, changeRequest := range changeRequests {
		responses[i] = dto.ToParameterChangeRequestResponse(changeRequest)
	}

	return responses, nil
}

// ApproveParameterChangeRequest handles approving a parameter change request
func (h *Handler) ApproveParameterChangeRequest(ctx context.Context, id uint, userID uint, req *dto.ApproveParameterChangeRequestRequest) (*dto.ParameterChangeRequestResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "approve-parameter-change-request").Uint("id", id).Logger()
	logger.Info().Msg("Approving parameter change request")

	changeRequest, err := h.service.ApproveParameterChangeRequest(ctx, id, userID, req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to approve parameter change request")
		return nil, err
	}

	response := dto.ToParameterChangeRequestResponse(changeRequest)
	return &response, nil
}

// RejectParameterChangeRequest handles rejecting a parameter change request
func (h *Handler) RejectParameterChangeRequest(ctx context.Context, id uint, userID uint, req *dto.RejectParameterChangeRequestRequest) (*dto.ParameterChangeRequestResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "reject-parameter-change-request").Uint("id", id).Logger()
	logger.Info().Msg("Rejecting parameter change request")

	changeRequest, err := h.service.RejectParameterChangeRequest(ctx, id, userID, req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to reject parameter change request")
		return nil, err
	}

	response := dto.ToParameterChangeRequestResponse(changeRequest)
	return &response, nil
}

// GetParameterChangeRequestsByStatus handles getting parameter change requests by status with pagination
func (h *Handler) GetParameterChangeRequestsByStatus(ctx context.Context, status model.ParameterChangeRequestStatus, limit, offset int) (*dto.ParameterChangeRequestListResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-parameter-change-requests-by-status").Str("status", string(status)).Logger()
	logger.Info().Msg("Getting parameter change requests by status")

	changeRequests, total, err := h.service.GetParameterChangeRequestsByStatus(ctx, status, limit, offset)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get parameter change requests by status")
		return nil, err
	}

	response := dto.ToParameterChangeRequestListResponse(changeRequests, total, limit, offset)
	return &response, nil
}

// GetParameterChangeRequestByIDWithDetails handles getting a detailed parameter change request by ID
func (h *Handler) GetParameterChangeRequestByIDWithDetails(ctx context.Context, id uint) (*dto.ParameterChangeRequestResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-parameter-change-request-by-id-with-details").Uint("id", id).Logger()
	logger.Info().Msg("Getting detailed parameter change request by ID")

	changeRequest, err := h.service.GetParameterChangeRequestByIDWithDetails(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get detailed parameter change request")
		return nil, err
	}

	response := dto.ToParameterChangeRequestResponse(changeRequest)
	return &response, nil
}
