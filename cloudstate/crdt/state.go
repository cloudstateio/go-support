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

func newFor(delta *entity.CrdtDelta) (CRDT, error) {
	switch t := delta.GetDelta().(type) {
	case *entity.CrdtDelta_Flag:
		return NewFlag(), nil
	case *entity.CrdtDelta_Gcounter:
		return NewGCounter(), nil
	case *entity.CrdtDelta_Gset:
		return NewGSet(), nil
	case *entity.CrdtDelta_Lwwregister:
		return NewLWWRegister(nil), nil
	case *entity.CrdtDelta_Ormap:
		return NewORMap(), nil
	case *entity.CrdtDelta_Orset:
		return NewORSet(), nil
	case *entity.CrdtDelta_Pncounter:
		return NewPNCounter(), nil
	case *entity.CrdtDelta_Vote:
		return NewVote(), nil
	default:
		return nil, fmt.Errorf("no CRDT type matched: %v", t)
	}
}
