package tftags

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Get(d *schema.ResourceData, v interface{}) error {
	rv := reflect.ValueOf(v).Elem()
	if !rv.CanSet() {
		return errors.New("only supports pointer structs")
	}
	fmt.Println(rv.String())
	if rv.Kind() != reflect.Struct {
		return errors.New("only struct type is supported")
	}

	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		if value, ok := t.Field(i).Tag.Lookup("tf"); ok {
			splitTags := strings.Split(value, ",")
			if len(splitTags) < 1 {
				return errors.New("no proper tag value")
			}

			if result, ok := d.GetOk(splitTags[0]); ok {
				rv.Field(i).Set(reflect.ValueOf(result))
			}
		}
	}

	return nil
}
