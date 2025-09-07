package repository

import (
	"api/internal/model"
	"context"

	"gorm.io/gorm"
)

// Repository defines the interface for all data operations
type Repository interface {
	// Attribute operations
	CreateAttribute(ctx context.Context, attribute *model.Attribute) error
	GetAttributeByID(ctx context.Context, id uint) (*model.Attribute, error)
	GetAttributeByName(ctx context.Context, name string) (*model.Attribute, error)
	GetAllAttributes(ctx context.Context, limit, offset int) ([]*model.Attribute, error)
	UpdateAttribute(ctx context.Context, attribute *model.Attribute) error
	DeleteAttribute(ctx context.Context, id uint) error
	GetAttributesByDataType(ctx context.Context, dataType model.DataType, limit, offset int) ([]*model.Attribute, error)
	IncrementAttributeUsageCount(ctx context.Context, id uint) error
	DecrementAttributeUsageCount(ctx context.Context, id uint) error
	CountAttributes(ctx context.Context) (int64, error)

	// Segment operations
	CreateSegment(ctx context.Context, segment *model.Segment) error
	GetSegmentByID(ctx context.Context, id uint) (*model.Segment, error)
	GetSegmentByName(ctx context.Context, name string) (*model.Segment, error)
	GetAllSegments(ctx context.Context, limit, offset int) ([]*model.Segment, error)
	UpdateSegment(ctx context.Context, segment *model.Segment) error
	DeleteSegment(ctx context.Context, id uint) error
	CountSegments(ctx context.Context) (int64, error)

	// Segment Rule operations
	CreateSegmentRule(ctx context.Context, rule *model.SegmentRule) error
	GetSegmentRulesBySegmentID(ctx context.Context, segmentID uint) ([]*model.SegmentRule, error)
	DeleteSegmentRulesBySegmentID(ctx context.Context, segmentID uint) error

	// Segment Rule Condition operations
	CreateSegmentRuleCondition(ctx context.Context, condition *model.SegmentRuleCondition) error
	GetSegmentRuleConditionsByRuleID(ctx context.Context, ruleID uint) ([]*model.SegmentRuleCondition, error)
	DeleteSegmentRuleConditionsByRuleID(ctx context.Context, ruleID uint) error
}

// repository implements Repository
type repository struct {
	db *gorm.DB
}

// New creates a new repository
func New(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

// CreateAttribute creates a new attribute
func (r *repository) CreateAttribute(ctx context.Context, attribute *model.Attribute) error {
	return r.db.WithContext(ctx).Create(attribute).Error
}

// GetAttributeByID retrieves an attribute by ID
func (r *repository) GetAttributeByID(ctx context.Context, id uint) (*model.Attribute, error) {
	var attribute model.Attribute
	err := r.db.WithContext(ctx).First(&attribute, id).Error
	if err != nil {
		return nil, err
	}
	return &attribute, nil
}

// GetAttributeByName retrieves an attribute by name
func (r *repository) GetAttributeByName(ctx context.Context, name string) (*model.Attribute, error) {
	var attribute model.Attribute
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&attribute).Error
	if err != nil {
		return nil, err
	}
	return &attribute, nil
}

// GetAllAttributes retrieves all attributes with pagination
func (r *repository) GetAllAttributes(ctx context.Context, limit, offset int) ([]*model.Attribute, error) {
	var attributes []*model.Attribute
	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&attributes).Error
	return attributes, err
}

// UpdateAttribute updates an existing attribute
func (r *repository) UpdateAttribute(ctx context.Context, attribute *model.Attribute) error {
	return r.db.WithContext(ctx).Save(attribute).Error
}

// DeleteAttribute deletes an attribute by ID
func (r *repository) DeleteAttribute(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Attribute{}, id).Error
}

// GetAttributesByDataType retrieves attributes by data type with pagination
func (r *repository) GetAttributesByDataType(ctx context.Context, dataType model.DataType, limit, offset int) ([]*model.Attribute, error) {
	var attributes []*model.Attribute
	query := r.db.WithContext(ctx).Where("data_type = ?", dataType).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&attributes).Error
	return attributes, err
}

// IncrementAttributeUsageCount increments the usage count for an attribute
func (r *repository) IncrementAttributeUsageCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Attribute{}).Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

// DecrementAttributeUsageCount decrements the usage count for an attribute
func (r *repository) DecrementAttributeUsageCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Attribute{}).Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count - ?", 1)).Error
}

// CountAttributes returns the total number of attributes
func (r *repository) CountAttributes(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Attribute{}).Count(&count).Error
	return count, err
}
