package repository

import (
	"api/internal/model"
	"context"
)

// CreateSegment creates a new segment with its rules and conditions
func (r *repository) CreateSegment(ctx context.Context, segment *model.Segment) error {
	return r.db.WithContext(ctx).Create(segment).Error
}

// GetSegmentByID retrieves a segment by ID with all its rules and conditions
func (r *repository) GetSegmentByID(ctx context.Context, id uint) (*model.Segment, error) {
	var segment model.Segment
	err := r.db.WithContext(ctx).
		Preload("Rules").
		Preload("Rules.Conditions").
		Preload("Rules.Conditions.Attribute").
		First(&segment, id).Error
	if err != nil {
		return nil, err
	}
	return &segment, nil
}

// GetSegmentByName retrieves a segment by name
func (r *repository) GetSegmentByName(ctx context.Context, name string) (*model.Segment, error) {
	var segment model.Segment
	err := r.db.WithContext(ctx).
		Preload("Rules").
		Preload("Rules.Conditions").
		Preload("Rules.Conditions.Attribute").
		Where("name = ?", name).
		First(&segment).Error
	if err != nil {
		return nil, err
	}
	return &segment, nil
}

// GetAllSegments retrieves all segments with pagination
func (r *repository) GetAllSegments(ctx context.Context, limit, offset int) ([]*model.Segment, error) {
	var segments []*model.Segment
	query := r.db.WithContext(ctx).
		Preload("Rules").
		Preload("Rules.Conditions").
		Preload("Rules.Conditions.Attribute").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&segments).Error
	return segments, err
}

// UpdateSegment updates an existing segment
func (r *repository) UpdateSegment(ctx context.Context, segment *model.Segment) error {
	return r.db.WithContext(ctx).Save(segment).Error
}

// DeleteSegment deletes a segment by ID (cascade will handle rules and conditions)
func (r *repository) DeleteSegment(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Segment{}, id).Error
}

// CountSegments returns the total number of segments
func (r *repository) CountSegments(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Segment{}).Count(&count).Error
	return count, err
}

// CreateSegmentRule creates a new segment rule
func (r *repository) CreateSegmentRule(ctx context.Context, rule *model.SegmentRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetSegmentRulesBySegmentID retrieves all rules for a segment
func (r *repository) GetSegmentRulesBySegmentID(ctx context.Context, segmentID uint) ([]*model.SegmentRule, error) {
	var rules []*model.SegmentRule
	err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Conditions.Attribute").
		Where("segment_id = ?", segmentID).
		Find(&rules).Error
	return rules, err
}

// DeleteSegmentRulesBySegmentID deletes all rules for a segment
func (r *repository) DeleteSegmentRulesBySegmentID(ctx context.Context, segmentID uint) error {
	return r.db.WithContext(ctx).Where("segment_id = ?", segmentID).Delete(&model.SegmentRule{}).Error
}

// CreateSegmentRuleCondition creates a new segment rule condition
func (r *repository) CreateSegmentRuleCondition(ctx context.Context, condition *model.SegmentRuleCondition) error {
	return r.db.WithContext(ctx).Create(condition).Error
}

// GetSegmentRuleConditionsByRuleID retrieves all conditions for a rule
func (r *repository) GetSegmentRuleConditionsByRuleID(ctx context.Context, ruleID uint) ([]*model.SegmentRuleCondition, error) {
	var conditions []*model.SegmentRuleCondition
	err := r.db.WithContext(ctx).
		Preload("Attribute").
		Where("rule_id = ?", ruleID).
		Find(&conditions).Error
	return conditions, err
}

// DeleteSegmentRuleConditionsByRuleID deletes all conditions for a rule
func (r *repository) DeleteSegmentRuleConditionsByRuleID(ctx context.Context, ruleID uint) error {
	return r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Delete(&model.SegmentRuleCondition{}).Error
}
