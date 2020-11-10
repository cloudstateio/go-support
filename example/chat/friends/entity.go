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

package friends

import (
	"fmt"

	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

type Entity struct {
	state *crdt.ORSet
}

func (e *Entity) HandleCommand(ctx *crdt.CommandContext, name string, msg proto.Message) (*any.Any, error) {
	switch name {
	case "Add":
		friend, err := encoding.MarshalAny(msg.(*FriendRequest).GetFriend())
		if err != nil {
			return nil, err
		}
		e.state.Add(friend)
	case "Remove":
		friend, err := encoding.MarshalAny(msg.(*FriendRequest).GetFriend())
		if err != nil {
			return nil, err
		}
		e.state.Remove(friend)
	case "GetFriends":
		var list FriendsList
		for _, f := range e.state.Value() {
			var friend Friend
			if err := encoding.UnmarshalAny(f, &friend); err != nil {
				return nil, err
			}
			list.Friends = append(list.Friends, &friend)
		}
		fmt.Printf("getFriends for user: %s, %+v\n", msg.(*FriendRequest).GetFriend().GetUser(), list)
		return encoding.MarshalAny(&list)
	}
	return encoding.Empty, nil
}

func (e *Entity) Default(ctx *crdt.Context) (crdt.CRDT, error) {
	return crdt.NewORSet(), nil
}

func (e *Entity) Set(ctx *crdt.Context, state crdt.CRDT) error {
	switch set := state.(type) {
	case *crdt.ORSet:
		e.state = set
	}
	return fmt.Errorf("unknown type: %v", state)
}
