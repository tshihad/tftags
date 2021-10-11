package tftags

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// convert to string
func toString(v interface{}) string {
	return fmt.Sprint(v)
}

// compare src and dest, if both kind are same, assign dest=src
// else try to convert src to dest type and set
func compareAndSet(dest reflect.Value, src interface{}) {
	srcRef := reflect.ValueOf(src)
	if srcRef.Kind() != dest.Kind() {
		switch dest.Kind() {
		case reflect.Int:
			intVal, err := strconv.Atoi(toString(src))
			if err != nil {
				log.Fatalf("failed to convert %T to int", src)
			}
			dest.Set(reflect.ValueOf(intVal))
		case reflect.String:
			dest.Set(reflect.ValueOf(toString(src)))
		default:
			log.Panicf("cannot convert %s to %s", srcRef.Kind().String(), dest.Kind().String())
		}
	} else {
		dest.Set(srcRef)
	}
}
