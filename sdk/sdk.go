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
	"fmt"
	"log/slog"
	"sdk/internal/client"
	"sdk/internal/config"
	"sdk/internal/engine"
	"sdk/internal/events"
	"sdk/internal/storage"
	"sdk/pkg/errors"
	"sdk/pkg/logger"
	"sdk/types"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dgraph-io/badger/v4"
)

// Client interface defines the main SDK operations
type Client interface {
	Start(ctx context.Context) error
	Stop()
	EvaluateParameter(ctx context.Context, parameterName string, attribute *Attribute) RolloutValue
	GetMetadata(ctx context.Context) (*types.MetadataResponse, error)
}

// Attribute represents a collection of key-value pairs used for evaluation
type Attribute struct {
	m map[string]interface{}
}

// NewAttribute creates a new Attribute instance
func NewAttribute() *Attribute {
	return &Attribute{
		m: make(map[string]interface{}),
	}
}

// SetString sets a string value for the given key
func (a *Attribute) SetString(key string, value string) *Attribute {
	a.m[key] = value
	return a
}

// SetBool sets a boolean value for the given key
func (a *Attribute) SetBool(key string, value bool) *Attribute {
	a.m[key] = value
	return a
}

// SetNumber sets a numeric value for the given key
func (a *Attribute) SetNumber(key string, value float64) *Attribute {
	a.m[key] = value
	return a
}

// Get retrieves the value for the given key
func (a *Attribute) Get(key string) interface{} {
	return a.m[key]
}

// Delete removes the key-value pair from the attribute
func (a *Attribute) Delete(key string) {
	delete(a.m, key)
}

// Clear removes all key-value pairs from the attribute
func (a *Attribute) Clear() {
	a.m = make(map[string]interface{})
}

// Keys returns all keys in the attribute
func (a *Attribute) Keys() []string {
	keys := make([]string, 0, len(a.m))
	for key := range a.m {
		keys = append(keys, key)
	}
	return keys
}

// Values returns all values in the attribute
func (a *Attribute) Values() []interface{} {
	values := make([]interface{}, 0, len(a.m))
	for _, value := range a.m {
		values = append(values, value)
	}
	return values
}

// Len returns the number of key-value pairs in the attribute
func (a *Attribute) Len() int {
	return len(a.m)
}

// ToMap returns a copy of the internal map
func (a *Attribute) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range a.m {
		result[k] = v
	}
	return result
}

// RolloutValue represents a value with its associated data type
type RolloutValue struct {
	value    *string
	dataType types.ParameterDataType
	err      error
}

// NewRolloutValue creates a new RolloutValue instance
func NewRolloutValue(value *string, dataType types.ParameterDataType) RolloutValue {
	return RolloutValue{
		value:    value,
		dataType: dataType,
		err:      nil,
	}
}

// NewRolloutValueWithError creates a new RolloutValue with an error
func NewRolloutValueWithError(err error) RolloutValue {
	return RolloutValue{
		value:    nil,
		dataType: "",
		err:      err,
	}
}

// HasError returns true if the RolloutValue contains an error
func (rv RolloutValue) HasError() bool {
	return rv.err != nil
}

// Error returns the error if present
func (rv RolloutValue) Error() error {
	return rv.err
}

// AsString returns the value as a string, or defaultValue if conversion fails or there's an error
func (rv RolloutValue) AsString(defaultValue string) string {
	if rv.HasError() || rv.dataType != types.ParameterDataTypeString {
		return defaultValue
	}
	if rv.value == nil {
		return defaultValue
	}
	return *rv.value
}

