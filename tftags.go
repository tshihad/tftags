package tftags

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	computedTag = "computed"
	subTag      = "sub"
)

// Get accepts two argument. d contians ResourceData and v is the output struct
func Get(d resourceData, output interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(output))
	if !rv.CanSet() {
		return errors.New("input is not settable")
	}
	// currently only struct type is supported
	if rv.Kind() != reflect.Struct {
		return errors.New("only struct type is supported")
	}

	recursiveGet(rv, d, "", nil, false)
	return nil
}

// recursively run over the schema and populate the ouput struct. SchemaMap maps all the
// values in schema into an interface. path will the complete path to a value
func recursiveGet(rv reflect.Value, d resourceData, path string, schemaMap interface{}, isSub bool) {
	switch rv.Kind() {
	case reflect.Struct:
		// for type struct loop through all values and check tags 'tf'
		t := rv.Type()
		for i := 0; i < t.NumField(); i++ {
			if value, ok := t.Field(i).Tag.Lookup("tf"); ok {
				splitTags := strings.Split(value, ",")
				var newPath string
				if path != "" {
					newPath = path + "." + splitTags[0]
					// if sub then it is *Set, find the hash value and include in path
					if isSub {
						setData := schemaMap.(*schema.Set)
						newPath = fmt.Sprintf("%s.%d.%s", path, setData.F(setData.List()[0]), splitTags[0])
					}
				} else {
					newPath = splitTags[0]
				}
				// Get corresponding data from schema
				if val, ok := d.GetOk(newPath); ok {
					// iterate to the corresponding field and call recursiveGet again
					recursiveGet(rv.Field(i), d, newPath, val, searchTags(splitTags, subTag))
				}
			}
		}
	case reflect.Slice:
		var sArray reflect.Value
		// if the output contains field slice, check is it is a set or list
		schemaSet, isSet := schemaMap.(*schema.Set)
		if isSet {
			sArray = reflect.ValueOf(schemaSet.List())
		} else {
			sArray = reflect.ValueOf(schemaMap)
		}
		slice := reflect.MakeSlice(rv.Type(), sArray.Len(), sArray.Cap())
		rv.Set(slice)

		for i := 0; i < rv.Len(); i++ {
			// set path
			var newPath string
			if isSet {
				newPath = fmt.Sprintf("%s.%d", path, schemaSet.F(schemaSet.List()[i]))
			} else {
				newPath = fmt.Sprintf("%s.%d", path, i)
			}
			// recursively set each elements in slice
			recursiveGet(
				rv.Index(i),
				d,
				newPath,
				sArray.Index(i).Interface(),
				false)
		}
	case reflect.Map:
		// if output is map and schemaMap also map then allocates new map
		// to output
		rvMap := reflect.ValueOf(schemaMap)
		rv.Set(reflect.MakeMap(rv.Type()))

		for _, key := range rvMap.MapKeys() {
			// Set index and value directly here
			rv.SetMapIndex(key, rvMap.MapIndex(key))
		}
	default:
		rv.Set(reflect.ValueOf(schemaMap))
	}
}

func Set(d resourceData, v interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	// currently only struct type is supported
	if rv.Kind() != reflect.Struct {
		return errors.New("only struct type is supported")
	}
	// var result interface{}
	recursiveSet(rv, d, false)
	return nil
}

func recursiveSet(rv reflect.Value, d resourceData, computed bool) interface{} {
	switch rv.Kind() {
	case reflect.Struct:
		if rv.IsZero() {
			return nil
		}
		t := rv.Type()
		rMap := make(map[string]interface{})
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if value, ok := field.Tag.Lookup("tf"); ok {
				splitTags := strings.Split(value, ",")

				isSub := searchTags(splitTags, subTag)
				// if computed is true then it indicates it is a child struct
				if computed {
					rMap[splitTags[0]] = checkSub(rv, i, d, isSub)

				} else if searchTags(splitTags, computedTag) { // Check computed tags
					result := checkSub(rv, i, d, isSub)
					if !isEmpty(result) {
						d.Set(splitTags[0], result)
					}
				}
			} // else {
			// For non computed fields iterate all elements recursively
			// 		recursiveSet(rv.Field(i), d, false)
			// }
		}
		return rMap

	case reflect.Slice:
		result := make([]interface{}, rv.Len())
		// iterate through array and figure it out values. Value can be map, struct,
		// slice or primitive data type
		for i := 0; i < rv.Len(); i++ {
			result[i] = recursiveSet(rv.Index(i), d, computed)
		}

		return result

	case reflect.Map:
		result := make(map[string]interface{})
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			result[k.String()] = v.Interface()
		}

		return result
	}

	// Primitive data type
	return rv.Interface()
}

// checkSub check whether this is a sub element and recusively
// allocates output struct regarding it
func checkSub(rv reflect.Value, i int, d resourceData, isSub bool) interface{} {
	var result interface{}
	// if isSub allocates as array
	// isSub will denotes this is a sub element where schema contains array with one element
	// but input data structure having struct
	if isSub {
		r := recursiveSet(rv.Field(i), d, true)
		if isEmpty(r) {
			return nil
		}
		result = []interface{}{r}
	} else {
		result = recursiveSet(rv.Field(i), d, true)
	}
	return result
}

// searchTags search items in tags
func searchTags(slice []string, item string) bool {
	for i := 1; i < len(slice); i++ {
		if slice[i] == item {
			return true
		}
	}
	return false
}

// isEmpty checks given iterface is empty or not
func isEmpty(n interface{}) bool {
	return n == nil || reflect.ValueOf(n).IsZero()
}
