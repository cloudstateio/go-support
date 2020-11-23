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

// GSet, or Grow-only Set, is a set that can only have items added to it.
// A GSet is a very simple CRDT, its merge function is defined by taking
// the union of the two GSets being merged.
type GSet struct {
	value map[uint64]*any.Any
	added map[uint64]*any.Any
	*anyHasher
}

var _ CRDT = (*GSet)(nil)

func NewGSet() *GSet {
	return &GSet{
		value:     make(map[uint64]*any.Any),
		added:     make(map[uint64]*any.Any),
		anyHasher: &anyHasher{},
	}
}

func (s GSet) Size() int {
	return len(s.value)
}

func (s *GSet) Add(a *any.Any) {
	h := s.hashAny(a)
	if _, exists := s.value[h]; exists {
		return
	}
	s.value[h] = a
	s.added[h] = a
}

// func (s GSet) State() *entity.CrdtState {
// 	return &entity.CrdtState{
// 		State: &entity.CrdtState_Gset{
// 			Gset: &entity.GSetState{
// 				Items: s.Value(),
// 			},
// 		},
// 	}
// }

func (s GSet) HasDelta() bool {
	return len(s.added) > 0
}

func (s GSet) Value() []*any.Any {
	val := make([]*any.Any, len(s.value))
	var i = 0
	for _, v := range s.value {
		val[i] = v
		i++
	}
	return val
}

func (s GSet) Added() []*any.Any {
	val := make([]*any.Any, len(s.added))
	var i = 0
	for _, v := range s.added {
		val[i] = v
		i++
	}
	return val
}

func (s GSet) Delta() *entity.CrdtDelta {
	if !s.HasDelta() {
		return nil
	}
	return &entity.CrdtDelta{
		Delta: &entity.CrdtDelta_Gset{
			Gset: &entity.GSetDelta{
				Added: s.Added(),
			},
		},
	}
}

func (s *GSet) resetDelta() {
	s.added = make(map[uint64]*any.Any)
}

func (s *GSet) applyDelta(delta *entity.CrdtDelta) error {
	d := delta.GetGset()
	if d == nil {
		return fmt.Errorf("unable to apply state %+v to GSet", delta)
	}
	for _, v := range d.GetAdded() {
		s.value[s.hashAny(v)] = v
	}
	return nil
}
