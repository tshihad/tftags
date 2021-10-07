package tftags

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type rdTestImp struct {
	vals   map[string]interface{}
	setErr bool
}

func (r *rdTestImp) GetOk(key string) (interface{}, bool) {
	paths := strings.Split(key, ".")
	val, ok := r.vals[paths[0]]
	setIter := 0

	for i := 1; i < len(paths); i++ {
		switch v := val.(type) {
		case map[string]interface{}:
			val, ok = v[paths[i]]
		case []interface{}:
			index, err := strconv.Atoi(paths[i])
			if err != nil {
				panic(err)
			}
			val = v[index]
		case *schema.Set:
			val = v.List()[setIter]
			// setIter++
		default:
			panic(fmt.Sprintf("unknown type %v", v))
		}
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

func (r *rdTestImp) GetId() string {
	return "1"
}

func (r *rdTestImp) SetId(v string) {
	r.vals["id"] = v
}

type TT0 struct {
	T1 *TT1 `tf:"t1,computed"`
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

type TT3 struct {
	T2 TT2 `tf:"t2,computed,sub"`
}

type TT4 struct {
	ID    string `tf:"id,computed"`
	Array []TT1  `tf:"array,computed"`
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
			name: "Normal test case 0: Get values with pointer",
			args: &TT0{},
			given: func(r *rdTestImp) {
				r.vals = map[string]interface{}{
					"t1": map[string]interface{}{
						"name": "test 1 name",
						"data": 123,
					},
				}
			},
			want: &TT0{
				T1: &TT1{
					Name: "test 1 name",
					Data: 123,
				},
			},
		},
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
		{
			name: "Normal test case 3: Get values having sub",
			args: &TT3{},
			given: func(r *rdTestImp) {
				set := schema.NewSet(func(data interface{}) int {
					m := data.(map[string]interface{})
					for k := range m {
						return len(k)
					}
					return 0
				}, []interface{}{
					map[string]interface{}{
						"m": map[string]int{
							"data": 2,
						},
						"t1": map[string]interface{}{
							"name": "test 1 name",
							"data": 123,
						},
						"array": []interface{}{"test1", "test2"},
					},
				})
				r.vals = map[string]interface{}{
					"t2": set,
				}
			},
			want: &TT3{
				T2: TT2{
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
		},
		{
			name: "Test case 4: Array of struct",
			args: &TT4{},
			given: func(r *rdTestImp) {
				r.vals = map[string]interface{}{
					"array": []interface{}{
						map[string]interface{}{
							"name": "test1",
							"data": 12,
						},
						map[string]interface{}{
							"name": "test2",
							"data": 13,
						},
					},
				}
			},
			want: &TT4{
				ID: "1",
				Array: []TT1{
					{
						Name: "test1",
						Data: 12,
					},
					{
						Name: "test2",
						Data: 13,
					},
				},
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
		{
			name: "Normal test case 0: Get values with pointer",
			args: &TT0{
				T1: &TT1{
					Name: "test 1 name",
					Data: 123,
				},
			},
			want: map[string]interface{}{
				"t1": map[string]interface{}{
					"name": "test 1 name",
					"data": 123,
				},
			},
		},
		{
			name: "Normal test case 1: Set linear struct",
			args: TT1{
				Name: "test1",
				Data: 54,
			},
			want: map[string]interface{}{
				"name": "test1",
				"data": 54,
			},
		},
		{
			name: "Normal test case 2: Set complex struct",
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
				"m": map[string]interface{}{
					"data": 2,
				},
				"t1": map[string]interface{}{
					"name": "test 1 name",
					"data": 123,
				},
				"array": []interface{}{"test1", "test2"},
			},
		},
		{
			name: "Normal test case 3: struct with sub item",
			args: TT3{
				T2: TT2{
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
			want: map[string]interface{}{
				"t2": []interface{}{
					map[string]interface{}{
						"m": map[string]interface{}{
							"data": 2,
						},
						"t1": map[string]interface{}{
							"name": "test 1 name",
							"data": 123,
						},
						"array": []interface{}{"test1", "test2"},
					},
				},
			},
		},
		{
			name: "Normal test case 4: Set ID",
			args: TT4{
				ID: "12",
			},
			want: map[string]interface{}{
				"id": "12",
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
				t.Errorf("want %#v,\n but got %#v", tt.want, d.vals)
			}
		})
	}
}
