package service

import (
	"api/internal/dto"
	"api/internal/model"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
		Status:          "draft", // Default status
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

	return "Experiment created successfully", nil
}

// GetAllExperiments retrieves all experiments without variants
func (s *service) GetAllExperiments(ctx context.Context) ([]*model.Experiment, error) {
	return s.repo.GetAllExperiments(ctx, 0, 0) // 0, 0 means no limit and no offset
}

// GetExperimentByID retrieves an experiment by ID with all variants and their parameters
func (s *service) GetExperimentByID(ctx context.Context, id uint) (*model.Experiment, []*model.ExperimentVariant, map[int][]*model.ExperimentVariantParameter, *model.Attribute, error) {
	// Get the experiment
	experiment, err := s.repo.GetExperimentByID(ctx, id)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	// Get the hash attribute
	hashAttribute, err := s.repo.GetAttributeByID(ctx, uint(experiment.HashAttributeID))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get hash attribute: %w", err)
	}

	// Get all variants for this experiment
	variants, err := s.repo.GetExperimentVariantsByExperimentID(ctx, id)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get experiment variants: %w", err)
	}

	// Get parameters for each variant
	variantParametersMap := make(map[int][]*model.ExperimentVariantParameter)
	for _, variant := range variants {
		parameters, err := s.repo.GetExperimentVariantParametersByVariantID(ctx, uint(variant.ID))
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to get variant parameters for variant %d: %w", variant.ID, err)
		}
		variantParametersMap[variant.ID] = parameters
	}

	return experiment, variants, variantParametersMap, hashAttribute, nil
}
