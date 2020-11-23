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

// GCounter, or Grow-only Counter, is a counter that can only be incremented.
// It works by tracking a separate counter value for each node, and taking the
// sum of the values for all the nodes to get the current counter value. Since
// each node only updates its own counter value, each node can coordinate those
// updates to ensure they are consistent. Then the merge function, if it sees
// two different values for the same node, simply takes the highest value,
// because that has to be the most recent value that the node published.
type GCounter struct {
	value uint64
	delta uint64
}

var _ CRDT = (*GCounter)(nil)

func NewGCounter() *GCounter {
	return &GCounter{}
}

func (c *GCounter) Value() uint64 {
	return c.value
}

func (c *GCounter) Increment(i uint64) {
	c.value += i
	c.delta += i
}

func (c GCounter) HasDelta() bool {
	return c.delta > 0
}

func (c *GCounter) Delta() *entity.CrdtDelta {
	if c.delta == 0 {
		return nil
	}
	return &entity.CrdtDelta{
		Delta: &entity.CrdtDelta_Gcounter{
			Gcounter: &entity.GCounterDelta{
				Increment: c.delta,
			},
		},
	}
}

func (c *GCounter) resetDelta() {
	c.delta = 0
}

func (c *GCounter) applyDelta(delta *entity.CrdtDelta) error {
	d := delta.GetGcounter()
	if d == nil {
		return fmt.Errorf("unable to apply delta %v to GCounter", delta)
	}
	c.value += d.GetIncrement()
	return nil
}
