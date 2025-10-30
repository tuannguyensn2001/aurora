package client

import (
	"context"
	"fmt"
	"sdk/pkg/errors"
	"sdk/pkg/logger"
	"sdk/types"

	"resty.dev/v3"
)

// HTTPDataFetcher implements DataFetcher using HTTP
type HTTPDataFetcher struct {
	endpointURL string
	logger      logger.Logger
}

// NewHTTPDataFetcher creates a new HTTP data fetcher
func NewHTTPDataFetcher(endpointURL string, logger logger.Logger) DataFetcher {
	return &HTTPDataFetcher{
		endpointURL: endpointURL,
		logger:      logger,
	}
}

// GetParameters fetches parameters from the upstream service
func (f *HTTPDataFetcher) GetParameters(ctx context.Context) ([]types.Parameter, error) {
	client := resty.New()
	defer client.Close()

	var res types.UpstreamParametersResponse
	response, err := client.R().
		SetContext(ctx).
		SetResult(&res).
		SetBody(map[string]interface{}{}).
		Post(fmt.Sprintf("%s/api/v1/sdk/parameters", f.endpointURL))

	f.logger.Debug("parameters from upstream", "response", response)

	if err != nil {
		f.logger.ErrorContext(ctx, "failed to get parameters from upstream", "error", err)
		return nil, errors.NewNetworkError("get parameters from upstream", err)
	}

	castResp := response.Result().(*types.UpstreamParametersResponse)
	return castResp.Parameters, nil
}

// GetExperiments fetches experiments from the upstream service
func (f *HTTPDataFetcher) GetExperiments(ctx context.Context) ([]types.Experiment, error) {
	client := resty.New()
	defer client.Close()

	var res types.UpstreamExperimentsResponse
	response, err := client.R().
		SetContext(ctx).
		SetResult(&res).
		SetBody(map[string]interface{}{}).
		Post(fmt.Sprintf("%s/api/v1/sdk/experiments", f.endpointURL))

	f.logger.Debug("experiments from upstream", "response", response)

	if err != nil {
		f.logger.ErrorContext(ctx, "failed to get experiments from upstream", "error", err)
		return nil, errors.NewNetworkError("get experiments from upstream", err)
	}

	castResp := response.Result().(*types.UpstreamExperimentsResponse)
	return castResp.Experiments, nil
}

// S3DataFetcher implements DataFetcher using S3
type S3DataFetcher struct {
	s3Client    S3Client
	bucketName  string
	logger      logger.Logger
	httpFetcher DataFetcher // Fallback to HTTP
}

// S3Client interface for dependency injection
type S3Client interface {
	GetObject(ctx context.Context, input interface{}) (interface{}, error)
}

// NewS3DataFetcher creates a new S3 data fetcher
func NewS3DataFetcher(s3Client S3Client, bucketName string, logger logger.Logger, httpFetcher DataFetcher) DataFetcher {
	return &S3DataFetcher{
		s3Client:    s3Client,
		bucketName:  bucketName,
		logger:      logger,
		httpFetcher: httpFetcher,
	}
}

// GetParameters fetches parameters from S3 with HTTP fallback
func (f *S3DataFetcher) GetParameters(ctx context.Context) ([]types.Parameter, error) {
	// Try S3 first
	parameters, err := f.getParametersFromS3(ctx)
	if err != nil {
		f.logger.WarnContext(ctx, "failed to get parameters from S3, falling back to HTTP", "error", err)
		return f.httpFetcher.GetParameters(ctx)
	}
	return parameters, nil
}

// GetExperiments fetches experiments from S3 with HTTP fallback
func (f *S3DataFetcher) GetExperiments(ctx context.Context) ([]types.Experiment, error) {
	// Try S3 first
	experiments, err := f.getExperimentsFromS3(ctx)
	if err != nil {
		f.logger.WarnContext(ctx, "failed to get experiments from S3, falling back to HTTP", "error", err)
		return f.httpFetcher.GetExperiments(ctx)
	}
	return experiments, nil
}

// getParametersFromS3 fetches parameters from S3
func (f *S3DataFetcher) getParametersFromS3(ctx context.Context) ([]types.Parameter, error) {
	// This would be implemented with actual S3 client calls
	// For now, return an error to trigger fallback
	return nil, errors.NewNetworkError("S3 not implemented", fmt.Errorf("S3 data fetcher not implemented"))
}

// getExperimentsFromS3 fetches experiments from S3
func (f *S3DataFetcher) getExperimentsFromS3(ctx context.Context) ([]types.Experiment, error) {
	// This would be implemented with actual S3 client calls
	// For now, return an error to trigger fallback
	return nil, errors.NewNetworkError("S3 not implemented", fmt.Errorf("S3 data fetcher not implemented"))
}
