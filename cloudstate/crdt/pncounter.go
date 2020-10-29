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
)

// PNCounter, or Positive-Negative Counter, is a counter that can both be incremented
// and decremented. It works by combining two GCounters, a positive one, that tracks
// increments, and a negative one, that tracks decrements. The final counter value is
// computed by subtracting the negative GCounter from the positive GCounter.
type PNCounter struct {
	value int64
	delta int64
}

var _ CRDT = (*PNCounter)(nil)

func NewPNCounter() *PNCounter {
	return &PNCounter{}
}

func (c *PNCounter) Value() int64 {
	return c.value
}

func (c *PNCounter) Increment(i int64) {
	c.value += i
	c.delta += i
}

func (c *PNCounter) Decrement(d int64) {
	c.value -= d
	c.delta -= d
}

func (c *PNCounter) State() *entity.CrdtState {
	return &entity.CrdtState{
		State: &entity.CrdtState_Pncounter{
			Pncounter: &entity.PNCounterState{
				Value: c.value,
			},
		},
	}
}

func (c *PNCounter) HasDelta() bool {
	return c.delta != 0
}

func (c *PNCounter) Delta() *entity.CrdtDelta {
	if c.delta == 0 {
		return nil
	}
	return &entity.CrdtDelta{
		Delta: &entity.CrdtDelta_Pncounter{
			Pncounter: &entity.PNCounterDelta{
				Change: c.delta,
			},
		},
	}
}

func (c *PNCounter) resetDelta() {
	c.delta = 0
}

func (c *PNCounter) applyState(state *entity.CrdtState) error {
	s := state.GetPncounter()
	if s == nil {
		return fmt.Errorf("unable to apply state %v to PNCounter", state)
	}
	c.value = s.GetValue()
	return nil
}

func (c *PNCounter) applyDelta(delta *entity.CrdtDelta) error {
	d := delta.GetPncounter()
	if d == nil {
		return fmt.Errorf("unable to apply delta %v to PNCounter", delta)
	}
	c.value += d.GetChange()
	return nil
}
