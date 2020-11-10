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

func voteRequest(messages ...proto.Message) *crdt.VoteRequest {
	r := &crdt.VoteRequest{
		Actions: make([]*crdt.VoteRequestAction, 0),
	}
	for _, i := range messages {
		switch t := i.(type) {
		case *crdt.Get:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.VoteRequestAction{Action: &crdt.VoteRequestAction_Get{Get: t}})
		case *crdt.Delete:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.VoteRequestAction{Action: &crdt.VoteRequestAction_Delete{Delete: t}})
		case *crdt.VoteVote:
			r.Id = t.Key
			r.Actions = append(r.Actions, &crdt.VoteRequestAction{Action: &crdt.VoteRequestAction_Vote{Vote: t}})
		default:
			panic("no type matched")
		}
	}
	return r
}
