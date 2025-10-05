package service

import (
	"api/internal/dto"
	"api/internal/mapper"
	"api/internal/model"
	"api/internal/repository"
	"context"
	"errors"
	"fmt"
	"sdk"
	"strconv"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CreateParameter creates a new parameter
func (s *service) CreateParameter(ctx context.Context, req *dto.CreateParameterRequest) (*model.Parameter, error) {
	// Check if parameter with same name already exists
	existing, err := s.repo.GetParameterByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("parameter with name '" + req.Name + "' already exists")
	}

	// Validate default rollout value based on data type
	if err := s.validateParameterValue(req.DefaultRolloutValue, req.DataType); err != nil {
		return nil, err
	}

	parameter := &model.Parameter{
		Name:        req.Name,
		Description: req.Description,
		DataType:    req.DataType,
		DefaultRolloutValue: model.RolloutValue{
			Data: req.DefaultRolloutValue,
		},
		UsageCount: 0,
	}

	if err := s.repo.CreateParameter(ctx, parameter); err != nil {
		return nil, err
	}

	return parameter, nil
}

// GetParameterByID retrieves a parameter by ID
func (s *service) GetParameterByID(ctx context.Context, id uint) (*model.Parameter, error) {
	parameter, err := s.repo.GetParameterByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("parameter with ID %d not found", id)
		}
		return nil, err
	}

	// Ensure conditions and rules are always arrays
	if parameter.Conditions == nil {
		parameter.Conditions = []model.ParameterCondition{}
	}
	if parameter.Rules == nil {
		parameter.Rules = []model.ParameterRule{}
	}

	return parameter, nil
}

// GetParameterByName retrieves a parameter by name
func (s *service) GetParameterByName(ctx context.Context, name string) (*model.Parameter, error) {
	return s.repo.GetParameterByName(ctx, name)
}

// GetAllParameters retrieves all parameters
func (s *service) GetAllParameters(ctx context.Context) ([]*model.Parameter, error) {
	parameters, err := s.repo.GetAllParameters(ctx, 0, 0) // No pagination for findAll equivalent
	if err != nil {
		return nil, err
	}

	// Ensure conditions and rules are always arrays for each parameter
	for _, parameter := range parameters {
		if parameter.Conditions == nil {
			parameter.Conditions = []model.ParameterCondition{}
		}
		if parameter.Rules == nil {
			parameter.Rules = []model.ParameterRule{}
		}
	}

	return parameters, nil
}

// UpdateParameter updates an existing parameter
func (s *service) UpdateParameter(ctx context.Context, id uint, req *dto.UpdateParameterRequest) (*model.Parameter, error) {
	logger := log.Ctx(ctx).With().Str("service", "update-parameter").Uint("id", id).Logger()
	parameter, err := s.GetParameterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if name is being updated and if it conflicts
	if req.Name != nil && *req.Name != parameter.Name {
		existing, err := s.repo.GetParameterByName(ctx, *req.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("parameter with name '" + *req.Name + "' already exists")
		}
		parameter.Name = *req.Name
	}

	// Update description if provided
	if req.Description != nil {
		parameter.Description = *req.Description
	}

	// Validate default rollout value if being updated
	if req.DefaultRolloutValue != nil {
		dataType := parameter.DataType
		if req.DataType != nil {
			dataType = *req.DataType
		}
		if err := s.validateParameterValue(req.DefaultRolloutValue, dataType); err != nil {
			return nil, err
		}
		parameter.DefaultRolloutValue = model.RolloutValue{
			Data: req.DefaultRolloutValue,
		}
	}

	// If data type is being changed, validate all existing condition values
	if req.DataType != nil && *req.DataType != parameter.DataType {
		for _, condition := range parameter.Conditions {
			if err := s.validateParameterValue(condition.RolloutValue.Data, *req.DataType); err != nil {
				return nil, fmt.Errorf("existing condition rollout value is invalid for new data type: %v", err)
			}
		}
		for _, rule := range parameter.Rules {
			if err := s.validateParameterValue(rule.RolloutValue.Data, *req.DataType); err != nil {
				return nil, fmt.Errorf("existing rule rollout value is invalid for new data type: %v", err)
			}
		}
		parameter.DataType = *req.DataType
	}

	if err := s.repo.UpdateParameter(ctx, parameter); err != nil {
		return nil, err
	}

	logger.Info().Msg("Enqueuing sync parameter job")
	_, err = s.riverClient.Insert(ctx, dto.SyncParameterArgs{
		ParameterID: int(id),
	}, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to enqueue sync parameter job")
		return nil, err
	}

	return parameter, nil
}

