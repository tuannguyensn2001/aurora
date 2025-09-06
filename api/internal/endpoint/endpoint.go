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

// Helper function to parse ID from string
func ParseID(idStr string) (uint, error) {
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
