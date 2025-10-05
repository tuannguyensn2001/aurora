package sdk

import (
	"context"
	"fmt"

	"resty.dev/v3"
)

type MetadataResponse struct {
	EnableS3 bool `json:"enableS3"`
}

func (c *client) GetMetadata(ctx context.Context) (*MetadataResponse, error) {
	client := resty.New()
	defer client.Close()
	response, err := client.R().
		SetContext(ctx).
		SetResult(&MetadataResponse{}).
		Post(fmt.Sprintf("%s/api/v1/sdk/metadata", c.endpointUrl))
	if err != nil {
		return nil, err
	}
	return response.Result().(*MetadataResponse), nil
}
