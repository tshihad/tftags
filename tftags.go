package tftags

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Get accepts two argument. d contians ResourceData and v is the output struct
func Get(d *schema.ResourceData, output interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(output))
	if !rv.CanSet() {
		return errors.New("input is not settable")
	}
	// currently only struct type is supported
	if rv.Kind() != reflect.Struct {
		return errors.New("only struct type is supported")
	}

	recursiveGet(rv, d, "", nil)
	return nil
}

// recursively run over the schema and populate the ouput struct. SchemaMap maps all the
// values in schema into an interface. path will the complete path to a value
func recursiveGet(rv reflect.Value, d *schema.ResourceData, path string, schemaMap interface{}) {
	switch rv.Kind() {
	case reflect.Struct:
		// for type struct loop through all values and check tags 'tf'
		t := rv.Type()
		for i := 0; i < t.NumField(); i++ {
			if value, ok := t.Field(i).Tag.Lookup("tf"); ok {
				splitTags := strings.Split(value, ",")
				if len(splitTags) < 1 {
					panic("no proper tag value")
				}
				var newPath string
				if path != "" {
					newPath = path + "." + splitTags[0]
				} else {
					newPath = splitTags[0]
				}
				// Get corresponding data from schema
				if val, ok := d.GetOk(newPath); ok {
					// iterate to the corresponding field and call recursiveGet again
					recursiveGet(rv.Field(i), d, newPath, val)
				}
			}
		}
	case reflect.Slice:
		// if the output contains field slice, check correspoding schemaMap is also
		// a slice, if so allocate new slice to the rv
		if array, ok := schemaMap.([]interface{}); ok {
			slice := reflect.MakeSlice(rv.Type(), len(array), cap(array))
			rv.Set(slice)
			for i := 0; i < rv.Len(); i++ {
				// recursively set each elements in slice
				recursiveGet(rv.Index(i), d, fmt.Sprintf("%s.%d", path, i), array[i])
			}
		}
	case reflect.Map:
		// if output is map and schemaMap also map then allocates new map
		// to output
		if m, ok := schemaMap.(map[string]interface{}); ok {
			rvMap := reflect.MakeMap(rv.Type())
			rv.Set(rvMap)

			for k, v := range m {
				// Set index and value directly here
				rv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
				// recursive(rv.MapIndex(reflect.ValueOf(k)), d, fmt.Sprintf("%s.%s", path, k), v)
			}
		}
	default:
		rv.Set(reflect.ValueOf(schemaMap))
	}
}
