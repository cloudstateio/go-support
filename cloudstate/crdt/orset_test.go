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

func TestORSet(t *testing.T) {
	t.Run("should reflect a state update", func(t *testing.T) {
		s := NewORSet()
		err := s.applyState(
			&entity.CrdtState{
				State: &entity.CrdtState_Orset{
					Orset: &entity.ORSetState{
						Items: append(make([]*any.Any, 0), encoding.String("one"), encoding.String("two")),
					},
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if s.Size() != 2 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 2)
		}
	})

	t.Run("should generate an add delta", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		delta := encDecDelta(s.Delta())
		s.resetDelta()
		if alen := len(delta.GetOrset().GetAdded()); alen != 1 {
			t.Fatalf("s.Delta()).GetAdded()): %v; want: %v", alen, 1)
		}
		if !contains(delta.GetOrset().GetAdded(), "one") {
			t.Fatal("did not found one")
		}
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
		s.Add(encoding.String("two"))
		s.Add(encoding.String("three"))
		if s.Size() != 3 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 3)
		}
		if alen := len(encDecDelta(s.Delta()).GetOrset().GetAdded()); alen != 2 {
			t.Fatalf("len(GetAdded()): %v; want: %v", alen, 2)
		}
		if !contains(s.Added(), "two", "three") {
			t.Fatal("did not found two and three")
		}
		s.resetDelta()
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
	})

	t.Run("should generate a remove delta", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.Add(encoding.String("two"))
		s.Add(encoding.String("three"))
		s.resetDelta()
		if !contains(s.Value(), "one", "two", "three") {
			t.Fatalf("removed does not include: one, two, three")
		}
		s.Remove(encoding.String("one"))
		s.Remove(encoding.String("two"))
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		if contains(s.Value(), "one") {
			t.Fatalf("set should not include one")
		}
		if contains(s.Value(), "two") {
			t.Fatalf("set should not include two")
		}
		if !contains(s.Value(), "three") {
			t.Fatalf("set should include three")
		}
		if dlen := len(encDecDelta(s.Delta()).GetOrset().GetRemoved()); dlen != 2 {
			t.Fatalf("len(delta.GetRemoved()): %v; want: %v", dlen, 2)
		}
		if !contains(s.Removed(), "one", "two") {
			t.Fatalf("removed does not include: one, two")
		}
	})

	t.Run("should generate a clear delta", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.Add(encoding.String("two"))
		_ = s.Delta()
		s.resetDelta()
		s.Clear()
		if s.Size() != 0 {
			t.Fatalf("s.Size(): %v; want: %v", len(s.Removed()), 0)
		}
		delta := encDecDelta(s.Delta())
		s.resetDelta()
		if !delta.GetOrset().GetCleared() {
			t.Fail()
		}
		s.Delta()
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
	})

	t.Run("should generate a clear delta when everything is removed", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.Add(encoding.String("two"))
		s.resetDelta()
		s.Remove(encoding.String("one"))
		s.Remove(encoding.String("two"))
		if s.Size() != 0 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 0)
		}
		delta := encDecDelta(s.Delta())
		s.resetDelta()
		if cleared := delta.GetOrset().GetCleared(); !cleared {
			t.Fatalf("delta.Cleared: %v; want: %v", cleared, true)
		}
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
	})

	t.Run("should not generate a delta when an added element is removed", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.Add(encoding.String("two"))
		s.Delta()
		s.resetDelta()
		s.Add(encoding.String("two"))
		s.Remove(encoding.String("two"))
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		// delta := encDecDelta(s.Delta())
		s.resetDelta()
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
	})

	t.Run("should not generate a delta when a removed element is added", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.Add(encoding.String("two"))
		s.Delta()
		s.resetDelta()
		s.Remove(encoding.String("two"))
		s.Add(encoding.String("two"))
		if s.Size() != 2 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 2)
		}
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
	})

	t.Run("should not generate a delta when an already existing element is added", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.resetDelta()
		s.Add(encoding.String("one"))
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
	})

	t.Run("should not generate a delta when a non existing element is removed", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.resetDelta()
		s.Remove(encoding.String("two"))
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
	})

	t.Run("clear all other deltas when the set is cleared", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.resetDelta()
		s.Add(encoding.String("two"))
		s.Remove(encoding.String("one"))
		s.Clear()
		if s.Size() != 0 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 0)
		}
		delta := encDecDelta(s.Delta())
		if cleared := delta.GetOrset().GetCleared(); !cleared {
			t.Fatalf("delta.Cleared: %v; want: %v", cleared, 0)
		}
		if alen := len(delta.GetOrset().GetAdded()); alen != 0 {
			t.Fatalf("len(delta.GetAdded()): %v; want: %v", alen, 0)
		}
		if rlen := len(delta.GetOrset().GetRemoved()); rlen != 0 {
			t.Fatalf("len(delta.GetRemoved): %v; want: %v", rlen, 0)
		}
	})
	t.Run("should reflect a delta add", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.resetDelta()
		if err := s.applyDelta(encDecDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Orset{
				Orset: &entity.ORSetDelta{
					Added: append(make([]*any.Any, 0), encoding.String("two")),
				},
			},
		})); err != nil {
			t.Fatal(err)
		}
		if s.Size() != 2 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 2)
		}
		s.resetDelta()
		if s.HasDelta() {
			t.Fatalf("set has delta")
		}
		stateLen := len(encDecState(s.State()).GetOrset().GetItems())
		if stateLen != 2 {
			t.Fatalf("len(GetItems()): %v; want: %v", stateLen, 2)
		}
	})

	t.Run("should reflect a delta remove", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.Add(encoding.String("two"))
		if err := s.applyDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Orset{
				Orset: &entity.ORSetDelta{
					Removed: append(make([]*any.Any, 0), encoding.String("two")),
				},
			},
		}); err != nil {
			t.Fatal(err)
		}
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}

		if slen := len(encDecState(s.State()).GetOrset().GetItems()); slen != 1 {
			t.Fatalf("len(GetItems()): %v; want: %v", slen, 1)
		}
	})

	t.Run("should reflect a delta clear", func(t *testing.T) {
		s := NewORSet()
		s.Add(encoding.String("one"))
		s.Add(encoding.String("two"))
		s.resetDelta()
		if err := s.applyDelta(encDecDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Orset{
				Orset: &entity.ORSetDelta{
					Cleared: true,
				},
			},
		})); err != nil {
			t.Fatal(err)
		}
		if s.Size() != 0 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 0)
		}
		if slen := len(encDecState(s.State()).GetOrset().GetItems()); slen != 0 {
			t.Fatalf("len(GetItems()): %v; want: %v", slen, 0)
		}
	})

	t.Run("should work with protobuf types", func(t *testing.T) {
		s := NewORSet()
		type Example struct {
			Field1 string
		}
		one, err := encoding.Struct(Example{Field1: "one"})
		if err != nil {
			t.Fatal(err)
		}
		s.Add(one)
		two, err := encoding.Struct(Example{Field1: "two"})
		if err != nil {
			t.Fatal(err)
		}
		s.Add(two)
		s.resetDelta()
		s.Remove(one)
		if s.Size() != 1 {
			t.Fatalf("s.Size(): %v; want: %v", s.Size(), 1)
		}
		delta := encDecDelta(s.Delta())

		if rlen := len(delta.GetOrset().GetRemoved()); rlen != 1 {
			t.Fatalf("rlen: %v; want: %v", rlen, 1)
		}
		e := &Example{}
		if err := encoding.UnmarshalJSON(delta.GetOrset().GetRemoved()[0], e); err != nil || e.Field1 != "one" {
			t.Fail()
		}
	})
}

func TestORSetAdditional(t *testing.T) {
	t.Run("apply invalid delta", func(t *testing.T) {
		s := NewORSet()
		if err := s.applyDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Flag{
				Flag: &entity.FlagDelta{
					Value: false,
				},
			},
		}); err == nil {
			t.Fatal("orset applyDelta should err but did not")
		}
	})
	t.Run("apply invalid state", func(t *testing.T) {
		s := NewORSet()
		if err := s.applyState(&entity.CrdtState{
			State: &entity.CrdtState_Flag{
				Flag: &entity.FlagState{
					Value: false,
				},
			},
		}); err == nil {
			t.Fatal("orset applyState should err but did not")
		}
	})
}
