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
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/cloudstateio/go-support/cloudstate/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the implementation of the Server server API for the CRDT service.
type Server struct {
	// mu protects the map below.
	mu sync.RWMutex
	// entities has descriptions of entities registered by service names
	entities map[ServiceName]*Entity

	entity.UnimplementedCrdtServer
}

// NewServer returns an initialized Server.
func NewServer() *Server {
	return &Server{
		entities: make(map[ServiceName]*Entity),
	}
}

// CrdtEntities can be registered to a server that handles crdt entities by a ServiceName.
// Whenever a internalCRDT.Server receives an CrdInit for an instance of a crdt entity identified by its
// EntityID and a ServiceName, the internalCRDT.Server handles such entities through their lifecycle.
// The handled entities value are captured by a context that is held fo each of them.
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

// After invoking handle, the first message sent will always be a CrdtInit message,
// containing the entity ID, and, if it exists or is available, the current value of
// the entity. After that, one or more commands may be sent, as well as deltas as
// they arrive, and the entire value if either the entity is created, or the proxy
// wishes the user function to replace its entire value. The user function must
// respond with one reply per command in. They do not necessarily have to be sent
// in the same order that the commands were sent, the command ID is used to correlate
// commands to replies.
func (s *Server) Handle(stream entity.Crdt_HandleServer) error {
	defer func() {
		if r := recover(); r != nil {
			_ = sendFailure(fmt.Errorf("CrdtServer.Handle panic-ked with: %v", r), stream)
			panic(r)
		}
	}()
	for {
		err := s.handle(stream)
		if err == nil {
			continue
		}
		if err == io.EOF {
			return nil
		}
		if status.Code(err) == codes.Canceled {
			return err
		}
		log.Print(err)
		if sendErr := sendFailure(err, stream); sendErr != nil {
			log.Print(sendErr)
		}
		return status.Error(codes.Aborted, err.Error())
	}
}

// handle handles a streams messages to be received.
// io.EOF returned will close the stream gracefully, other errors will be sent
// to the proxy as a failure and a nil error value restarts the stream to be
// reused.
func (s *Server) handle(stream entity.Crdt_HandleServer) error {
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	r := &runner{stream: stream}
	switch m := first.GetMessage().(type) {
	case *entity.CrdtStreamIn_Init:
		// First, always a CrdtInit message must be received.
		if err = s.handleInit(m.Init, r); err != nil {
			return fmt.Errorf("handling of CrdtInit failed with: %w", err)
		}
	default:
		return fmt.Errorf("a message was received without having a CrdtInit message first: %v", m)
	}
	// Handle all other messages after a CrdtInit message has been received.
	for {
		if r.context.deleted {
			// With a context flagged deleted, a CrdtDelete
			// was received or delete state action was sent.
			// Here we return no error and left the stream open.
			return nil
		}
		if r.context.failed != nil {
			// failed means deactivated. We may never get this far.
			return nil
		}
		msg, err := r.stream.Recv()
		if err != nil {
			return err
		}
		switch m := msg.GetMessage().(type) {
		case *entity.CrdtStreamIn_State:
			if err := r.handleState(m.State); err != nil {
				return err
			}
			if err := r.handleChange(); err != nil {
				return err
			}
		case *entity.CrdtStreamIn_Changed:
			if err := r.handleDelta(m.Changed); err != nil {
				return err
			}
			if err := r.handleChange(); err != nil {
				return err
			}
		case *entity.CrdtStreamIn_Deleted:
			// Delete the entity. May be sent at any time. The user function should clear its value when it receives this.
			// A proxy may decide to terminate the stream after sending this.
			r.context.Delete()
		case *entity.CrdtStreamIn_Command:
			// A command, may be sent at any time.
			// The CRDT is allowed to be changed.
			if err := r.handleCommand(m.Command); err != nil {
				return err
			}
		case *entity.CrdtStreamIn_StreamCancelled:
			// The CRDT is allowed to be changed.
			if err := r.handleCancellation(m.StreamCancelled); err != nil {
				return err
			}
		case *entity.CrdtStreamIn_Init:
			if EntityID(m.Init.EntityId) == r.context.EntityID {
				return errors.New("duplicate init message for the same entity")
			}
			return fmt.Errorf("duplicate init message for a new entity: %q", m.Init.EntityId)
		case nil:
			return errors.New("empty message received")
		default:
			return fmt.Errorf("unknown message received: %+v", msg.GetMessage())
		}
	}
}

func (s *Server) handleInit(init *entity.CrdtInit, r *runner) error {
	if init.GetServiceName() == "" || init.GetEntityId() == "" {
		return fmt.Errorf("no service name or entity id was defined for init: %+v", init)
	}
	serviceName := ServiceName(init.GetServiceName())
	s.mu.RLock()
	entity, ok := s.entities[serviceName]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("received CrdtInit for an unknown crdt service: %q", serviceName)
	}
	if entity.EntityFunc == nil {
		return fmt.Errorf("entity.EntityFunc is not defined for service: %q", serviceName)
	}
	id := EntityID(init.GetEntityId())
	r.context = &Context{
		EntityID:    id,
		Entity:      entity,
		Instance:    entity.EntityFunc(id),
		created:     false,
		ctx:         r.stream.Context(), // This context is stable as long as the runner runs.
		streamedCtx: make(map[CommandID]*CommandContext),
	}
	// The init message may have an initial state.
	if state := init.GetState(); state != nil {
		if err := r.handleState(state); err != nil {
			return err
		}
	}
	// The user entity can provide a CRDT through a default function if none is set.
	return r.context.initDefault()
}
