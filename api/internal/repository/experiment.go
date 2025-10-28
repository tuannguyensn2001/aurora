package repository

import (
	"api/internal/model"
	"context"

	"api/internal/constant"

	"gorm.io/gorm"
)

// CreateExperiment creates a new experiment
func (r *repository) CreateExperiment(ctx context.Context, experiment *model.Experiment) error {
	// Check if there's a transaction in context
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx.WithContext(ctx).Create(experiment).Error
	}
	return r.db.WithContext(ctx).Create(experiment).Error
}

// GetExperimentByID retrieves an experiment by ID with all related data preloaded
func (r *repository) GetExperimentByID(ctx context.Context, id uint) (*model.Experiment, error) {
	var experiment model.Experiment
	err := r.db.WithContext(ctx).
		Preload("Segment").
		Preload("Segment.Rules").
		Preload("Segment.Rules.Conditions").
		Preload("Segment.Rules.Conditions.Attribute").
		Preload("HashAttribute").
		Preload("Variants").
		Preload("Variants.Parameters").
		First(&experiment, id).Error
	if err != nil {
		return nil, err
	}
	return &experiment, nil
}

// GetExperimentByUuid retrieves an experiment by UUID
func (r *repository) GetExperimentByUuid(ctx context.Context, uuid string) (*model.Experiment, error) {
	var experiment model.Experiment
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&experiment).Error
	if err != nil {
		return nil, err
	}
	return &experiment, nil
}

// GetAllExperiments retrieves all experiments with pagination
func (r *repository) GetAllExperiments(ctx context.Context, limit, offset int) ([]*model.Experiment, error) {
	var experiments []*model.Experiment
	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&experiments).Error
	return experiments, err
}

// UpdateExperiment updates an existing experiment
func (r *repository) UpdateExperiment(ctx context.Context, experiment *model.Experiment) error {
	return r.db.WithContext(ctx).Save(experiment).Error
}

// DeleteExperiment deletes an experiment by ID
func (r *repository) DeleteExperiment(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Experiment{}, id).Error
}

// CountExperiments returns the total number of experiments
func (r *repository) CountExperiments(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Experiment{}).Count(&count).Error
	return count, err
}

// CreateExperimentVariant creates a new experiment variant
func (r *repository) CreateExperimentVariant(ctx context.Context, variant *model.ExperimentVariant) error {
	// Check if there's a transaction in context
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx.WithContext(ctx).Create(variant).Error
	}
	return r.db.WithContext(ctx).Create(variant).Error
}

// GetExperimentVariantsByExperimentID retrieves all variants for an experiment
func (r *repository) GetExperimentVariantsByExperimentID(ctx context.Context, experimentID uint) ([]*model.ExperimentVariant, error) {
	var variants []*model.ExperimentVariant
	err := r.db.WithContext(ctx).Where("experiment_id = ?", experimentID).Order("created_at ASC").Find(&variants).Error
	return variants, err
}

// DeleteExperimentVariantsByExperimentID deletes all variants for an experiment
func (r *repository) DeleteExperimentVariantsByExperimentID(ctx context.Context, experimentID uint) error {
	return r.db.WithContext(ctx).Where("experiment_id = ?", experimentID).Delete(&model.ExperimentVariant{}).Error
}

// CreateExperimentVariantParameter creates a new experiment variant parameter
func (r *repository) CreateExperimentVariantParameter(ctx context.Context, parameter *model.ExperimentVariantParameter) error {
	// Check if there's a transaction in context
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx.WithContext(ctx).Create(parameter).Error
	}
	return r.db.WithContext(ctx).Create(parameter).Error
}

// GetExperimentVariantParametersByVariantID retrieves all parameters for a variant
func (r *repository) GetExperimentVariantParametersByVariantID(ctx context.Context, variantID uint) ([]*model.ExperimentVariantParameter, error) {
	var parameters []*model.ExperimentVariantParameter
	err := r.db.WithContext(ctx).Where("experiment_variant_id = ?", variantID).Order("created_at ASC").Find(&parameters).Error
	return parameters, err
}

