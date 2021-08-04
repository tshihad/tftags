package tftags

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Get(d *schema.ResourceData, v interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.CanSet() {
		return errors.New("only supports pointer structs")
	}
	fmt.Println(rv.String())
	if rv.Kind() != reflect.Struct {
		return errors.New("only struct type is supported")
	}

	recursive(rv, d, "", nil)
	return nil
}

func recursive(rv reflect.Value, d *schema.ResourceData, path string, schemaMap interface{}) {
	switch rv.Kind() {
	case reflect.Struct:
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
					recursive(rv.Field(i), d, newPath, val)
				}
			}
		}
	case reflect.Slice:
		if array, ok := schemaMap.([]interface{}); ok {
			slice := reflect.MakeSlice(rv.Type(), len(array), cap(array))
			rv.Set(slice)
			for i := 0; i < rv.Len(); i++ {
				recursive(rv.Index(i), d, fmt.Sprintf("%s.%d", path, i), array[i])
			}
		}
	default:
		fmt.Println(rv.Kind())
		rv.Set(reflect.ValueOf(schemaMap))
	}
}
