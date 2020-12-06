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

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
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

	failure     error
	response    *any.Any
	forward     *protocol.Forward
	sideEffects []*protocol.SideEffect

	// respond is the function to be used for streamed responses.
	respond RespondFunc
	// cancel is the function called when a client closes a stream. With the
	// then half-closed stream, the user function can choose to respond with
	// an action response.
	cancel CancelFunc
	// close is called whenever a client closes a stream.
	close     CloseFunc
	cancelled bool
}

func (c *Context) RespondWith(reply *any.Any) {
	c.failure = nil
	c.response = reply
	c.forward = nil
}

func (c *Context) Forward(forward *protocol.Forward) {
	c.failure = nil
	c.response = nil
	c.forward = forward
}

func (c *Context) SideEffect(effect *protocol.SideEffect) {
	c.sideEffects = append(c.sideEffects, effect)
}

// CloseFunc registers a function that is called whenever a client closes a
// stream.
func (c *Context) CloseFunc(close CloseFunc) {
	c.close = close
}

func (c *Context) CancellationFunc(cancel CancelFunc) {
	c.cancel = cancel
}

// Cancel cancels server command streaming.
func (c *Context) Cancel() {
	c.cancelled = true
}

func (c *Context) Command() *entity.ActionCommand {
	return c.command
}

func (c *Context) Metadata() *protocol.Metadata {
	return c.metadata
}

func (c *Context) Respond(err error) error {
	if c.respond != nil {
		c.failure = err
		return c.respond(c)
	}
	return nil
}

func (c *Context) respondFunc(respond RespondFunc) {
	c.respond = respond
}
