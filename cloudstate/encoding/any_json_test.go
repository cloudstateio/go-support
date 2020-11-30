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
	"testing"

	"github.com/golang/protobuf/ptypes/any"
)

func TestMarshalling(t *testing.T) {
	t.Run("marshals to cloudstate any.Any", func(t *testing.T) {
		s := &a{B: "29", C: 29}
		x, err := MarshalJSON(s)
		if err != nil {
			t.Fail()
		}
		url := "json.cloudstate.io/github.com/cloudstateio/go-support/cloudstate/encoding.a"
		if x.GetTypeUrl() != url {
			t.Fail()
		}
	})
	t.Run("marshal/unmarshal pointer struct", func(t *testing.T) {
		s := &a{B: "29", C: 29}
		x, err := MarshalJSON(s)
		if err != nil {
			t.Error(err)
		}
		s1 := &a{}
		err = UnmarshalJSON(x, s1)
		if err != nil {
			t.Error(err)
		}
		if *s != *s1 {
			t.Fail()
		}
	})
	t.Run("marshal/unmarshal struct", func(t *testing.T) {
		s := a{B: "29", C: 29}
		x, err := MarshalJSON(s)
		if err != nil {
			t.Error(err)
		}
		s1 := a{}
		err = UnmarshalJSON(x, &s1)
		if err != nil {
			t.Error(err)
		}
		if s != s1 {
			t.Fail()
		}
	})
	// t.Run("marshal/unmarshal proto message", func(t *testing.T) {
	// 	s := entity.CrdtState_Flag{Flag: &entity.FlagState{
	// 		Value: true,
	// 	}}
	// 	x, err := MarshalJSON(s)
	// 	if err != nil {
	// 		t.Error(err)
	// 	}
	// 	s1 := entity.CrdtState_Flag{Flag: &entity.FlagState{
	// 		Value: true,
	// 	}}
	// 	err = UnmarshalJSON(x, &s1)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	if s.Flag.Value != s1.Flag.Value {
	// 		t.Fatal()
	// 	}
	// })
}

var testsJSON = []struct {
	name       string
	value      interface{}
	zero       interface{}
	typeURL    string
	shouldFail bool
}{
	{JSONTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.a",
		&a{B: "29", C: 29}, &a{}, JSONTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.a", false},
	{JSONTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.a",
		a{B: "29", C: 29}, a{}, JSONTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.a", false},
	{JSONTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.aDefault" + "_defaultValue",
		aDefault{}, aDefault{}, JSONTypeURLPrefix + "/github.com/cloudstateio/go-support/cloudstate/encoding.aDefault", false},
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
				any0 = any1 // prevent the call optimized away
			})
		}
	}
	_ = any0 == nil // use any0
}
