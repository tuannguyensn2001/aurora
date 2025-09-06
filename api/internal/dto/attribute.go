package dto

import (
	"api/internal/model"
	"time"
)

// CreateAttributeRequest represents the request to create an attribute
type CreateAttributeRequest struct {
	Name          string         `json:"name" validate:"required"`
	Description   string         `json:"description" validate:"required"`
	DataType      model.DataType `json:"dataType" validate:"required,oneof=boolean string number enum"`
	HashAttribute *bool          `json:"hashAttribute,omitempty"`
	EnumOptions   []string       `json:"enumOptions,omitempty"`
}

// UpdateAttributeRequest represents the request to update an attribute
type UpdateAttributeRequest struct {
	Name          *string         `json:"name,omitempty"`
	Description   *string         `json:"description,omitempty"`
	DataType      *model.DataType `json:"dataType,omitempty"`
	HashAttribute *bool           `json:"hashAttribute,omitempty"`
	EnumOptions   []string        `json:"enumOptions,omitempty"`
}

// AttributeResponse represents the response for attribute operations
type AttributeResponse struct {
	ID            uint           `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	DataType      model.DataType `json:"dataType"`
	HashAttribute bool           `json:"hashAttribute"`
	EnumOptions   []string       `json:"enumOptions"`
	UsageCount    int            `json:"usageCount"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// AttributeListResponse represents the response for listing attributes
type AttributeListResponse struct {
	Attributes []AttributeResponse `json:"attributes"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ToAttributeResponse converts model.Attribute to AttributeResponse
func ToAttributeResponse(attr *model.Attribute) AttributeResponse {
	return AttributeResponse{
		ID:            attr.ID,
		Name:          attr.Name,
		Description:   attr.Description,
		DataType:      attr.DataType,
		HashAttribute: attr.HashAttribute,
		EnumOptions:   attr.EnumOptions,
		UsageCount:    attr.UsageCount,
		CreatedAt:     attr.CreatedAt,
		UpdatedAt:     attr.UpdatedAt,
	}
}

// ToAttributeListResponse converts slice of model.Attribute to AttributeListResponse
func ToAttributeListResponse(attrs []*model.Attribute) AttributeListResponse {
	responses := make([]AttributeResponse, len(attrs))
	for i, attr := range attrs {
		responses[i] = ToAttributeResponse(attr)
	}
	return AttributeListResponse{
		Attributes: responses,
	}
}
