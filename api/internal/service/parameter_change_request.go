package service

import (
	"api/internal/dto"
	"api/internal/model"
	"api/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CreateParameterChangeRequest creates a new parameter change request
func (s *service) CreateParameterChangeRequest(ctx context.Context, userID uint, req *dto.CreateParameterChangeRequestRequest) (*model.ParameterChangeRequest, error) {
	logger := log.Ctx(ctx).With().Str("service", "create-parameter-change-request").Uint("parameterId", req.ParameterID).Logger()

	// Check if parameter exists
	parameter, err := s.repo.GetParameterByID(ctx, req.ParameterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("parameter with ID %d not found", req.ParameterID)
		}
		return nil, err
	}

	// Check if there's already a pending change request for this parameter
	existing, err := s.repo.GetPendingParameterChangeRequestByParameterID(ctx, req.ParameterID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("parameter '%s' already has a pending change request (ID: %d). Please approve or reject it before creating a new one", parameter.Name, existing.ID)
	}

	// Build the change data from the request
	changeData := model.ParameterChangeData{
		Name:                req.Name,
		Description:         req.ParameterDescription,
		DataType:            req.DataType,
		DefaultRolloutValue: req.DefaultRolloutValue,
	}

	// Convert rules if provided
	if len(req.Rules) > 0 {
		changeData.Rules = make([]model.ParameterRuleRequest, len(req.Rules))
		for i, rule := range req.Rules {
			ruleReq := model.ParameterRuleRequest{
				Name:         rule.Name,
				Description:  rule.Description,
				Type:         rule.Type,
				RolloutValue: rule.RolloutValue,
				SegmentID:    rule.SegmentID,
				MatchType:    rule.MatchType,
			}

			// Convert conditions if provided
			if len(rule.Conditions) > 0 {
				ruleReq.Conditions = make([]model.ParameterRuleConditionRequest, len(rule.Conditions))
				for j, cond := range rule.Conditions {
					ruleReq.Conditions[j] = model.ParameterRuleConditionRequest{
						AttributeID: cond.AttributeID,
						Operator:    cond.Operator,
						Value:       cond.Value,
					}
				}
			}

			changeData.Rules[i] = ruleReq
		}
	}

	// Create the change request
	changeRequest := &model.ParameterChangeRequest{
		ParameterID:       req.ParameterID,
		RequestedByUserID: userID,
		Status:            model.ChangeRequestStatusPending,
		Description:       req.Description,
		ChangeData:        changeData,
	}

	if err := s.repo.CreateParameterChangeRequest(ctx, changeRequest); err != nil {
		logger.Error().Err(err).Msg("Failed to create parameter change request")
		return nil, err
	}

	// Reload with relationships
	return s.repo.GetParameterChangeRequestByID(ctx, changeRequest.ID)
}

// GetParameterChangeRequestByID retrieves a parameter change request by ID
func (s *service) GetParameterChangeRequestByID(ctx context.Context, id uint) (*model.ParameterChangeRequest, error) {
	changeRequest, err := s.repo.GetParameterChangeRequestByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("parameter change request with ID %d not found", id)
		}
		return nil, err
	}
	return changeRequest, nil
}

// GetPendingParameterChangeRequestByParameterID retrieves pending change request for a parameter
func (s *service) GetPendingParameterChangeRequestByParameterID(ctx context.Context, parameterID uint) (*model.ParameterChangeRequest, error) {
	changeRequest, err := s.repo.GetPendingParameterChangeRequestByParameterID(ctx, parameterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No pending request is not an error
		}
		return nil, err
	}
	return changeRequest, nil
}

// GetParameterChangeRequestsByParameterID retrieves all change requests for a parameter
func (s *service) GetParameterChangeRequestsByParameterID(ctx context.Context, parameterID uint) ([]*model.ParameterChangeRequest, error) {
	return s.repo.GetParameterChangeRequestsByParameterID(ctx, parameterID)
}

