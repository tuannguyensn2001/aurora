package service

import (
	"api/internal/dto"
	"api/internal/model"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// CreateSegment creates a new segment with its rules and conditions
func (s *service) CreateSegment(ctx context.Context, req *dto.CreateSegmentRequest) (*model.Segment, error) {
	// Check if segment with same name already exists
	existing, err := s.repo.GetSegmentByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("segment with name '" + req.Name + "' already exists")
	}

	// Validate that all referenced attributes exist if rules are provided
	if len(req.Rules) > 0 {
		for _, ruleReq := range req.Rules {
			for _, conditionReq := range ruleReq.Conditions {
				_, err := s.repo.GetAttributeByID(ctx, conditionReq.AttributeID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, fmt.Errorf("attribute with ID %d not found", conditionReq.AttributeID)
					}
					return nil, err
				}
			}
		}
	}

	// Create segment with rules and conditions
	segment := &model.Segment{
		Name:        req.Name,
		Description: req.Description,
		Rules:       make([]model.SegmentRule, len(req.Rules)),
	}

	// Create rules and conditions
	for i, ruleReq := range req.Rules {
		rule := model.SegmentRule{
			Name:        ruleReq.Name,
			Description: ruleReq.Description,
			Conditions:  make([]model.SegmentRuleCondition, len(ruleReq.Conditions)),
		}

		for j, conditionReq := range ruleReq.Conditions {
			condition := model.SegmentRuleCondition{
				AttributeID: conditionReq.AttributeID,
				Operator:    conditionReq.Operator,
				Value:       conditionReq.Value,
			}
			rule.Conditions[j] = condition
		}

		segment.Rules[i] = rule
	}

	if err := s.repo.CreateSegment(ctx, segment); err != nil {
		return nil, err
	}

	return segment, nil
}

// GetSegmentByID retrieves a segment by ID
func (s *service) GetSegmentByID(ctx context.Context, id uint) (*model.Segment, error) {
	segment, err := s.repo.GetSegmentByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("segment with ID %d not found", id)
		}
		return nil, err
	}
	return segment, nil
}

// GetSegmentByName retrieves a segment by name
func (s *service) GetSegmentByName(ctx context.Context, name string) (*model.Segment, error) {
	return s.repo.GetSegmentByName(ctx, name)
}

// GetAllSegments retrieves all segments
func (s *service) GetAllSegments(ctx context.Context) ([]*model.Segment, error) {
	return s.repo.GetAllSegments(ctx, 0, 0) // No pagination for findAll equivalent
}

// UpdateSegment updates an existing segment
func (s *service) UpdateSegment(ctx context.Context, id uint, req *dto.UpdateSegmentRequest) (*model.Segment, error) {
	segment, err := s.GetSegmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if name is being updated and if it conflicts
	if req.Name != nil && *req.Name != segment.Name {
		existing, err := s.repo.GetSegmentByName(ctx, *req.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("segment with name '" + *req.Name + "' already exists")
		}
		segment.Name = *req.Name
	}

	// Update description if provided
	if req.Description != nil {
		segment.Description = *req.Description
	}

	// If rules are being updated, replace all existing rules
	if len(req.Rules) > 0 {
		// Validate that all referenced attributes exist
		for _, ruleReq := range req.Rules {
			for _, conditionReq := range ruleReq.Conditions {
				_, err := s.repo.GetAttributeByID(ctx, conditionReq.AttributeID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, fmt.Errorf("attribute with ID %d not found", conditionReq.AttributeID)
					}
					return nil, err
				}
			}
		}

		// Remove existing rules (cascade will handle conditions)
		if err := s.repo.DeleteSegmentRulesBySegmentID(ctx, id); err != nil {
			return nil, err
		}

		// Create new rules
		newRules := make([]model.SegmentRule, len(req.Rules))
		for i, ruleReq := range req.Rules {
			rule := model.SegmentRule{
				Name:        ruleReq.Name,
				Description: ruleReq.Description,
				SegmentID:   id,
				Conditions:  make([]model.SegmentRuleCondition, len(ruleReq.Conditions)),
			}

			for j, conditionReq := range ruleReq.Conditions {
				condition := model.SegmentRuleCondition{
					AttributeID: conditionReq.AttributeID,
					Operator:    conditionReq.Operator,
					Value:       conditionReq.Value,
				}
				rule.Conditions[j] = condition
			}

			newRules[i] = rule
		}

		segment.Rules = newRules
	}

	if err := s.repo.UpdateSegment(ctx, segment); err != nil {
		return nil, err
	}

	return segment, nil
}

// DeleteSegment deletes a segment
func (s *service) DeleteSegment(ctx context.Context, id uint) error {
	segment, err := s.GetSegmentByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.DeleteSegment(ctx, segment.ID)
}
