package repository

import (
	"api/internal/model"
	"context"
)

// CreateParameterChangeRequest creates a new parameter change request
func (r *repository) CreateParameterChangeRequest(ctx context.Context, changeRequest *model.ParameterChangeRequest) error {
	return r.db.WithContext(ctx).Create(changeRequest).Error
}

// GetParameterChangeRequestByID retrieves a parameter change request by ID with all related data
func (r *repository) GetParameterChangeRequestByID(ctx context.Context, id uint) (*model.ParameterChangeRequest, error) {
	var changeRequest model.ParameterChangeRequest
	err := r.db.WithContext(ctx).
		Preload("Parameter").
		Preload("RequestedByUser").
		Preload("ReviewedByUser").
		First(&changeRequest, id).Error
	if err != nil {
		return nil, err
	}
	return &changeRequest, nil
}

// GetPendingParameterChangeRequestByParameterID retrieves pending change request for a specific parameter
func (r *repository) GetPendingParameterChangeRequestByParameterID(ctx context.Context, parameterID uint) (*model.ParameterChangeRequest, error) {
	var changeRequest model.ParameterChangeRequest
	err := r.db.WithContext(ctx).
		Where("parameter_id = ? AND status = ?", parameterID, model.ChangeRequestStatusPending).
		Preload("Parameter").
		Preload("RequestedByUser").
		First(&changeRequest).Error
	if err != nil {
		return nil, err
	}
	return &changeRequest, nil
}

// GetParameterChangeRequestsByParameterID retrieves all change requests for a specific parameter
func (r *repository) GetParameterChangeRequestsByParameterID(ctx context.Context, parameterID uint) ([]*model.ParameterChangeRequest, error) {
	var changeRequests []*model.ParameterChangeRequest
	err := r.db.WithContext(ctx).
		Where("parameter_id = ?", parameterID).
		Preload("Parameter").
		Preload("RequestedByUser").
		Preload("ReviewedByUser").
		Order("created_at DESC").
		Find(&changeRequests).Error
	return changeRequests, err
}

// UpdateParameterChangeRequest updates an existing parameter change request
func (r *repository) UpdateParameterChangeRequest(ctx context.Context, changeRequest *model.ParameterChangeRequest) error {
	return r.db.WithContext(ctx).Save(changeRequest).Error
}

// GetParameterChangeRequestsByStatus retrieves parameter change requests by status with basic info
func (r *repository) GetParameterChangeRequestsByStatus(ctx context.Context, status model.ParameterChangeRequestStatus, limit, offset int) ([]*model.ParameterChangeRequest, error) {
	var changeRequests []*model.ParameterChangeRequest
	query := r.db.WithContext(ctx).
		Where("status = ?", status).
		Preload("Parameter").
		Preload("RequestedByUser").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&changeRequests).Error
	return changeRequests, err
}

// GetParameterChangeRequestByIDWithDetails retrieves a parameter change request by ID with all related data for detailed view
func (r *repository) GetParameterChangeRequestByIDWithDetails(ctx context.Context, id uint) (*model.ParameterChangeRequest, error) {
	var changeRequest model.ParameterChangeRequest
	err := r.db.WithContext(ctx).
		Preload("Parameter").
		Preload("Parameter.Rules").
		Preload("Parameter.Rules.Conditions").
		Preload("Parameter.Rules.Conditions.Attribute").
		Preload("Parameter.Rules.Segment").
		Preload("RequestedByUser").
		Preload("ReviewedByUser").
		First(&changeRequest, id).Error
	if err != nil {
		return nil, err
	}
	return &changeRequest, nil
}

// GetParameterChangeRequestsByStatusWithPagination retrieves parameter change requests by status with pagination and count
func (r *repository) GetParameterChangeRequestsByStatusWithPagination(ctx context.Context, status model.ParameterChangeRequestStatus, limit, offset int) ([]*model.ParameterChangeRequest, int64, error) {
	var changeRequests []*model.ParameterChangeRequest
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).
		Model(&model.ParameterChangeRequest{}).
		Where("status = ?", status).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := r.db.WithContext(ctx).
		Where("status = ?", status).
		Preload("Parameter").
		Preload("RequestedByUser").
		Preload("ReviewedByUser").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err = query.Find(&changeRequests).Error
	return changeRequests, total, err
}