// AsNumber returns the value as a float64, or defaultValue if conversion fails or there's an error
func (rv RolloutValue) AsNumber(defaultValue float64) float64 {
	if rv.HasError() || rv.dataType != types.ParameterDataTypeNumber {
		return defaultValue
	}
	if rv.value == nil {
		return defaultValue
	}
	value, err := strconv.ParseFloat(*rv.value, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// AsInt returns the value as an int, or defaultValue if conversion fails or there's an error
func (rv RolloutValue) AsInt(defaultValue int) int {
	if rv.HasError() || rv.dataType != types.ParameterDataTypeNumber {
		return defaultValue
	}
	if rv.value == nil {
		return defaultValue
	}
	value, err := strconv.ParseInt(*rv.value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(value)
}

// AsBool returns the value as a bool, or defaultValue if conversion fails or there's an error
func (rv RolloutValue) AsBool(defaultValue bool) bool {
	if rv.HasError() || rv.dataType != types.ParameterDataTypeBoolean {
		return defaultValue
	}
	if rv.value == nil {
		return defaultValue
	}
	value, err := strconv.ParseBool(*rv.value)
	if err != nil {
		return defaultValue
	}
	return value
}

func (rv RolloutValue) raw() *string {
	return rv.value
}

// ClientOptions holds the required configuration options for the client
type ClientOptions struct {
	S3BucketName string
	EndpointURL  string
	ServiceName  string
}

// Option represents a functional option for configuring the client
type Option func(*config.Config)

// WithRefreshRate sets the refresh rate for parameter updates
func WithRefreshRate(refreshRate time.Duration) Option {
	return func(c *config.Config) {
		c.RefreshRate = refreshRate
	}
}

// WithLogLevel sets the logging level
func WithLogLevel(logLevel slog.Level) Option {
	return func(c *config.Config) {
		c.LogLevel = logLevel
	}
}

// WithS3Client sets a custom S3 client
func WithS3Client(s3Client *s3.Client) Option {
	return func(c *config.Config) {
		c.S3Client = s3Client
	}
}

// WithInMemoryOnly configures the client to use in-memory storage only
func WithInMemoryOnly(inMemoryOnly bool) Option {
	return func(c *config.Config) {
		c.InMemoryOnly = inMemoryOnly
	}
}

// WithPath sets the storage path for the BadgerDB
func WithPath(path string) Option {
	return func(c *config.Config) {
		c.Path = path
	}
}

// WithEnableS3 enables or disables S3 usage
func WithEnableS3(enableS3 bool) Option {
	return func(c *config.Config) {
		c.EnableS3 = enableS3
	}
}

func WithOnEvaluate(onEvaluate func(source string, parameterName string, attribute *Attribute, rolloutValueRaw *string, err error)) Option {
	return func(c *config.Config) {
		// Convert the public Attribute to the internal interface
		c.OnEvaluate = func(source string, parameterName string, attr config.Attribute, rolloutValueRaw *string, err error) {
			if attrAdapter, ok := attr.(*attributeAdapter); ok {
				onEvaluate(source, parameterName, attrAdapter.attribute, rolloutValueRaw, err)
			}
		}
	}
}

// WithBatchMaxSize sets the maximum number of events per batch
func WithBatchMaxSize(maxSize int) Option {
	return func(c *config.Config) {
		c.BatchConfig.MaxSize = maxSize
	}
}

// WithBatchMaxBytes sets the maximum bytes per batch
func WithBatchMaxBytes(maxBytes int) Option {
	return func(c *config.Config) {
		c.BatchConfig.MaxBytes = maxBytes
	}
}

// WithBatchMaxWaitTime sets the maximum wait time before flushing batch
func WithBatchMaxWaitTime(maxWaitTime time.Duration) Option {
	return func(c *config.Config) {
		c.BatchConfig.MaxWaitTime = maxWaitTime
	}
}

// WithBatchFlushSize sets the size at which to flush batch immediately
func WithBatchFlushSize(flushSize int) Option {
	return func(c *config.Config) {
		c.BatchConfig.FlushSize = flushSize
	}
}

// WithBatchFlushBytes sets the bytes at which to flush batch immediately
func WithBatchFlushBytes(flushBytes int) Option {
	return func(c *config.Config) {
		c.BatchConfig.FlushBytes = flushBytes
	}
}

// NewClient creates a new Aurora SDK client with the given options
func NewClient(clientOptions ClientOptions, options ...Option) (Client, error) {
	// Validate required fields
	if clientOptions.EndpointURL == "" {
		return nil, errors.NewConfigurationError("endpoint URL is required", nil)
	}

	// Create configuration
	cfg := config.DefaultConfig()
	cfg.EndpointURL = clientOptions.EndpointURL
	cfg.S3BucketName = clientOptions.S3BucketName
	cfg.ServiceName = clientOptions.ServiceName

	// Apply options
	for _, option := range options {
		option(cfg)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Initialize logger
	if cfg.Logger == nil {
		cfg.Logger = logger.NewDefaultLogger(cfg.LogLevel)
	}

	// Initialize BadgerDB
	opt := badger.DefaultOptions(cfg.Path)
	if cfg.InMemoryOnly {
		opt = opt.WithInMemory(true)
	}
	db, err := badger.Open(opt)
	if err != nil {
		return nil, errors.NewConfigurationError("failed to open storage", err)
	}

	// Initialize components
	storage := storage.NewBadgerStorage(db, cfg.Logger)
	engineImpl := engine.NewEvaluationEngine(cfg.Logger)

	// Create engine adapter
	engineAdapter := &engineAdapter{engine: engineImpl}

	// Initialize data fetcher
	var dataFetcher client.DataFetcher
	if cfg.EnableS3 && cfg.S3Client != nil {
		// Create S3 adapter
		s3Adapter := &s3ClientAdapter{client: cfg.S3Client}
		dataFetcher = client.NewS3DataFetcher(s3Adapter, cfg.S3BucketName, cfg.Logger, client.NewHTTPDataFetcher(cfg.EndpointURL, cfg.Logger))
	} else {
		dataFetcher = client.NewHTTPDataFetcher(cfg.EndpointURL, cfg.Logger)
	}

	// Initialize event tracker
	eventSender := events.NewHTTPEventSender(cfg.EndpointURL, cfg.Logger)
	eventTrackerImpl := events.NewBatchEventTracker(cfg.EndpointURL, cfg.ServiceName, cfg.Logger, cfg.BatchConfig, eventSender)

	// Create event tracker adapter
	eventTrackerAdapter := &eventTrackerAdapter{tracker: eventTrackerImpl}

	// Create client
	auroraClient := client.NewAuroraClient(cfg, storage, engineAdapter, eventTrackerAdapter, dataFetcher)

	// Wrap with adapter to match public interface
	return &clientAdapter{
		client: auroraClient,
	}, nil
}

// clientAdapter adapts the internal client to the public interface
type clientAdapter struct {
	client client.Client
}

func (a *clientAdapter) Start(ctx context.Context) error {
	return a.client.Start(ctx)
}

func (a *clientAdapter) Stop() {
	a.client.Stop()
}

func (a *clientAdapter) EvaluateParameter(ctx context.Context, parameterName string, attribute *Attribute) RolloutValue {
	// Convert public Attribute to internal interface
	internalAttr := &attributeAdapter{attribute: attribute}
	result := a.client.EvaluateParameter(ctx, parameterName, internalAttr)

	// Convert internal RolloutValue to public type
	if impl, ok := result.(*client.RolloutValueImpl); ok {
		return RolloutValue{
			value:    impl.Raw(),
			dataType: impl.DataType,
			err:      impl.Error(),
		}
	}

	// Fallback for other implementations
	return RolloutValue{
		value:    result.Raw(),
		dataType: types.ParameterDataTypeString, // Default type
		err:      result.Error(),
	}
}

func (a *clientAdapter) GetMetadata(ctx context.Context) (*types.MetadataResponse, error) {
	return a.client.GetMetadata(ctx)
}

// attributeAdapter adapts public Attribute to internal interface
type attributeAdapter struct {
	attribute *Attribute
}

func (a *attributeAdapter) Get(key string) interface{} {
	return a.attribute.Get(key)
}

func (a *attributeAdapter) ToMap() map[string]interface{} {
	return a.attribute.ToMap()
}

// s3ClientAdapter adapts AWS S3 client to internal interface
type s3ClientAdapter struct {
	client *s3.Client
}

func (a *s3ClientAdapter) GetObject(ctx context.Context, input interface{}) (interface{}, error) {
	// This would need to be implemented with actual S3 calls
	// For now, return an error to trigger fallback
	return nil, errors.NewNetworkError("S3 not implemented", fmt.Errorf("S3 client adapter not implemented"))
}

// engineAdapter adapts engine.Engine to client.Engine
type engineAdapter struct {
	engine engine.Engine
}

func (a *engineAdapter) EvaluateParameter(parameter *types.Parameter, attribute client.Attribute) string {
	// Convert client.Attribute to engine.Attribute
	attrAdapter := &attributeToEngineAdapter{attr: attribute}
	return a.engine.EvaluateParameter(parameter, attrAdapter)
}

func (a *engineAdapter) EvaluateExperiment(experiment *types.Experiment, attribute client.Attribute, parameterName string) (string, types.ParameterDataType, bool) {
	// Convert client.Attribute to engine.Attribute
	attrAdapter := &attributeToEngineAdapter{attr: attribute}
	return a.engine.EvaluateExperiment(experiment, attrAdapter, parameterName)
}

func (a *engineAdapter) EvaluateExperimentDetailed(experiment *types.Experiment, attribute client.Attribute, parameterName string) *types.ExperimentEvaluationResult {
	// Convert client.Attribute to engine.Attribute
	attrAdapter := &attributeToEngineAdapter{attr: attribute}
	return a.engine.EvaluateExperimentDetailed(experiment, attrAdapter, parameterName)
}

// eventTrackerAdapter adapts events.EventTracker to client.EventTracker
type eventTrackerAdapter struct {
	tracker events.EventTracker
}

func (a *eventTrackerAdapter) TrackEvent(ctx context.Context, event types.EvaluationEvent) {
	a.tracker.TrackEvent(ctx, event)
}

func (a *eventTrackerAdapter) CreateParameterEvaluationEvent(parameterName string, attribute client.Attribute, rolloutValue *string, err error) types.EvaluationEvent {
	// Convert client.Attribute to events.Attribute
	attrAdapter := &attributeToEventsAdapter{attr: attribute}
	return a.tracker.CreateParameterEvaluationEvent(parameterName, attrAdapter, rolloutValue, err)
}

func (a *eventTrackerAdapter) CreateExperimentEvaluationEvent(parameterName string, attribute client.Attribute, rolloutValue *string, err error, experimentID *int, experimentUUID *string, variantID *int, variantName *string) types.EvaluationEvent {
	// Convert client.Attribute to events.Attribute
	attrAdapter := &attributeToEventsAdapter{attr: attribute}
	return a.tracker.CreateExperimentEvaluationEvent(parameterName, attrAdapter, rolloutValue, err, experimentID, experimentUUID, variantID, variantName)
}

func (a *eventTrackerAdapter) Stop(ctx context.Context) {
	a.tracker.Stop(ctx)
}

// attributeToEngineAdapter adapts client.Attribute to engine.Attribute
type attributeToEngineAdapter struct {
	attr client.Attribute
}

func (a *attributeToEngineAdapter) Get(key string) interface{} {
	return a.attr.Get(key)
}

func (a *attributeToEngineAdapter) ToMap() map[string]interface{} {
	return a.attr.ToMap()
}

// attributeToEventsAdapter adapts client.Attribute to events.Attribute
type attributeToEventsAdapter struct {
	attr client.Attribute
}

func (a *attributeToEventsAdapter) Get(key string) interface{} {
	return a.attr.Get(key)
}

func (a *attributeToEventsAdapter) ToMap() map[string]interface{} {
	return a.attr.ToMap()
}
