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

package presence

import (
	"fmt"

	"github.com/cloudstateio/go-support/cloudstate/crdt"
	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
)

type Entity struct {
	state *crdt.Vote
	users int
}

func (p *Entity) HandleCommand(ctx *crdt.CommandContext, name string, msg proto.Message) (*any.Any, error) {
	switch name {
	case "Connect":
		return p.Connect(ctx, msg.(*User))
	case "Monitor":
		return p.Monitor(ctx, msg.(*User))
	}
	return nil, nil
}

func (p *Entity) Connect(ctx *crdt.CommandContext, user *User) (*any.Any, error) {
	if ctx.Streamed() {
		ctx.CancelFunc(func(c *crdt.CommandContext) error {
			p.disconnect()
			return nil
		})
		p.connect()
	}
	return encoding.MarshalAny(&empty.Empty{})
}

func (p *Entity) Monitor(ctx *crdt.CommandContext, user *User) (*any.Any, error) {
	online := p.state.AtLeastOne()
	if ctx.Streamed() {
		ctx.ChangeFunc(func(c *crdt.CommandContext) (*any.Any, error) {
			if online != p.state.AtLeastOne() {
				online = p.state.AtLeastOne()
			}
			fmt.Printf("onStateChange: %s return: {%v}", user.Name, online)
			return encoding.MarshalAny(&OnlineStatus{Online: online})
		})
	}
	fmt.Printf("onStateChange: %s return: {%v}", user.Name, online)
	return encoding.MarshalAny(&OnlineStatus{Online: online})
}

func (p *Entity) connect() {
	p.users += 1
	if p.users == 1 {
		p.state.Vote(true)
	}
}

func (p *Entity) disconnect() {
	p.users -= 1
	if p.users == 0 {
		p.state.Vote(false)
	}
}

func (p *Entity) Default(ctx *crdt.Context) (crdt.CRDT, error) {
	return crdt.NewVote(), nil
}
func (p *Entity) Set(ctx *crdt.Context, state crdt.CRDT) error {
	p.state = state.(*crdt.Vote)
	p.users = 0
	return nil
}
