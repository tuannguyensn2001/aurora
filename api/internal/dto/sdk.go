package dto

type GetMetadataSDKRequest struct {
}

type GetMetadataSDKResponse struct {
	EnableS3 bool `json:"enableS3"`
}
