package model

import (
	"encoding/json"
)

type Experiment struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Uuid            string `json:"uuid"`
	Hypothesis      string `json:"hypothesis"`
	Description     string `json:"description"`
	StartDate       int64  `json:"startDate"`
	EndDate         int64
	HashAttributeID int                 `json:"hashAttributeId"`
	PopulationSize  int                 `json:"populationSize"`
	Strategy        string              `json:"strategy"`
	CreatedAt       int64               `json:"createdAt"`
	UpdatedAt       int64               `json:"updatedAt"`
	Status          string              `json:"status"`
	SegmentID       int                 `json:"segmentId"`
	RawValue        json.RawMessage     `gorm:"type:jsonb;column:raw_value" json:"rawValue,omitempty"`
	Segment         *Segment            `json:"segment,omitempty"`
	HashAttribute   *Attribute          `json:"hashAttribute,omitempty"`
	Variants        []ExperimentVariant `json:"variants"`
}

func (e *Experiment) TableName() string {
	return "experiments"
}

// PopulateRawValue creates a JSON representation of the experiment with all related fields
// This should be called after all related entities are loaded via preloads
func (e *Experiment) PopulateRawValue() error {
	// Create a map with all experiment fields for JSON serialization
	rawData := map[string]interface{}{
		"id":              e.ID,
		"name":            e.Name,
		"uuid":            e.Uuid,
		"hypothesis":      e.Hypothesis,
		"description":     e.Description,
		"startDate":       e.StartDate,
		"endDate":         e.EndDate,
		"hashAttributeId": e.HashAttributeID,
		"populationSize":  e.PopulationSize,
		"strategy":        e.Strategy,
		"createdAt":       e.CreatedAt,
		"updatedAt":       e.UpdatedAt,
		"status":          e.Status,
		"segmentId":       e.SegmentID,
		"segment":         e.Segment,
		"hashAttribute":   e.HashAttribute,
		"variants":        e.Variants,
	}

	// Marshal to JSON
	rawJSON, err := json.Marshal(rawData)
	if err != nil {
		return err
	}

	e.RawValue = json.RawMessage(rawJSON)
	return nil
}

type ExperimentVariant struct {
	ID                int                          `json:"id"`
	ExperimentID      int                          `json:"experimentId"`
	Name              string                       `json:"name"`
	Description       string                       `json:"description"`
	CreatedAt         int64                        `json:"createdAt"`
	UpdatedAt         int64                        `json:"updatedAt"`
	TrafficAllocation int                          `json:"trafficAllocation"`
	Parameters        []ExperimentVariantParameter `json:"parameters"`
}

func (e *ExperimentVariant) TableName() string {
	return "experiment_variants"
}

type ExperimentVariantParameter struct {
	ID                  int    `json:"id"`
	ExperimentVariantID int    `json:"experimentVariantId"`
	ParameterDataType   string `json:"parameterDataType"`
	ParameterID         int    `json:"parameterId"`
	ParameterName       string `json:"parameterName"`
	RolloutValue        string `json:"rolloutValue"`
	CreatedAt           int64  `json:"createdAt"`
	UpdatedAt           int64  `json:"updatedAt"`
	ExperimentID        int    `json:"experimentId"`
}

func (e *ExperimentVariantParameter) TableName() string {
	return "experiment_variant_parameters"
}
