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

package synth

import (
	"reflect"
	"testing"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

type tester struct {
	t *testing.T
}

func (t *tester) toStruct(x *any.Any, i interface{}) {
	t.t.Helper()
	if err := encoding.DecodeStruct(x, i); err != nil {
		t.t.Fatal(err)
	}
}

func (t *tester) toProto(x *any.Any, p proto.Message) {
	t.t.Helper()
	if err := encoding.UnmarshalAny(x, p); err != nil {
		t.t.Fatal(err)
	}
}

func (t *tester) unexpected(i ...interface{}) {
	t.t.Helper()
	t.t.Fatalf("got unexpected: %+v", i...)
}

func (t *tester) expectedInt(got int, want int) {
	t.t.Helper()
	if got != want {
		t.t.Fatalf("got = %v; wanted: %d", got, want)
	}
}

func (t *tester) expectedInt64(got int64, want int64) {
	t.t.Helper()
	if got != want {
		t.t.Fatalf("got = %v; wanted: %d", got, want)
	}
}

func (t *tester) expectedUInt64(got uint64, want uint64) {
	t.t.Helper()
	if got != want {
		t.t.Fatalf("got = %v; wanted: %d", got, want)
	}
}

func (t *tester) expectedUInt32(got uint32, want uint32) {
	t.t.Helper()
	if got != want {
		t.t.Fatalf("got = %v; wanted: %d", got, want)
	}
}

func (t *tester) expectedInt32(got int32, want int32) {
	t.t.Helper()
	if got != want {
		t.t.Fatalf("got = %v; wanted: %d", got, want)
	}
}

func (t *tester) expectedString(got string, want string) {
	t.t.Helper()
	if got != want {
		t.t.Fatalf("got = %v; wanted: %s", got, want)
	}
}

func (t *tester) expectedNoError(got error) {
	t.t.Helper()
	if got != nil {
		t.t.Fatalf("got = %v; wanted: nil", got)
	}
}

func (t *tester) expectedNil(got interface{}) {
	t.t.Helper()
	if got == nil {
		return
	}
	if reflect.ValueOf(got).IsNil() {
		return
	}
	t.t.Fatalf("got = %v; wanted: nil", got)
}

func (t *tester) expectedNotNil(got interface{}) {
	t.t.Helper()
	if got == nil {
		t.t.Fatalf("got = %v; wanted: not nil", got)
	}
	if reflect.ValueOf(got).IsNil() {
		t.t.Fatalf("got = %v; wanted: not nil", got)
	}
}

func (t *tester) expectedBool(got bool, want bool) {
	t.t.Helper()
	if got != want {
		t.t.Fatalf("got = %v; wanted: %v", got, want)
	}
}

func (t *tester) expectedTrue(got bool) {
	t.t.Helper()
	if !got {
		t.t.Fatalf("got = %v; wanted: true", got)
	}
}

func (t *tester) expectedFalse(got bool) {
	t.t.Helper()
	if got {
		t.t.Fatalf("got = %v; wanted: false", got)
	}
}

func (t *tester) expectedSame(x *any.Any, i interface{}) {
	t.t.Helper()
	if !oneEquals([]*any.Any{x}, i) {
		t.t.Fatalf("none of %+v found in %+v", i, x)
	}
}

func (t *tester) expectedOneIn(x []*any.Any, i interface{}) {
	t.t.Helper()
	if !oneEquals(x, i) {
		t.t.Fatalf("none of %+v found in %+v", i, x)
	}
}

func oneEquals(x []*any.Any, i interface{}) bool {
	for _, a := range x {
		if reflect.DeepEqual(a, i) {
			return true
		}
	}
	return false
}
