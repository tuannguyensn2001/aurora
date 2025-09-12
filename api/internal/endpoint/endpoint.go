package endpoint

import (
	"context"
	"strconv"

	"api/internal/dto"
	"api/internal/handler"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints holds all endpoint functions
type Endpoints struct {
	CreateAttribute              endpoint.Endpoint
	GetAllAttributes             endpoint.Endpoint
	GetAttributeByID             endpoint.Endpoint
	UpdateAttribute              endpoint.Endpoint
	DeleteAttribute              endpoint.Endpoint
	IncrementAttributeUsageCount endpoint.Endpoint
	DecrementAttributeUsageCount endpoint.Endpoint
	CreateSegment                endpoint.Endpoint
	GetAllSegments               endpoint.Endpoint
	GetSegmentByID               endpoint.Endpoint
	UpdateSegment                endpoint.Endpoint
	DeleteSegment                endpoint.Endpoint
	CreateParameter              endpoint.Endpoint
	GetAllParameters             endpoint.Endpoint
	GetParameterByID             endpoint.Endpoint
	UpdateParameter              endpoint.Endpoint
	UpdateParameterWithRules     endpoint.Endpoint
	DeleteParameter              endpoint.Endpoint
	AddParameterRule             endpoint.Endpoint
	UpdateParameterRule          endpoint.Endpoint
	DeleteParameterRule          endpoint.Endpoint
	CreateExperiment             endpoint.Endpoint
	GetAllExperiments            endpoint.Endpoint
	GetExperimentByID            endpoint.Endpoint
	RejectExperiment             endpoint.Endpoint
}

// MakeEndpoints creates all endpoints
func MakeEndpoints(h *handler.Handler) Endpoints {
	return Endpoints{
		CreateAttribute:              makeCreateAttributeEndpoint(h),
		GetAllAttributes:             makeGetAllAttributesEndpoint(h),
		GetAttributeByID:             makeGetAttributeByIDEndpoint(h),
		UpdateAttribute:              makeUpdateAttributeEndpoint(h),
		DeleteAttribute:              makeDeleteAttributeEndpoint(h),
		IncrementAttributeUsageCount: makeIncrementAttributeUsageCountEndpoint(h),
		DecrementAttributeUsageCount: makeDecrementAttributeUsageCountEndpoint(h),
		CreateSegment:                makeCreateSegmentEndpoint(h),
		GetAllSegments:               makeGetAllSegmentsEndpoint(h),
		GetSegmentByID:               makeGetSegmentByIDEndpoint(h),
		UpdateSegment:                makeUpdateSegmentEndpoint(h),
		DeleteSegment:                makeDeleteSegmentEndpoint(h),
		CreateParameter:              makeCreateParameterEndpoint(h),
		GetAllParameters:             makeGetAllParametersEndpoint(h),
		GetParameterByID:             makeGetParameterByIDEndpoint(h),
		UpdateParameter:              makeUpdateParameterEndpoint(h),
		UpdateParameterWithRules:     makeUpdateParameterWithRulesEndpoint(h),
		DeleteParameter:              makeDeleteParameterEndpoint(h),
		AddParameterRule:             makeAddParameterRuleEndpoint(h),
		UpdateParameterRule:          makeUpdateParameterRuleEndpoint(h),
		DeleteParameterRule:          makeDeleteParameterRuleEndpoint(h),
		CreateExperiment:             makeCreateExperimentEndpoint(h),
		GetAllExperiments:            makeGetAllExperimentsEndpoint(h),
		GetExperimentByID:            makeGetExperimentByIDEndpoint(h),
		RejectExperiment:             makeRejectExperimentEndpoint(h),
	}
}

// CreateAttributeRequest represents the endpoint request
type CreateAttributeRequest struct {
	Request dto.CreateAttributeRequest `json:"request"`
}

func makeCreateAttributeEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateAttributeRequest)
		return h.CreateAttribute(ctx, &req.Request)
	}
}

// GetAllAttributesRequest represents the endpoint request
type GetAllAttributesRequest struct{}

func makeGetAllAttributesEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return h.GetAllAttributes(ctx)
	}
}

// GetAttributeByIDRequest represents the endpoint request
type GetAttributeByIDRequest struct {
	ID uint `json:"id"`
}

func makeGetAttributeByIDEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetAttributeByIDRequest)
		return h.GetAttributeByID(ctx, req.ID)
	}
}

// UpdateAttributeRequest represents the endpoint request
type UpdateAttributeRequest struct {
	ID      uint                       `json:"id"`
	Request dto.UpdateAttributeRequest `json:"request"`
}

func makeUpdateAttributeEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateAttributeRequest)
		return h.UpdateAttribute(ctx, req.ID, &req.Request)
	}
}

// DeleteAttributeRequest represents the endpoint request
type DeleteAttributeRequest struct {
	ID uint `json:"id"`
}

func makeDeleteAttributeEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteAttributeRequest)
		err := h.DeleteAttribute(ctx, req.ID)
		return nil, err
	}
}

// IncrementAttributeUsageCountRequest represents the endpoint request
type IncrementAttributeUsageCountRequest struct {
	ID uint `json:"id"`
}

func makeIncrementAttributeUsageCountEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(IncrementAttributeUsageCountRequest)
		err := h.IncrementAttributeUsageCount(ctx, req.ID)
		return nil, err
	}
}

// DecrementAttributeUsageCountRequest represents the endpoint request
type DecrementAttributeUsageCountRequest struct {
	ID uint `json:"id"`
}

func makeDecrementAttributeUsageCountEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DecrementAttributeUsageCountRequest)
		err := h.DecrementAttributeUsageCount(ctx, req.ID)
		return nil, err
	}
}

// CreateSegmentRequest represents the endpoint request
type CreateSegmentRequest struct {
	Request dto.CreateSegmentRequest `json:"request"`
}

func makeCreateSegmentEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateSegmentRequest)
		return h.CreateSegment(ctx, &req.Request)
	}
}

// GetAllSegmentsRequest represents the endpoint request
type GetAllSegmentsRequest struct{}

func makeGetAllSegmentsEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return h.GetAllSegments(ctx)
	}
}

// GetSegmentByIDRequest represents the endpoint request
type GetSegmentByIDRequest struct {
	ID uint `json:"id"`
}

func makeGetSegmentByIDEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetSegmentByIDRequest)
		return h.GetSegmentByID(ctx, req.ID)
	}
}

// UpdateSegmentRequest represents the endpoint request
type UpdateSegmentRequest struct {
	ID      uint                     `json:"id"`
	Request dto.UpdateSegmentRequest `json:"request"`
}

func makeUpdateSegmentEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateSegmentRequest)
		return h.UpdateSegment(ctx, req.ID, &req.Request)
	}
}

// DeleteSegmentRequest represents the endpoint request
type DeleteSegmentRequest struct {
	ID uint `json:"id"`
}

func makeDeleteSegmentEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteSegmentRequest)
		err := h.DeleteSegment(ctx, req.ID)
		return nil, err
	}
}

// Parameter endpoint requests and implementations

// CreateParameterRequest represents the endpoint request
type CreateParameterRequest struct {
	Request dto.CreateParameterRequest `json:"request"`
}

func makeCreateParameterEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateParameterRequest)
		return h.CreateParameter(ctx, &req.Request)
	}
}

// GetAllParametersRequest represents the endpoint request
type GetAllParametersRequest struct{}

func makeGetAllParametersEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return h.GetAllParameters(ctx)
	}
}

// GetParameterByIDRequest represents the endpoint request
type GetParameterByIDRequest struct {
	ID uint `json:"id"`
}

func makeGetParameterByIDEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetParameterByIDRequest)
		return h.GetParameterByID(ctx, req.ID)
	}
}

// UpdateParameterRequest represents the endpoint request
type UpdateParameterRequest struct {
	ID      uint                       `json:"id"`
	Request dto.UpdateParameterRequest `json:"request"`
}

func makeUpdateParameterEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateParameterRequest)
		return h.UpdateParameter(ctx, req.ID, &req.Request)
	}
}

// UpdateParameterWithRulesRequest represents the endpoint request
type UpdateParameterWithRulesRequest struct {
	ID      uint                                `json:"id"`
	Request dto.UpdateParameterWithRulesRequest `json:"request"`
}

func makeUpdateParameterWithRulesEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateParameterWithRulesRequest)
		return h.UpdateParameterWithRules(ctx, req.ID, &req.Request)
	}
}

// DeleteParameterRequest represents the endpoint request
type DeleteParameterRequest struct {
	ID uint `json:"id"`
}

func makeDeleteParameterEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteParameterRequest)
		err := h.DeleteParameter(ctx, req.ID)
		return nil, err
	}
}

// AddParameterRuleRequest represents the endpoint request
type AddParameterRuleRequest struct {
	ID      uint                           `json:"id"`
	Request dto.CreateParameterRuleRequest `json:"request"`
}

func makeAddParameterRuleEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AddParameterRuleRequest)
		return h.AddParameterRule(ctx, req.ID, &req.Request)
	}
}

// UpdateParameterRuleRequest represents the endpoint request
type UpdateParameterRuleRequest struct {
	ID      uint                           `json:"id"`
	RuleID  uint                           `json:"ruleId"`
	Request dto.UpdateParameterRuleRequest `json:"request"`
}

func makeUpdateParameterRuleEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateParameterRuleRequest)
		return h.UpdateParameterRule(ctx, req.ID, req.RuleID, &req.Request)
	}
}

// DeleteParameterRuleRequest represents the endpoint request
type DeleteParameterRuleRequest struct {
	ID     uint `json:"id"`
	RuleID uint `json:"ruleId"`
}

func makeDeleteParameterRuleEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteParameterRuleRequest)
		return h.DeleteParameterRule(ctx, req.ID, req.RuleID)
	}
}

// CreateExperimentRequest represents the endpoint request
type CreateExperimentRequest struct {
	Request dto.CreateExperimentRequest `json:"request"`
}

func makeCreateExperimentEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateExperimentRequest)
		return h.CreateExperiment(ctx, &req.Request)
	}
}

// GetAllExperimentsRequest represents the endpoint request
type GetAllExperimentsRequest struct{}

func makeGetAllExperimentsEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return h.GetAllExperiments(ctx)
	}
}

// GetExperimentByIDRequest represents the endpoint request
type GetExperimentByIDRequest struct {
	ID uint `json:"id"`
}

func makeGetExperimentByIDEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetExperimentByIDRequest)
		return h.GetExperimentByID(ctx, req.ID)
	}
}

// RejectExperimentRequest represents the endpoint request
type RejectExperimentRequest struct {
	ID      uint                        `json:"id"`
	Request dto.RejectExperimentRequest `json:"request"`
}

func makeRejectExperimentEndpoint(h *handler.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RejectExperimentRequest)
		return h.RejectExperiment(ctx, req.ID, &req.Request)
	}
}

// Helper function to parse ID from string
func ParseID(idStr string) (uint, error) {
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
