package service

import (
	"api/internal/constant"
	"api/internal/dto"
	"api/internal/mapper"
	"api/internal/model"
	"context"
	"errors"
	"fmt"
	sdk "sdk/types"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const txContextKey contextKey = "tx"

// CreateExperiment creates a new experiment with its variants and parameters
func (s *service) CreateExperiment(ctx context.Context, req *dto.CreateExperimentRequest) (string, error) {

	if err := req.Validate(); err != nil {
		return "", err
	}

	_, err := s.repo.GetAttributeByID(ctx, uint(req.HashAttributeID))
	if err != nil {
		return "", err
	}

	// Collect unique parameter IDs
	parameterIDSSet := make(map[int]bool)
	for _, variant := range req.Variants {
		for _, parameter := range variant.Parameters {
			parameterIDSSet[parameter.ParameterID] = true
		}
	}

	parameterIDS := make([]int, 0, len(parameterIDSSet))
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
			// Validate rollout value based on data type
			if verifiedParameter.DataType == model.ParameterDataTypeString || verifiedParameter.DataType == model.ParameterDataTypeNumber {
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

	// Check for conflicting experiments with sophisticated segment analysis
	conflictingExperiments, err := s.repo.FindConflictingExperiments(ctx, parameterIDS, req.SegmentID, req.StartDate, req.EndDate)
	if err != nil {
		return "", fmt.Errorf("failed to check for conflicting experiments: %w", err)
	}

	// Filter experiments based on sophisticated segment overlap analysis
	var actualConflicts []*model.Experiment
	for _, exp := range conflictingExperiments {
		hasOverlap, err := s.checkSegmentOverlap(ctx, req.SegmentID, exp.SegmentID)
		if err != nil {
			return "", fmt.Errorf("failed to check segment overlap: %w", err)
		}
		if hasOverlap {
			actualConflicts = append(actualConflicts, exp)
		}
	}

	if len(actualConflicts) > 0 {
		// Build detailed conflict message
		var conflictDetails []string
		for _, exp := range actualConflicts {
			conflictDetails = append(conflictDetails, fmt.Sprintf("Experiment '%s' (ID: %d, Status: %s, Segment: %d, Period: %d-%d)",
				exp.Name, exp.ID, exp.Status, exp.SegmentID, exp.StartDate, exp.EndDate))
		}

		return "", fmt.Errorf("experiment conflicts detected with %d existing experiment(s): [%s]",
			len(actualConflicts), strings.Join(conflictDetails, ", "))
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
			panic(r)
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
	txCtx := context.WithValue(ctx, txContextKey, tx)
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
	if err := s.updateExperimentStatusIfNeeded(ctx, experiment); err != nil {
		return nil, nil, nil, nil, err
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

	// Save the updated experiment and update raw_value
	if err := s.updateExperimentAndRawValue(ctx, experiment); err != nil {
		return nil, err
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

	// Extract parameter IDs from experiment variants
	parameterIDS := s.extractParameterIDsFromExperiment(experiment)

	// Check for conflicting experiments with sophisticated segment analysis
	conflictingExperiments, err := s.repo.FindConflictingExperiments(ctx, parameterIDS, experiment.SegmentID, experiment.StartDate, experiment.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check for conflicting experiments: %w", err)
	}

	// Filter experiments based on sophisticated segment overlap analysis and exclude current experiment
	var actualConflicts []*model.Experiment
	for _, exp := range conflictingExperiments {
		// Exclude the current experiment from conflict check
		if exp.ID == experiment.ID {
			continue
		}
		hasOverlap, err := s.checkSegmentOverlap(ctx, experiment.SegmentID, exp.SegmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to check segment overlap: %w", err)
		}
		if hasOverlap {
			actualConflicts = append(actualConflicts, exp)
		}
	}

	if len(actualConflicts) > 0 {
		// Build detailed conflict message
		var conflictDetails []string
		for _, exp := range actualConflicts {
			conflictDetails = append(conflictDetails, fmt.Sprintf("Experiment '%s' (ID: %d, Status: %s, Segment: %d, Period: %d-%d)",
				exp.Name, exp.ID, exp.Status, exp.SegmentID, exp.StartDate, exp.EndDate))
		}

		return nil, fmt.Errorf("experiment conflicts detected with %d existing experiment(s): [%s]",
			len(actualConflicts), strings.Join(conflictDetails, ", "))
	}

	// Update the experiment status to approved
	experiment.Status = constant.ExperimentStatusSchedule
	experiment.UpdatedAt = time.Now().Unix()

	// Save the updated experiment and update raw_value
	if err := s.updateExperimentAndRawValue(ctx, experiment); err != nil {
		return nil, err
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

	// Check if experiment can be aborted (only scheduled or running experiments can be aborted)
	if experiment.Status != constant.ExperimentStatusSchedule && experiment.Status != constant.ExperimentStatusRunning {
		return nil, fmt.Errorf("experiment is not in schedule status")
	}

	// Update the experiment status to abort
	experiment.Status = constant.ExperimentStatusAbort
	experiment.UpdatedAt = time.Now().Unix()

	// Save the updated experiment and update raw_value
	if err := s.updateExperimentAndRawValue(ctx, experiment); err != nil {
		return nil, err
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

// checkSegmentOverlap determines if two segments can have overlapping users
func (s *service) checkSegmentOverlap(ctx context.Context, segmentID1, segmentID2 int) (bool, error) {
	// Case 1: Both segments are empty (no segment)
	if segmentID1 == 0 && segmentID2 == 0 {
		return true, nil // All users are in both segments
	}

	// Case 2: One segment is empty, one is specific
	if segmentID1 == 0 || segmentID2 == 0 {
		return true, nil // Empty segment includes all users, so there's always overlap
	}

	// Case 3: Both segments are specific - need to analyze their conditions
	if segmentID1 == segmentID2 {
		return true, nil // Same segment
	}

	// Load both segments with their rules and conditions
	segment1, err := s.repo.GetSegmentByID(ctx, uint(segmentID1))
	if err != nil {
		return false, fmt.Errorf("failed to load segment %d: %w", segmentID1, err)
	}

	segment2, err := s.repo.GetSegmentByID(ctx, uint(segmentID2))
	if err != nil {
		return false, fmt.Errorf("failed to load segment %d: %w", segmentID2, err)
	}

	res, err := s.solver.CheckSegmentsConflict([]model.Segment{*segment1, *segment2})
	if err != nil {
		return false, fmt.Errorf("failed to check segments conflict: %w", err)
	}

	return !res.Valid, nil
}

// updateExperimentStatusIfNeeded updates the experiment status based on current date
// and updates the raw_value field accordingly
func (s *service) updateExperimentStatusIfNeeded(ctx context.Context, experiment *model.Experiment) error {
	now := time.Now().Unix()
	var newStatus string

	if experiment.Status == constant.ExperimentStatusSchedule && experiment.StartDate < now {
		newStatus = constant.ExperimentStatusRunning
	} else if experiment.Status == constant.ExperimentStatusRunning && experiment.EndDate < now {
		newStatus = constant.ExperimentStatusFinish
	} else {
		return nil
	}

	experiment.Status = newStatus
	experiment.UpdatedAt = now

	if err := s.repo.UpdateExperiment(ctx, experiment); err != nil {
		return fmt.Errorf("failed to update experiment: %w", err)
	}

	// Update raw_value field - log error but don't fail the operation
	if err := s.repo.UpdateExperimentRawValue(ctx, uint(experiment.ID)); err != nil {
		log.Ctx(ctx).Error().Err(err).Int("experimentId", experiment.ID).Msg("Failed to update experiment raw_value")
	}

	return nil
}

// updateExperimentAndRawValue updates the experiment and its raw_value field
// It logs errors for raw_value updates but doesn't fail the operation
func (s *service) updateExperimentAndRawValue(ctx context.Context, experiment *model.Experiment) error {
	if err := s.repo.UpdateExperiment(ctx, experiment); err != nil {
		return fmt.Errorf("failed to update experiment: %w", err)
	}

	// Update raw_value field - log error but don't fail the update
	if err := s.repo.UpdateExperimentRawValue(ctx, uint(experiment.ID)); err != nil {
		log.Ctx(ctx).Error().Err(err).Int("experimentId", experiment.ID).Msg("Failed to update experiment raw_value")
	}

	return nil
}

// extractParameterIDsFromExperiment extracts unique parameter IDs from an experiment's variants
func (s *service) extractParameterIDsFromExperiment(experiment *model.Experiment) []int {
	parameterIDSSet := make(map[int]bool)
	for _, variant := range experiment.Variants {
		for _, parameter := range variant.Parameters {
			parameterIDSSet[parameter.ParameterID] = true
		}
	}

	parameterIDS := make([]int, 0, len(parameterIDSSet))
	for parameterID := range parameterIDSSet {
		parameterIDS = append(parameterIDS, parameterID)
	}

	return parameterIDS
}
