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

package synth

import (
	"github.com/cloudstateio/go-support/tck/crdt"
	"github.com/golang/protobuf/proto"
)

func lwwRegisterRequest(messages ...proto.Message) *crdt.LWWRegisterRequest {
	r := &crdt.LWWRegisterRequest{
		Actions: make([]*crdt.LWWRegisterRequestAction, 0),
	}
	for _, i := range messages {
		switch t := i.(type) {
		case *crdt.Get:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.LWWRegisterRequestAction{Action: &crdt.LWWRegisterRequestAction_Get{Get: t}})
		case *crdt.Delete:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.LWWRegisterRequestAction{Action: &crdt.LWWRegisterRequestAction_Delete{Delete: t}})
		case *crdt.LWWRegisterSet:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.LWWRegisterRequestAction{Action: &crdt.LWWRegisterRequestAction_Set{Set: t}})
		case *crdt.LWWRegisterSetWithClock:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.LWWRegisterRequestAction{Action: &crdt.LWWRegisterRequestAction_SetWithClock{SetWithClock: t}})
		default:
			panic("no type matched")
		}
	}
	return r
}
