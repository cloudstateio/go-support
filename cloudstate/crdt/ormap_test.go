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

func TestORMap(t *testing.T) {
	t.Run("should have no elements when instantiated", func(t *testing.T) {
		m := NewORMap()
		if got, want := m.Size(), 0; got != want {
			t.Fatalf("got: %v; want: %v", got, want)
		}
		if m.Delta() != nil {
			t.Fatal("m.Delta() is not nil but should")
		}
		m.resetDelta()
		m.Entries()
		// if got, want := len(encDecState(m.State()).GetOrmap().Entries), 0; got != want {
		if got, want := len(m.Entries()), 0; got != want {
			t.Fatalf("got: %v; want: %v", got, want)
		}
	})
	t.Run("should generate an add delta", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		if !m.HasKey(encoding.String("one")) {
			t.Fatal("m has no 'one' key")
		}
		if s := m.Size(); s != 1 {
			t.Fatalf("m.Size(): %v; want: %v", s, 1)
		}
		delta := m.Delta()
		m.resetDelta()
		if l := len(encDecDelta(delta).GetOrmap().GetAdded()); l != 1 {
			t.Fatalf("delta added length: %v; want: %v", l, 1)
		}
		entry := delta.GetOrmap().GetAdded()[0]
		if k := encoding.DecodeString(entry.GetKey()); k != "one" {
			t.Fatalf("key: %v; want: %v", k, "one")
		}
		// if v := entry.GetValue().GetGcounter().GetValue(); v != 0 {
		if v := entry.GetDelta().GetGcounter().GetIncrement(); v != 0 {
			t.Fatalf("GCounter.Value: %v; want: %v", v, 0)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}

		m.Set(encoding.String("two"), NewGCounter())
		counter, err := m.GCounter(encoding.String("two"))
		if err != nil {
			t.Fatal(err)
		}
		counter.Increment(10)
		if s := m.Size(); s != 2 {
			t.Fatalf("m.Size(): %v; want: %v", s, 2)
		}
		delta2 := encDecDelta(m.Delta())
		m.resetDelta()
		if l := len(delta2.GetOrmap().GetAdded()); l != 1 {
			t.Fatalf("delta added length: %v; want: %v", l, 1)
		}
		entry2 := delta2.GetOrmap().GetAdded()[0]
		if k := encoding.DecodeString(entry2.GetKey()); k != "two" {
			t.Fatalf("key: %v; want: %v", k, "two")
		}
		if v := entry2.GetDelta().GetGcounter().GetIncrement(); v != 10 {
			t.Fatalf("GCounter.Value: %v; want: %v", v, 10)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}

		if l := len(delta2.GetOrmap().GetUpdated()); l != 0 {
			t.Fatalf("length of delta: %v; want: %v", l, 0)
		}
	})

	t.Run("should generate a remove delta", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.Set(encoding.String("two"), NewGCounter())
		m.Set(encoding.String("three"), NewGCounter())
		m.Delta()
		m.resetDelta()
		if !m.HasKey(encoding.String("one")) {
			t.Fatalf("map should have key: %v but had not", "one")
		}
		if !m.HasKey(encoding.String("two")) {
			t.Fatalf("map should have key: %v but had not", "two")
		}
		if !m.HasKey(encoding.String("three")) {
			t.Fatalf("map should have key: %v but had not", "three")
		}
		if s := m.Size(); s != 3 {
			t.Fatalf("m.Size(): %v; want: %v", s, 3)
		}
		m.Delete(encoding.String("one"))
		m.Delete(encoding.String("two"))
		if s := m.Size(); s != 1 {
			t.Fatalf("m.Size(): %v; want: %v", s, 1)
		}
		if m.HasKey(encoding.String("one")) {
			t.Fatalf("map should not have key: %v but had", "one")
		}
		if m.HasKey(encoding.String("two")) {
			t.Fatalf("map should not have key: %v but had", "two")
		}
		if !m.HasKey(encoding.String("three")) {
			t.Fatalf("map should have key: %v but had not", "three")
		}
		delta := encDecDelta(m.Delta())
		m.resetDelta()
		if l := len(delta.GetOrmap().GetRemoved()); l != 2 {
			t.Fatalf("length of delta.removed: %v; want: %v", l, 2)
		}
		if !contains(delta.GetOrmap().GetRemoved(), "one", "two") {
			t.Fatalf("delta.removed should contain keys 'one','two' but did not: %v", delta.GetOrmap().GetRemoved())
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
	})
	t.Run("should generate an update delta", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.resetDelta()
		counter, err := m.GCounter(encoding.String("one"))
		if err != nil {
			t.Fatal(err)
		}
		counter.Increment(5)
		delta := encDecDelta(m.Delta())
		m.resetDelta()
		if l := len(delta.GetOrmap().GetUpdated()); l != 1 {
			t.Fatalf("length of delta.updated: %v; want: %v", l, 1)
		}
		entry := delta.GetOrmap().GetUpdated()[0]
		if k := encoding.DecodeString(entry.GetKey()); k != "one" {
			t.Fatalf("key of updated entry was: %v; want: %v", k, "one")
		}
		if i := entry.GetDelta().GetGcounter().GetIncrement(); i != 5 {
			t.Fatalf("increment: %v; want: %v", i, 5)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
	})
	t.Run("should generate a clear delta", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.Set(encoding.String("two"), NewGCounter())
		m.resetDelta()
		m.Clear()
		if s := m.Size(); s != 0 {
			t.Fatalf("m.Size(): %v; want: %v", s, 0)
		}
		delta := encDecDelta(m.Delta())
		m.resetDelta()
		if c := delta.GetOrmap().GetCleared(); !c {
			t.Fatalf("delta cleared: %v; want: %v", c, true)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
	})
	t.Run("should generate a clear delta when everything is removed", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.Set(encoding.String("two"), NewGCounter())
		m.resetDelta()
		m.Delete(encoding.String("one"))
		m.Delete(encoding.String("two"))
		if s := m.Size(); s != 0 {
			t.Fatalf("m.Size(): %v; want: %v", s, 0)
		}
		delta := encDecDelta(m.Delta())
		m.resetDelta()
		if c := delta.GetOrmap().GetCleared(); !c {
			t.Fatalf("delta cleared: %v; want: %v", c, true)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
	})
	t.Run("should not generate a delta when an added element is removed", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.resetDelta()
		m.Set(encoding.String("two"), NewGCounter())
		m.Delete(encoding.String("two"))
		if s := m.Size(); s != 1 {
			t.Fatalf("m.Size(): %v; want: %v", s, 1)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
	})
	t.Run("should generate a delta when a removed element is added", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.Set(encoding.String("two"), NewGCounter())
		m.resetDelta()
		m.Delete(encoding.String("two"))
		m.Set(encoding.String("two"), NewGCounter())
		if s := m.Size(); s != 2 {
			t.Fatalf("m.Size(): %v; want: %v", s, 2)
		}
		delta := encDecDelta(m.Delta())
		m.resetDelta()
		if l := len(delta.GetOrmap().GetRemoved()); l != 1 {
			t.Fatalf("length of delta.removed: %v; want: %v", l, 1)
		}
		if l := len(delta.GetOrmap().GetAdded()); l != 1 {
			t.Fatalf("length of delta.added: %v; want: %v", l, 1)
		}
		if l := len(delta.GetOrmap().GetUpdated()); l != 0 {
			t.Fatalf("length of delta.updated: %v; want: %v", l, 0)
		}
	})
	t.Run("should not generate a delta when a non existing element is removed", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.resetDelta()
		m.Delete(encoding.String("two"))
		if s := m.Size(); s != 1 {
			t.Fatalf("m.Size(): %v; want: %v", s, 1)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
	})
	t.Run("should generate a delta when an already existing element is set", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.resetDelta()
		m.Set(encoding.String("one"), NewGCounter())
		if s := m.Size(); s != 1 {
			t.Fatalf("m.Size(): %v; want: %v", s, 1)
		}
		delta := encDecDelta(m.Delta())
		m.resetDelta()
		if l := len(delta.GetOrmap().GetRemoved()); l != 1 {
			t.Fatalf("length of delta.removed: %v; want: %v", l, 1)
		}
		if l := len(delta.GetOrmap().GetAdded()); l != 1 {
			t.Fatalf("length of delta.added: %v; want: %v", l, 1)
		}
		if l := len(delta.GetOrmap().GetUpdated()); l != 0 {
			t.Fatalf("length of delta.updated: %v; want: %v", l, 0)
		}
	})
	t.Run("clear all other deltas when the set is cleared", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.Set(encoding.String("two"), NewGCounter())
		m.resetDelta()
		counter, err := m.GCounter(encoding.String("two"))
		if err != nil {
			t.Fatal(err)
		}
		counter.Increment(10)
		m.Set(encoding.String("one"), NewGCounter())
		if s := m.Size(); s != 2 {
			t.Fatalf("m.Size(): %v; want: %v", s, 2)
		}
		m.Clear()
		if s := m.Size(); s != 0 {
			t.Fatalf("m.Size(): %v; want: %v", s, 0)
		}
		delta := encDecDelta(m.Delta())
		m.resetDelta()
		if c := delta.GetOrmap().GetCleared(); !c {
			t.Fatalf("ormap cleared: %v; want: %v", c, true)
		}
		if l := len(delta.GetOrmap().GetAdded()); l != 0 {
			t.Fatalf("added len: %v; want: %v", l, 0)
		}
		if l := len(delta.GetOrmap().GetRemoved()); l != 0 {
			t.Fatalf("added len: %v; want: %v", l, 0)
		}
		if l := len(delta.GetOrmap().GetUpdated()); l != 0 {
			t.Fatalf("added len: %v; want: %v", l, 0)
		}
	})
	t.Run("should reflect a delta add", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.resetDelta()
		err := m.applyDelta(encDecDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Ormap{Ormap: &entity.ORMapDelta{
				Added: append(make([]*entity.ORMapEntryDelta, 0), &entity.ORMapEntryDelta{
					Key: encoding.String("two"),
					Delta: &entity.CrdtDelta{
						Delta: &entity.CrdtDelta_Gcounter{
							Gcounter: &entity.GCounterDelta{
								Increment: 4,
							},
						},
					},
				}),
			}},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if s := m.Size(); s != 2 {
			t.Fatalf("m.Size(): %v; want: %v", s, 2)
		}
		if !contains(m.Keys(), "one", "two") {
			t.Fatalf("m.Keys() should include 'one','two' but did not: %v", m.Keys())
		}
		counter, err := m.GCounter(encoding.String("two"))
		if err != nil {
			t.Fatal(err)
		}
		if v := counter.Value(); v != 4 {
			t.Fatalf("counter.Value(): %v; want: %v", v, 4)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
		m.resetDelta()
		// if l := len(encDecState(m.State()).GetOrmap().GetEntries()); l != 2 {
		if l := len(m.Entries()); l != 2 {
			t.Fatalf("state entries len: %v; want: %v", l, 2)
		}
	})
	t.Run("should reflect a delta remove", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.Set(encoding.String("two"), NewGCounter())
		m.resetDelta()
		err := m.applyDelta(encDecDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Ormap{
				Ormap: &entity.ORMapDelta{
					Removed: append(make([]*any.Any, 0), encoding.String("two")),
				}},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if s := m.Size(); s != 1 {
			t.Fatalf("m.Size(): %v; want: %v", s, 1)
		}
		if !contains(m.Keys(), "one") {
			t.Fatalf("m.Keys() should contain 'one' but did not: %v", m.Keys())
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
		m.resetDelta()
		if l := len(m.Entries()); l != 1 {
			t.Fatalf("state entries len: %v; want: %v", l, 1)
		}
	})
	t.Run("should reflect a delta clear", func(t *testing.T) {
		m := NewORMap()
		m.Set(encoding.String("one"), NewGCounter())
		m.Set(encoding.String("two"), NewGCounter())
		m.resetDelta()
		err := m.applyDelta(encDecDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Ormap{
				Ormap: &entity.ORMapDelta{
					Cleared: true,
				}},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if s := m.Size(); s != 0 {
			t.Fatalf("m.Size(): %v; want: %v", s, 0)
		}
		if d := m.Delta(); d != nil {
			t.Fatalf("m.Delta(): %v; want: %v", d, nil)
		}
		m.resetDelta()
		if l := len(m.Entries()); l != 0 {
			t.Fatalf("state entries len: %v; want: %v", l, 0)
		}
	})
	t.Run("should work with protobuf keys", func(t *testing.T) {
		m := NewORMap()
		type c struct {
			Field1 string
		}
		one, err := encoding.Struct(&c{Field1: "one"})
		if err != nil {
			t.Fatal(err)
		}
		m.Set(one, NewGCounter())
		two, err := encoding.Struct(&c{Field1: "two"})
		if err != nil {
			t.Fatal(err)
		}
		m.Set(two, NewGCounter())
		m.resetDelta()
		m.Delete(one)
		if s := m.Size(); s != 1 {
			t.Fatalf("m.Size(): %v; want: %v", s, 1)
		}
		delta := encDecDelta(m.Delta())
		if l := len(delta.GetOrmap().GetRemoved()); l != 1 {
			t.Fatalf("added len: %v; want: %v", l, 1)
		}
		c0 := &c{}
		if err := encoding.DecodeStruct(m.Delta().GetOrmap().GetRemoved()[0], c0); err != nil {
			t.Fatal(err)
		}
		if f1 := c0.Field1; f1 != "one" {
			t.Fatalf("c0.Field1: %v; want: %v", f1, "one")
		}
	})
	t.Run("should work with json types", func(t *testing.T) {
		m := NewORMap()
		bar, err := encoding.Struct(struct{ Foo string }{Foo: "bar"})
		if err != nil {
			t.Fatal(err)
		}
		m.Set(bar, NewGCounter())
		baz, err := encoding.Struct(struct{ Foo string }{Foo: "baz"})
		if err != nil {
			t.Fatal(err)
		}
		m.Set(baz, NewGCounter())
		m.resetDelta()
		m.Delete(bar)
		if s := m.Size(); s != 1 {
			t.Fatalf("m.Size(): %v; want: %v", s, 1)
		}
		delta := encDecDelta(m.Delta())
		if l := len(delta.GetOrmap().GetRemoved()); l != 1 {
			t.Fatalf("added len: %v; want: %v", l, 1)
		}
		c0 := &struct{ Foo string }{}
		if err := encoding.DecodeStruct(m.Delta().GetOrmap().GetRemoved()[0], c0); err != nil {
			t.Fatal(err)
		}
		if f1 := c0.Foo; f1 != "bar" {
			t.Fatalf("c0.Field1: %v; want: %v", f1, "bar")
		}
	})
}

func TestORMapAdditional(t *testing.T) {
	t.Run("should return values", func(t *testing.T) {
		s := NewORMap()
		s.Set(encoding.String("one"), NewFlag())
		s.Set(encoding.String("two"), NewFlag())
		s.Set(encoding.String("three"), NewFlag())
		s.Entries()
	})
	t.Run("apply invalid delta", func(t *testing.T) {
		s := NewORMap()
		if err := s.applyDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Flag{
				Flag: &entity.FlagDelta{
					Value: false,
				},
			},
		}); err == nil {
			t.Fatal("ormap applyDelta should err but did not")
		}
	})
}