// GetExperimentVariantParametersByExperimentID retrieves all parameters for an experiment
func (r *repository) GetExperimentVariantParametersByExperimentID(ctx context.Context, experimentID uint) ([]*model.ExperimentVariantParameter, error) {
	var parameters []*model.ExperimentVariantParameter
	err := r.db.WithContext(ctx).Where("experiment_id = ?", experimentID).Order("created_at ASC").Find(&parameters).Error
	return parameters, err
}

// DeleteExperimentVariantParametersByVariantID deletes all parameters for a variant
func (r *repository) DeleteExperimentVariantParametersByVariantID(ctx context.Context, variantID uint) error {
	return r.db.WithContext(ctx).Where("experiment_variant_id = ?", variantID).Delete(&model.ExperimentVariantParameter{}).Error
}

func (r *repository) GetExperimentByName(ctx context.Context, name string) (*model.Experiment, error) {
	var experiment model.Experiment
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&experiment).Error
	if err != nil {
		return nil, err
	}
	return &experiment, nil
}

func (r *repository) GetExperimentsActive(ctx context.Context) ([]model.Experiment, error) {

	result := make([]model.Experiment, 0)
	err := r.db.WithContext(ctx).
		Where("status in (?)", []string{constant.ExperimentStatusSchedule, constant.ExperimentStatusRunning}).
		Preload("HashAttribute").
		Preload("Variants").
		Preload("Variants.Parameters").
		Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindConflictingExperiments finds experiments that conflict with the given parameters
// Conflicts occur when:
// 1. Same parameter IDs are used
// 2. Segments can have overlapping users (sophisticated segment analysis)
// 3. Time periods overlap
// 4. Status is schedule or running
func (r *repository) FindConflictingExperiments(ctx context.Context, parameterIDs []int, segmentID int, startDate, endDate int64) ([]*model.Experiment, error) {
	var experiments []*model.Experiment

	// Build the query to find experiments with same parameters and overlapping time
	// Time overlap logic: NOT (b < x OR y < a) = (b >= x AND y >= a)
	// Where [a,b] is existing experiment time window and [x,y] is new experiment time window
	query := r.db.WithContext(ctx).
		Joins("JOIN experiment_variant_parameters evp ON experiments.id = evp.experiment_id").
		Where("experiments.status IN (?)", []string{constant.ExperimentStatusSchedule, constant.ExperimentStatusRunning}).
		Where("evp.parameter_id IN (?)", parameterIDs).
		Where("experiments.end_date >= ? AND ? >= experiments.start_date", startDate, endDate)

	// Load experiments with segment details for sophisticated overlap analysis
	err := query.
		Preload("Segment").
		Preload("Segment.Rules").
		Preload("Segment.Rules.Conditions").
		Preload("Segment.Rules.Conditions.Attribute").
		Preload("Variants").
		Preload("Variants.Parameters").
		Distinct("experiments.*").
		Find(&experiments).Error

	return experiments, err
}

// UpdateExperimentRawValue updates the raw_value field for an experiment after loading all related data
func (r *repository) UpdateExperimentRawValue(ctx context.Context, id uint) error {
	// Get the experiment with all related data loaded
	var experiment model.Experiment
	err := r.db.WithContext(ctx).
		Preload("Segment").
		Preload("Segment.Rules").
		Preload("Segment.Rules.Conditions").
		Preload("Segment.Rules.Conditions.Attribute").
		Preload("HashAttribute").
		Preload("Variants").
		Preload("Variants.Parameters").
		First(&experiment, id).Error
	if err != nil {
		return err
	}

	// Populate raw value with all related data
	if err := experiment.PopulateRawValue(); err != nil {
		return err
	}

	// Update only the raw_value field
	return r.db.WithContext(ctx).Model(&experiment).Select("raw_value").Updates(map[string]interface{}{
		"raw_value": experiment.RawValue,
	}).Error
}
