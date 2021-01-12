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

package eventsourced

import (
	"context"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/ptypes/any"
)

type Context struct {
	EntityID EntityID
	// EventSourcedEntity describes the instance hold by the EntityInstance.
	EventSourcedEntity *Entity
	// Instance is an instance of the registered entity.
	Instance EntityHandler

	ctx            context.Context
	events         []interface{}
	failed         error
	eventSequence  int64
	shouldSnapshot bool
	forward        *protocol.Forward
	sideEffects    []*protocol.SideEffect
}

// Emit is called by a command handler.
func (c *Context) Emit(event interface{}) {
	if c.failed != nil {
		// We can't fail sooner but won't handle events after one failed anymore.
		return
	}
	if err := c.Instance.HandleEvent(c, event); err != nil {
		c.fail(err)
		return
	}
	c.events = append(c.events, event)
	c.eventSequence++
	c.shouldSnapshot = c.shouldSnapshot || (c.eventSequence%c.EventSourcedEntity.SnapshotEvery == 0)
}

// Effect adds a side effect to be emitted. An effect is something whose
// result has no impact on the result of the current command - if it fails,
// the current command still succeeds. The result of the effect is therefore
// ignored. Effects are only performed after the successful completion of any
// state actions requested by the command handler.
//
// Effects may be declared as synchronous or asynchronous. Asynchronous commands
// run in a "fire and forget" fashion. The code flow of the caller (the command
// handler of the entity which emitted the async command) continues while the
// command is being asynchronously processed. Meanwhile, synchronous commands
// run in "blocking" mode, ie. the commands are processed in order, one at a time.
// The final result of the command handler, either a reply or a forward, is not
// sent until all synchronous commands are completed.
func (c *Context) Effect(effect *protocol.SideEffect) {
	c.sideEffects = append(c.sideEffects, effect)
}

// Forward sets a protocol.Forward to where a command is forwarded to.
//
// An entity may, rather than sending a reply to a command, forward it to another entity.
// This is done by sending a forward message back to the proxy, instructing the proxy which
// call on which entity should be invoked, and passing the message to invoke it with.
//
// The command won’t be forwarded until any state actions requested by the command handler
// have successfully completed. It is the responsibility of the forwarded action to return
// a reply that matches the type of the original command handler. Forwards can be chained
// arbitrarily long.
func (c *Context) Forward(forward *protocol.Forward) {
	c.forward = forward
}

// StreamCtx returns the context.Context for the contexts' current running stream.
func (c *Context) StreamCtx() context.Context {
	return c.ctx
}

func (c *Context) fail(err error) {
	c.failed = err
}

func (c *Context) reset() {
	c.events = nil
	c.failed = nil
	c.forward = nil
	c.sideEffects = nil
}

func (c *Context) resetSnapshotEvery() {
	c.shouldSnapshot = false
}

// marshalEventsAny marshals and the clears events emitted through the context.
func (c *Context) marshalEventsAny() ([]*any.Any, error) {
	events := make([]*any.Any, len(c.events))
	for i, evt := range c.events {
		event, err := encoding.MarshalAny(evt)
		if err != nil {
			return nil, err
		}
		events[i] = event
	}
	c.events = make([]interface{}, 0)
	return events, nil
}
