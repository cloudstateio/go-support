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

	"github.com/cloudstateio/go-support/cloudstate/entity"
)

func TestFlag(t *testing.T) {
	delta := func(value bool) *entity.CrdtDelta {
		return &entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Flag{
				Flag: &entity.FlagDelta{
					Value: value,
				},
			},
		}
	}

	t.Run("should be disabled when instantiated", func(t *testing.T) {
		f := NewFlag()
		if f.Value() {
			t.Fatal("flag should be false but was not")
		}
	})
	t.Run("should reflect a delta update", func(t *testing.T) {
		f := NewFlag()
		if err := f.applyDelta(delta(true)); err != nil {
			t.Fatal(err)
		}
		if !f.Value() {
			t.Fatal("flag should be true but was not")
		}
	})
	t.Run("should generate deltas", func(t *testing.T) {
		f := NewFlag()
		f.Enable()
		if !encDecDelta(f.Delta()).GetFlag().GetValue() {
			t.Fatal("flag delta should be true but was not")
		}
		f.resetDelta()
		if f.HasDelta() {
			t.Fatal("flag should have no delta")
		}
	})
	t.Run("should return its state", func(t *testing.T) {
		f := NewFlag()
		// if encDecState(f.State()).GetFlag().GetValue() {
		// 	t.Fatal("value should be false but was not")
		// }
		f.resetDelta()
		f.Enable()
		if !f.Value() {
			t.Fatal("delta should be true but was not")
		}
		f.resetDelta()
		if f.HasDelta() {
			t.Fatal("flag should have no delta")
		}
	})
}

func TestFlagAdditional(t *testing.T) {
	t.Run("should return correct delta on zero value", func(t *testing.T) {
		f := NewFlag()
		if f.Delta().GetFlag().GetValue() != false {
			t.Fatal("flag delta should be false but was not")
		}
		f.Enable()
		if f.Delta().GetFlag().GetValue() != true {
			t.Fatal("flag delta should be true but was not")
		}
	})
	t.Run("apply invalid delta", func(t *testing.T) {
		f := NewFlag()
		if err := f.applyDelta(&entity.CrdtDelta{
			Delta: &entity.CrdtDelta_Gcounter{
				Gcounter: &entity.GCounterDelta{
					Increment: 11,
				},
			},
		}); err == nil {
			t.Fatal("flag applyDelta should err but did not")
		}
	})
}
