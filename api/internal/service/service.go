package service

import (
	"api/config"
	"api/internal/dto"
	"api/internal/external/solver"
	"api/internal/model"
	"api/internal/repository"
	"context"
	"sdk"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
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
	CheckSegmentOverlap(ctx context.Context, segmentIDs []uint) (bool, error)

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

	// Parameter Change Request operations
	CreateParameterChangeRequest(ctx context.Context, userID uint, req *dto.CreateParameterChangeRequestRequest) (*model.ParameterChangeRequest, error)
	GetParameterChangeRequestByID(ctx context.Context, id uint) (*model.ParameterChangeRequest, error)
	GetParameterChangeRequestByIDWithDetails(ctx context.Context, id uint) (*model.ParameterChangeRequest, error)
	GetPendingParameterChangeRequestByParameterID(ctx context.Context, parameterID uint) (*model.ParameterChangeRequest, error)
	GetParameterChangeRequestsByParameterID(ctx context.Context, parameterID uint) ([]*model.ParameterChangeRequest, error)
	GetParameterChangeRequestsByStatus(ctx context.Context, status model.ParameterChangeRequestStatus, limit, offset int) ([]*model.ParameterChangeRequest, int64, error)
	ApproveParameterChangeRequest(ctx context.Context, id uint, userID uint, req *dto.ApproveParameterChangeRequestRequest) (*model.ParameterChangeRequest, error)
	RejectParameterChangeRequest(ctx context.Context, id uint, userID uint, req *dto.RejectParameterChangeRequestRequest) (*model.ParameterChangeRequest, error)

	// Experiment operations
	CreateExperiment(ctx context.Context, req *dto.CreateExperimentRequest) (string, error)
	GetAllExperiments(ctx context.Context) ([]*model.Experiment, error)
	GetExperimentByID(ctx context.Context, id uint) (*model.Experiment, []*model.ExperimentVariant, map[int][]*model.ExperimentVariantParameter, *model.Attribute, error)
	RejectExperiment(ctx context.Context, id uint, req *dto.RejectExperimentRequest) (*model.Experiment, error)
	ApproveExperiment(ctx context.Context, id uint, req *dto.ApproveExperimentRequest) (*model.Experiment, error)
	AbortExperiment(ctx context.Context, id uint, req *dto.AbortExperimentRequest) (*model.Experiment, error)
	SimulateParameter(ctx context.Context, req *dto.SimulateParameterRequest) (dto.SimulateParameterResponse, error)
	GetActiveExperimentsSDK(ctx context.Context) ([]sdk.Experiment, error)

	// Auth operations
	GetGoogleOAuthConfig(cfg *config.Config) *oauth2.Config
	GenerateStateToken() (string, error)
	GetGoogleLoginURL(ctx context.Context, cfg *config.Config) (string, string, error)
	HandleGoogleCallback(ctx context.Context, cfg *config.Config, code string, state string) (*dto.AuthResponse, error)
	GetGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error)
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	RefreshToken(ctx context.Context, cfg *config.Config, refreshToken string) (*dto.AuthResponse, error)

	// Event operations
	TrackEvent(ctx context.Context, req *dto.TrackEventRequest) (*dto.TrackEventResponse, error)
	TrackBatchEvent(ctx context.Context, req *dto.TrackBatchEventRequest) (*dto.TrackBatchEventResponse, error)
}

// service implements Service
type service struct {
	repo         repository.Repository
	riverClient  *river.Client[pgx.Tx]
	auroraClient sdk.Client
	solver       solver.Solver
	eventService *EventService
}

// New creates a new service
func New(repo repository.Repository, riverClient *river.Client[pgx.Tx], auroraClient sdk.Client, solver solver.Solver) Service {
	// Create event service
	eventRepo := repository.NewEventRepository(repo.GetDB())
	eventService := NewEventService(eventRepo, log.Logger)

	return &service{
		repo:         repo,
		riverClient:  riverClient,
		auroraClient: auroraClient,
		solver:       solver,
		eventService: eventService,
	}
}

// TrackEvent tracks an evaluation event
func (s *service) TrackEvent(ctx context.Context, req *dto.TrackEventRequest) (*dto.TrackEventResponse, error) {
	return s.eventService.TrackEvent(ctx, req)
}

// TrackBatchEvent tracks multiple evaluation events in batch
func (s *service) TrackBatchEvent(ctx context.Context, req *dto.TrackBatchEventRequest) (*dto.TrackBatchEventResponse, error) {
	return s.eventService.TrackBatchEvent(ctx, req)
}
