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

// Helper function to parse ID from string
func ParseID(idStr string) (uint, error) {
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
