//
// Copyright 2019 Lightbend Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package encoding

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"reflect"
	"testing"
)

var testsJSON = []struct {
	name       string
	value      interface{}
	zero       interface{}
	typeURL    string
	shouldFail bool
}{
	{jsonTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.a",
		a{B: "29", C: 29}, a{}, jsonTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.a", false},
	{jsonTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.aDefault" + "_defaultValue",
		aDefault{}, aDefault{}, jsonTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.aDefault", false},
}

func TestMarshallerJSON(t *testing.T) {
	for _, test := range testsJSON {
		test0 := test
		t.Run(fmt.Sprintf("%v", test0.name), func(t *testing.T) {
			any0, err := MarshalJSON(test0.value)
			hasErr := err != nil
			if hasErr && !test0.shouldFail {
				t.Errorf("err: %v, got: %+v, expected: %+v", err, any0, test0)
				return
			} else if !hasErr && test0.shouldFail {
				t.Errorf("err: %v, got: %+v, expected: %+v", err, any0, test0)
				return
			}
			failed := any0.GetTypeUrl() != test0.typeURL
			if failed && !test0.shouldFail {
				t.Errorf("err: %v, got: %+v, expected: %+v", err, any0, test0)
			} else if !failed && test0.shouldFail {
				t.Errorf("err: %v, got: %+v, expected: %+v", err, any0, test0)
			}
			value := reflect.New(reflect.TypeOf(test0.value))
			err = UnmarshalJSON(any0, value.Interface())
			if err != nil {
				t.Error(err)
			}
			if test0.value != value.Elem().Interface() {
				t.Errorf("err: %v. got: %+v, expected: %+v", err, value.Elem().Interface(), test0.value)
			}
		})
	}
}

func BenchmarkMarshallerJSON(b *testing.B) {
	var any0 *any.Any
	for _, bench := range testsJSON {
		if !bench.shouldFail {
			b.Run(bench.name, func(b *testing.B) {
				b.ReportAllocs()
				var any1 *any.Any
				for i := 0; i < b.N; i++ {
					any0, err := MarshalJSON(bench.value)
					if err != nil {
						b.Error(err)
					}
					value := bench.zero
					err = UnmarshalJSON(any0, &value)
					if err != nil {
						b.Error(err)
					}
				}
				any0 = any1 //prevent the call optimized away
			})
		}
	}
	_ = any0 == nil //use any0
}
