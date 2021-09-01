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

type TT1 struct {
	Name string `tf:"name,computed"`
	Data int    `tf:"data,computed"`
}
type TT2 struct {
	M     map[string]int `tf:"m,computed"`
	T1    TT1            `tf:"t1,computed"`
	Array []string       `tf:"array,computed"`
}

func TestGet(t *testing.T) {
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
			if !reflect.DeepEqual(tt.want, tt.args) {
				t.Errorf("want %v, but got %v", tt.want, tt.args)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		// given   func(r *rdTestImp)
		wantErr bool
		want    map[string]interface{}
	}{
		// {
		// 	name: "Normal test case 1: Set linear struct",
		// 	args: TT1{
		// 		Name: "test1",
		// 		Data: 54,
		// 	},
		// 	want: map[string]interface{}{
		// 		"name": "test1",
		// 		"data": 54,
		// 	},
		// },
		{
			name: "Normal test case 12: Set complex struct",
			args: TT2{
				M: map[string]int{
					"data": 2,
				},
				T1: TT1{
					Name: "test 1 name",
					Data: 123,
				},
				Array: []string{"test1", "test2"},
			},
			want: map[string]interface{}{
				"m": map[string]int{
					"data": 2,
				},
				"t1": map[string]interface{}{
					"name": "test 1 name",
					"data": 123,
				},
				"array": []string{"test1", "test2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &rdTestImp{
				vals: make(map[string]interface{}),
			}
			if err := Set(d, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, d.vals) {
				t.Errorf("want %v, but got %v", tt.want, d.vals)
			}
		})
	}
}