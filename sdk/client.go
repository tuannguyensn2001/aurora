// Package sdk provides the Aurora A/B Testing SDK for Go applications.
//
// This SDK allows applications to:
// - Evaluate feature flags and parameters based on user attributes
// - Fetch experiment configurations from Aurora backend or S3
// - Cache configurations locally for performance
// - Handle errors gracefully with proper error types
//
// Basic usage:
//
//	clientOptions := sdk.ClientOptions{
//		EndpointURL: "https://your-aurora-instance.com",
//		S3BucketName: "your-s3-bucket", // optional
//	}
//
//	client, err := sdk.NewClient(clientOptions,
//		sdk.WithRefreshRate(30*time.Second),
//		sdk.WithLogLevel(slog.LevelInfo),
//		sdk.WithInMemoryOnly(true),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = client.Start(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Stop()
//
//	// Create user attributes
//	attrs := sdk.NewAttribute().
//		SetString("user_id", "12345").
//		SetString("country", "US").
//		SetNumber("age", 25)
//
//	// Evaluate a parameter
//	result := client.EvaluateParameter(context.Background(), "welcome_message", attrs)
//	if result.HasError() {
//		log.Printf("Error: %v", result.Error())
//	} else {
//		message := result.AsString("Hello!")
//		fmt.Println(message)
//	}
package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dgraph-io/badger/v4"
	"resty.dev/v3"
)

const (
	defaultRefreshRate = 1 * time.Minute
	defaultLogLevel    = slog.LevelDebug
	defaultPath        = "/sdk-dump"
)

// Client interface defines the main SDK operations
type Client interface {
	Start(ctx context.Context) error
	Stop()
	EvaluateParameter(ctx context.Context, parameterName string, attribute *Attribute) RolloutValue
	GetMetadata(ctx context.Context) (*MetadataResponse, error)
}

// AuroraClient implements the Client interface
type AuroraClient struct {
	s3BucketName string
	refreshRate  time.Duration
	logLevel     slog.Level
	logger       *slog.Logger
	quit         chan struct{}
	s3Client     *s3.Client
	inMemoryOnly bool
	path         string
	engine       *engine
	endpointUrl  string
	enableS3     bool
	serviceName  string
	eventTracker *EventTracker
	onEvaluate   func(source string, parameterName string, attribute *Attribute, rolloutValueRaw *string, err error)
	storage      storage
}

// ClientOptions holds the required configuration options for the client
type ClientOptions struct {
	S3BucketName string
	EndpointURL  string
	ServiceName  string
}

// Option represents a functional option for configuring the client
type Option func(*AuroraClient)

// WithRefreshRate sets the refresh rate for parameter updates
func WithRefreshRate(refreshRate time.Duration) Option {
	return func(c *AuroraClient) {
		c.refreshRate = refreshRate
	}
}

// WithLogLevel sets the logging level
func WithLogLevel(logLevel slog.Level) Option {
	return func(c *AuroraClient) {
		c.logLevel = logLevel
	}
}

// WithS3Client sets a custom S3 client
func WithS3Client(s3Client *s3.Client) Option {
	return func(c *AuroraClient) {
		c.s3Client = s3Client
	}
}

// WithInMemoryOnly configures the client to use in-memory storage only
func WithInMemoryOnly(inMemoryOnly bool) Option {
	return func(c *AuroraClient) {
		c.inMemoryOnly = inMemoryOnly
	}
}

// WithPath sets the storage path for the BadgerDB
func WithPath(path string) Option {
	return func(c *AuroraClient) {
		c.path = path
	}
}

// WithEnableS3 enables or disables S3 usage
func WithEnableS3(enableS3 bool) Option {
	return func(c *AuroraClient) {
		c.enableS3 = enableS3
	}
}

func WithOnEvaluate(onEvaluate func(source string, parameterName string, attribute *Attribute, rolloutValueRaw *string, err error)) Option {
	return func(c *AuroraClient) {
		c.onEvaluate = onEvaluate
	}
}

func (c *AuroraClient) applyDefaults() {
	// Set defaults
	c.refreshRate = defaultRefreshRate
	c.logLevel = defaultLogLevel
	c.inMemoryOnly = false
	c.path = defaultPath
	c.enableS3 = true
	c.quit = make(chan struct{})

	// Set up AWS S3 client
	if c.enableS3 {
		cfg, err := awsconfig.LoadDefaultConfig(context.Background())
		if err != nil {
			// Log warning but don't fail - fallback to upstream API
			c.enableS3 = false
		} else {
			c.s3Client = s3.NewFromConfig(cfg)
		}
	}
}

