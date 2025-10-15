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
