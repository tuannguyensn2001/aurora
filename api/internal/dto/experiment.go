package dto

import (
	"errors"

	"api/internal/model"

	"github.com/go-playground/validator/v10"
)

type CreateExperimentRequest struct {
	Name            string                           `json:"name" binding:"required" validate:"required"`
	Hypothesis      string                           `json:"hypothesis" binding:"required" validate:"required"`
	Description     string                           `json:"description" binding:"required" validate:"required"`
	StartDate       int64                            `json:"startDate" binding:"required" validate:"required"`
	EndDate         int64                            `json:"endDate" binding:"required" validate:"required"`
	HashAttributeID int                              `json:"hashAttributeId" binding:"required" validate:"required"`
	PopulationSize  int                              `json:"populationSize" binding:"required" validate:"required,min=1,max=100"`
	Strategy        string                           `json:"strategy" binding:"required" validate:"required,oneof=percentage_split"`
	Variants        []CreateExperimentVariantRequest `json:"variants" binding:"required" validate:"required"`
}

func (r *CreateExperimentRequest) Validate() error {

	if err := validator.New().Struct(r); err != nil {
		return err
	}

	if r.StartDate >= r.EndDate {
		return errors.New("startDate must be before endDate")
	}

	if len(r.Variants) == 0 {
		return errors.New("variants must be at least one")
	}

	for _, variant := range r.Variants {
		if len(variant.Parameters) == 0 {
			return errors.New("parameters must be at least one")
		}
	}

	totalTrafficAllocation := 0
	for _, variant := range r.Variants {
		totalTrafficAllocation += variant.TrafficAllocation
	}

	if totalTrafficAllocation != 100 {
		return errors.New("total traffic allocation must be 100")
	}

	return nil
}

type CreateExperimentVariantRequest struct {
	Name              string                                    `json:"name" binding:"required" validate:"required"`
	Description       string                                    `json:"description" binding:"required" validate:"required"`
	TrafficAllocation int                                       `json:"trafficAllocation" binding:"required" validate:"required,min=1,max=100"`
	Parameters        []CreateExperimentVariantParameterRequest `json:"parameters" binding:"required" validate:"required"`
}

type CreateExperimentVariantParameterRequest struct {
	ParameterDataType string `json:"parameterDataType" binding:"required" validate:"required,oneof=string number boolean"`
	ParameterID       int    `json:"parameterId" binding:"required" validate:"required"`
	ParameterName     string `json:"parameterName" binding:"required" validate:"required"`
	RolloutValue      string `json:"rolloutValue" binding:"required" validate:"required"`
}

// ExperimentResponse represents the response for experiment operations (without variants)
type ExperimentResponse struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Uuid            string `json:"uuid"`
	Hypothesis      string `json:"hypothesis"`
	Description     string `json:"description"`
	StartDate       int64  `json:"startDate"`
	EndDate         int64  `json:"endDate"`
	HashAttributeID int    `json:"hashAttributeId"`
	PopulationSize  int    `json:"populationSize"`
	Strategy        string `json:"strategy"`
	CreatedAt       int64  `json:"createdAt"`
	UpdatedAt       int64  `json:"updatedAt"`
	Status          string `json:"status"`
}

// HashAttributeResponse represents the hash attribute in experiment responses
type HashAttributeResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ExperimentVariantResponse represents the response for experiment variant
type ExperimentVariantResponse struct {
	ID                int                                  `json:"id"`
	Name              string                               `json:"name"`
	Description       string                               `json:"description"`
	TrafficAllocation int                                  `json:"trafficAllocation"`
	CreatedAt         int64                                `json:"createdAt"`
	UpdatedAt         int64                                `json:"updatedAt"`
	Parameters        []ExperimentVariantParameterResponse `json:"parameters"`
}

// ExperimentVariantParameterResponse represents the response for experiment variant parameter
type ExperimentVariantParameterResponse struct {
	ID                int    `json:"id"`
	ParameterDataType string `json:"parameterDataType"`
	ParameterID       int    `json:"parameterId"`
	ParameterName     string `json:"parameterName"`
	RolloutValue      string `json:"rolloutValue"`
	CreatedAt         int64  `json:"createdAt"`
	UpdatedAt         int64  `json:"updatedAt"`
}

