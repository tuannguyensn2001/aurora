package model

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
	Segment         *Segment            `json:"segment,omitempty"`
	HashAttribute   *Attribute          `json:"hashAttribute,omitempty"`
	Variants        []ExperimentVariant `json:"variants"`
}

func (e *Experiment) TableName() string {
	return "experiments"
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
