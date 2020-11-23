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

// ORSet, or Observed-Removed Set, is a set that can have items both added
// and removed from it. It is implemented by maintaining a set of unique tags
// for each element which are generated on addition into the set. When an
// element is removed, all the tags that that node currently observes are added
// to the removal set, so as long as there haven’t been any new additions that
// the node hasn’t seen when it removed the element, the element will be removed.
type ORSet struct {
	value   map[uint64]*any.Any
	added   map[uint64]*any.Any
	removed map[uint64]*any.Any
	cleared bool
	*anyHasher
}

var _ CRDT = (*ORSet)(nil)

func NewORSet() *ORSet {
	return &ORSet{
		value:     make(map[uint64]*any.Any),
		added:     make(map[uint64]*any.Any),
		removed:   make(map[uint64]*any.Any),
		cleared:   false,
		anyHasher: &anyHasher{},
	}
}

func (s *ORSet) Size() int {
	return len(s.value)
}

func (s *ORSet) Add(a *any.Any) {
	h := s.hashAny(a)
	if _, ok := s.value[h]; ok {
		return
	}
	if _, ok := s.removed[h]; ok {
		delete(s.removed, h)
	} else {
		s.added[h] = a
	}
	s.value[h] = a
}

func (s *ORSet) Remove(a *any.Any) {
	h := s.hashAny(a)
	if _, ok := s.value[h]; !ok {
		return
	}
	if len(s.value) == 1 {
		s.Clear()
		return
	}
	delete(s.value, h)
	if _, ok := s.added[h]; ok {
		delete(s.added, h)
	} else {
		s.removed[h] = a
	}
}

func (s *ORSet) Clear() {
	s.value = make(map[uint64]*any.Any)
	s.added = make(map[uint64]*any.Any)
	s.removed = make(map[uint64]*any.Any)
	s.cleared = true
}

func (s ORSet) Value() []*any.Any {
	val := make([]*any.Any, len(s.value))
	var i = 0
	for _, v := range s.value {
		val[i] = v
		i++
	}
	return val
}

func (s ORSet) Added() []*any.Any {
	val := make([]*any.Any, len(s.added))
	var i = 0
	for _, v := range s.added {
		val[i] = v
		i++
	}
	return val
}

func (s ORSet) Removed() []*any.Any {
	val := make([]*any.Any, len(s.removed))
	var i = 0
	for _, v := range s.removed {
		val[i] = v
		i++
	}
	return val
}

func (s *ORSet) Delta() *entity.CrdtDelta {
	return &entity.CrdtDelta{
		Delta: &entity.CrdtDelta_Orset{
			Orset: &entity.ORSetDelta{
				Added:   s.Added(),
				Removed: s.Removed(),
				Cleared: s.cleared,
			},
		},
	}
}

func (s *ORSet) HasDelta() bool {
	return s.cleared || len(s.added) > 0 || len(s.removed) > 0
}

func (s *ORSet) resetDelta() {
	s.cleared = false
	s.added = make(map[uint64]*any.Any)
	s.removed = make(map[uint64]*any.Any)
}

func (s *ORSet) applyDelta(delta *entity.CrdtDelta) error {
	d := delta.GetOrset()
	if d == nil {
		return fmt.Errorf("unable to delta %v to ORSet", delta)
	}
	if d.GetCleared() {
		s.value = make(map[uint64]*any.Any)
	}
	for _, r := range d.GetRemoved() {
		delete(s.value, s.hashAny(r))
	}
	for _, a := range d.GetAdded() {
		h := s.hashAny(a)
		if _, ok := s.value[h]; !ok {
			s.value[h] = a
		}
	}
	return nil
}
