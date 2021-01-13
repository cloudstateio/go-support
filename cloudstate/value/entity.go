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

package value

import (
	"time"

	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

type Entity struct {
	// ServiceName is the fully qualified name of the service that implements
	// this entities interface. Setting it is mandatory.
	ServiceName ServiceName
	// EntityFunc creates a new entity.
	EntityFunc    func(EntityID) EntityHandler
	PersistenceID string

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

type EntityHandler interface {
	HandleCommand(ctx *Context, name string, msg proto.Message) (*any.Any, error)
	HandleState(ctx *Context, state *any.Any) error
}