// UpdateParameterWithRules updates a parameter and completely replaces all its rules
func (s *service) UpdateParameterWithRules(ctx context.Context, id uint, req *dto.UpdateParameterWithRulesRequest) (*model.Parameter, error) {
	logger := log.Ctx(ctx).With().Str("service", "update-parameter-with-rules").Uint("id", id).Logger()
	// Use database transaction to ensure atomicity
	return s.withTransaction(ctx, func(txRepo repository.Repository) (*model.Parameter, error) {
		parameter, err := txRepo.GetParameterByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("parameter with ID %d not found", id)
			}
			return nil, err
		}

		// Check if name is being updated and if it conflicts
		if req.Name != nil && *req.Name != parameter.Name {
			existing, err := txRepo.GetParameterByName(ctx, *req.Name)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			if existing != nil {
				return nil, errors.New("parameter with name '" + *req.Name + "' already exists")
			}
			parameter.Name = *req.Name
		}

		// Update description if provided
		if req.Description != nil {
			parameter.Description = *req.Description
		}

		// Determine the final data type for validation
		finalDataType := parameter.DataType
		if req.DataType != nil {
			finalDataType = *req.DataType
		}

		// Validate default rollout value if being updated
		if req.DefaultRolloutValue != nil {
			if err := s.validateParameterValue(req.DefaultRolloutValue, finalDataType); err != nil {
				return nil, err
			}
			parameter.DefaultRolloutValue = model.RolloutValue{
				Data: req.DefaultRolloutValue,
			}
		}

		// If data type is being changed, validate all existing condition values
		if req.DataType != nil && *req.DataType != parameter.DataType {
			for _, condition := range parameter.Conditions {
				if err := s.validateParameterValue(condition.RolloutValue.Data, *req.DataType); err != nil {
					return nil, fmt.Errorf("existing condition rollout value is invalid for new data type: %v", err)
				}
			}
			parameter.DataType = *req.DataType
		}

		// Update parameter metadata first
		if err := txRepo.UpdateParameter(ctx, parameter); err != nil {
			return nil, err
		}

		// Handle rules replacement if provided
		if req.Rules != nil {
			// Delete all existing rules for this parameter
			if err := txRepo.DeleteParameterRulesByParameterID(ctx, id); err != nil {
				return nil, fmt.Errorf("failed to delete existing rules: %v", err)
			}

			// Create new rules
			for _, ruleReq := range req.Rules {
				// Validate rollout value for the rule
				if err := s.validateParameterValue(ruleReq.RolloutValue, finalDataType); err != nil {
					return nil, fmt.Errorf("invalid rollout value for rule '%s': %v", ruleReq.Name, err)
				}

				rule := &model.ParameterRule{
					Name:         ruleReq.Name,
					Description:  ruleReq.Description,
					Type:         ruleReq.Type,
					ParameterID:  id,
					RolloutValue: model.RolloutValue{Data: ruleReq.RolloutValue},
					SegmentID:    ruleReq.SegmentID,
					MatchType:    ruleReq.MatchType,
				}

				// Validate segment-based rule requirements
				if ruleReq.Type == model.RuleTypeSegment {
					if ruleReq.SegmentID == nil || ruleReq.MatchType == nil {
						return nil, fmt.Errorf("segment ID and match type are required for segment-based rule '%s'", ruleReq.Name)
					}
					// Validate that segment exists
					_, err := txRepo.GetSegmentByID(ctx, *ruleReq.SegmentID)
					if err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							return nil, fmt.Errorf("segment with ID %d not found for rule '%s'", *ruleReq.SegmentID, ruleReq.Name)
						}
						return nil, err
					}
				}

				// Create the rule
				if err := txRepo.CreateParameterRule(ctx, rule); err != nil {
					return nil, fmt.Errorf("failed to create rule '%s': %v", ruleReq.Name, err)
				}

				// Add conditions if it's an attribute-based rule
				if ruleReq.Type == model.RuleTypeAttribute && len(ruleReq.Conditions) > 0 {
					for _, conditionReq := range ruleReq.Conditions {
						// Validate that attribute exists
						_, err := txRepo.GetAttributeByID(ctx, conditionReq.AttributeID)
						if err != nil {
							if errors.Is(err, gorm.ErrRecordNotFound) {
								return nil, fmt.Errorf("attribute with ID %d not found for rule '%s'", conditionReq.AttributeID, ruleReq.Name)
							}
							return nil, err
						}

						condition := &model.ParameterRuleCondition{
							RuleID:      rule.ID,
							AttributeID: conditionReq.AttributeID,
							Operator:    conditionReq.Operator,
							Value:       conditionReq.Value,
						}
						if err := txRepo.CreateParameterRuleCondition(ctx, condition); err != nil {
							return nil, fmt.Errorf("failed to create condition for rule '%s': %v", ruleReq.Name, err)
						}
					}
				}
			}
		}

		// Return updated parameter with all rules
		logger.Info().Msg("Enqueuing sync parameter job")
		_, err = s.riverClient.Insert(ctx, dto.SyncParameterArgs{
			ParameterID: int(id),
		}, nil)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to enqueue sync parameter job")
			return nil, fmt.Errorf("failed to enqueue sync parameter job: %v", err)
		}

		return txRepo.GetParameterByID(ctx, id)
	})
}

