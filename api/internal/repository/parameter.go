package repository

import (
	"api/internal/model"
	"context"
)

// CreateParameter creates a new parameter with its rules and conditions
func (r *repository) CreateParameter(ctx context.Context, parameter *model.Parameter) error {
	return r.db.WithContext(ctx).Create(parameter).Error
}

// GetParameterByID retrieves a parameter by ID with all its rules and conditions
func (r *repository) GetParameterByID(ctx context.Context, id uint) (*model.Parameter, error) {
	var parameter model.Parameter
	err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Conditions.Segment").
		Preload("Rules").
		Preload("Rules.Segment").
		Preload("Rules.Conditions").
		Preload("Rules.Conditions.Attribute").
		First(&parameter, id).Error
	if err != nil {
		return nil, err
	}
	return &parameter, nil
}

// GetParameterByName retrieves a parameter by name
func (r *repository) GetParameterByName(ctx context.Context, name string) (*model.Parameter, error) {
	var parameter model.Parameter
	err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Conditions.Segment").
		Preload("Rules").
		Preload("Rules.Segment").
		Preload("Rules.Conditions").
		Preload("Rules.Conditions.Attribute").
		Where("name = ?", name).
		First(&parameter).Error
	if err != nil {
		return nil, err
	}
	return &parameter, nil
}

// GetAllParameters retrieves all parameters with pagination
func (r *repository) GetAllParameters(ctx context.Context, limit, offset int) ([]*model.Parameter, error) {
	var parameters []*model.Parameter
	query := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Conditions.Segment").
		Preload("Rules").
		Preload("Rules.Segment").
		Preload("Rules.Segment.Rules").
		Preload("Rules.Segment.Rules.Conditions").
		Preload("Rules.Segment.Rules.Conditions.Attribute").
		Preload("Rules.Conditions").
		Preload("Rules.Conditions.Attribute").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&parameters).Error
	return parameters, err
}

// UpdateParameter updates an existing parameter
func (r *repository) UpdateParameter(ctx context.Context, parameter *model.Parameter) error {
	return r.db.WithContext(ctx).Save(parameter).Error
}

// DeleteParameter deletes a parameter by ID (cascade will handle rules and conditions)
func (r *repository) DeleteParameter(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Parameter{}, id).Error
}

// IncrementParameterUsageCount increments the usage count for a parameter
func (r *repository) IncrementParameterUsageCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Parameter{}).Where("id = ?", id).
		UpdateColumn("usage_count", r.db.Raw("usage_count + ?", 1)).Error
}

// DecrementParameterUsageCount decrements the usage count for a parameter
func (r *repository) DecrementParameterUsageCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Parameter{}).Where("id = ?", id).
		UpdateColumn("usage_count", r.db.Raw("usage_count - ?", 1)).Error
}

// CountParameters returns the total number of parameters
func (r *repository) CountParameters(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Parameter{}).Count(&count).Error
	return count, err
}