// ExperimentDetailResponse represents the detailed response for experiment operations (with variants and parameters)
type ExperimentDetailResponse struct {
	ID              int                         `json:"id"`
	Name            string                      `json:"name"`
	Uuid            string                      `json:"uuid"`
	Hypothesis      string                      `json:"hypothesis"`
	Description     string                      `json:"description"`
	StartDate       int64                       `json:"startDate"`
	EndDate         int64                       `json:"endDate"`
	HashAttributeID int                         `json:"hashAttributeId"`
	HashAttribute   HashAttributeResponse       `json:"hashAttribute"`
	PopulationSize  int                         `json:"populationSize"`
	Strategy        string                      `json:"strategy"`
	CreatedAt       int64                       `json:"createdAt"`
	UpdatedAt       int64                       `json:"updatedAt"`
	Status          string                      `json:"status"`
	Variants        []ExperimentVariantResponse `json:"variants"`
}

// ToExperimentResponse converts a model.Experiment to ExperimentResponse
func ToExperimentResponse(experiment *model.Experiment) ExperimentResponse {
	return ExperimentResponse{
		ID:              experiment.ID,
		Name:            experiment.Name,
		Uuid:            experiment.Uuid,
		Hypothesis:      experiment.Hypothesis,
		Description:     experiment.Description,
		StartDate:       experiment.StartDate,
		EndDate:         experiment.EndDate,
		HashAttributeID: experiment.HashAttributeID,
		PopulationSize:  experiment.PopulationSize,
		Strategy:        experiment.Strategy,
		CreatedAt:       experiment.CreatedAt,
		UpdatedAt:       experiment.UpdatedAt,
		Status:          experiment.Status,
	}
}

// ToExperimentVariantParameterResponse converts a model.ExperimentVariantParameter to ExperimentVariantParameterResponse
func ToExperimentVariantParameterResponse(parameter *model.ExperimentVariantParameter) ExperimentVariantParameterResponse {
	return ExperimentVariantParameterResponse{
		ID:                parameter.ID,
		ParameterDataType: parameter.ParameterDataType,
		ParameterID:       parameter.ParameterID,
		ParameterName:     parameter.ParameterName,
		RolloutValue:      parameter.RolloutValue,
		CreatedAt:         parameter.CreatedAt,
		UpdatedAt:         parameter.UpdatedAt,
	}
}

// ToExperimentVariantResponse converts a model.ExperimentVariant to ExperimentVariantResponse
func ToExperimentVariantResponse(variant *model.ExperimentVariant, parameters []*model.ExperimentVariantParameter) ExperimentVariantResponse {
	parameterResponses := make([]ExperimentVariantParameterResponse, len(parameters))
	for i, param := range parameters {
		parameterResponses[i] = ToExperimentVariantParameterResponse(param)
	}

	return ExperimentVariantResponse{
		ID:                variant.ID,
		Name:              variant.Name,
		Description:       variant.Description,
		TrafficAllocation: variant.TrafficAllocation,
		CreatedAt:         variant.CreatedAt,
		UpdatedAt:         variant.UpdatedAt,
		Parameters:        parameterResponses,
	}
}

// ToExperimentDetailResponse converts a model.Experiment with variants and parameters to ExperimentDetailResponse
func ToExperimentDetailResponse(experiment *model.Experiment, variants []*model.ExperimentVariant, variantParametersMap map[int][]*model.ExperimentVariantParameter, hashAttribute *model.Attribute) ExperimentDetailResponse {
	variantResponses := make([]ExperimentVariantResponse, len(variants))
	for i, variant := range variants {
		parameters := variantParametersMap[variant.ID]
		variantResponses[i] = ToExperimentVariantResponse(variant, parameters)
	}

	hashAttrResponse := HashAttributeResponse{
		ID:   experiment.HashAttributeID,
		Name: "",
	}
	if hashAttribute != nil {
		hashAttrResponse.Name = hashAttribute.Name
	}

	return ExperimentDetailResponse{
		ID:              experiment.ID,
		Name:            experiment.Name,
		Uuid:            experiment.Uuid,
		Hypothesis:      experiment.Hypothesis,
		Description:     experiment.Description,
		StartDate:       experiment.StartDate,
		EndDate:         experiment.EndDate,
		HashAttributeID: experiment.HashAttributeID,
		HashAttribute:   hashAttrResponse,
		PopulationSize:  experiment.PopulationSize,
		Strategy:        experiment.Strategy,
		CreatedAt:       experiment.CreatedAt,
		UpdatedAt:       experiment.UpdatedAt,
		Status:          experiment.Status,
		Variants:        variantResponses,
	}
}

// RejectExperimentRequest represents the request to reject an experiment
type RejectExperimentRequest struct {
	Reason string `json:"reason,omitempty"` // Optional reason for rejection
}

// ApproveExperimentRequest represents the request to approve an experiment
type ApproveExperimentRequest struct {
	Notes string `json:"notes,omitempty"` // Optional notes for approval
}

// AbortExperimentRequest represents the request to abort an experiment
type AbortExperimentRequest struct {
	Reason string `json:"reason,omitempty"` // Optional reason for aborting
}
