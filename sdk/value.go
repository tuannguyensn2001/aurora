package sdk

import "strconv"

// RolloutValue represents a value with its associated data type
type RolloutValue struct {
	value    *string
	dataType ParameterDataType
	err      error
}

// NewRolloutValue creates a new RolloutValue instance
func NewRolloutValue(value *string, dataType ParameterDataType) RolloutValue {
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
	if rv.HasError() || rv.dataType != ParameterDataTypeString {
		return defaultValue
	}
	if rv.value == nil {
		return defaultValue
	}
	return *rv.value
}

// AsNumber returns the value as a float64, or defaultValue if conversion fails or there's an error
func (rv RolloutValue) AsNumber(defaultValue float64) float64 {
	if rv.HasError() || rv.dataType != ParameterDataTypeNumber {
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
	if rv.HasError() || rv.dataType != ParameterDataTypeNumber {
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
	if rv.HasError() || rv.dataType != ParameterDataTypeBoolean {
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
