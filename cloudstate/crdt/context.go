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
	"context"
	"errors"
)

// Context holds the context of a running entity.
type Context struct {
	// EntityID is the ID of the entity.
	EntityID EntityID
	// Entity describes the instance that is used as an entity.
	Entity *Entity
	// Instance is the instance of the entity this context is for.
	Instance EntityHandler
	// the root CRDT managed by this user function.
	crdt CRDT
	// ctx is the context.Context from the stream this context is assigned to.
	ctx context.Context
	// streamedCtx are command contexts of streamed commands.
	streamedCtx map[CommandID]*CommandContext
	// created defines if the CRDT was created by the user function.
	created bool
	deleted bool
	// failed holds an internal error occurred during message processing where no error path was possible.
	// user function Emit calls are an example.
	failed error
}

// StreamCtx returns the context.Context from the transport stream this context is assigned to.
func (c *Context) StreamCtx() context.Context {
	return c.ctx
}

// TODO: do we really need that?
func (c *Context) CRDT() CRDT {
	return c.crdt
}

// Delete marks the CRDT to be deleted initiated by the user function.
func (c *Context) Delete() {
	c.deleted = true
}

// fail fails the command context with the given message.
func (c *Context) fail(err error) {
	if c.failed != nil {
		return
	}
	c.failed = err
}

// initDefault initializes the CRDT with a default value if it's not already set.
func (c *Context) initDefault() error {
	// with a handled state, the CRDT might already be set.
	if c.crdt != nil {
		// TODO: the type of c.Instance.Default(c) and c.crdt have to match. should we check that?
		return c.Instance.Set(c, c.crdt)
	}
	// with no state given, the entity instance can provide one.
	var err error
	if c.crdt, err = c.Instance.Default(c); err != nil {
		return err
	}
	if c.crdt == nil {
		return errors.New("no default CRDT set by the entities default method")
	}
	// the entity gets the CRDT to be set.
	if err := c.Instance.Set(c, c.crdt); err != nil {
		return err
	}
	c.created = true
	return nil
}
