package dto

import "sdk/types"

type GetMetadataSDKRequest struct {
}

type GetMetadataSDKResponse struct {
	EnableS3 bool `json:"enableS3"`
}

type GetAllParametersSDKRequest struct {
}

type GetAllParametersSDKResponse struct {
	Parameters []types.Parameter `json:"parameters"`
}

type GetAllExperimentsSDKRequest struct {
}

type GetAllExperimentsSDKResponse struct {
	Experiments []types.Experiment `json:"experiments"`
}
