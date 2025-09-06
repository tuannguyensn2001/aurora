package service

import (
	"api/internal/dto"
	"api/internal/model"
	"api/internal/repository"
	"context"
	"errors"

	"gorm.io/gorm"
)

// Service defines the interface for all business logic operations
type Service interface {
	// Attribute operations
	CreateAttribute(ctx context.Context, req *dto.CreateAttributeRequest) (*model.Attribute, error)
	GetAttributeByID(ctx context.Context, id uint) (*model.Attribute, error)
	GetAttributeByName(ctx context.Context, name string) (*model.Attribute, error)
	GetAllAttributes(ctx context.Context) ([]*model.Attribute, error)
	UpdateAttribute(ctx context.Context, id uint, req *dto.UpdateAttributeRequest) (*model.Attribute, error)
	DeleteAttribute(ctx context.Context, id uint) error
	IncrementAttributeUsageCount(ctx context.Context, id uint) error
	DecrementAttributeUsageCount(ctx context.Context, id uint) error
}

// service implements Service
type service struct {
	repo repository.Repository
}

// New creates a new service
func New(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}

// CreateAttribute creates a new attribute
func (s *service) CreateAttribute(ctx context.Context, req *dto.CreateAttributeRequest) (*model.Attribute, error) {
	// Check if attribute with same name already exists
	existing, err := s.repo.GetAttributeByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("attribute with name '" + req.Name + "' already exists")
	}

	// Validate enum options for enum data type
	if req.DataType == model.DataTypeEnum {
		if req.EnumOptions == nil || len(req.EnumOptions) == 0 {
			return nil, errors.New("enum options are required for enum data type")
		}
	}

	// Set default values
	hashAttribute := false
	if req.HashAttribute != nil {
		hashAttribute = *req.HashAttribute
	}

	enumOptions := []string{}
	if req.DataType == model.DataTypeEnum && req.EnumOptions != nil {
		enumOptions = req.EnumOptions
	}

	attribute := &model.Attribute{
		Name:          req.Name,
		Description:   req.Description,
		DataType:      req.DataType,
		HashAttribute: hashAttribute,
		EnumOptions:   enumOptions,
		UsageCount:    0,
	}

	if err := s.repo.CreateAttribute(ctx, attribute); err != nil {
		return nil, err
	}

	return attribute, nil
}

// GetAttributeByID retrieves an attribute by ID
func (s *service) GetAttributeByID(ctx context.Context, id uint) (*model.Attribute, error) {
	attribute, err := s.repo.GetAttributeByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attribute with ID " + string(rune(id)) + " not found")
		}
		return nil, err
	}
	return attribute, nil
}

// GetAttributeByName retrieves an attribute by name
func (s *service) GetAttributeByName(ctx context.Context, name string) (*model.Attribute, error) {
	return s.repo.GetAttributeByName(ctx, name)
}

// GetAllAttributes retrieves all attributes
func (s *service) GetAllAttributes(ctx context.Context) ([]*model.Attribute, error) {
	return s.repo.GetAllAttributes(ctx, 0, 0) // No pagination for findAll equivalent
}

// UpdateAttribute updates an existing attribute
func (s *service) UpdateAttribute(ctx context.Context, id uint, req *dto.UpdateAttributeRequest) (*model.Attribute, error) {
	attribute, err := s.GetAttributeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if name is being updated and if it conflicts
	if req.Name != nil && *req.Name != attribute.Name {
		existing, err := s.repo.GetAttributeByName(ctx, *req.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("attribute with name '" + *req.Name + "' already exists")
		}
		attribute.Name = *req.Name
	}

	// Update other fields if provided
	if req.Description != nil {
		attribute.Description = *req.Description
	}
	if req.HashAttribute != nil {
		attribute.HashAttribute = *req.HashAttribute
	}

	// Handle data type changes
	if req.DataType != nil {
		// Validate enum options if data type is being changed to enum
		if *req.DataType == model.DataTypeEnum {
			if req.EnumOptions == nil || len(req.EnumOptions) == 0 {
				return nil, errors.New("enum options are required for enum data type")
			}
		}

		// Clear enum options if data type is being changed from enum
		if *req.DataType != model.DataTypeEnum {
			attribute.EnumOptions = []string{}
		}

		attribute.DataType = *req.DataType
	}

	// Update enum options if provided
	if req.EnumOptions != nil {
		attribute.EnumOptions = req.EnumOptions
	}

	if err := s.repo.UpdateAttribute(ctx, attribute); err != nil {
		return nil, err
	}

	return attribute, nil
}

// DeleteAttribute deletes an attribute
func (s *service) DeleteAttribute(ctx context.Context, id uint) error {
	attribute, err := s.GetAttributeByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if attribute is being used in experiments
	if attribute.UsageCount > 0 {
		return errors.New("cannot delete attribute '" + attribute.Name + "' as it is being used in " + string(rune(attribute.UsageCount)) + " experiment(s)")
	}

	return s.repo.DeleteAttribute(ctx, id)
}

// IncrementAttributeUsageCount increments the usage count for an attribute
func (s *service) IncrementAttributeUsageCount(ctx context.Context, id uint) error {
	// Check if attribute exists
	_, err := s.GetAttributeByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.IncrementAttributeUsageCount(ctx, id)
}

// DecrementAttributeUsageCount decrements the usage count for an attribute
func (s *service) DecrementAttributeUsageCount(ctx context.Context, id uint) error {
	// Check if attribute exists
	_, err := s.GetAttributeByID(ctx, id)
	if err != nil {
		return err
	}

	// Get current attribute to check usage count
	attribute, err := s.repo.GetAttributeByID(ctx, id)
	if err != nil {
		return err
	}

	// Don't decrement below 0
	if attribute.UsageCount <= 0 {
		return nil
	}

	return s.repo.DecrementAttributeUsageCount(ctx, id)
}
