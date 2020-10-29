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

func ormapRequest(messages ...proto.Message) *crdt.ORMapRequest {
	r := &crdt.ORMapRequest{
		Actions: make([]*crdt.ORMapRequestAction, 0),
	}
	for _, i := range messages {
		switch t := i.(type) {
		case *crdt.Get:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.ORMapRequestAction{Action: &crdt.ORMapRequestAction_Get{Get: t}})
		case *crdt.Delete:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.ORMapRequestAction{Action: &crdt.ORMapRequestAction_Delete{Delete: t}})
		case *crdt.ORMapSet:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.ORMapRequestAction{Action: &crdt.ORMapRequestAction_SetKey{SetKey: t}})
		case *crdt.ORMapDelete:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.ORMapRequestAction{Action: &crdt.ORMapRequestAction_DeleteKey{DeleteKey: t}})
		case *crdt.ORMapActionRequest:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.ORMapRequestAction{Action: &crdt.ORMapRequestAction_Request{Request: t}})
		default:
			panic("no type matched")
		}
	}
	return r
}
