package client

import (
	"sdk/types"
	"strconv"
)

// RolloutValueImpl implements the RolloutValue interface
type RolloutValueImpl struct {
	value    *string
	DataType types.ParameterDataType
	err      error
}

// NewRolloutValue creates a new RolloutValue instance
func NewRolloutValue(value *string, dataType types.ParameterDataType) RolloutValue {
	return &RolloutValueImpl{
		value:    value,
		DataType: dataType,
		err:      nil,
	}
}

// NewRolloutValueWithError creates a new RolloutValue with an error
func NewRolloutValueWithError(err error) RolloutValue {
	return &RolloutValueImpl{
		value:    nil,
		DataType: "",
		err:      err,
	}
}

// HasError returns true if the RolloutValue contains an error
func (rv *RolloutValueImpl) HasError() bool {
	return rv.err != nil
}

// Error returns the error if present
func (rv *RolloutValueImpl) Error() error {
	return rv.err
}

// AsString returns the value as a string, or defaultValue if conversion fails or there's an error
func (rv *RolloutValueImpl) AsString(defaultValue string) string {
	if rv.HasError() || rv.DataType != types.ParameterDataTypeString {
		return defaultValue
	}
	if rv.value == nil {
		return defaultValue
	}
	return *rv.value
}

// AsNumber returns the value as a float64, or defaultValue if conversion fails or there's an error
func (rv *RolloutValueImpl) AsNumber(defaultValue float64) float64 {
	if rv.HasError() || rv.DataType != types.ParameterDataTypeNumber {
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
func (rv *RolloutValueImpl) AsInt(defaultValue int) int {
	if rv.HasError() || rv.DataType != types.ParameterDataTypeNumber {
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
func (rv *RolloutValueImpl) AsBool(defaultValue bool) bool {
	if rv.HasError() || rv.DataType != types.ParameterDataTypeBoolean {
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

// Raw returns the raw string value
func (rv *RolloutValueImpl) Raw() *string {
	return rv.value
}
