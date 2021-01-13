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
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
)

type (
	ServiceName string
	EntityID    string
	CommandID   int64
)

func (sn ServiceName) String() string {
	return string(sn)
}

type Server struct {
	// mu protects the map below.
	mu sync.RWMutex
	// entities has descriptions of entities registered by service names
	entities map[ServiceName]*Entity

	entity.UnimplementedValueEntityServer
}

func NewServer() *Server {
	return &Server{
		entities: make(map[ServiceName]*Entity),
	}
}

func (s *Server) Register(e *Entity) error {
	if e.EntityFunc == nil {
		return errors.New("the entity has to define an EntityFunc but did not")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entities[e.ServiceName]; exists {
		return fmt.Errorf("an entity with service name: %q is already registered", e.ServiceName)
	}
	s.entities[e.ServiceName] = e
	return nil
}

func (s *Server) Handle(stream entity.ValueEntity_HandleServer) error {
	init, err := stream.Recv()
	if err != nil {
		return err
	}
	switch m := init.GetMessage().(type) {
	case *entity.ValueEntityStreamIn_Init:
	default:
		return fmt.Errorf("a message was received without having an Init message first: %v", m)
	}

	e, err := s.entityFor(ServiceName(init.GetInit().GetServiceName()))
	if err != nil {
		return err
	}
	id := EntityID(init.GetInit().GetEntityId())
	c := &Context{
		EntityID: id,
		Entity:   e,
		Instance: e.EntityFunc(id),
		ctx:      stream.Context(),
	}

	if state := init.GetInit().GetState().GetValue(); state != nil {
		err = c.Instance.HandleState(c, state)
		if err != nil {
			return err
		}
		c.state = state
	}
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch m := msg.GetMessage().(type) {
		case *entity.ValueEntityStreamIn_Command:
			reply, err := c.runCommand(m.Command)
			if err != nil && !errors.Is(err, protocol.ClientError{}) {
				return err
			}
			c.failure = err
			err = stream.Send(&entity.ValueEntityStreamOut{
				Message: &entity.ValueEntityStreamOut_Reply{
					Reply: c.entityReply(m.Command, reply),
				},
			})
			if err != nil {
				return err
			}
			c.reset()
		case *entity.ValueEntityStreamIn_Init:
			if EntityID(m.Init.EntityId) == c.EntityID {
				return errors.New("duplicate init message for the same entity")
			}
			return fmt.Errorf("duplicate init message for a new entity: %q", m.Init.EntityId)
		}
	}
}

func (s *Server) entityFor(service ServiceName) (*Entity, error) {
	s.mu.RLock()
	e, ok := s.entities[service]
	s.mu.RUnlock()
	if !ok {
		return e, fmt.Errorf("unknown service: %q", service)
	}
	return e, nil
}
