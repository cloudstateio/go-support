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
	"time"

	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
)

// Entity describes an event sourced entity. It is used to be registered as
// an event sourced entity on a CloudState instance.
type Entity struct {
	// ServiceName is the fully qualified name of the service that implements this
	// entities interface.
	// Setting it is mandatory.
	ServiceName ServiceName
	// PersistenceID is used to namespace events in the journal, useful for
	// when you share the same database between multiple entities. It defaults to
	// the simple name for the entity type.
	// It’s good practice to select one explicitly, this means your database
	// isn’t depend on type names in your code.
	// Setting it is mandatory.
	PersistenceID string
	// SnapshotEvery controls how often snapshots are taken,
	// so that the entity doesn't need to be recovered from the whole journal
	// each time it’s loaded. If left unset, it defaults to 100.
	// Setting it to a negative number will result in snapshots never being taken.
	SnapshotEvery int64
	// EntityFunc is a factory method which generates a new Entity.
	EntityFunc func(id EntityID) EntityHandler

	PassivationStrategy protocol.EntityPassivationStrategy
}

type Option func(s *Entity)

func (e *Entity) Options(options ...Option) {
	for _, opt := range options {
		opt(e)
	}
}

func WithPassivationStrategyTimeout(duration time.Duration) Option {
	return func(e *Entity) {
		e.PassivationStrategy = protocol.EntityPassivationStrategy{
			Strategy: &protocol.EntityPassivationStrategy_Timeout{
				Timeout: &protocol.TimeoutPassivationStrategy{
					Timeout: duration.Milliseconds(),
				},
			},
		}
	}
}

type (
	ServiceName string
	EntityID    string
	CommandID   int64
)

func (sn ServiceName) String() string {
	return string(sn)
}

func (id CommandID) Value() int64 {
	return int64(id)
}

// tag::entity-type[]
// An EntityHandler implements methods to handle commands and events.
type EntityHandler interface {
	// HandleCommand is the code that handles a command. It
	// may validate the command using the current state, and
	// may emit events as part of its processing. A command
	// handler must not update the state of the entity directly,
	// only indirectly by emitting events. If a command handler
	// does update the state, then when the entity is passivated
	// (removed from memory), those updates will be lost.
	HandleCommand(ctx *Context, name string, cmd proto.Message) (reply proto.Message, err error)
	// HandleEvent is the only piece of code that is allowed
	// to update the state of the entity. It receives events,
	// and, according to the event, updates the state.
	HandleEvent(ctx *Context, event interface{}) error
}

// end::entity-type[]

// tag::snapshooter[]
// A Snapshooter enables eventsourced snapshots to be taken and as well
// handling snapshots provided.
type Snapshooter interface {
	// Snapshot is a recording of the entire current state of an entity,
	// persisted periodically (eg, every 100 events), as an optimization.
	// With snapshots, when the entity is reloaded from the journal, the
	// entire journal doesn't need to be replayed, just the changes since
	// the last snapshot.
	Snapshot(ctx *Context) (snapshot interface{}, err error)
	// HandleSnapshot is used to apply snapshots provided by the Cloudstate
	// proxy.
	HandleSnapshot(ctx *Context, snapshot interface{}) error
}

// end::snapshooter[]
