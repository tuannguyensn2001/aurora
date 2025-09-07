package dto

import (
	"errors"

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
