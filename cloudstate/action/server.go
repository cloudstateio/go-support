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

// Server is the implementation of the Server server API for the CRDT service.
type Server struct {
	// mu protects the map below.
	mu sync.RWMutex
	// entities has descriptions of entities registered by service names
	entities map[ServiceName]*Entity

	entity.UnimplementedActionProtocolServer
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

func (s *Server) entityFor(service ServiceName) (*Entity, error) {
	s.mu.RLock()
	e, ok := s.entities[service]
	s.mu.RUnlock()
	if !ok {
		return e, fmt.Errorf("unknown service: %q", service)
	}
	return e, nil
}

type runner struct {
	context  *Context
	response *entity.ActionResponse
}

func (s *Server) HandleUnary(ctx context.Context, command *entity.ActionCommand) (*entity.ActionResponse, error) {
	e, err := s.entityFor(ServiceName(command.ServiceName))
	if err != nil {
		return nil, err
	}
	r := runner{context: &Context{
		Entity:      e,
		Instance:    e.EntityFunc(),
		ctx:         ctx,
		command:     command,
		metadata:    command.Metadata,
		sideEffects: make([]*protocol.SideEffect, 0),
	}}
	err = r.context.runCommand(command)
	if err != nil && !errors.Is(err, protocol.ClientError{}) {
		return nil, err
	}
	if err != nil {
		r.context.failure = err
	}
	return r.actionResponse()
}

func (s *Server) HandleStreamedIn(stream entity.ActionProtocol_HandleStreamedInServer) error {
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	e, err := s.entityFor(ServiceName(first.ServiceName))
	if err != nil {
		return err
	}
	r := runner{context: &Context{
		Entity:      e,
		Instance:    e.EntityFunc(),
		ctx:         stream.Context(),
		command:     first,
		metadata:    first.Metadata,
		sideEffects: make([]*protocol.SideEffect, 0),
	}}
	for {
		command, err := stream.Recv()
		if err == io.EOF {
			// The client closed the stream.
			if r.context.close != nil {
				err := r.context.close(r.context)
				if err != nil {
					r.context.failure = err
				}
			}
			r.response, err = r.actionResponse()
			if err != nil {
				return err
			}
			if r.response != nil {
				return stream.SendAndClose(r.response)
			}
			return nil
		}
		if err != nil {
			return err
		}
		err = r.context.runCommand(command)
		if err != nil {
			r.context.failure = err
		}
	}
}

func (s *Server) HandleStreamedOut(command *entity.ActionCommand, stream entity.ActionProtocol_HandleStreamedOutServer) error {
	e, err := s.entityFor(ServiceName(command.ServiceName))
	if err != nil {
		return err
	}
	r := runner{context: &Context{
		Entity:      e,
		Instance:    e.EntityFunc(),
		ctx:         stream.Context(),
		command:     command,
		metadata:    command.Metadata,
		sideEffects: make([]*protocol.SideEffect, 0),
	}}
	r.context.RespondFunc(func(c *Context) error {
		response, err := r.actionResponse()
		if err != nil {
			return err
		}
		if err = stream.Send(response); err != nil {
			return err
		}
		r.response = nil
		r.context.reply = nil
		r.context.forward = nil
		r.context.failure = nil
		r.context.sideEffects = make([]*protocol.SideEffect, 0)
		return nil
	})
	for {
		if err := r.context.runCommand(command); err != nil {
			r.context.failure = err
		}
		if r.context.cancelled {
			return nil
		}
	}
}

func (s *Server) HandleStreamed(stream entity.ActionProtocol_HandleStreamedServer) error {
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	e, err := s.entityFor(ServiceName(first.ServiceName))
	if err != nil {
		return err
	}
	r := runner{context: &Context{
		Entity:      e,
		Instance:    e.EntityFunc(),
		ctx:         stream.Context(),
		command:     first,
		metadata:    first.Metadata,
		sideEffects: make([]*protocol.SideEffect, 0),
	}}
	r.context.RespondFunc(func(c *Context) error {
		response, err := r.actionResponse()
		if err != nil {
			return err
		}
		err = stream.Send(response)
		if err != nil {
			return err
		}
		r.response = nil
		r.context.reply = nil
		r.context.forward = nil
		r.context.failure = nil
		r.context.sideEffects = make([]*protocol.SideEffect, 0)
		return nil
	})
	for {
		command, err := stream.Recv()
		if err == io.EOF {
			// The client closed the stream.
			if r.context.close != nil {
				err := r.context.close(r.context)
				if err != nil {
					r.context.failure = err
				}
			}
			return nil
		}
		if err != nil {
			return err
		}
		command.ServiceName = r.context.command.ServiceName
		command.Name = r.context.command.Name
		command.Metadata = r.context.command.Metadata
		err = r.context.runCommand(command)
		if err != nil {
			r.context.failure = err
		}
	}
}

func (r *runner) actionResponse() (*entity.ActionResponse, error) {
	if r.context.failure != nil {
		return &entity.ActionResponse{
			Response: &entity.ActionResponse_Failure{
				Failure: &protocol.Failure{
					Description: r.context.failure.Error(),
				},
			},
			SideEffects: r.context.sideEffects,
		}, nil
	}
	if r.context.reply != nil {
		return &entity.ActionResponse{
			Response: &entity.ActionResponse_Reply{
				Reply: &protocol.Reply{
					Payload:  r.context.reply,
					Metadata: r.context.command.Metadata,
				},
			},
			SideEffects: r.context.sideEffects,
		}, nil
	}
	if r.context.forward != nil {
		r.context.forward.Metadata = r.context.command.Metadata
		return &entity.ActionResponse{
			Response: &entity.ActionResponse_Forward{
				Forward: r.context.forward,
			},
			SideEffects: r.context.sideEffects,
		}, nil
	}
	if len(r.context.sideEffects) > 0 {
		return &entity.ActionResponse{
			SideEffects: r.context.sideEffects,
		}, nil
	}
	return &entity.ActionResponse{}, nil
}