// ApproveParameterChangeRequest approves a change request and applies the changes
func (s *service) ApproveParameterChangeRequest(ctx context.Context, id uint, userID uint, req *dto.ApproveParameterChangeRequestRequest) (*model.ParameterChangeRequest, error) {
	logger := log.Ctx(ctx).With().Str("service", "approve-parameter-change-request").Uint("id", id).Logger()

	// Get the change request
	changeRequest, err := s.repo.GetParameterChangeRequestByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("parameter change request with ID %d not found", id)
		}
		return nil, err
	}

	// Check if it's still pending
	if changeRequest.Status != model.ChangeRequestStatusPending {
		return nil, fmt.Errorf("change request is not pending (current status: %s)", changeRequest.Status)
	}

	// Get the underlying GORM DB
	db := s.repo.GetDB()

	// Start transaction
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Create a new repository instance with the transaction
	txRepo := repository.New(tx)

	// Execute the transaction logic
	err = func() error {
		// Get parameter with transaction
		parameter, err := txRepo.GetParameterByID(ctx, changeRequest.ParameterID)
		if err != nil {
			return fmt.Errorf("failed to get parameter: %w", err)
		}

		// Apply the changes from changeData
		if changeRequest.ChangeData.Name != nil {
			// Check if new name conflicts with existing
			if *changeRequest.ChangeData.Name != parameter.Name {
				existing, err := txRepo.GetParameterByName(ctx, *changeRequest.ChangeData.Name)
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if existing != nil && existing.ID != parameter.ID {
					return fmt.Errorf("parameter with name '%s' already exists", *changeRequest.ChangeData.Name)
				}
				parameter.Name = *changeRequest.ChangeData.Name
			}
		}

		if changeRequest.ChangeData.Description != nil {
			parameter.Description = *changeRequest.ChangeData.Description
		}

		if changeRequest.ChangeData.DataType != nil {
			parameter.DataType = *changeRequest.ChangeData.DataType
		}

		if changeRequest.ChangeData.DefaultRolloutValue != nil {
			parameter.DefaultRolloutValue = model.RolloutValue{
				Data: changeRequest.ChangeData.DefaultRolloutValue,
			}
		}

		// Update parameter metadata
		if err := txRepo.UpdateParameter(ctx, parameter); err != nil {
			return fmt.Errorf("failed to update parameter: %w", err)
		}

		// Handle rules replacement if provided
		if len(changeRequest.ChangeData.Rules) > 0 {
			// Delete all existing rules
			if err := txRepo.DeleteParameterRulesByParameterID(ctx, parameter.ID); err != nil {
				return fmt.Errorf("failed to delete existing rules: %w", err)
			}

			// Create new rules
			for _, ruleReq := range changeRequest.ChangeData.Rules {
				rule := &model.ParameterRule{
					Name:         ruleReq.Name,
					Description:  ruleReq.Description,
					Type:         ruleReq.Type,
					ParameterID:  parameter.ID,
					RolloutValue: model.RolloutValue{Data: ruleReq.RolloutValue},
					SegmentID:    ruleReq.SegmentID,
					MatchType:    ruleReq.MatchType,
				}

				if err := txRepo.CreateParameterRule(ctx, rule); err != nil {
					return fmt.Errorf("failed to create rule '%s': %w", ruleReq.Name, err)
				}

				// Add conditions for attribute-based rules
				if ruleReq.Type == model.RuleTypeAttribute && len(ruleReq.Conditions) > 0 {
					for _, conditionReq := range ruleReq.Conditions {
						condition := &model.ParameterRuleCondition{
							RuleID:      rule.ID,
							AttributeID: conditionReq.AttributeID,
							Operator:    conditionReq.Operator,
							Value:       conditionReq.Value,
						}
						if err := txRepo.CreateParameterRuleCondition(ctx, condition); err != nil {
							return fmt.Errorf("failed to create condition for rule '%s': %w", ruleReq.Name, err)
						}
					}
				}
			}
		}

		// Update parameter raw value
		if err := txRepo.UpdateParameterRawValue(ctx, parameter.ID); err != nil {
			logger.Error().Err(err).Msg("Failed to update parameter raw_value")
		}

		// Update the change request status
		now := time.Now()
		changeRequest.Status = model.ChangeRequestStatusApproved
		changeRequest.ReviewedByUserID = &userID
		changeRequest.ReviewedAt = &now
		if err := txRepo.UpdateParameterChangeRequest(ctx, changeRequest); err != nil {
			return fmt.Errorf("failed to update change request status: %w", err)
		}

		// Enqueue sync parameter job
		logger.Info().Msg("Enqueuing sync parameter job")
		_, err = s.riverClient.Insert(ctx, dto.SyncParameterArgs{
			ParameterID: int(parameter.ID),
		}, nil)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to enqueue sync parameter job")
		}

		return nil
	}()

	if err != nil {
		// Rollback on error
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			return nil, fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rollbackErr)
		}
		return nil, err
	}

	// Commit the transaction
	if commitErr := tx.Commit().Error; commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	// Reload with relationships
	return s.repo.GetParameterChangeRequestByID(ctx, changeRequest.ID)
}

// RejectParameterChangeRequest rejects a change request
func (s *service) RejectParameterChangeRequest(ctx context.Context, id uint, userID uint, req *dto.RejectParameterChangeRequestRequest) (*model.ParameterChangeRequest, error) {
	logger := log.Ctx(ctx).With().Str("service", "reject-parameter-change-request").Uint("id", id).Logger()

	// Get the change request
	changeRequest, err := s.repo.GetParameterChangeRequestByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("parameter change request with ID %d not found", id)
		}
		return nil, err
	}

	// Check if it's still pending
	if changeRequest.Status != model.ChangeRequestStatusPending {
		return nil, fmt.Errorf("change request is not pending (current status: %s)", changeRequest.Status)
	}

	// Update status to rejected
	now := time.Now()
	changeRequest.Status = model.ChangeRequestStatusRejected
	changeRequest.ReviewedByUserID = &userID
	changeRequest.ReviewedAt = &now

	if err := s.repo.UpdateParameterChangeRequest(ctx, changeRequest); err != nil {
		logger.Error().Err(err).Msg("Failed to update change request status")
		return nil, err
	}

	// Reload with relationships
	return s.repo.GetParameterChangeRequestByID(ctx, changeRequest.ID)
}
