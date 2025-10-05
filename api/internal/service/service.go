package service

import (
	"api/internal/dto"
	"api/internal/model"
	"api/internal/repository"
	"context"
	"sdk"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
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

	// Parameter operations
	CreateParameter(ctx context.Context, req *dto.CreateParameterRequest) (*model.Parameter, error)
	GetParameterByID(ctx context.Context, id uint) (*model.Parameter, error)
	GetParameterByName(ctx context.Context, name string) (*model.Parameter, error)
	GetAllParameters(ctx context.Context) ([]*model.Parameter, error)
	GetAllParametersSDK(ctx context.Context) ([]sdk.Parameter, error)
	UpdateParameter(ctx context.Context, id uint, req *dto.UpdateParameterRequest) (*model.Parameter, error)
	UpdateParameterWithRules(ctx context.Context, id uint, req *dto.UpdateParameterWithRulesRequest) (*model.Parameter, error)
	DeleteParameter(ctx context.Context, id uint) error
	AddParameterRule(ctx context.Context, parameterID uint, req *dto.CreateParameterRuleRequest) (*model.Parameter, error)
	UpdateParameterRule(ctx context.Context, parameterID uint, ruleID uint, req *dto.UpdateParameterRuleRequest) (*model.Parameter, error)
	DeleteParameterRule(ctx context.Context, parameterID uint, ruleID uint) (*model.Parameter, error)
	IncrementParameterUsageCount(ctx context.Context, id uint) error
	DecrementParameterUsageCount(ctx context.Context, id uint) error

	// Experiment operations
	CreateExperiment(ctx context.Context, req *dto.CreateExperimentRequest) (string, error)
	GetAllExperiments(ctx context.Context) ([]*model.Experiment, error)
	GetExperimentByID(ctx context.Context, id uint) (*model.Experiment, []*model.ExperimentVariant, map[int][]*model.ExperimentVariantParameter, *model.Attribute, error)
	RejectExperiment(ctx context.Context, id uint, req *dto.RejectExperimentRequest) (*model.Experiment, error)
	ApproveExperiment(ctx context.Context, id uint, req *dto.ApproveExperimentRequest) (*model.Experiment, error)
	AbortExperiment(ctx context.Context, id uint, req *dto.AbortExperimentRequest) (*model.Experiment, error)
	SimulateParameter(ctx context.Context, req *dto.SimulateParameterRequest) (dto.SimulateParameterResponse, error)
}

// service implements Service
type service struct {
	repo         repository.Repository
	riverClient  *river.Client[pgx.Tx]
	auroraClient sdk.Client
}

// New creates a new service
func New(repo repository.Repository, riverClient *river.Client[pgx.Tx], auroraClient sdk.Client) Service {
	return &service{
		repo:         repo,
		riverClient:  riverClient,
		auroraClient: auroraClient,
	}
}