// CreateParameterRule creates a new parameter rule
func (r *repository) CreateParameterRule(ctx context.Context, rule *model.ParameterRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetParameterRuleByID retrieves a parameter rule by ID with all its conditions
func (r *repository) GetParameterRuleByID(ctx context.Context, id uint) (*model.ParameterRule, error) {
	var rule model.ParameterRule
	err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Conditions.Attribute").
		Preload("Segment").
		First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetParameterRulesByParameterID retrieves all rules for a parameter
func (r *repository) GetParameterRulesByParameterID(ctx context.Context, parameterID uint) ([]*model.ParameterRule, error) {
	var rules []*model.ParameterRule
	err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Conditions.Attribute").
		Preload("Segment").
		Where("parameter_id = ?", parameterID).
		Find(&rules).Error
	return rules, err
}

// UpdateParameterRule updates an existing parameter rule
func (r *repository) UpdateParameterRule(ctx context.Context, rule *model.ParameterRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// DeleteParameterRule deletes a parameter rule by ID
func (r *repository) DeleteParameterRule(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ParameterRule{}, id).Error
}

// DeleteParameterRulesByParameterID deletes all rules for a parameter
func (r *repository) DeleteParameterRulesByParameterID(ctx context.Context, parameterID uint) error {
	return r.db.WithContext(ctx).Where("parameter_id = ?", parameterID).Delete(&model.ParameterRule{}).Error
}

// CreateParameterRuleCondition creates a new parameter rule condition
func (r *repository) CreateParameterRuleCondition(ctx context.Context, condition *model.ParameterRuleCondition) error {
	return r.db.WithContext(ctx).Create(condition).Error
}

// GetParameterRuleConditionsByRuleID retrieves all conditions for a rule
func (r *repository) GetParameterRuleConditionsByRuleID(ctx context.Context, ruleID uint) ([]*model.ParameterRuleCondition, error) {
	var conditions []*model.ParameterRuleCondition
	err := r.db.WithContext(ctx).
		Preload("Attribute").
		Where("rule_id = ?", ruleID).
		Find(&conditions).Error
	return conditions, err
}

// DeleteParameterRuleConditionsByRuleID deletes all conditions for a rule
func (r *repository) DeleteParameterRuleConditionsByRuleID(ctx context.Context, ruleID uint) error {
	return r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Delete(&model.ParameterRuleCondition{}).Error
}

// CreateParameterCondition creates a new parameter condition (legacy)
func (r *repository) CreateParameterCondition(ctx context.Context, condition *model.ParameterCondition) error {
	return r.db.WithContext(ctx).Create(condition).Error
}

// GetParameterConditionsByParameterID retrieves all conditions for a parameter (legacy)
func (r *repository) GetParameterConditionsByParameterID(ctx context.Context, parameterID uint) ([]*model.ParameterCondition, error) {
	var conditions []*model.ParameterCondition
	err := r.db.WithContext(ctx).
		Preload("Segment").
		Where("parameter_id = ?", parameterID).
		Find(&conditions).Error
	return conditions, err
}

// DeleteParameterConditionsByParameterID deletes all conditions for a parameter (legacy)
func (r *repository) DeleteParameterConditionsByParameterID(ctx context.Context, parameterID uint) error {
	return r.db.WithContext(ctx).Where("parameter_id = ?", parameterID).Delete(&model.ParameterCondition{}).Error
}

func (r *repository) GetParametersByIDs(ctx context.Context, ids []int) ([]model.Parameter, error) {
	var parameters []model.Parameter
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&parameters).Error
	return parameters, err
}

// GetAllParametersForSDK retrieves all parameters with only raw_value for SDK conversion
// This is optimized for SDK usage since raw_value already contains all necessary data
func (r *repository) GetAllParametersForSDK(ctx context.Context) ([]*model.Parameter, error) {
	var parameters []*model.Parameter
	err := r.db.WithContext(ctx).
		Select("id, raw_value").
		Order("id").
		Find(&parameters).Error
	return parameters, err
}

// UpdateParameterRawValue updates the raw_value field for a parameter after loading all related data
func (r *repository) UpdateParameterRawValue(ctx context.Context, id uint) error {
	// Get the parameter with all related data loaded
	var parameter model.Parameter
	err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Conditions.Segment").
		Preload("Rules").
		Preload("Rules.Segment").
		Preload("Rules.Segment.Rules").
		Preload("Rules.Segment.Rules.Conditions").
		Preload("Rules.Segment.Rules.Conditions.Attribute").
		Preload("Rules.Conditions").
		Preload("Rules.Conditions.Attribute").
		First(&parameter, id).Error
	if err != nil {
		return err
	}

	// Populate raw value with all related data
	if err := parameter.PopulateRawValue(); err != nil {
		return err
	}

	// Update only the raw_value field
	return r.db.WithContext(ctx).Model(&parameter).Select("raw_value").Updates(map[string]interface{}{
		"raw_value": parameter.RawValue,
	}).Error
}
