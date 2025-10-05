package sdk

import "fmt"

// SDKError represents different types of errors that can occur in the SDK
type SDKError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// ErrorType represents the category of error
type ErrorType string

const (
	// ErrorTypeParameterNotFound indicates a parameter was not found
	ErrorTypeParameterNotFound ErrorType = "parameter_not_found"
	// ErrorTypeInvalidAttribute indicates an attribute is invalid
	ErrorTypeInvalidAttribute ErrorType = "invalid_attribute"
	// ErrorTypeEvaluationFailed indicates parameter evaluation failed
	ErrorTypeEvaluationFailed ErrorType = "evaluation_failed"
	// ErrorTypeStorageError indicates a storage-related error
	ErrorTypeStorageError ErrorType = "storage_error"
	// ErrorTypeNetworkError indicates a network-related error
	ErrorTypeNetworkError ErrorType = "network_error"
	// ErrorTypeConfigurationError indicates a configuration error
	ErrorTypeConfigurationError ErrorType = "configuration_error"
)

// Error implements the error interface
func (e SDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause of the error
func (e SDKError) Unwrap() error {
	return e.Cause
}

// IsType checks if the error is of a specific type
func (e SDKError) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// NewSDKError creates a new SDK error
func NewSDKError(errorType ErrorType, message string, cause error) *SDKError {
	return &SDKError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// NewParameterNotFoundError creates a parameter not found error
func NewParameterNotFoundError(parameterName string) *SDKError {
	return NewSDKError(
		ErrorTypeParameterNotFound,
		fmt.Sprintf("parameter '%s' not found", parameterName),
		nil,
	)
}

// NewInvalidAttributeError creates an invalid attribute error
func NewInvalidAttributeError(attributeName string, reason string) *SDKError {
	return NewSDKError(
		ErrorTypeInvalidAttribute,
		fmt.Sprintf("invalid attribute '%s': %s", attributeName, reason),
		nil,
	)
}

// NewEvaluationFailedError creates an evaluation failed error
func NewEvaluationFailedError(parameterName string, cause error) *SDKError {
	return NewSDKError(
		ErrorTypeEvaluationFailed,
		fmt.Sprintf("failed to evaluate parameter '%s'", parameterName),
		cause,
	)
}

// NewStorageError creates a storage error
func NewStorageError(operation string, cause error) *SDKError {
	return NewSDKError(
		ErrorTypeStorageError,
		fmt.Sprintf("storage operation '%s' failed", operation),
		cause,
	)
}

// NewNetworkError creates a network error
func NewNetworkError(operation string, cause error) *SDKError {
	return NewSDKError(
		ErrorTypeNetworkError,
		fmt.Sprintf("network operation '%s' failed", operation),
		cause,
	)
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(message string, cause error) *SDKError {
	return NewSDKError(
		ErrorTypeConfigurationError,
		message,
		cause,
	)
}
