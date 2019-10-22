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
	"bytes"
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"testing"
)

type a struct {
	B string `json:"b"`
	C int32  `json:"c"`
}

type aDefault struct {
}

var tests = []struct {
	name       string
	value      interface{}
	typeURL    string
	shouldFail bool
}{
	{primitiveTypeURLPrefixInt32, uint32(28), primitiveTypeURLPrefixInt32, true},
	{primitiveTypeURLPrefixInt32 + "_defaultValue", uint32(0), primitiveTypeURLPrefixInt32, true},
	{primitiveTypeURLPrefixInt32, int32(29), primitiveTypeURLPrefixInt32, false},
	{primitiveTypeURLPrefixInt32 + "_defaultValue", int32(0), primitiveTypeURLPrefixInt32, false},
	{primitiveTypeURLPrefixInt64, int64(29), primitiveTypeURLPrefixInt64, false},
	{primitiveTypeURLPrefixInt64 + "_defaultValue", int64(0), primitiveTypeURLPrefixInt64, false},
	{primitiveTypeURLPrefixFloat, float32(2.9), primitiveTypeURLPrefixFloat, false},
	{primitiveTypeURLPrefixFloat + "_defaultValue", float32(2.9), primitiveTypeURLPrefixFloat, false},
	{primitiveTypeURLPrefixDouble, float64(2.9), primitiveTypeURLPrefixDouble, false},
	{primitiveTypeURLPrefixDouble + "_defaultValue", float64(0), primitiveTypeURLPrefixDouble, false},
	{primitiveTypeURLPrefixString, "29", primitiveTypeURLPrefixString, false},
	{primitiveTypeURLPrefixString + "_defaultValue", "", primitiveTypeURLPrefixString, false},
	{primitiveTypeURLPrefixBool + "_true", true, primitiveTypeURLPrefixBool, false},
	{primitiveTypeURLPrefixBool + "_false", false, primitiveTypeURLPrefixBool, false},
	{primitiveTypeURLPrefixBool + "_defaultValue", false, primitiveTypeURLPrefixBool, false},
	{primitiveTypeURLPrefixBytes, make([]byte, 29), primitiveTypeURLPrefixBytes, false},
	{primitiveTypeURLPrefixBytes + "_defaultValue", make([]byte, 0), primitiveTypeURLPrefixBytes, false},
}

func TestMarshallerPrimitives(t *testing.T) {
	for _, test := range tests {
		tc := test
		t.Run(fmt.Sprintf("%v", tc.name), func(t *testing.T) {
			any0, err := MarshalPrimitive(tc.value)
			hasErr := err != nil
			if hasErr && !tc.shouldFail {
				t.Errorf("err: %v, got: %+v, expected: %+v", err, any0, tc)
				return
			} else if !hasErr && tc.shouldFail {
				t.Errorf("err: %v, got: %+v, expected: %+v", err, any0, tc)
				return
			}
			failed := any0.GetTypeUrl() != tc.typeURL
			if failed && !tc.shouldFail {
				t.Errorf("err: %v, got: %+v, expected: %+v", err, any0, tc)
			} else if !failed && tc.shouldFail {
				t.Errorf("err: %v, got: %+v, expected: %+v", err, any0, tc)
			}
		})
	}
}

func TestMarshalUnmarshalPrimitive(t *testing.T) {
	for _, test := range tests {
		if test.shouldFail {
			continue
		}
		tc := test
		t.Run(fmt.Sprintf("%v", tc.name), func(t *testing.T) {
			a, err := MarshalPrimitive(tc.value)
			if err != nil {
				t.Error(err)
			}
			u, err := UnmarshalPrimitive(a)
			if err != nil {
				t.Error(err)
			}
			switch ut := u.(type) {
			case []byte:
				byt := tc.value.([]byte)
				if bytes.Compare(byt, ut) != 0 {
					t.Errorf("err: %v. got: %+v, expected: %+v", err, u, tc.value)
				}
			default:
				if tc.value != u {
					t.Errorf("err: %v. got: %+v, expected: %+v", err, u, tc.value)
				}
			}
		})
	}
}

func BenchmarkMarshallerPrimitives(b *testing.B) {
	var any0 *any.Any
	for _, i := range tests {
		tc := i
		if !tc.shouldFail {
			b.Run(tc.name, func(b *testing.B) {
				b.ReportAllocs()
				var any1 *any.Any
				for i := 0; i < b.N; i++ {
					any1, _ = MarshalPrimitive(tc.value)
				}
				any0 = any1 //prevent the call optimized away
			})
		}
	}
	_ = any0 == nil //use any0
}

func BenchmarkMarshalUnmarshal(b *testing.B) {
	var any0 *any.Any
	for _, i := range tests {
		tc := i
		if !tc.shouldFail {
			b.Run(tc.name, func(b *testing.B) {
				b.ReportAllocs()
				var any1 *any.Any
				for i := 0; i < b.N; i++ {
					a, err := MarshalPrimitive(tc.value)
					if err != nil {
						b.Error(err)
					}
					u, err := UnmarshalPrimitive(a)
					if err != nil {
						b.Error(err)
					}
					switch ut := u.(type) {
					case []byte:
						byt := tc.value.([]byte)
						if bytes.Compare(byt, ut) != 0 {
							b.Errorf("err: %v. got: %+v, expected: %+v", err, u, tc.value)
						}
					default:
						if tc.value != u {
							b.Errorf("err: %v. got: %+v, expected: %+v", err, u, tc.value)
						}
					}
				}
				any0 = any1 //prevent the call optimized away
			})
		}
	}
	_ = any0 == nil //use any0
}
