package sdk

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
