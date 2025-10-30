package config

import (
	"sdk/types"
	"time"

	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config holds all configuration for the SDK
type Config struct {
	// Client configuration
	EndpointURL  string
	S3BucketName string
	ServiceName  string

	// Storage configuration
	InMemoryOnly bool
	Path         string

	// S3 configuration
	EnableS3 bool
	S3Client *s3.Client

	// Refresh configuration
	RefreshRate time.Duration

	// Logging configuration
	LogLevel slog.Level
	Logger   Logger

	// Event tracking configuration
	BatchConfig types.BatchConfig

	// Callback configuration
	OnEvaluate func(source string, parameterName string, attribute Attribute, rolloutValueRaw *string, err error)
}

// Logger interface for dependency injection
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	DebugContext(ctx interface{}, msg string, args ...any)
	InfoContext(ctx interface{}, msg string, args ...any)
	WarnContext(ctx interface{}, msg string, args ...any)
	ErrorContext(ctx interface{}, msg string, args ...any)
}

// Attribute interface for dependency injection
type Attribute interface {
	Get(key string) interface{}
	ToMap() map[string]interface{}
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		InMemoryOnly: false,
		Path:         "/sdk-dump",
		EnableS3:     true,
		RefreshRate:  1 * time.Minute,
		LogLevel:     slog.LevelDebug,
		BatchConfig: types.BatchConfig{
			MaxSize:     100,              // 100 events per batch
			MaxBytes:    1048576,          // 1MB per batch
			MaxWaitTime: 30 * time.Second, // 30 seconds max wait
			FlushSize:   10,               // Flush at 10 events
			FlushBytes:  104857,           // Flush at 100KB
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.EndpointURL == "" {
		return NewValidationError("endpoint URL is required", nil)
	}
	if c.ServiceName == "" {
		return NewValidationError("service name is required", nil)
	}
	if c.RefreshRate <= 0 {
		return NewValidationError("refresh rate must be positive", nil)
	}
	if c.BatchConfig.MaxSize <= 0 {
		return NewValidationError("batch max size must be positive", nil)
	}
	if c.BatchConfig.MaxBytes <= 0 {
		return NewValidationError("batch max bytes must be positive", nil)
	}
	if c.BatchConfig.MaxWaitTime <= 0 {
		return NewValidationError("batch max wait time must be positive", nil)
	}
	return nil
}

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) error {
	// This will be replaced with the proper error type from pkg/errors
	return &validationError{message: message, cause: cause}
}

type validationError struct {
	message string
	cause   error
}

func (e *validationError) Error() string {
	if e.cause != nil {
		return e.message + " (caused by: " + e.cause.Error() + ")"
	}
	return e.message
}

func (e *validationError) Unwrap() error {
	return e.cause
}
