package service

import (
	"api/internal/dto"
	"api/internal/model"
	"api/internal/repository"
	"context"
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

	// Segment operations
	CreateSegment(ctx context.Context, req *dto.CreateSegmentRequest) (*model.Segment, error)
	GetSegmentByID(ctx context.Context, id uint) (*model.Segment, error)
	GetSegmentByName(ctx context.Context, name string) (*model.Segment, error)
	GetAllSegments(ctx context.Context) ([]*model.Segment, error)
	UpdateSegment(ctx context.Context, id uint, req *dto.UpdateSegmentRequest) (*model.Segment, error)
	DeleteSegment(ctx context.Context, id uint) error
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
