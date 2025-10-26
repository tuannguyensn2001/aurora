# Aurora A/B Testing SDK Documentation

## Overview

The Aurora A/B Testing SDK is a Go library that enables applications to evaluate feature flags and parameters based on user attributes. It provides a robust, high-performance solution for A/B testing and feature flag management with local caching, multiple data sources, and comprehensive error handling.

## Table of Contents

1. [Architecture](#architecture)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [Usage Examples](#usage-examples)
6. [API Reference](#api-reference)
7. [Data Types](#data-types)
8. [Error Handling](#error-handling)
9. [Storage](#storage)
10. [Performance](#performance)
11. [Troubleshooting](#troubleshooting)

## Architecture

### Core Components

The SDK consists of several key components working together:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AuroraClient  │    │     Engine      │    │    Storage      │
│                 │    │                 │    │                 │
│ - Configuration │◄──►│ - Evaluation    │◄──►│ - BadgerDB      │
│ - HTTP Client   │    │ - Conditions    │    │ - Persistence   │
│ - S3 Client     │    │ - Experiments   │    │ - Caching      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Attributes    │    │   RolloutValue  │    │   Metadata      │
│                 │    │                 │    │                 │
│ - User Context  │    │ - Type Safety   │    │ - Configuration│
│ - Key-Value     │    │ - Error Handling│    │ - Status        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Data Flow

1. **Initialization**: Client connects to Aurora backend and optionally S3
2. **Configuration Fetch**: Parameters and experiments are fetched and cached locally
3. **Evaluation**: User attributes are evaluated against rules and conditions
4. **Result**: Type-safe rollout values are returned with proper error handling

### Key Features

- **Multi-source Configuration**: Supports both Aurora API and S3 as configuration sources
- **Local Caching**: Uses BadgerDB for high-performance local storage
- **Type Safety**: Strongly typed parameter values with automatic conversion
- **Error Handling**: Comprehensive error types with detailed error information
- **Performance**: Optimized evaluation engine with minimal latency
- **Flexibility**: Supports both in-memory and persistent storage modes

## Installation

### Prerequisites

- Go 1.25.1 or later
- Access to Aurora backend service
- Optional: AWS S3 bucket for configuration storage

### Go Module Installation

```bash
go get sdk
```

### Dependencies

The SDK automatically manages these dependencies:
- `github.com/aws/aws-sdk-go-v2` - AWS SDK for S3 integration
- `github.com/dgraph-io/badger/v4` - High-performance key-value database
- `resty.dev/v3` - HTTP client for API communication
- `github.com/spaolacci/murmur3` - Hashing for consistent user bucketing

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "sdk"
)

func main() {
    // Create client options
    clientOptions := sdk.ClientOptions{
        EndpointURL:   "https://your-aurora-instance.com",
        S3BucketName:  "your-s3-bucket", // optional
    }

    // Initialize client with options
    client, err := sdk.NewClient(clientOptions,
        sdk.WithRefreshRate(30*time.Second),
        sdk.WithLogLevel(slog.LevelInfo),
        sdk.WithInMemoryOnly(false),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Start the client
    err = client.Start(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    // Create user attributes
    attrs := sdk.NewAttribute().
        SetString("user_id", "12345").
        SetString("country", "US").
        SetNumber("age", 25)

    // Evaluate a parameter
    result := client.EvaluateParameter(context.Background(), "welcome_message", attrs)
    if result.HasError() {
        log.Printf("Error: %v", result.Error())
    } else {
        message := result.AsString("Hello!")
        fmt.Println(message)
    }
}
```

## Configuration

### Client Options

#### Required Configuration

```go
type ClientOptions struct {
    EndpointURL   string // Aurora backend URL (required)
    S3BucketName  string // S3 bucket name (optional)
}
```

#### Optional Configuration

The SDK provides several configuration options through functional options:

```go
// Refresh rate for configuration updates
sdk.WithRefreshRate(30*time.Second)

// Logging level
sdk.WithLogLevel(slog.LevelInfo)

// Storage mode (in-memory vs persistent)
sdk.WithInMemoryOnly(true)

// Storage path for persistent mode
sdk.WithPath("/custom/sdk-dump")

// Enable/disable S3 integration
sdk.WithS3Enabled(false)

// Custom evaluation callback
sdk.WithOnEvaluate(func(ctx context.Context, parameterName string, attribute *Attribute, result RolloutValue) {
    // Custom logic here
})
```

### Environment Variables

The SDK respects standard AWS environment variables for S3 configuration:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_REGION`
- `AWS_PROFILE`

### Default Values

```go
const (
    defaultRefreshRate = 1 * time.Minute
    defaultLogLevel    = slog.LevelDebug
    defaultPath        = "/sdk-dump"
)
```

## Usage Examples

### Feature Flag Evaluation

```go
// Boolean feature flag
result := client.EvaluateParameter(ctx, "enable_new_feature", userAttrs)
if result.HasError() {
    // Handle error
    return
}

isEnabled := result.AsBool(false) // Default to false
if isEnabled {
    // Show new feature
}
```

### A/B Testing

```go
// String parameter for A/B testing
result := client.EvaluateParameter(ctx, "button_text", userAttrs)
buttonText := result.AsString("Click Me") // Default text

// Number parameter for configuration
result = client.EvaluateParameter(ctx, "discount_percentage", userAttrs)
discount := result.AsNumber(0.0) // Default discount
```

### Complex User Attributes

```go
// Create comprehensive user attributes
attrs := sdk.NewAttribute().
    SetString("user_id", "user_12345").
    SetString("email", "user@example.com").
    SetString("country", "US").
    SetString("city", "San Francisco").
    SetString("user_type", "premium").
    SetNumber("age", 28).
    SetNumber("account_balance", 1500.50).
    SetBool("is_vip", true).
    SetBool("has_subscription", true)

// Evaluate multiple parameters
welcomeMessage := client.EvaluateParameter(ctx, "welcome_message", attrs)
themeColor := client.EvaluateParameter(ctx, "theme_color", attrs)
maxUploadSize := client.EvaluateParameter(ctx, "max_upload_size", attrs)
```

### Error Handling Patterns

```go
result := client.EvaluateParameter(ctx, "parameter_name", attrs)

// Check for errors first
if result.HasError() {
    switch {
    case result.Error().(*sdk.SDKError).IsType(sdk.ErrorTypeParameterNotFound):
        // Parameter doesn't exist, use default
        value := getDefaultValue()
    case result.Error().(*sdk.SDKError).IsType(sdk.ErrorTypeEvaluationFailed):
        // Evaluation failed, log and use default
        log.Printf("Evaluation failed: %v", result.Error())
        value := getDefaultValue()
    default:
        // Other errors
        log.Printf("Unexpected error: %v", result.Error())
        value := getDefaultValue()
    }
} else {
    // Success - use evaluated value
    value := result.AsString("default")
}
```

## API Reference

### Client Interface

```go
type Client interface {
    Start(ctx context.Context) error
    Stop()
    EvaluateParameter(ctx context.Context, parameterName string, attribute *Attribute) RolloutValue
    GetMetadata(ctx context.Context) (*MetadataResponse, error)
}
```

### Attribute Methods

```go
type Attribute struct {
    // Private fields
}

// Constructor
func NewAttribute() *Attribute

// Setters (chainable)
func (a *Attribute) SetString(key string, value string) *Attribute
func (a *Attribute) SetBool(key string, value bool) *Attribute
func (a *Attribute) SetNumber(key string, value float64) *Attribute

// Getters
func (a *Attribute) Get(key string) interface{}
func (a *Attribute) Keys() []string
func (a *Attribute) Values() []interface{}
func (a *Attribute) Len() int

// Modifiers
func (a *Attribute) Delete(key string)
func (a *Attribute) Clear()
```

### RolloutValue Methods

```go
type RolloutValue struct {
    // Private fields
}

// Error checking
func (rv RolloutValue) HasError() bool
func (rv RolloutValue) Error() error

// Type conversion with defaults
func (rv RolloutValue) AsString(defaultValue string) string
func (rv RolloutValue) AsNumber(defaultValue float64) float64
func (rv RolloutValue) AsInt(defaultValue int) int
func (rv RolloutValue) AsBool(defaultValue bool) bool
```

## Data Types

The SDK supports three parameter data types:
- **String**: Text values for messages, themes, and configurations
- **Number**: Numeric values for counts, percentages, and measurements  
- **Boolean**: True/false values for feature toggles and flags

All parameter values are automatically converted to the appropriate Go type when using the `AsString()`, `AsNumber()`, `AsInt()`, and `AsBool()` methods.

## Error Handling

### Error Types

The SDK provides comprehensive error handling with specific error types:

```go
type ErrorType string

const (
    ErrorTypeParameterNotFound    ErrorType = "parameter_not_found"
    ErrorTypeInvalidAttribute     ErrorType = "invalid_attribute"
    ErrorTypeEvaluationFailed     ErrorType = "evaluation_failed"
    ErrorTypeStorageError        ErrorType = "storage_error"
    ErrorTypeNetworkError         ErrorType = "network_error"
    ErrorTypeConfigurationError  ErrorType = "configuration_error"
)
```

### Error Structure

```go
type SDKError struct {
    Type    ErrorType
    Message string
    Cause   error
}

func (e SDKError) Error() string
func (e SDKError) Unwrap() error
func (e SDKError) IsType(errorType ErrorType) bool
```

### Error Handling Best Practices

1. **Always check for errors first**:
   ```go
   result := client.EvaluateParameter(ctx, "param", attrs)
   if result.HasError() {
       // Handle error
   }
   ```

2. **Use type-specific error handling**:
   ```go
   if err := result.Error(); err != nil {
       if sdkErr, ok := err.(*sdk.SDKError); ok {
           switch sdkErr.Type {
           case sdk.ErrorTypeParameterNotFound:
               // Parameter doesn't exist
           case sdk.ErrorTypeNetworkError:
               // Network issue
           }
       }
   }
   ```

3. **Provide meaningful defaults**:
   ```go
   value := result.AsString("default_value")
   ```

## Storage

### Storage Modes

#### In-Memory Mode
- **Use case**: Testing, development, or when persistence isn't needed
- **Performance**: Fastest access
- **Limitation**: Data lost on restart

```go
client, err := sdk.NewClient(options, sdk.WithInMemoryOnly(true))
```

#### Persistent Mode (Default)
- **Use case**: Production environments
- **Performance**: Fast with persistence
- **Storage**: BadgerDB key-value store

```go
client, err := sdk.NewClient(options, sdk.WithInMemoryOnly(false))
```

### Storage Structure

The SDK uses BadgerDB with the following key structure:

```
parameters:{parameter_name}     -> Parameter JSON
experiments:parameters:{name}   -> []string (experiment names)
{experiment_name}              -> Experiment JSON
```

### Storage Configuration

```go
// Custom storage path
client, err := sdk.NewClient(options, sdk.WithPath("/custom/path"))

// Default path: /sdk-dump
```

## Performance

### Optimization Features

1. **Local Caching**: All configurations cached locally for sub-millisecond access
2. **Efficient Evaluation**: Optimized condition evaluation engine
3. **Minimal Memory Footprint**: Efficient data structures and memory management
4. **Background Refresh**: Non-blocking configuration updates

### Performance Characteristics

- **Evaluation Latency**: < 1ms for cached parameters
- **Memory Usage**: ~10-50MB depending on configuration size
- **Storage Size**: ~1-10MB for typical configurations
- **Refresh Rate**: Configurable (default: 1 minute)

### Best Practices

1. **Reuse Attributes**: Create attribute objects once and reuse them
2. **Batch Evaluations**: Evaluate multiple parameters in sequence
3. **Monitor Performance**: Use logging to track evaluation times
4. **Optimize Refresh Rate**: Balance freshness vs performance

```go
// Good: Reuse attributes
attrs := sdk.NewAttribute().SetString("user_id", userID)
result1 := client.EvaluateParameter(ctx, "param1", attrs)
result2 := client.EvaluateParameter(ctx, "param2", attrs)

// Good: Appropriate refresh rate
client, err := sdk.NewClient(options, 
    sdk.WithRefreshRate(5*time.Minute)) // Less frequent for stable configs
```

## Troubleshooting

### Common Issues

#### 1. Configuration Errors

**Problem**: `configuration_error: endpoint URL is required`

**Solution**: Ensure EndpointURL is provided in ClientOptions:
```go
options := sdk.ClientOptions{
    EndpointURL: "https://your-aurora-instance.com", // Required
}
```

#### 2. Storage Errors

**Problem**: `storage_error: failed to open storage`

**Solution**: Check permissions and disk space:
```bash
# Check disk space
df -h

# Check permissions
ls -la /sdk-dump
```

#### 3. Network Errors

**Problem**: `network_error: connection refused`

**Solution**: Verify network connectivity and endpoint URL:
```bash
# Test connectivity
curl -I https://your-aurora-instance.com/api/v1/sdk/metadata
```

#### 4. Parameter Not Found

**Problem**: `parameter_not_found: parameter 'xyz' not found`

**Solution**: Verify parameter exists in Aurora backend or check parameter name spelling.

#### 5. S3 Configuration Issues

**Problem**: S3 integration not working

**Solution**: Verify AWS credentials and bucket permissions:
```bash
# Test AWS credentials
aws s3 ls s3://your-bucket-name
```

### Debugging

#### Enable Debug Logging

```go
client, err := sdk.NewClient(options, 
    sdk.WithLogLevel(slog.LevelDebug))
```

#### Monitor Evaluation

```go
client, err := sdk.NewClient(options,
    sdk.WithOnEvaluate(func(ctx context.Context, parameterName string, attribute *Attribute, result RolloutValue) {
        log.Printf("Evaluated %s: %v", parameterName, result)
    }))
```

#### Check Metadata

```go
metadata, err := client.GetMetadata(ctx)
if err != nil {
    log.Printf("Metadata error: %v", err)
} else {
    log.Printf("S3 enabled: %v", metadata.EnableS3)
}
```

### Performance Monitoring

```go
// Monitor evaluation performance
start := time.Now()
result := client.EvaluateParameter(ctx, "param", attrs)
duration := time.Since(start)
log.Printf("Evaluation took %v", duration)
```

### Health Checks

```go
// Simple health check
func healthCheck(client sdk.Client) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, err := client.GetMetadata(ctx)
    return err
}
```

## Advanced Usage

### Custom Evaluation Logic

```go
client, err := sdk.NewClient(options,
    sdk.WithOnEvaluate(func(ctx context.Context, parameterName string, attribute *Attribute, result RolloutValue) {
        // Log all evaluations
        log.Printf("Parameter: %s, User: %v, Result: %v", 
            parameterName, attribute.Get("user_id"), result)
        
        // Send metrics
        metrics.IncrementCounter("sdk.evaluation", map[string]string{
            "parameter": parameterName,
            "has_error": strconv.FormatBool(result.HasError()),
        })
    }))
```

### Graceful Shutdown

```go
func main() {
    client, err := sdk.NewClient(options)
    if err != nil {
        log.Fatal(err)
    }
    
    err = client.Start(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    // Handle shutdown signals
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    <-c
    log.Println("Shutting down...")
    
    client.Stop()
    log.Println("Shutdown complete")
}
```

### Context Management

```go
// Use context for timeouts and cancellation
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result := client.EvaluateParameter(ctx, "param", attrs)
if result.HasError() {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Evaluation timed out")
    }
}
```

This documentation provides comprehensive coverage of the Aurora A/B Testing SDK, including architecture, configuration, usage patterns, and troubleshooting guidance. The SDK is designed to be robust, performant, and easy to integrate into Go applications.
