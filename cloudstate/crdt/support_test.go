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
	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

func contains(in []*any.Any, all ...string) bool {
	seen := 0
	for _, x := range in {
		dec := encoding.DecodeString(x)
		for _, one := range all {
			if dec == one {
				seen++
			}
			if seen == len(all) {
				return true
			}
		}
	}
	return false
}

func encDecDelta(s *entity.CrdtDelta) *entity.CrdtDelta {
	marshal, err := proto.Marshal(s)
	if err != nil {
		// we panic for convenience in test
		panic(err)
	}
	out := &entity.CrdtDelta{}
	if err := proto.Unmarshal(marshal, out); err != nil {
		panic(err)
	}
	return out
}
