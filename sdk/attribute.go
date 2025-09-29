package sdk

type attribute struct {
	m map[string]interface{}
}

func NewAttribute() *attribute {
	return &attribute{
		m: make(map[string]interface{}),
	}
}

func (a *attribute) SetString(key string, value string) *attribute {
	a.m[key] = value
	return a
}

func (a *attribute) SetBool(key string, value bool) *attribute {
	a.m[key] = value
	return a
}

func (a *attribute) SetNumber(key string, value float64) *attribute {
	a.m[key] = value
	return a
}

func (a *attribute) Get(key string) interface{} {
	return a.m[key]
}

func (a *attribute) Delete(key string) {
	delete(a.m, key)
}

func (a *attribute) Clear() {
	a.m = make(map[string]interface{})
}

func (a *attribute) Keys() []string {
	keys := make([]string, 0, len(a.m))
	for key := range a.m {
		keys = append(keys, key)
	}
	return keys
}

func (a *attribute) Values() []interface{} {
	values := make([]interface{}, 0, len(a.m))
	for _, value := range a.m {
		values = append(values, value)
	}
	return values
}

func (a *attribute) Len() int {
	return len(a.m)
}
