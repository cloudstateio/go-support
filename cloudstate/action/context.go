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

package action

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

type CloseFunc func(c *Context) error
type CancelFunc func(c *Context) error
type RespondFunc func(c *Context) error

type Context struct {
	// Entity describes the instance that is used as an entity.
	Entity *Entity
	// Instance is the instance of the entity this context is for.
	Instance EntityHandler
	// ctx is the context.Context from the stream this context is assigned to.
	ctx      context.Context
	command  *entity.ActionCommand
	metadata *protocol.Metadata

	respond RespondFunc
	cancel  CancelFunc
	close   CloseFunc

	cancelled   bool
	reply       *any.Any
	forward     *protocol.Forward
	failure     error
	sideEffects []*protocol.SideEffect
}

func (c *Context) RespondWith(reply *any.Any) {
	c.forward = nil
	c.failure = nil
	c.reply = reply
}

func (c *Context) Forward(forward *protocol.Forward) {
	c.forward = forward
	c.failure = nil
	c.reply = nil
}

func (c *Context) SideEffect(effect *protocol.SideEffect) {
	c.sideEffects = append(c.sideEffects, effect)
}

func (c *Context) RespondFunc(respond RespondFunc) {
	c.respond = respond
}

func (c *Context) CloseFunc(close CloseFunc) {
	c.close = close
}

func (c *Context) CancellationFunc(cancel CancelFunc) {
	c.cancel = cancel
}

func (c *Context) Cancel() {
	c.cancelled = true
}

func (c *Context) Command() *entity.ActionCommand {
	return c.command
}

func (c *Context) Metadata() *protocol.Metadata {
	return c.metadata
}

func (c *Context) Response() *any.Any {
	return c.reply
}

func (c *Context) runCommand(cmd *entity.ActionCommand) error {
	// unmarshal the commands message
	msgName := strings.TrimPrefix(cmd.GetPayload().GetTypeUrl(), "type.googleapis.com/")
	messageType := proto.MessageType(msgName)
	message, ok := reflect.New(messageType.Elem()).Interface().(proto.Message)
	if !ok {
		return fmt.Errorf("messageType is no proto.Message: %v", messageType)
	}
	if err := proto.Unmarshal(cmd.Payload.Value, message); err != nil {
		return err
	}
	return c.Instance.HandleCommand(c, cmd.Name, message)
}

func (c *Context) Respond(err error) error {
	if c.respond != nil {
		c.failure = err
		return c.respond(c)
	}
	return nil
}