// DeleteParameter deletes a parameter
func (s *service) DeleteParameter(ctx context.Context, id uint) error {
	parameter, err := s.GetParameterByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if parameter is being used in experiments
	if parameter.UsageCount > 0 {
		return fmt.Errorf("cannot delete parameter '%s' as it is being used in %d experiment(s)", parameter.Name, parameter.UsageCount)
	}

	return s.repo.DeleteParameter(ctx, id)
}

// AddParameterRule adds a rule to a parameter
func (s *service) AddParameterRule(ctx context.Context, parameterID uint, req *dto.CreateParameterRuleRequest) (*model.Parameter, error) {
	parameter, err := s.GetParameterByID(ctx, parameterID)
	if err != nil {
		return nil, err
	}

	// Validate rollout value based on parameter data type
	if err := s.validateParameterValue(req.RolloutValue, parameter.DataType); err != nil {
		return nil, err
	}

	// For segment-based rules
	if req.Type == model.RuleTypeSegment {
		if req.SegmentID == nil || req.MatchType == nil {
			return nil, errors.New("segment ID and match type are required for segment-based rules")
		}

		// Validate that segment exists
		_, err := s.repo.GetSegmentByID(ctx, *req.SegmentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("segment with ID %d not found", *req.SegmentID)
			}
			return nil, err
		}

		// Check if rule for this segment already exists
		for _, rule := range parameter.Rules {
			if rule.Type == model.RuleTypeSegment && rule.SegmentID != nil && *rule.SegmentID == *req.SegmentID {
				return nil, fmt.Errorf("rule for segment ID %d already exists for this parameter", *req.SegmentID)
			}
		}
	}

	// For attribute-based rules
	if req.Type == model.RuleTypeAttribute && len(req.Conditions) > 0 {
		// Validate that attributes exist
		for _, condition := range req.Conditions {
			_, err := s.repo.GetAttributeByID(ctx, condition.AttributeID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, fmt.Errorf("attribute with ID %d not found", condition.AttributeID)
				}
				return nil, err
			}
		}
	}

	rule := &model.ParameterRule{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		RolloutValue: model.RolloutValue{
			Data: req.RolloutValue,
		},
		ParameterID: parameterID,
		SegmentID:   req.SegmentID,
		MatchType:   req.MatchType,
	}

	if err := s.repo.CreateParameterRule(ctx, rule); err != nil {
		return nil, err
	}

	// Add conditions if it's an attribute-based rule
	if req.Type == model.RuleTypeAttribute && len(req.Conditions) > 0 {
		for _, conditionReq := range req.Conditions {
			condition := &model.ParameterRuleCondition{
				RuleID:      rule.ID,
				AttributeID: conditionReq.AttributeID,
				Operator:    conditionReq.Operator,
				Value:       conditionReq.Value,
			}
			if err := s.repo.CreateParameterRuleCondition(ctx, condition); err != nil {
				return nil, err
			}
		}
	}

	// Return updated parameter with rules
	return s.GetParameterByID(ctx, parameterID)
}

