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

// A Flag is a boolean value that starts as false, and can be set to true.
// Once set to true, it cannot be set back to false. A flag is a very simple CRDT,
// the merge function is simply a boolean or over the two flag values being merged.
type Flag struct {
	value bool
	delta bool
}

var _ CRDT = (*Flag)(nil)

func NewFlag() *Flag {
	return &Flag{}
}

func (f Flag) Value() bool {
	return f.value
}

// Enables enables this flag. Once enabled, it can't be disabled.
func (f *Flag) Enable() {
	if !f.value {
		f.value, f.delta = true, true
	}
}

func (f Flag) Delta() *entity.CrdtDelta {
	return &entity.CrdtDelta{
		Delta: &entity.CrdtDelta_Flag{
			Flag: &entity.FlagDelta{
				Value: f.delta,
			},
		},
	}
}

func (f *Flag) HasDelta() bool {
	return f.delta
}

func (f *Flag) resetDelta() {
	f.delta = false
}

func (f *Flag) applyDelta(delta *entity.CrdtDelta) error {
	d := delta.GetFlag()
	if d == nil {
		return fmt.Errorf("unable to apply delta %+v to Flag", delta)
	}
	f.value = f.value || d.Value
	return nil
}
