package dto

import "sdk"

type GetMetadataSDKRequest struct {
}

type GetMetadataSDKResponse struct {
	EnableS3 bool `json:"enableS3"`
}

type GetAllParametersSDKRequest struct {
}

type GetAllParametersSDKResponse struct {
	Parameters []sdk.Parameter `json:"parameters"`
}