// UpdateParameterRule updates a parameter rule
func (s *service) UpdateParameterRule(ctx context.Context, parameterID uint, ruleID uint, req *dto.UpdateParameterRuleRequest) (*model.Parameter, error) {
	parameter, err := s.GetParameterByID(ctx, parameterID)
	if err != nil {
		return nil, err
	}

	rule, err := s.repo.GetParameterRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("rule with ID %d not found for parameter %d", ruleID, parameterID)
		}
		return nil, err
	}

	if rule.ParameterID != parameterID {
		return nil, fmt.Errorf("rule with ID %d does not belong to parameter %d", ruleID, parameterID)
	}

	// Validate rollout value if being updated
	if req.RolloutValue != nil {
		if err := s.validateParameterValue(req.RolloutValue, parameter.DataType); err != nil {
			return nil, err
		}
		rule.RolloutValue = model.RolloutValue{
			Data: req.RolloutValue,
		}
	}

	// Update rule properties
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.Type != nil {
		rule.Type = *req.Type
	}
	if req.SegmentID != nil {
		rule.SegmentID = req.SegmentID
	}
	if req.MatchType != nil {
		rule.MatchType = req.MatchType
	}

	// If type is being changed, validate new type requirements
	if req.Type != nil && *req.Type != rule.Type {
		if *req.Type == model.RuleTypeSegment {
			segmentID := rule.SegmentID
			matchType := rule.MatchType
			if req.SegmentID != nil {
				segmentID = req.SegmentID
			}
			if req.MatchType != nil {
				matchType = req.MatchType
			}
			if segmentID == nil || matchType == nil {
				return nil, errors.New("segment ID and match type are required for segment-based rules")
			}
			// Validate that segment exists
			_, err := s.repo.GetSegmentByID(ctx, *segmentID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, fmt.Errorf("segment with ID %d not found", *segmentID)
				}
				return nil, err
			}
		}
	}

	// For segment-based rules, validate segment if being updated
	if req.SegmentID != nil && (rule.SegmentID == nil || *req.SegmentID != *rule.SegmentID) {
		_, err := s.repo.GetSegmentByID(ctx, *req.SegmentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("segment with ID %d not found", *req.SegmentID)
			}
			return nil, err
		}

		// Check if rule for this segment already exists
		for _, existingRule := range parameter.Rules {
			if existingRule.Type == model.RuleTypeSegment && existingRule.SegmentID != nil &&
				*existingRule.SegmentID == *req.SegmentID && existingRule.ID != ruleID {
				return nil, fmt.Errorf("rule for segment ID %d already exists for this parameter", *req.SegmentID)
			}
		}
	}

	if err := s.repo.UpdateParameterRule(ctx, rule); err != nil {
		return nil, err
	}

	// Handle conditions update for attribute-based rules
	if len(req.Conditions) > 0 {
		// Remove existing conditions
		if err := s.repo.DeleteParameterRuleConditionsByRuleID(ctx, ruleID); err != nil {
			return nil, err
		}

		// Add new conditions
		for _, conditionReq := range req.Conditions {
			// Validate that attribute exists
			_, err := s.repo.GetAttributeByID(ctx, conditionReq.AttributeID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, fmt.Errorf("attribute with ID %d not found", conditionReq.AttributeID)
				}
				return nil, err
			}

			condition := &model.ParameterRuleCondition{
				RuleID:      ruleID,
				AttributeID: conditionReq.AttributeID,
				Operator:    conditionReq.Operator,
				Value:       conditionReq.Value,
			}
			if err := s.repo.CreateParameterRuleCondition(ctx, condition); err != nil {
				return nil, err
			}
		}
	}

	// Return updated parameter with rules
	return s.GetParameterByID(ctx, parameterID)
}

// DeleteParameterRule deletes a parameter rule
func (s *service) DeleteParameterRule(ctx context.Context, parameterID uint, ruleID uint) (*model.Parameter, error) {
	rule, err := s.repo.GetParameterRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("rule with ID %d not found for parameter %d", ruleID, parameterID)
		}
		return nil, err
	}

	if rule.ParameterID != parameterID {
		return nil, fmt.Errorf("rule with ID %d does not belong to parameter %d", ruleID, parameterID)
	}

	if err := s.repo.DeleteParameterRule(ctx, ruleID); err != nil {
		return nil, err
	}

	// Return updated parameter with rules
	return s.GetParameterByID(ctx, parameterID)
}

