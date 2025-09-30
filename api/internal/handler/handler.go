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

// CreateSegment handles the business logic for creating a segment
func (h *Handler) CreateSegment(ctx context.Context, req *dto.CreateSegmentRequest) (*dto.SegmentResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "create-segment").Logger()
	logger.Info().Msg("Creating segment")

	segment, err := h.service.CreateSegment(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create segment")
		return nil, err
	}

	response := dto.ToSegmentResponse(segment)
	return &response, nil
}

// GetAllSegments handles the business logic for getting all segments
func (h *Handler) GetAllSegments(ctx context.Context) ([]dto.SegmentResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-all-segments").Logger()
	logger.Info().Msg("Getting all segments")

	segments, err := h.service.GetAllSegments(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all segments")
		return nil, err
	}

	responses := make([]dto.SegmentResponse, len(segments))
	for i, segment := range segments {
		responses[i] = dto.ToSegmentResponse(segment)
	}

	return responses, nil
}

// GetSegmentByID handles the business logic for getting a segment by ID
func (h *Handler) GetSegmentByID(ctx context.Context, id uint) (*dto.SegmentResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-segment-by-id").Uint("id", id).Logger()
	logger.Info().Msg("Getting segment by ID")

	segment, err := h.service.GetSegmentByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to get segment by ID")
		return nil, err
	}

	response := dto.ToSegmentResponse(segment)
	return &response, nil
}

// UpdateSegment handles the business logic for updating a segment
func (h *Handler) UpdateSegment(ctx context.Context, id uint, req *dto.UpdateSegmentRequest) (*dto.SegmentResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "update-segment").Uint("id", id).Logger()
	logger.Info().Msg("Updating segment")

	segment, err := h.service.UpdateSegment(ctx, id, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to update segment")
		return nil, err
	}

	response := dto.ToSegmentResponse(segment)
	return &response, nil
}

// DeleteSegment handles the business logic for deleting a segment
func (h *Handler) DeleteSegment(ctx context.Context, id uint) error {
	logger := log.Ctx(ctx).With().Str("handler", "delete-segment").Uint("id", id).Logger()
	logger.Info().Msg("Deleting segment")

	err := h.service.DeleteSegment(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to delete segment")
		return err
	}

	return nil
}

// CreateParameter handles the business logic for creating a parameter
func (h *Handler) CreateParameter(ctx context.Context, req *dto.CreateParameterRequest) (*dto.ParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "create-parameter").Logger()
	logger.Info().Msg("Creating parameter")

	parameter, err := h.service.CreateParameter(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create parameter")
		return nil, err
	}

	response := dto.ToParameterResponse(parameter)
	return &response, nil
}

// GetAllParameters handles the business logic for getting all parameters
func (h *Handler) GetAllParameters(ctx context.Context) ([]dto.ParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-all-parameters").Logger()
	logger.Info().Msg("Getting all parameters")

	parameters, err := h.service.GetAllParameters(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all parameters")
		return nil, err
	}

	responses := make([]dto.ParameterResponse, len(parameters))
	for i, parameter := range parameters {
		responses[i] = dto.ToParameterResponse(parameter)
	}

	return responses, nil
}

// GetParameterByID handles the business logic for getting a parameter by ID
func (h *Handler) GetParameterByID(ctx context.Context, id uint) (*dto.ParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-parameter-by-id").Uint("id", id).Logger()
	logger.Info().Msg("Getting parameter by ID")

	parameter, err := h.service.GetParameterByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to get parameter by ID")
		return nil, err
	}

	response := dto.ToParameterResponse(parameter)
	return &response, nil
}

// UpdateParameter handles the business logic for updating a parameter
func (h *Handler) UpdateParameter(ctx context.Context, id uint, req *dto.UpdateParameterRequest) (*dto.ParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "update-parameter").Uint("id", id).Logger()
	logger.Info().Msg("Updating parameter")

	parameter, err := h.service.UpdateParameter(ctx, id, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to update parameter")
		return nil, err
	}

	response := dto.ToParameterResponse(parameter)
	return &response, nil
}

// UpdateParameterWithRules handles the business logic for comprehensive parameter update with rules
func (h *Handler) UpdateParameterWithRules(ctx context.Context, id uint, req *dto.UpdateParameterWithRulesRequest) (*dto.ParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "update-parameter-with-rules").Uint("id", id).Logger()
	logger.Info().Msg("Updating parameter with rules")

	parameter, err := h.service.UpdateParameterWithRules(ctx, id, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to update parameter with rules")
		return nil, err
	}

	response := dto.ToParameterResponse(parameter)
	return &response, nil
}

// DeleteParameter handles the business logic for deleting a parameter
func (h *Handler) DeleteParameter(ctx context.Context, id uint) error {
	logger := log.Ctx(ctx).With().Str("handler", "delete-parameter").Uint("id", id).Logger()
	logger.Info().Msg("Deleting parameter")

	err := h.service.DeleteParameter(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to delete parameter")
		return err
	}

	return nil
}

