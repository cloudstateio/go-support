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
	"testing"

	"github.com/golang/protobuf/ptypes/any"
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
	{PrimitiveTypeURLPrefixInt32, uint32(28), PrimitiveTypeURLPrefixInt32, true},
	{PrimitiveTypeURLPrefixInt32 + "_defaultValue", uint32(0), PrimitiveTypeURLPrefixInt32, true},
	{PrimitiveTypeURLPrefixInt32, int32(29), PrimitiveTypeURLPrefixInt32, false},
	{PrimitiveTypeURLPrefixInt32 + "_defaultValue", int32(0), PrimitiveTypeURLPrefixInt32, false},
	{PrimitiveTypeURLPrefixInt64, int64(29), PrimitiveTypeURLPrefixInt64, false},
	{PrimitiveTypeURLPrefixInt64 + "_defaultValue", int64(0), PrimitiveTypeURLPrefixInt64, false},
	{PrimitiveTypeURLPrefixFloat, float32(2.9), PrimitiveTypeURLPrefixFloat, false},
	{PrimitiveTypeURLPrefixFloat + "_defaultValue", float32(2.9), PrimitiveTypeURLPrefixFloat, false},
	{PrimitiveTypeURLPrefixDouble, float64(2.9), PrimitiveTypeURLPrefixDouble, false},
	{PrimitiveTypeURLPrefixDouble + "_defaultValue", float64(0), PrimitiveTypeURLPrefixDouble, false},
	{PrimitiveTypeURLPrefixString, "29", PrimitiveTypeURLPrefixString, false},
	{PrimitiveTypeURLPrefixString + "_defaultValue", "", PrimitiveTypeURLPrefixString, false},
	{PrimitiveTypeURLPrefixBool + "_true", true, PrimitiveTypeURLPrefixBool, false},
	{PrimitiveTypeURLPrefixBool + "_false", false, PrimitiveTypeURLPrefixBool, false},
	{PrimitiveTypeURLPrefixBool + "_defaultValue", false, PrimitiveTypeURLPrefixBool, false},
	{PrimitiveTypeURLPrefixBytes, make([]byte, 29), PrimitiveTypeURLPrefixBytes, false},
	{PrimitiveTypeURLPrefixBytes + "_defaultValue", make([]byte, 0), PrimitiveTypeURLPrefixBytes, false},
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
				if !bytes.Equal(byt, ut) {
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
				any0 = any1 // prevent the call optimized away
			})
		}
	}
	_ = any0 == nil // use any0
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
						if !bytes.Equal(byt, ut) {
							b.Errorf("err: %v. got: %+v, expected: %+v", err, u, tc.value)
						}
					default:
						if tc.value != u {
							b.Errorf("err: %v. got: %+v, expected: %+v", err, u, tc.value)
						}
					}
				}
				any0 = any1 // prevent the call optimized away
			})
		}
	}
	_ = any0 == nil // use any0
}
