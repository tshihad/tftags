package tftags

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type rdTestImp struct {
	vals   map[string]interface{}
	setErr bool
}

func (r *rdTestImp) GetOk(key string) (interface{}, bool) {
	paths := strings.Split(key, ".")
	val, ok := r.vals[paths[0]]
	for i := 1; i < len(paths); i++ {
		val, ok = val.(map[string]interface{})[paths[i]]
	}
	return val, ok
}
func (r *rdTestImp) Set(key string, value interface{}) error {
	if r.setErr {
		return errors.New("error")
	}
	r.vals[key] = value
	return nil
}

func TestGet(t *testing.T) {
	type TT1 struct {
		Name string `tf:"name"`
		Data int    `tf:"data"`
	}
	type TT2 struct {
		M     map[string]int `tf:"m"`
		T1    TT1            `tf:"t1"`
		Array []string       `tf:"array"`
	}
	tests := []struct {
		name    string
		args    interface{}
		given   func(r *rdTestImp)
		want    interface{}
		wantErr bool
	}{
		{
			name: "Normal test case 1: Get values for a linear struct",
			args: &TT1{},
			given: func(r *rdTestImp) {
				r.vals = map[string]interface{}{
					"name": "test 1 name",
					"data": 123,
				}
			},
			want: &TT1{
				Name: "test 1 name",
				Data: 123,
			},
		},
		{
			name: "Normal test case 2: Get values for a complex struct",
			args: &TT2{},
			given: func(r *rdTestImp) {
				r.vals = map[string]interface{}{
					"m": map[string]int{
						"data": 2,
					},
					"t1": map[string]interface{}{
						"name": "test 1 name",
						"data": 123,
					},
					"array": []string{"test1", "test2"},
				}
			},
			want: &TT2{
				M: map[string]int{
					"data": 2,
				},
				T1: TT1{
					Name: "test 1 name",
					Data: 123,
				},
				Array: []string{"test1", "test2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &rdTestImp{}
			tt.given(d)
			if err := Get(d, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args, tt.want) {
				t.Errorf("want %v, but got %v", tt.want, tt.args)
			}
		})
	}
}
