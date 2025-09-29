package sdk

import "strconv"

type rolloutValue struct {
	value    *string
	dataType ParameterDataType
}

func NewRolloutValue(value *string, dataType ParameterDataType) rolloutValue {
	return rolloutValue{
		value:    value,
		dataType: dataType,
	}
}

func (rv rolloutValue) AsString(defaultValue string) string {
	if rv.dataType != ParameterDataTypeString {
		return defaultValue
	}
	if rv.value == nil {
		return defaultValue
	}
	return *rv.value
}

func (rv rolloutValue) AsNumber(defaultValue float64) float64 {
	if rv.dataType != ParameterDataTypeNumber {
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

func (rv rolloutValue) AsInt(defaultValue int) int {
	if rv.dataType != ParameterDataTypeNumber {
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

func (rv rolloutValue) AsBool(defaultValue bool) bool {
	if rv.dataType != ParameterDataTypeBoolean {
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
