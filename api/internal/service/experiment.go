package service

import (
	"api/internal/constant"
	"api/internal/dto"
	"api/internal/mapper"
	"api/internal/model"
	"context"
	"errors"
	"fmt"
	"sdk"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CreateExperiment creates a new experiment with its variants and parameters
func (s *service) CreateExperiment(ctx context.Context, req *dto.CreateExperimentRequest) (string, error) {

	if err := req.Validate(); err != nil {
		return "", err
	}

	_, err := s.repo.GetAttributeByID(ctx, uint(req.HashAttributeID))
	if err != nil {
		return "", err
	}

	parameterIDSSet := make(map[int]bool)
	for _, variant := range req.Variants {
		for _, parameter := range variant.Parameters {
			parameterIDSSet[parameter.ParameterID] = true
		}
	}

	parameterIDS := make([]int, 0)
	for parameterID := range parameterIDSSet {
		parameterIDS = append(parameterIDS, parameterID)
	}
	parameters, err := s.repo.GetParametersByIDs(ctx, parameterIDS)

	if err != nil {
		return "", err
	}

	if len(parameters) != len(parameterIDS) {
		return "", fmt.Errorf("some parameters do not exist")
	}

	mapParameters := make(map[int]model.Parameter)
	for _, parameter := range parameters {
		mapParameters[int(parameter.ID)] = parameter
	}

	for _, variant := range req.Variants {
		for _, parameter := range variant.Parameters {
			verifiedParameter, ok := mapParameters[parameter.ParameterID]
			if !ok {
				return "", fmt.Errorf("parameter %d does not exist", parameter.ParameterID)
			}
			if verifiedParameter.DataType != model.ParameterDataType(parameter.ParameterDataType) {
				return "", fmt.Errorf("parameter %d has invalid data type", parameter.ParameterID)
			}
			if verifiedParameter.Name != parameter.ParameterName {
				return "", fmt.Errorf("parameter %d has invalid name", parameter.ParameterID)
			}
			if verifiedParameter.DataType == model.ParameterDataTypeString {
				if parameter.RolloutValue == "" {
					return "", fmt.Errorf("parameter %d has invalid rollout value", parameter.ParameterID)
				}
			}
			if verifiedParameter.DataType == model.ParameterDataTypeNumber {
				if parameter.RolloutValue == "" {
					return "", fmt.Errorf("parameter %d has invalid rollout value", parameter.ParameterID)
				}
			}
			if verifiedParameter.DataType == model.ParameterDataTypeBoolean {
				if parameter.RolloutValue != "true" && parameter.RolloutValue != "false" {
					return "", fmt.Errorf("parameter %d has invalid rollout value", parameter.ParameterID)
				}
			}
		}
	}

	exp, err := s.repo.GetExperimentByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	if exp != nil {
		return "", fmt.Errorf("experiment with name %s already exists", req.Name)
	}

	if req.SegmentID != 0 {
		_, err = s.repo.GetSegmentByID(ctx, uint(req.SegmentID))
		if err != nil {
			return "", fmt.Errorf("segment with id %d not found", req.SegmentID)
		}
	}

	// Start a transaction
	tx := s.repo.GetDB().Begin()
	if tx.Error != nil {
		return "", fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Rollback transaction on any error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the experiment with business logic
	now := time.Now().Unix()
	experiment := &model.Experiment{
		Name:            req.Name,
		Uuid:            uuid.New().String(), // Generate UUID
		Hypothesis:      req.Hypothesis,
		Description:     req.Description,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		HashAttributeID: req.HashAttributeID,
		PopulationSize:  req.PopulationSize,
		Strategy:        req.Strategy,
		CreatedAt:       now,
		UpdatedAt:       now,
		Status:          constant.ExperimentStatusDraft, // Default status
		SegmentID:       req.SegmentID,
	}

	// Use transaction context
	txCtx := context.WithValue(ctx, "tx", tx)
	if err := s.repo.CreateExperiment(txCtx, experiment); err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to create experiment: %w", err)
	}

	// Create variants and their parameters
	for _, variantReq := range req.Variants {
		variant := &model.ExperimentVariant{
			ExperimentID:      experiment.ID,
			Name:              variantReq.Name,
			Description:       variantReq.Description,
			TrafficAllocation: variantReq.TrafficAllocation,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := s.repo.CreateExperimentVariant(txCtx, variant); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to create experiment variant: %w", err)
		}

		// Create variant parameters
		for _, paramReq := range variantReq.Parameters {
			parameter := &model.ExperimentVariantParameter{
				ExperimentVariantID: variant.ID,
				ParameterDataType:   paramReq.ParameterDataType,
				ParameterID:         paramReq.ParameterID,
				ParameterName:       paramReq.ParameterName,
				RolloutValue:        paramReq.RolloutValue,
				ExperimentID:        experiment.ID,
				CreatedAt:           now,
				UpdatedAt:           now,
			}

			if err := s.repo.CreateExperimentVariantParameter(txCtx, parameter); err != nil {
				tx.Rollback()
				return "", fmt.Errorf("failed to create experiment variant parameter: %w", err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update raw_value field with all related data
	if err := s.repo.UpdateExperimentRawValue(ctx, uint(experiment.ID)); err != nil {
		// Log error but don't fail the creation since experiment was already created successfully
		log.Ctx(ctx).Error().Err(err).Int("experimentId", experiment.ID).Msg("Failed to update experiment raw_value")
	}

	return "Experiment created successfully", nil
}

// GetAllExperiments retrieves all experiments without variants
func (s *service) GetAllExperiments(ctx context.Context) ([]*model.Experiment, error) {
	return s.repo.GetAllExperiments(ctx, 0, 0) // 0, 0 means no limit and no offset
}

// GetExperimentByID retrieves an experiment by ID with all variants and their parameters
func (s *service) GetExperimentByID(ctx context.Context, id uint) (*model.Experiment, []*model.ExperimentVariant, map[int][]*model.ExperimentVariantParameter, *model.Attribute, error) {
	// Get the experiment with all preloaded data
	experiment, err := s.repo.GetExperimentByID(ctx, id)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	// Convert variants slice to pointer slice for compatibility
	variants := make([]*model.ExperimentVariant, len(experiment.Variants))
	for i, variant := range experiment.Variants {
		variants[i] = &variant
	}

	// Build parameters map from preloaded data
	variantParametersMap := make(map[int][]*model.ExperimentVariantParameter)
	for _, variant := range experiment.Variants {
		parameters := make([]*model.ExperimentVariantParameter, len(variant.Parameters))
		for i, param := range variant.Parameters {
			parameters[i] = &param
		}
		variantParametersMap[variant.ID] = parameters
	}

	// Update experiment status based on dates
	if experiment.Status == constant.ExperimentStatusSchedule {
		if experiment.StartDate < time.Now().Unix() {
			experiment.Status = constant.ExperimentStatusRunning
			experiment.UpdatedAt = time.Now().Unix()
			err = s.repo.UpdateExperiment(ctx, experiment)
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("failed to update experiment: %w", err)
			}
			// Update raw_value field with all related data
			if err := s.repo.UpdateExperimentRawValue(ctx, uint(experiment.ID)); err != nil {
				// Log error but don't fail the get operation
				log.Ctx(ctx).Error().Err(err).Int("experimentId", experiment.ID).Msg("Failed to update experiment raw_value")
			}
		}
	} else if experiment.Status == constant.ExperimentStatusRunning {
		if experiment.EndDate < time.Now().Unix() {
			experiment.Status = constant.ExperimentStatusFinish
			experiment.UpdatedAt = time.Now().Unix()
			err = s.repo.UpdateExperiment(ctx, experiment)
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("failed to update experiment: %w", err)
			}
			// Update raw_value field with all related data
			if err := s.repo.UpdateExperimentRawValue(ctx, uint(experiment.ID)); err != nil {
				// Log error but don't fail the get operation
				log.Ctx(ctx).Error().Err(err).Int("experimentId", experiment.ID).Msg("Failed to update experiment raw_value")
			}
		}
	}

	return experiment, variants, variantParametersMap, experiment.HashAttribute, nil
}

// RejectExperiment rejects an experiment by updating its status to "cancel"
func (s *service) RejectExperiment(ctx context.Context, id uint, req *dto.RejectExperimentRequest) (*model.Experiment, error) {
	// Get the experiment first to ensure it exists
	experiment, err := s.repo.GetExperimentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	// Check if experiment can be rejected (business logic based on ExperimentStatus enum)
	if experiment.Status != constant.ExperimentStatusDraft {
		return nil, fmt.Errorf("experiment is not in draft status")
	}

	// Update the experiment status to cancel (reject)
	experiment.Status = constant.ExperimentStatusCancel
	experiment.UpdatedAt = time.Now().Unix()

	// Save the updated experiment
	err = s.repo.UpdateExperiment(ctx, experiment)
	if err != nil {
		return nil, fmt.Errorf("failed to update experiment: %w", err)
	}

	// Update raw_value field with all related data
	if err := s.repo.UpdateExperimentRawValue(ctx, uint(experiment.ID)); err != nil {
		// Log error but don't fail the update
		log.Ctx(ctx).Error().Err(err).Int("experimentId", experiment.ID).Msg("Failed to update experiment raw_value")
	}

	s.riverClient.Insert(ctx, dto.SyncExperimentArgs{}, nil)

	return experiment, nil
}

// ApproveExperiment approves an experiment by updating its status to "approved"
func (s *service) ApproveExperiment(ctx context.Context, id uint, req *dto.ApproveExperimentRequest) (*model.Experiment, error) {
	// Get the experiment first to ensure it exists
	experiment, err := s.repo.GetExperimentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	if experiment.Status != constant.ExperimentStatusDraft {
		return nil, fmt.Errorf("experiment is not in draft status")
	}

	// Update the experiment status to approved
	experiment.Status = constant.ExperimentStatusSchedule
	experiment.UpdatedAt = time.Now().Unix()

	// Save the updated experiment
	err = s.repo.UpdateExperiment(ctx, experiment)
	if err != nil {
		return nil, fmt.Errorf("failed to update experiment: %w", err)
	}

	// Update raw_value field with all related data
	if err := s.repo.UpdateExperimentRawValue(ctx, uint(experiment.ID)); err != nil {
		// Log error but don't fail the update
		log.Ctx(ctx).Error().Err(err).Int("experimentId", experiment.ID).Msg("Failed to update experiment raw_value")
	}

	s.riverClient.Insert(ctx, dto.SyncExperimentArgs{}, nil)

	return experiment, nil
}

// AbortExperiment aborts an experiment by updating its status to "abort"
func (s *service) AbortExperiment(ctx context.Context, id uint, req *dto.AbortExperimentRequest) (*model.Experiment, error) {
	// Get the experiment first to ensure it exists
	experiment, err := s.repo.GetExperimentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	// Check if experiment can be aborted (business logic based on ExperimentStatus enum)
	// if experiment.Status == constant.ExperimentStatusCancel {
	// 	return nil, fmt.Errorf("cannot abort a canceled experiment")
	// }

	// if experiment.Status == constant.ExperimentStatusAbort {
	// 	return nil, fmt.Errorf("experiment is already aborted")
	// }

	// if experiment.Status == constant.ExperimentStatusFinish {
	// 	return nil, fmt.Errorf("cannot abort a finished experiment")
	// }

	// if experiment.Status == constant.ExperimentStatusDraft {
	// 	return nil, fmt.Errorf("cannot abort a draft experiment, reject it instead")
	// }

	if experiment.Status != constant.ExperimentStatusSchedule && experiment.Status != constant.ExperimentStatusRunning {
		return nil, fmt.Errorf("experiment is not in schedule status")
	}

	// Update the experiment status to abort
	experiment.Status = constant.ExperimentStatusAbort
	experiment.UpdatedAt = time.Now().Unix()

	// Save the updated experiment
	err = s.repo.UpdateExperiment(ctx, experiment)
	if err != nil {
		return nil, fmt.Errorf("failed to update experiment: %w", err)
	}

	// Update raw_value field with all related data
	if err := s.repo.UpdateExperimentRawValue(ctx, uint(experiment.ID)); err != nil {
		// Log error but don't fail the update
		log.Ctx(ctx).Error().Err(err).Int("experimentId", experiment.ID).Msg("Failed to update experiment raw_value")
	}

	s.riverClient.Insert(ctx, dto.SyncExperimentArgs{}, nil)

	return experiment, nil
}

func (s *service) GetActiveExperimentsSDK(ctx context.Context) ([]sdk.Experiment, error) {
	experiments, err := s.repo.GetExperimentsActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active experiments: %w", err)
	}

	// Use raw value conversion for better performance
	result, err := mapper.ExperimentsToSDKFromRawValue(experiments)
	if err != nil {
		return nil, fmt.Errorf("failed to convert experiments to sdk: %w", err)
	}
	return result, nil
}