// IncrementParameterUsageCount increments the usage count for a parameter
func (s *service) IncrementParameterUsageCount(ctx context.Context, id uint) error {
	return s.repo.IncrementParameterUsageCount(ctx, id)
}

// DecrementParameterUsageCount decrements the usage count for a parameter
func (s *service) DecrementParameterUsageCount(ctx context.Context, id uint) error {
	return s.repo.DecrementParameterUsageCount(ctx, id)
}

// validateParameterValue validates a parameter value based on its data type
func (s *service) validateParameterValue(value interface{}, dataType model.ParameterDataType) error {
	switch dataType {
	case model.ParameterDataTypeBoolean:
		if _, ok := value.(bool); !ok {
			return errors.New("value must be a boolean for boolean parameter")
		}
	case model.ParameterDataTypeString:
		if _, ok := value.(string); !ok {
			return errors.New("value must be a string for string parameter")
		}
	case model.ParameterDataTypeNumber:
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid number types
		case nil:
			return errors.New("value cannot be nil for number parameter")
		default:
			return fmt.Errorf("value must be a valid number for number parameter, got %T", v)
		}
	default:
		return fmt.Errorf("unsupported parameter data type: %s", dataType)
	}
	return nil
}

// withTransaction executes a function within a database transaction
func (s *service) withTransaction(ctx context.Context, fn func(repository.Repository) (*model.Parameter, error)) (*model.Parameter, error) {
	// Get the underlying GORM DB from the repository
	db := s.getDB()

	// Start transaction
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Create a new repository instance with the transaction
	txRepo := repository.New(tx)

	// Execute the function with the transaction repository
	result, err := fn(txRepo)
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

	return result, nil
}

// getDB extracts the underlying GORM DB from the repository
func (s *service) getDB() *gorm.DB {
	// This is a temporary solution - we need to expose the DB from the repository
	// For now, we'll use reflection or add a method to the repository interface
	type dbGetter interface {
		GetDB() *gorm.DB
	}

	if getter, ok := s.repo.(dbGetter); ok {
		return getter.GetDB()
	}

	// Fallback - this shouldn't happen in a well-designed system
	panic("repository does not expose underlying database connection")
}

func (s *service) SimulateParameter(ctx context.Context, req *dto.SimulateParameterRequest) (dto.SimulateParameterResponse, error) {

	logger := log.Ctx(ctx).With().Str("service", "simulate-parameter").Logger()

	attribute := sdk.NewAttribute()
	for _, attributeReq := range req.Attributes {
		logger.Info().Str("attribute", attributeReq.Name).Str("dataType", string(attributeReq.DataType)).Str("value", attributeReq.Value).Msg("Simulating attribute")
		switch attributeReq.DataType {
		case model.DataTypeBoolean:
			value, err := strconv.ParseBool(attributeReq.Value)
			if err == nil {
				attribute.SetBool(attributeReq.Name, value)
			}
		case model.DataTypeString:
			attribute.SetString(attributeReq.Name, attributeReq.Value)
		case model.DataTypeNumber:
			value, err := strconv.ParseFloat(attributeReq.Value, 64)
			if err == nil {
				attribute.SetNumber(attributeReq.Name, value)
			}
		case model.DataTypeEnum:
			attribute.SetString(attributeReq.Name, attributeReq.Value)
		}

	}

	logger.Info().Interface("attribute", attribute.Keys()).Msg("Simulating parameter")

	rolloutValue := s.auroraClient.EvaluateParameter(ctx, req.ParameterName, attribute)
	var value interface{}
	switch req.ParameterType {
	case model.ParameterDataTypeBoolean:
		value = rolloutValue.AsBool(false)
	case model.ParameterDataTypeString:
		value = rolloutValue.AsString("")
	case model.ParameterDataTypeNumber:
		value = rolloutValue.AsNumber(0)
	}

	return dto.SimulateParameterResponse{
		Value: value,
	}, nil
}

func (s *service) GetAllParametersSDK(ctx context.Context) ([]sdk.Parameter, error) {
	parameters, err := s.repo.GetAllParameters(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	return mapper.ParametersToSDK(parameters)
}
