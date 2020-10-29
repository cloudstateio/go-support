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
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const snapshotEveryDefault = 100

// Server is the implementation of the Server server API for the EventSourced service.
type Server struct {
	// mu protects the map below.
	mu sync.RWMutex
	// entities are indexed by their service name.
	entities map[ServiceName]*Entity
}

// NewServer returns a new event sourced server.
func NewServer() *Server {
	return &Server{
		entities: make(map[ServiceName]*Entity),
	}
}

// Register registers an Entity a an event sourced entity for CloudState.
func (s *Server) Register(entity *Entity) error {
	if entity.EntityFunc == nil {
		return errors.New("the entity has to define an EntityFunc but did not")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entities[entity.ServiceName]; exists {
		return fmt.Errorf("an entity with service name: %s is already registered", entity.ServiceName)
	}
	if entity.SnapshotEvery == 0 {
		entity.SnapshotEvery = snapshotEveryDefault
	}
	s.entities[entity.ServiceName] = entity
	return nil
}

// Handle handles the stream. One stream will be established per active entity.
// Once established, the first message sent will be Init, which contains the entity ID, and,
// if the entity has previously persisted a snapshot, it will contain that snapshot. It will
// then send zero to many event messages, one for each event previously persisted. The entity
// is expected to apply these to its state in a deterministic fashion. Once all the events
// are sent, one to many commands are sent, with new commands being sent as new requests for
// the entity come in. The entity is expected to reply to each command with exactly one reply
// message. The entity should reply in order, and any events that the entity requests to be
// persisted the entity should handle itself, applying them to its own state, as if they had
// arrived as events when the event stream was being replayed on load.
//
// ClientError handling is done so that any error returned, triggers the stream to be closed.
// If an error is a client failure, a ClientAction_Failure is sent with a command id set
// if provided by the error. If an error is a protocol failure or any other error, a
// EventSourcedStreamOut_Failure is sent. A protocol failure might provide a command id to
// be included.
// TODO: rephrase this to the new atomic failure pattern.
func (s *Server) Handle(stream entity.EventSourced_HandleServer) error {
	defer func() {
		if r := recover(); r != nil {
			// on a panic we try to tell the proxy and panic again.
			_ = sendProtocolFailure(fmt.Errorf("Server.Handle panic-ked with: %v", r), stream)
			panic(r)
		}
	}()
	// For any error we get other than codes.Canceled,
	// we send a protocol.Failure and close the stream.
	if err := s.handle(stream); err != nil {
		if status.Code(err) == codes.Canceled {
			return err
		}
		log.Print(err)
		if sendErr := sendProtocolFailure(err, stream); sendErr != nil {
			log.Print(sendErr)
		}
		return status.Error(codes.Aborted, err.Error())
	}
	return nil
}

func (s *Server) handle(stream entity.EventSourced_HandleServer) error {
	first, err := stream.Recv()
	switch err {
	case nil:
		break
	case io.EOF:
		return nil
	default:
		return err
	}
	r := &runner{stream: stream}
	switch m := first.GetMessage().(type) {
	case *entity.EventSourcedStreamIn_Init:
		if err := s.handleInit(m.Init, r); err != nil {
			return err
		}
	default:
		return fmt.Errorf("a message was received without having an EventSourcedInit message handled before: %+v", first.GetMessage())
	}
	for {
		if r.context.failed != nil {
			// failed means deactivated. We may never get this far.
			// context.failed should have been sent as a client reply failure.
			// see: https://github.com/cloudstateio/cloudstate/pull/119#discussion_r444851439
			return fmt.Errorf("failed context was not reported: %w", r.context.failed)
		}
		msg, err := r.stream.Recv()
		switch err {
		case nil:
			break
		case io.EOF:
			return nil
		default:
			return err
		}
		switch m := msg.GetMessage().(type) {
		case *entity.EventSourcedStreamIn_Command:
			err := r.handleCommand(m.Command)
			if err == nil {
				continue
			}
			if _, ok := err.(protocol.ServerError); !ok {
				return protocol.ServerError{
					Failure: &protocol.Failure{CommandId: m.Command.Id},
					Err:     err,
				}
			}
			return err
		case *entity.EventSourcedStreamIn_Event:
			if err := r.handleEvent(m.Event); err != nil {
				return err
			}
		case *entity.EventSourcedStreamIn_Init:
			return errors.New("duplicate init message for the same entity")
		case nil:
			return errors.New("empty message received")
		default:
			return fmt.Errorf("unknown message received: %+v", msg.GetMessage())
		}
	}
}

func (s *Server) handleInit(init *entity.EventSourcedInit, r *runner) error {
	service := ServiceName(init.GetServiceName())
	s.mu.RLock()
	e, exists := s.entities[service]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("received a command for an unknown eventsourced service: %q", service)
	}
	if e.EntityFunc == nil {
		return fmt.Errorf("entity.EntityFunc not defined: %q", service)
	}

	id := EntityID(init.GetEntityId())
	r.context = &Context{
		EntityID:           id,
		EventSourcedEntity: e,
		Instance:           e.EntityFunc(id),
		eventSequence:      0,
		ctx:                r.stream.Context(),
	}
	if snapshot := init.GetSnapshot(); snapshot != nil {
		if err := r.handleInitSnapshot(snapshot); err != nil {
			return err
		}
	}
	return nil
}
