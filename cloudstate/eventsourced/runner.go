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
	"net/url"
	"reflect"
	"strings"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

// runner attaches a eventsourced.Context to a stream and runs it.
type runner struct {
	stream  entity.EventSourced_HandleServer
	context *Context
}

// handleCommand handles a command received from the Cloudstate proxy.
func (r *runner) handleCommand(cmd *protocol.Command) error {
	msgName := strings.TrimPrefix(cmd.Payload.GetTypeUrl(), encoding.ProtoAnyBase+"/")
	msgType := proto.MessageType(msgName)
	if msgType.Kind() != reflect.Ptr {
		return fmt.Errorf("msgType: %s is of non Ptr kind", msgType)
	}
	// Get a zero-ed message of this type.
	message, ok := reflect.New(msgType.Elem()).Interface().(proto.Message)
	if !ok {
		return fmt.Errorf("msgType is no proto.Message: %v", msgType)
	}
	// Unmarshal the payload onto the zero-ed message.
	err := proto.Unmarshal(cmd.Payload.Value, message)
	if err != nil {
		return fmt.Errorf("%s, %w", err, encoding.ErrMarshal)
	}
	// The gRPC implementation returns the service method return and an error as a second return value.
	cmdReply, errReturned := r.context.Instance.HandleCommand(r.context, cmd.Name, message)
	// We the take error returned as a client failure except if it's a protocol.ServerError.
	if errReturned != nil {
		// If the error is a ServerError, we return this error and the stream will end.
		if _, ok := errReturned.(protocol.ServerError); ok {
			return errReturned
		}
		r.context.failed = nil
		return r.sendClientActionFailure(&protocol.Failure{
			CommandId:   cmd.Id,
			Description: errReturned.Error(),
			Restart:     len(r.context.events) > 0,
		})
	}
	// The context may have failed.
	if r.context.failed != nil {
		return r.context.failed
	}
	// Get the reply.
	reply, err := encoding.MarshalAny(cmdReply)
	if err != nil { // this should never happen
		return protocol.ServerError{
			Failure: &protocol.Failure{CommandId: cmd.GetId()},
			Err:     fmt.Errorf("marshalling of reply failed: %w", err),
		}
	}
	// Get the events emitted.
	events, err := r.context.marshalEventsAny()
	if err != nil {
		return protocol.ServerError{
			Failure: &protocol.Failure{CommandId: cmd.GetId()},
			Err:     fmt.Errorf("marshalling of events failed: %w", err),
		}
	}
	// Handle the snapshot.
	snapshot, err := r.handleSnapshot()
	if err != nil {
		return protocol.ServerError{
			Failure: &protocol.Failure{CommandId: cmd.GetId()},
			Err:     fmt.Errorf("marshalling of the snapshot failed: %w", err),
		}
	}
	// Spec: It is illegal to send a snapshot without sending any events.
	if snapshot != nil && len(events) == 0 {
		return errors.New("it is illegal to send a snapshot without sending any events")
	}
	if r.context.forward != nil {
		return r.sendEventSourcedReply(&entity.EventSourcedReply{
			CommandId: cmd.GetId(),
			ClientAction: &protocol.ClientAction{
				Action: &protocol.ClientAction_Forward{
					Forward: r.context.forward,
				},
			},
			Events:      events,
			Snapshot:    snapshot,
			SideEffects: r.context.sideEffects,
		})
	}
	return r.sendEventSourcedReply(&entity.EventSourcedReply{
		CommandId: cmd.GetId(),
		ClientAction: &protocol.ClientAction{
			Action: &protocol.ClientAction_Reply{
				Reply: &protocol.Reply{
					Payload: reply,
				},
			},
		},
		Events:      events,
		Snapshot:    snapshot,
		SideEffects: r.context.sideEffects,
	})
}

func (r *runner) handleInitSnapshot(snapshot *entity.EventSourcedSnapshot) error {
	s, err := r.unmarshalSnapshot(snapshot)
	if s == nil || err != nil {
		return fmt.Errorf("handling snapshot failed with: %w", err)
	}
	sh, ok := r.context.Instance.(Snapshooter)
	if !ok {
		return fmt.Errorf("entity instance does not implement eventsourced.Snapshooter")
	}
	err = sh.HandleSnapshot(r.context, s)
	if err != nil {
		return fmt.Errorf("handling snapshot failed with: %w", err)
	}
	r.context.eventSequence = snapshot.SnapshotSequence
	return nil
}