// NewClient creates a new Aurora SDK client with the given options
func NewClient(clientOptions ClientOptions, options ...Option) (*AuroraClient, error) {
	// Validate required fields
	if clientOptions.EndpointURL == "" {
		return nil, NewConfigurationError("endpoint URL is required", nil)
	}

	c := &AuroraClient{
		s3BucketName: clientOptions.S3BucketName,
		endpointUrl:  clientOptions.EndpointURL,
		serviceName:  clientOptions.ServiceName,
	}

	// Apply defaults
	c.applyDefaults()

	// Apply options
	for _, option := range options {
		option(c)
	}

	// Initialize logger
	opts := &slog.HandlerOptions{
		Level: c.logLevel,
	}
	c.logger = slog.New(slog.NewTextHandler(os.Stdout, opts))

	// Initialize BadgerDB
	opt := badger.DefaultOptions(c.path)
	if c.inMemoryOnly {
		opt = opt.WithInMemory(true)
	}
	db, err := badger.Open(opt)
	if err != nil {
		return nil, NewConfigurationError("failed to open storage", err)
	}
	c.storage = newStorage(db, c.logger)
	c.engine = newEngine(c.logger)
	c.eventTracker = NewEventTracker(c.endpointUrl, c.serviceName, c.logger)
	return c, nil
}

// Start initializes and starts the client
func (c *AuroraClient) Start(ctx context.Context) error {
	c.logger.Info("starting")
	if ctx.Err() != nil {
		c.logger.ErrorContext(ctx, "context is done")
		return ctx.Err()
	}
	err := c.persist(ctx)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to persist parameters", "error", err)

	}
	go c.dispatch(ctx)
	return nil
}

// Stop shuts down the client gracefully
func (c *AuroraClient) Stop() {
	c.storage.close(context.Background())
	close(c.quit)
}

func (c *AuroraClient) dispatch(ctx context.Context) {
	c.logger.Info("dispatching")
	ticket := time.NewTicker(c.refreshRate)

	for {
		select {
		case <-ticket.C:
			c.logger.Info("dispatching")
			err := c.persist(ctx)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to get parameters", "error", err)
			}

		case <-ctx.Done():
			ticket.Stop()
			return
		case <-c.quit:
			ticket.Stop()
			return
		}

	}
}

func (c *AuroraClient) persist(ctx context.Context) error {

	experiments, err := c.getExperiments(ctx)
	if err != nil {
		return err
	}
	err = c.storage.persistExperiments(ctx, experiments)
	if err != nil {
		return err
	}

	parameters, err := c.getParameters(ctx)
	if err != nil {
		return err
	}
	err = c.storage.persistParameters(ctx, parameters)
	if err != nil {
		return err
	}
	c.logger.Info("parameters persisted", "count", len(parameters))
	return nil
}

func (c *AuroraClient) getParameters(ctx context.Context) ([]Parameter, error) {

	if !c.enableS3 {
		return c.getParametersFromUpstream(ctx)
	}

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(c.s3BucketName),
		Key:    aws.String("parameters.json"),
	}
	getObjectOutput, err := c.s3Client.GetObject(ctx, getObjectInput)
	if err != nil {
		return nil, NewNetworkError("get parameters from s3", err)
	}

	parameters := []Parameter{}
	err = json.NewDecoder(getObjectOutput.Body).Decode(&parameters)
	if err != nil {
		return nil, NewNetworkError("decode parameters from s3", err)
	}
	return parameters, nil
}

// UpstreamParametersResponse represents the response from the upstream parameters API
type UpstreamParametersResponse struct {
	Parameters []Parameter `json:"parameters"`
}

type UpstreamExperimentsResponse struct {
	Experiments []Experiment `json:"experiments"`
}

func (c *AuroraClient) getParametersFromUpstream(ctx context.Context) ([]Parameter, error) {
	client := resty.New()
	defer client.Close()
	var res UpstreamParametersResponse
	response, err := client.R().
		SetContext(ctx).
		SetResult(&res).
		SetBody(map[string]interface{}{}).
		Post(fmt.Sprintf("%s/api/v1/sdk/parameters", c.endpointUrl))
	c.logger.Debug("parameters from upstream", "response", response)

	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get parameters from upstream", "error", err)
		return nil, NewNetworkError("get parameters from upstream", err)
	}

	castResp := response.Result().(*UpstreamParametersResponse)
	return castResp.Parameters, nil

}

