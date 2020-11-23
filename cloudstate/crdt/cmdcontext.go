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
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

var ErrCtxFailCalled = errors.New("context failed")
var ErrStateChanged = errors.New("CRDT change not allowed")

type ChangeFunc func(c *CommandContext) (*any.Any, error)
type CancelFunc func(c *CommandContext) error

// A CommandContext carries change and cancel function handlers and other
// values to handle a command over different phases of a commands lifecycle.
type CommandContext struct {
	*Context
	CommandID   CommandID
	change      ChangeFunc
	cancel      CancelFunc
	cmd         *protocol.Command
	forward     *protocol.Forward
	sideEffects []*protocol.SideEffect
	// ended means, we will send a streamed message where we mark the message
	// as the last one in the stream and therefore, the streamed command has ended.
	ended bool
}

// Command returns the protobuf message the context is handling as a command.
func (c *CommandContext) Command() *protocol.Command {
	return c.cmd
}

// Streamed returns whether the command handled by the context is streamed.
func (c *CommandContext) Streamed() bool {
	if c.cmd == nil {
		return false
	}
	return c.cmd.Streamed
}

// ChangeFunc sets the function to be called whenever the CRDT is changed.
// For non-streamed contexts this is a `no operation`.
func (c *CommandContext) ChangeFunc(f ChangeFunc) {
	if !c.Streamed() {
		return
	}
	c.change = f
}

// CancelFunc registers an on cancel handler for this command.
// The registered function will be invoked if the client initiates a stream cancel.
// It will not be invoked if the entity cancels the stream itself.
// The CancelFunc may update the CRDT, and may emit side effects.
func (c *CommandContext) CancelFunc(f CancelFunc) {
	if !c.Streamed() {
		return
	}
	c.cancel = f
}

// EndStream marks a command stream to be ended.
func (c *CommandContext) EndStream() {
	if !c.Streamed() {
		return
	}
	c.ended = true
}

// Forward forwards this command to another service.
// The protocol.Forward provided has to ensure it references a valid service and command.
func (c *CommandContext) Forward(forward *protocol.Forward) {
	if c.forward != nil {
		c.fail(errors.New("this context has already forwarded"))
	}
	c.forward = forward
}

// SideEffect adds a side effect to being emitted after the current command successfully has completed.
func (c *CommandContext) SideEffect(effect *protocol.SideEffect) {
	c.sideEffects = append(c.sideEffects, effect)
}

func (c *CommandContext) runCommand(cmd *protocol.Command) (*any.Any, error) {
	// unmarshal the commands message
	msgName := strings.TrimPrefix(cmd.GetPayload().GetTypeUrl(), "type.googleapis.com/")
	messageType := proto.MessageType(msgName)
	message, ok := reflect.New(messageType.Elem()).Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("messageType is no proto.Message: %v", messageType)
	}
	if err := proto.Unmarshal(cmd.Payload.Value, message); err != nil {
		return nil, err
	}
	return c.Instance.HandleCommand(c, cmd.Name, message)
}

func (c *CommandContext) clientActionFor(reply *any.Any) (*protocol.ClientAction, error) {
	if c.failed != nil {
		return &protocol.ClientAction{
			Action: &protocol.ClientAction_Failure{
				Failure: &protocol.Failure{
					CommandId:   c.CommandID.Value(),
					Description: c.failed.Error(),
				},
			},
		}, nil
	}
	if reply != nil {
		if c.forward != nil {
			return nil, errors.New("this context has already been forwarded")
		}
		return &protocol.ClientAction{
			Action: &protocol.ClientAction_Reply{
				Reply: &protocol.Reply{
					Payload: reply,
				},
			},
		}, nil
	}
	if c.forward != nil {
		return &protocol.ClientAction{
			Action: &protocol.ClientAction_Forward{
				Forward: c.forward,
			},
		}, nil
	}
	return nil, nil
}

func (c *CommandContext) stateAction() *entity.CrdtStateAction {
	if c.created && c.crdt.HasDelta() {
		c.created = false
		if c.deleted {
			c.crdt = nil
			return nil
		}
		action := &entity.CrdtStateAction{
			Action: &entity.CrdtStateAction_Update{
				Update: &entity.CrdtDelta{
					Delta: c.crdt.Delta().GetDelta(),
				},
			},
		}
		c.crdt.resetDelta()
		return action
	}
	if c.created && c.deleted {
		c.created = false
		c.crdt = nil
		return nil
	}
	if c.deleted {
		c.crdt = nil
		return &entity.CrdtStateAction{
			Action: &entity.CrdtStateAction_Delete{Delete: &entity.CrdtDelete{}},
		}
	}
	if c.crdt.HasDelta() {
		delta := c.crdt.Delta()
		c.crdt.resetDelta()
		return &entity.CrdtStateAction{
			Action: &entity.CrdtStateAction_Update{Update: delta},
		}
	}
	return nil
}

func (c *CommandContext) clearSideEffect() {
	c.sideEffects = make([]*protocol.SideEffect, 0)
}

func (c *CommandContext) changed() (reply *any.Any, err error) {
	reply, err = c.change(c)
	if c.crdt.HasDelta() {
		// the user is not allowed to change the CRDT.
		err = ErrStateChanged
	}
	return
}

func (c *CommandContext) cancelled() error {
	return c.cancel(c)
}

func (c *Context) commandContextFor(cmd *protocol.Command) *CommandContext {
	return &CommandContext{
		Context:     c,
		cmd:         cmd,
		CommandID:   CommandID(cmd.Id),
		sideEffects: make([]*protocol.SideEffect, 0),
	}
}

func (c *CommandContext) trackChanges() {
	c.streamedCtx[c.CommandID] = c
}
