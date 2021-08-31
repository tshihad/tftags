package tftags

// resourceData is the representation of *schema.ResourceData
type resourceData interface {
	// Get(key string) interface{}
	GetOk(key string) (interface{}, bool)
	Set(key string, value interface{}) error
}