func (c *AuroraClient) getExperimentsFromUpstream(ctx context.Context) ([]Experiment, error) {
	client := resty.New()
	defer client.Close()
	var res UpstreamExperimentsResponse
	response, err := client.R().
		SetContext(ctx).
		SetResult(&res).
		SetBody(map[string]interface{}{}).
		Post(fmt.Sprintf("%s/api/v1/sdk/experiments", c.endpointUrl))
	c.logger.Debug("experiments from upstream", "response", response)

	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get experiments from upstream", "error", err)
		return nil, NewNetworkError("get experiments from upstream", err)
	}

	castResp := response.Result().(*UpstreamExperimentsResponse)
	return castResp.Experiments, nil
}

func (c *AuroraClient) getExperiments(ctx context.Context) ([]Experiment, error) {
	if !c.enableS3 {
		return c.getExperimentsFromUpstream(ctx)
	}
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(c.s3BucketName),
		Key:    aws.String("experiments.json"),
	}
	getObjectOutput, err := c.s3Client.GetObject(ctx, getObjectInput)
	if err != nil {
		return nil, NewNetworkError("get experiments from s3", err)
	}
	experiments := []Experiment{}
	err = json.NewDecoder(getObjectOutput.Body).Decode(&experiments)
	if err != nil {
		return nil, NewNetworkError("decode experiments from s3", err)
	}
	return experiments, nil
}

// EvaluateParameter evaluates a parameter against the given attributes
func (c *AuroraClient) EvaluateParameter(ctx context.Context, parameterName string, attribute *Attribute) RolloutValue {

	experimentResult, resExperiments := c.resolveFromExperiments(ctx, parameterName, attribute)
	if !resExperiments.HasError() {
		if c.onEvaluate != nil {
			c.onEvaluate("experiment", parameterName, attribute, resExperiments.raw(), resExperiments.Error())
		}

		// Track experiment evaluation event
		if c.eventTracker != nil {
			event := c.eventTracker.CreateExperimentEvaluationEvent(
				parameterName,
				attribute,
				resExperiments.raw(),
				resExperiments.Error(),
				experimentResult.ExperimentID,
				experimentResult.ExperimentUUID,
				experimentResult.VariantID,
				experimentResult.VariantName,
			)
			c.eventTracker.TrackEvent(ctx, event)
		}

		return resExperiments
	}

	res := c.resolveFromParameter(ctx, parameterName, attribute)

	if c.onEvaluate != nil {
		c.onEvaluate("parameter", parameterName, attribute, res.raw(), res.Error())
	}

	// Track parameter evaluation event
	if c.eventTracker != nil {
		event := c.eventTracker.CreateParameterEvaluationEvent(
			parameterName,
			attribute,
			res.raw(),
			res.Error(),
		)
		c.eventTracker.TrackEvent(ctx, event)
	}

	return res
}

func (c *AuroraClient) resolveFromExperiments(ctx context.Context, parameterName string, attribute *Attribute) (*ExperimentEvaluationResult, RolloutValue) {
	experiments, err := c.storage.getExperimentsByParameterName(ctx, parameterName)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get experiments by parameter name", "error", err)
		return nil, NewRolloutValueWithError(NewParameterNotFoundError(parameterName))
	}
	if len(experiments) == 0 {
		return nil, NewRolloutValueWithError(NewParameterNotFoundError(parameterName))
	}

	for _, experiment := range experiments {
		result := c.engine.evaluateExperimentDetailed(&experiment, attribute, parameterName)
		if result.Success {
			return result, NewRolloutValue(&result.Value, result.DataType)
		}
	}
	return nil, NewRolloutValueWithError(NewParameterNotFoundError(parameterName))
}

func (c *AuroraClient) resolveFromParameter(ctx context.Context, parameterName string, attribute *Attribute) RolloutValue {

	c.logger.InfoContext(ctx, "resolving parameter", "parameterName", parameterName)
	parameter, err := c.storage.getParameterByName(ctx, parameterName)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get parameter", "error", err)
		return NewRolloutValueWithError(NewParameterNotFoundError(parameterName))
	}

	rolloutValueStr := c.engine.evaluateParameter(&parameter, attribute)
	c.logger.InfoContext(ctx, "resolved parameter", "parameterName", parameterName, "rolloutValue", rolloutValueStr, "dataType", parameter.DataType, "rules count", len(parameter.Rules))
	return NewRolloutValue(&rolloutValueStr, parameter.DataType)

}