// AddParameterRule handles the business logic for adding a rule to a parameter
func (h *Handler) AddParameterRule(ctx context.Context, id uint, req *dto.CreateParameterRuleRequest) (*dto.ParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "add-parameter-rule").Uint("id", id).Logger()
	logger.Info().Msg("Adding parameter rule")

	parameter, err := h.service.AddParameterRule(ctx, id, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to add parameter rule")
		return nil, err
	}

	response := dto.ToParameterResponse(parameter)
	return &response, nil
}

// UpdateParameterRule handles the business logic for updating a parameter rule
func (h *Handler) UpdateParameterRule(ctx context.Context, id uint, ruleID uint, req *dto.UpdateParameterRuleRequest) (*dto.ParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "update-parameter-rule").Uint("id", id).Uint("ruleId", ruleID).Logger()
	logger.Info().Msg("Updating parameter rule")

	parameter, err := h.service.UpdateParameterRule(ctx, id, ruleID, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Uint("ruleId", ruleID).Msg("Failed to update parameter rule")
		return nil, err
	}

	response := dto.ToParameterResponse(parameter)
	return &response, nil
}

// DeleteParameterRule handles the business logic for deleting a parameter rule
func (h *Handler) DeleteParameterRule(ctx context.Context, id uint, ruleID uint) (*dto.ParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "delete-parameter-rule").Uint("id", id).Uint("ruleId", ruleID).Logger()
	logger.Info().Msg("Deleting parameter rule")

	parameter, err := h.service.DeleteParameterRule(ctx, id, ruleID)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Uint("ruleId", ruleID).Msg("Failed to delete parameter rule")
		return nil, err
	}

	response := dto.ToParameterResponse(parameter)
	return &response, nil
}

// CreateExperiment handles the business logic for creating an experiment
func (h *Handler) CreateExperiment(ctx context.Context, req *dto.CreateExperimentRequest) (string, error) {
	logger := log.Ctx(ctx).With().Str("handler", "create-experiment").Logger()
	logger.Info().Msg("Creating experiment")

	message, err := h.service.CreateExperiment(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create experiment")
		return "", err
	}

	logger.Info().Msg("Experiment created successfully")
	return message, nil
}

// GetAllExperiments handles the business logic for getting all experiments
func (h *Handler) GetAllExperiments(ctx context.Context) ([]dto.ExperimentResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-all-experiments").Logger()
	logger.Info().Msg("Getting all experiments")

	experiments, err := h.service.GetAllExperiments(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all experiments")
		return nil, err
	}

	responses := make([]dto.ExperimentResponse, len(experiments))
	for i, experiment := range experiments {
		responses[i] = dto.ToExperimentResponse(experiment)
	}

	return responses, nil
}

// GetExperimentByID handles the business logic for getting an experiment by ID with variants and parameters
func (h *Handler) GetExperimentByID(ctx context.Context, id uint) (*dto.ExperimentDetailResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "get-experiment-by-id").Uint("id", id).Logger()
	logger.Info().Msg("Getting experiment by ID with details")

	experiment, variants, variantParametersMap, hashAttribute, err := h.service.GetExperimentByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to get experiment by ID")
		return nil, err
	}

	response := dto.ToExperimentDetailResponse(experiment, variants, variantParametersMap, hashAttribute)
	return &response, nil
}

// RejectExperiment handles the business logic for rejecting an experiment
func (h *Handler) RejectExperiment(ctx context.Context, id uint, req *dto.RejectExperimentRequest) (*dto.ExperimentResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "reject-experiment").Uint("id", id).Logger()
	logger.Info().Msg("Rejecting experiment")

	experiment, err := h.service.RejectExperiment(ctx, id, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to reject experiment")
		return nil, err
	}

	response := dto.ToExperimentResponse(experiment)
	return &response, nil
}

// ApproveExperiment handles the business logic for approving an experiment
func (h *Handler) ApproveExperiment(ctx context.Context, id uint, req *dto.ApproveExperimentRequest) (*dto.ExperimentResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "approve-experiment").Uint("id", id).Logger()
	logger.Info().Msg("Approving experiment")

	experiment, err := h.service.ApproveExperiment(ctx, id, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to approve experiment")
		return nil, err
	}

	response := dto.ToExperimentResponse(experiment)
	return &response, nil
}

// AbortExperiment handles the business logic for aborting an experiment
func (h *Handler) AbortExperiment(ctx context.Context, id uint, req *dto.AbortExperimentRequest) (*dto.ExperimentResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "abort-experiment").Uint("id", id).Logger()
	logger.Info().Msg("Aborting experiment")

	experiment, err := h.service.AbortExperiment(ctx, id, req)
	if err != nil {
		logger.Error().Err(err).Uint("id", id).Msg("Failed to abort experiment")
		return nil, err
	}

	response := dto.ToExperimentResponse(experiment)
	return &response, nil
}

func (h *Handler) SimulateParameter(ctx context.Context, req *dto.SimulateParameterRequest) (*dto.SimulateParameterResponse, error) {
	logger := log.Ctx(ctx).With().Str("handler", "simulate-parameter").Logger()
	logger.Info().Msg("Simulating parameter")

	response, err := h.service.SimulateParameter(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to simulate parameter")
		return nil, err
	}

	return &response, nil
}
