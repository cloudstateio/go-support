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
	"fmt"

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/golang/protobuf/ptypes/any"
)

// ORMap, or Observed-Removed Map, is similar to an ORSet, with the addition
// that the values of the set serve as keys for a map, and the values of the
// map are themselves, CRDTs. When a value for the same key in an ORMap is
// modified concurrently on two different nodes, the values from the two nodes
// are merged together.
type ORMap struct {
	value map[uint64]*orMapValue
	delta orMapDelta
	*anyHasher
}

type orMapValue struct {
	key   *any.Any
	value CRDT
}

var _ CRDT = (*ORMap)(nil)

type orMapDelta struct {
	added   map[uint64]*any.Any
	removed map[uint64]*any.Any
	cleared bool
}

func NewORMap() *ORMap {
	return &ORMap{
		value: make(map[uint64]*orMapValue),
		delta: orMapDelta{
			added:   make(map[uint64]*any.Any),
			removed: make(map[uint64]*any.Any),
			cleared: false,
		},
		anyHasher: &anyHasher{},
	}
}

func (m *ORMap) HasKey(x *any.Any) (hasKey bool) {
	_, hasKey = m.value[m.hashAny(x)]
	return
}

func (m *ORMap) Size() int {
	return len(m.value)
}

func (m *ORMap) Values() []*entity.CrdtState {
	values := make([]*entity.CrdtState, len(m.value))
	var i = 0
	for _, v := range m.value {
		values[i] = v.value.State()
		i++
	}
	return values
}

func (m *ORMap) Keys() []*any.Any {
	keys := make([]*any.Any, len(m.value))
	var i = 0
	for _, v := range m.value {
		keys[i] = v.key
		i++
	}
	return keys
}

func (m *ORMap) Get(key *any.Any) CRDT {
	if s, ok := m.value[m.hashAny(key)]; ok {
		return s.value
	}
	return nil
}

func (m *ORMap) Set(key *any.Any, value CRDT) {
	k := m.hashAny(key)
	// from ref. impl: Setting an existing key to a new value
	// can have unintended effects, as the old value may end
	// up being merged with the new.
	if _, has := m.value[k]; has {
		if _, has := m.delta.added[k]; !has {
			m.delta.removed[k] = key
		}
	}
	m.value[k] = &orMapValue{
		key:   key,
		value: value,
	}
	m.delta.added[k] = key
}

func (m *ORMap) Delete(key *any.Any) {
	k := m.hashAny(key)
	if _, has := m.value[k]; !has {
		return
	}
	if len(m.value) == 1 {
		m.Clear()
		return
	}
	delete(m.value, k)
	if _, has := m.delta.added[k]; has {
		delete(m.delta.added, k)
		return
	}
	m.delta.removed[k] = key
}

func (d *orMapDelta) clear() {
	d.added = make(map[uint64]*any.Any)
	d.removed = make(map[uint64]*any.Any)
	d.cleared = true
}

func (m *ORMap) Clear() {
	if len(m.value) == 0 {
		return
	}
	m.value = make(map[uint64]*orMapValue)
	m.delta.clear()
}

func (m *ORMap) HasDelta() bool {
	if m.delta.cleared || len(m.delta.added) > 0 || len(m.delta.removed) > 0 {
		return true
	}
	for _, v := range m.value {
		if v.value.HasDelta() {
			return true
		}
	}
	return false
}

func (m *ORMap) Delta() *entity.CrdtDelta {
	if !m.HasDelta() {
		return nil
	}
	added := make([]*entity.ORMapEntry, 0)
	updated := make([]*entity.ORMapEntryDelta, 0)
	for _, v := range m.value {
		if _, has := m.delta.added[m.hashAny(v.key)]; has {
			added = append(added, &entity.ORMapEntry{
				Key:   v.key,
				Value: v.value.State(),
			})
		} else if v.value.HasDelta() {
			updated = append(updated, &entity.ORMapEntryDelta{
				Key:   v.key,
				Delta: v.value.Delta(),
			})
		}
	}
	removed := make([]*any.Any, len(m.delta.removed))
	var i = 0
	for _, e := range m.delta.removed {
		removed[i] = e
		i++
	}
	return &entity.CrdtDelta{
		Delta: &entity.CrdtDelta_Ormap{
			Ormap: &entity.ORMapDelta{
				Cleared: m.delta.cleared,
				Removed: removed,
				Updated: updated,
				Added:   added,
			},
		},
	}
}

func (m *ORMap) applyDelta(delta *entity.CrdtDelta) error {
	d := delta.GetOrmap()
	if d == nil {
		return fmt.Errorf("unable to apply delta %v to the ORMap", delta)
	}
	if d.GetCleared() {
		m.value = make(map[uint64]*orMapValue)
	}
	for _, r := range d.GetRemoved() {
		delete(m.value, m.hashAny(r))
	}
	for _, a := range d.Added {
		if m.HasKey(a.GetKey()) {
			continue
		}
		state, err := newFor(a.GetValue())
		if err != nil {
			return err
		}
		if err := state.applyState(a.GetValue()); err != nil {
			return err
		}
		m.value[m.hashAny(a.GetKey())] = &orMapValue{
			key:   a.GetKey(),
			value: state,
		}
	}
	for _, u := range d.Updated {
		if v, has := m.value[m.hashAny(u.GetKey())]; has {
			if err := v.value.applyDelta(u.GetDelta()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *ORMap) resetDelta() {
	for _, v := range m.value {
		v.value.resetDelta()
	}
	m.delta.cleared = false // TODO: what's the thing with cleared to be different to orMapDelta.clear()?
	m.delta.added = make(map[uint64]*any.Any)
	m.delta.removed = make(map[uint64]*any.Any)
}

func (m *ORMap) State() *entity.CrdtState {
	entries := make([]*entity.ORMapEntry, len(m.value))
	var i = 0
	for _, v := range m.value {
		entries[i] = &entity.ORMapEntry{
			Key:   v.key,
			Value: v.value.State(),
		}
		i++
	}
	return &entity.CrdtState{
		State: &entity.CrdtState_Ormap{
			Ormap: &entity.ORMapState{
				Entries: entries,
			},
		},
	}
}

func (m *ORMap) applyState(state *entity.CrdtState) error {
	s := state.GetOrmap()
	if s == nil {
		return fmt.Errorf("unable to apply state %v to the ORMap", state)
	}
	m.value = make(map[uint64]*orMapValue, len(s.GetEntries()))
	for _, entry := range s.GetEntries() {
		value, err := newFor(entry.GetValue())
		if err != nil {
			return err
		}
		v := &orMapValue{
			key:   entry.GetKey(),
			value: value,
		}
		if err := v.value.applyState(entry.GetValue()); err != nil {
			return err
		}
		m.value[m.hashAny(v.key)] = v
	}
	return nil
}
