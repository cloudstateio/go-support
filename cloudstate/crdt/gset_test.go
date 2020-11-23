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

package crdt

import (
	"testing"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/golang/protobuf/ptypes/any"
)

func TestGset(t *testing.T) {
	delta := func(x []*any.Any) *entity.CrdtDelta {
		return &entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Gset{
				Gset: &entity.GSetDelta{
					Added: x,
				},
			},
		}
	}

	t.Run("should have no elements when instantiated", func(t *testing.T) {
		s := NewGSet()
		if s.Size() != 0 {
			t.FailNow()
		}
		if s.HasDelta() {
			t.Fatal("has delta but should not")
		}
		itemsLen := len(s.Value())
		if itemsLen != 0 {
			t.Fatalf("len(items): %v; want: %v", itemsLen, 0)
		}
	})

	t.Run("should generate an add delta", func(t *testing.T) {
		s := NewGSet()
		s.Add(encoding.String("one"))
		if !contains(s.Value(), "one") {
			t.Fatal("set should have a: one")
		}
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		delta := encDecDelta(s.Delta())
		s.resetDelta()
		addedLen := len(delta.GetGset().GetAdded())
		if addedLen != 1 {
			t.Fatalf("s.Size(): %v; want: %v", addedLen, 1)
		}
		if !contains(delta.GetGset().GetAdded(), "one") {
			t.Fatalf("set should have a: one")
		}
		if s.HasDelta() {
			t.Fatalf("has but should not")
		}
		s.Add(encoding.String("two"))
		s.Add(encoding.String("three"))
		if s.Size() != 3 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 3)
		}
		delta2 := encDecDelta(s.Delta())
		addedLen2 := len(delta2.GetGset().GetAdded())
		if addedLen2 != 2 {
			t.Fatalf("s.Size(): %v; want: %v", addedLen2, 2)
		}
		if !contains(delta2.GetGset().GetAdded(), "two", "three") {
			t.Fatalf("delta should include two, three")
		}
		s.resetDelta()
		if s.HasDelta() {
			t.Fatalf("has delta but should not")
		}
	})

	t.Run("should not generate a delta when an already existing element is added", func(t *testing.T) {
		s := NewGSet()
		s.Add(encoding.String("one"))
		s.resetDelta()
		s.Add(encoding.String("one"))
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		if s.HasDelta() {
			t.Fatalf("has delta but should not")
		}
	})

	t.Run("should reflect a delta add", func(t *testing.T) {
		s := NewGSet()
		s.Add(encoding.String("one"))
		s.resetDelta()
		if err := s.applyDelta(delta(append(make([]*any.Any, 0), encoding.String("two")))); err != nil {
			t.Fatal(err)
		}
		if s.Size() != 2 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 2)
		}
		if !contains(s.Value(), "one", "two") {
			t.Fatalf("delta should include two, three")
		}
		if s.HasDelta() {
			t.Fatalf("has delta but should not")
		}
		state := s.Value()
		if len(state) != 2 {
			t.Fatalf("state.GetItems(): %v; want: %v", state, 2)
		}
	})

	t.Run("should work with protobuf types", func(t *testing.T) {
		s := NewGSet()
		type Example struct {
			Field1 string
		}
		field1, err := encoding.Struct(&Example{Field1: "one"})
		if err != nil {
			t.Fatal(err)
		}
		s.Add(field1)
		s.resetDelta()
		s.Add(field1)
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		field2, err := encoding.Struct(&Example{Field1: "two"})
		if err != nil {
			t.Fatal(err)
		}
		s.Add(field2)
		if s.Size() != 2 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 2)
		}
		delta := encDecDelta(s.Delta())
		if len(delta.GetGset().GetAdded()) != 1 {
			t.Fatalf("s.Size(): %v; want: %v", len(delta.GetGset().GetAdded()), 1)
		}
		foundOne := false
		for _, v := range delta.GetGset().GetAdded() {
			e := Example{}
			if err := encoding.UnmarshalJSON(v, &e); err != nil {
				t.Fatal(err)
			}
			if e.Field1 == "two" {
				foundOne = true
			}
		}
		if !foundOne {
			t.Fatalf("delta should include two")
		}
	})

	type a struct {
		B string
		C int
	}

	t.Run("add primitive type", func(t *testing.T) {
		s := NewGSet()
		s.Add(encoding.Int32(5))
		for _, any := range s.value {
			p, err := encoding.UnmarshalPrimitive(any)
			if err != nil {
				t.FailNow()
			}
			i, ok := p.(int32)
			if !ok {
				t.FailNow()
			}
			if i != 5 {
				t.FailNow()
			}
		}
	})

	t.Run("add struct stable", func(t *testing.T) {
		s := NewGSet()
		json, err := encoding.JSON(
			a{
				B: "hupps",
				C: 7,
			})
		if err != nil {
			t.Error(err)
		}
		s.Add(json)
		if s.Size() != 1 {
			t.Fatalf("s.Size %v; want: %v", s.Size(), 1)
		}
		json, err = encoding.JSON(
			a{
				B: "hupps",
				C: 7,
			})
		if err != nil {
			t.Error(err)
		}
		s.Add(json)
		if s.Size() != 1 {
			t.Fatalf("s.Size %v; want: %v", s.Size(), 1)
		}
	})
}

func TestGSetAdditional(t *testing.T) {
	t.Run("apply invalid delta", func(t *testing.T) {
		s := NewGSet()
		if err := s.applyDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Flag{
				Flag: &entity.FlagDelta{
					Value: false,
				},
			},
		}); err == nil {
			t.Fatal("gset applyDelta should err but did not")
		}
	})

}
