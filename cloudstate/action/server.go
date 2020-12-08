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
	"reflect"
	"strings"
	"sync"

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
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

	// internal marker enforced by go-grpc.
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
	if _, ok := s.entities[e.ServiceName]; ok {
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

// HandleUnary handles an unary command. The input command will contain the
// service name, command name, request metadata and the command payload. The
// reply may contain a direct reply, a forward or a failure, and it may contain
// many side effects.
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
	err = r.runCommand(command)
	if err != nil && !errors.Is(err, protocol.ClientError{}) {
		return nil, err
	}
	if err != nil {
		r.context.failure = err
	}
	return r.actionResponse()
}

// HandleStreamedIn handles a streamed in command. The first message in will
// contain the request metadata, including the service name and command name.
// It will not have an associated payload set. This will be followed by zero to
// many messages in with a payload, but no service name or command name set.
//
// If the underlying transport supports per stream metadata, rather than per
// message metadata, then that metadata will only be included in the metadata
// of the first message. In contrast, if the underlying transport supports per
// message metadata, there will be no metadata on the first message, the
// metadata will instead be found on each subsequent message.
//
// The semantics of stream closure in this protocol map 1:1 with the semantics
// of gRPC stream closure, that is, when the client closes the stream, the
// stream is considered half closed, and the server should eventually, but not
// necessarily immediately, send a response message with a status code and
// trailers.
// If however the server sends a response message before the client closes the
// stream, the stream is completely closed, and the client should handle this
// and stop sending more messages.
//
// Either the client or the server may cancel the stream at any time,
// cancellation is indicated through an HTTP2 stream RST message.
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
		cmd, err := stream.Recv()
		if err == io.EOF {
			// The client closed the stream.
			if r.context.close != nil {
				if err := r.context.close(r.context); err != nil {
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
		err = r.runCommand(cmd)
		if err != nil {
			r.context.failure = err
		}
	}
}

// HandleStreamedOut handles a streamed out command. The input command will
// contain the service name, command name, request metadata and the command
// payload. Zero or more replies may be sent, each containing either a direct
// reply, a forward or a failure, and each may contain many side effects. The
// stream to the client will be closed when the this stream is closed, with the
// same status as this stream is closed with.
//
// Either the client or the server may cancel the stream at any time,
// cancellation is indicated through an HTTP2 stream RST message.
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
	r.context.respondFunc(func(c *Context) error {
		r.response, err = r.actionResponse()
		if err != nil {
			return err
		}
		if err = stream.Send(r.response); err != nil {
			return err
		}
		r.response = nil
		r.context.response = nil
		r.context.forward = nil
		r.context.failure = nil
		r.context.sideEffects = make([]*protocol.SideEffect, 0)
		return nil
	})
	for {
		// No matter what error runCommand returns here, we take it as an error
		// to stop the stream as errors are sent through action.Context.Respond.
		if err = r.runCommand(command); err != nil {
			return err
		}
		if r.context.cancelled {
			return nil
		}
	}
}

// HandleStreamed handles a full duplex streamed command.
//
// The first message in will contain the request metadata, including the
// service name and command name. It will not have an associated payload set.
// This will be followed by zero to many messages in with a payload, but no
// service name or command name set.
//
// Zero or more replies may be sent, each containing either a direct reply, a
// forward or a failure, and each may contain many side effects.
//
// If the underlying transport supports per stream metadata, rather than per
// message metadata, then that metadata will only be included in the metadata
// of the first message. In contrast, if the underlying transport supports per
// message metadata, there will be no metadata on the first message, the
// metadata will instead be found on each subsequent message.
//
// The semantics of stream closure in this protocol map 1:1 with the semantics
// of gRPC stream closure, that is, when the client closes the stream, the
// stream is considered half closed, and the server should eventually, but not
// necessarily immediately, close the stream with a status code and trailers.
//
// If however the server closes the stream with a status code and trailers, the
// stream is immediately considered completely closed, and no further messages
// sent by the client will be handled by the server.
//
// Either the client or the server may cancel the stream at any time,
// cancellation is indicated through an HTTP2 stream RST message.
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
	r.context.respondFunc(func(c *Context) error {
		r.response, err = r.actionResponse()
		if err != nil {
			return err
		}
		if err = stream.Send(r.response); err != nil {
			return err
		}
		r.response = nil
		r.context.failure = nil
		r.context.response = nil
		r.context.forward = nil
		r.context.sideEffects = make([]*protocol.SideEffect, 0)
		return nil
	})
	for {
		cmd, err := stream.Recv()
		if err == io.EOF {
			// The client closed the stream.
			if r.context.close != nil {
				if err := r.context.close(r.context); err != nil {
					r.context.failure = err
				}
			}
			return nil
		}
		if err != nil {
			return err
		}
		cmd.ServiceName = r.context.command.ServiceName
		cmd.Name = r.context.command.Name
		cmd.Metadata = r.context.command.Metadata
		if err = r.runCommand(cmd); err != nil {
			r.context.failure = err
		}
	}
}

type runner struct {
	context  *Context
	response *entity.ActionResponse
}

// runCommand responds with effects, a response, a forward or a
// failure using the action.Context passed to the command handler.
func (r *runner) runCommand(cmd *entity.ActionCommand) error {
	// unmarshal the commands message
	msgName := strings.TrimPrefix(cmd.GetPayload().GetTypeUrl(), "type.googleapis.com/")
	if strings.HasPrefix(msgName, "json.cloudstate.io/") {
		return r.context.Instance.HandleCommand(r.context, cmd.Name, cmd.Payload)
	}
	messageType := proto.MessageType(msgName)
	message, ok := reflect.New(messageType.Elem()).Interface().(proto.Message)
	if !ok {
		return fmt.Errorf("messageType is no proto.Message: %v", messageType)
	}
	if err := proto.Unmarshal(cmd.Payload.Value, message); err != nil {
		return err
	}
	return r.context.Instance.HandleCommand(r.context, cmd.Name, message)
}

// actionResponse returns an action response depending on the runners
// current state.
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
	if r.context.response != nil {
		return &entity.ActionResponse{
			Response: &entity.ActionResponse_Reply{
				Reply: &protocol.Reply{
					Payload:  r.context.response,
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