func (r *runner) handleSnapshot() (*any.Any, error) {
	if !r.context.shouldSnapshot {
		return nil, nil
	}
	sh, ok := r.context.Instance.(Snapshooter)
	if !ok {
		return nil, nil
	}
	s, err := sh.Snapshot(r.context)
	if err != nil {
		return nil, fmt.Errorf("getting a snapshot has failed: %w", err)
	}
	// TODO: we expect a proto.Message but can support other format.
	snapshot, err := encoding.MarshalAny(s)
	if err != nil {
		return nil, err
	}
	r.context.resetSnapshotEvery()
	return snapshot, nil
}

func (r *runner) handleEvent(event *entity.EventSourcedEvent) error {
	// TODO: here's the point where events can be protobufs, serialized as json or other formats
	msgName := strings.TrimPrefix(event.Payload.GetTypeUrl(), encoding.ProtoAnyBase+"/")
	msgType := proto.MessageType(msgName)
	if msgType.Kind() != reflect.Ptr {
		return fmt.Errorf("msgType.Kind() is not a pointer type: %v", msgType)
	}
	// Get a zero-ed message of this type.
	message, ok := reflect.New(msgType.Elem()).Interface().(proto.Message)
	if !ok {
		return fmt.Errorf("unable to create a new zero-ed message of type: %v", msgType)
	}
	// Marshal what we got as an any.Any onto it.
	if err := proto.Unmarshal(event.Payload.Value, message); err != nil {
		return fmt.Errorf("%s: %w", err, encoding.ErrMarshal)
	}
	// We're ready to handle the proto message.
	if err := r.context.Instance.HandleEvent(r.context, message); err != nil {
		return err
	}
	r.context.eventSequence = event.Sequence
	return r.context.failed
}

// applyEvent applies an event to a local entity.
func (r *runner) applyEvent(event interface{}) error {
	payload, err := encoding.MarshalAny(event)
	if err != nil {
		return err
	}
	return r.handleEvent(&entity.EventSourcedEvent{Payload: payload})
}

func (*runner) unmarshalSnapshot(snapshot *entity.EventSourcedSnapshot) (interface{}, error) {
	// see: https://developers.google.com/protocol-buffers/docs/reference/csharp/class/google/protobuf/well-known-types/any#typeurl
	typeURL := snapshot.Snapshot.GetTypeUrl()
	if !strings.Contains(typeURL, "://") {
		typeURL = "https://" + typeURL
	}
	parsedURL, err := url.Parse(typeURL)
	if err != nil {
		return nil, err
	}
	switch parsedURL.Host {
	case encoding.PrimitiveTypeURLPrefix:
		return encoding.UnmarshalPrimitive(snapshot.Snapshot)
	case encoding.ProtoAnyBase:
		// TODO: this might be something else than a proto message
		msgName := strings.TrimPrefix(snapshot.Snapshot.GetTypeUrl(), encoding.ProtoAnyBase+"/")
		msgType := proto.MessageType(msgName)
		if msgType.Kind() != reflect.Ptr {
			return nil, err
		}
		message, ok := reflect.New(msgType.Elem()).Interface().(proto.Message)
		if !ok {
			return nil, err
		}
		if err := proto.Unmarshal(snapshot.Snapshot.Value, message); err != nil {
			return nil, err
		}
		return message, nil
	}
	return nil, fmt.Errorf("no snapshot unmarshaller found for: %q", parsedURL.String())
}

func (r *runner) sendEventSourcedReply(reply *entity.EventSourcedReply) error {
	return r.stream.Send(&entity.EventSourcedStreamOut{
		Message: &entity.EventSourcedStreamOut_Reply{
			Reply: reply,
		},
	})
}

func (r *runner) sendClientActionFailure(failure *protocol.Failure) error {
	return r.sendEventSourcedReply(&entity.EventSourcedReply{
		CommandId: failure.CommandId,
		ClientAction: &protocol.ClientAction{
			Action: &protocol.ClientAction_Failure{
				Failure: failure,
			},
		},
	})
}
