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

package cloudstate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type TestEntity struct {
	Value int64
	EventEmitter
}

func (inc TestEntity) HandleCommand(_ context.Context, command interface{}) (handled bool, reply interface{}, err error) {
	switch cmd := command.(type) {
	case *IncrementByCommand:
		reply, err := inc.IncrementByCommand(nil, cmd)
		return true, reply, err
	case *DecrementByCommand:
		reply, err := inc.DecrementByCommand(nil, cmd)
		return true, reply, err
	}
	return
}

func (inc TestEntity) String() string {
	return proto.CompactTextString(inc)
}

func (inc TestEntity) ProtoMessage() {
}

func (inc TestEntity) Reset() {
}

func (te *TestEntity) Snapshot() (snapshot interface{}, err error) {
	return encoding.MarshalPrimitive(te.Value)
}

func (te *TestEntity) HandleSnapshot(snapshot interface{}) (handled bool, err error) {
	switch v := snapshot.(type) {
	case int64:
		te.Value = v
	}
	return true, nil
}

func (te *TestEntity) IncrementBy(n int64) (int64, error) {
	te.Value += n
	return te.Value, nil
}

func (te *TestEntity) DecrementBy(n int64) (int64, error) {
	te.Value -= n
	return te.Value, nil
}

// initialize value to <0 let us check whether an initCommand works
var testEntity = &TestEntity{
	Value:        -1,
	EventEmitter: NewEmitter(),
}

func resetTestEntity() {
	testEntity = &TestEntity{
		Value:        -1,
		EventEmitter: NewEmitter(),
	}
}

// IncrementByCommand with value receiver
func (te TestEntity) IncrementByCommand(_ context.Context, ibc *IncrementByCommand) (*empty.Empty, error) {
	te.Emit(&IncrementByEvent{
		Value: ibc.Amount,
	})
	return &empty.Empty{}, nil
}

// DecrementByCommand with pointer receiver
func (te *TestEntity) DecrementByCommand(_ context.Context, ibc *DecrementByCommand) (*empty.Empty, error) {
	te.Emit(&DecrementByEvent{
		Value: ibc.Amount,
	})
	return &empty.Empty{}, nil
}

type IncrementByEvent struct {
	Value int64 `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (inc IncrementByEvent) String() string {
	return proto.CompactTextString(inc)
}

func (inc IncrementByEvent) ProtoMessage() {
}

func (inc IncrementByEvent) Reset() {
}

type DecrementByEvent struct {
	Value int64 `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (inc DecrementByEvent) String() string {
	return proto.CompactTextString(inc)
}

func (inc DecrementByEvent) ProtoMessage() {
}

func (inc DecrementByEvent) Reset() {
}

func (te *TestEntity) DecrementByEvent(d *DecrementByEvent) error {
	_, err := te.DecrementBy(d.Value)
	return err
}

func (te *TestEntity) HandleEvent(_ context.Context, event interface{}) (handled bool, err error) {
	switch e := event.(type) {
	case *IncrementByEvent:
		_, err := te.IncrementBy(e.Value)
		return true, err
	case *DecrementByEvent:
		_, err := te.DecrementBy(e.Value)
		return true, err
	default:
		return false, nil
	}
}

type IncrementByCommand struct {
	Amount int64 `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (inc IncrementByCommand) String() string {
	return proto.CompactTextString(inc)
}

func (inc IncrementByCommand) ProtoMessage() {
}

func (inc IncrementByCommand) Reset() {
}

type DecrementByCommand struct {
	Amount int64 `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (inc DecrementByCommand) String() string {
	return proto.CompactTextString(inc)
}

func (inc DecrementByCommand) ProtoMessage() {
}

func (inc DecrementByCommand) Reset() {
}

// TestEventSourcedHandleServer is a grpc.ServerStream mock
type TestEventSourcedHandleServer struct {
	grpc.ServerStream
}

func (t TestEventSourcedHandleServer) Context() context.Context {
	return context.Background()
}

func (t TestEventSourcedHandleServer) Send(out *protocol.EventSourcedStreamOut) error {
	return nil
}
func (t TestEventSourcedHandleServer) Recv() (*protocol.EventSourcedStreamIn, error) {
	return nil, nil
}

func newHandler(t *testing.T) *EventSourcedServer {
	handler := newEventSourcedServer()
	entity := EventSourcedEntity{
		EntityFunc: func() Entity {
			resetTestEntity()
			testEntity.Value = 0
			return testEntity
		},
		ServiceName:   "TestEventSourcedServer-Service",
		SnapshotEvery: 0,
		registerOnce:  sync.Once{},
	}
	err := entity.init()
	if err != nil {
		t.Errorf("%v", err)
	}
	err = handler.registerEntity(&entity)
	if err != nil {
		t.Errorf("%v", err)
	}
	return handler
}

func initHandler(handler *EventSourcedServer, t *testing.T) {
	err := handler.handleInit(&protocol.EventSourcedInit{
		ServiceName: "TestEventSourcedServer-Service",
		EntityId:    "entity-0",
	})
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
}

func marshal(msg proto.Message, t *testing.T) ([]byte, error) {
	cmd, err := proto.Marshal(msg)
	if err != nil {
		t.Errorf("%v", err)
	}
	return cmd, err
}

func TestMain(m *testing.M) {
	proto.RegisterType((*IncrementByEvent)(nil), "IncrementByEvent")
	proto.RegisterType((*DecrementByEvent)(nil), "DecrementByEvent")
	proto.RegisterType((*TestEntity)(nil), "TestEntity")
	proto.RegisterType((*IncrementByCommand)(nil), "IncrementByCommand")
	proto.RegisterType((*DecrementByCommand)(nil), "DecrementByCommand")
	defer resetTestEntity()
	os.Exit(m.Run())
}

func TestErrSend(t *testing.T) {
	err0 := ErrSendFailure
	err1 := fmt.Errorf("on reply: %w", ErrSendFailure)
	if !errors.Is(err1, err0) {
		t.Fatalf("err1 is no err0 but should")
	}
}
func TestSnapshot(t *testing.T) {
	resetTestEntity()
	handler := newHandler(t)
	primitive, err := encoding.MarshalPrimitive(int64(987))
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = handler.handleInit(&protocol.EventSourcedInit{
		ServiceName: "TestEventSourcedServer-Service",
		EntityId:    "entity-0",
		Snapshot: &protocol.EventSourcedSnapshot{
			SnapshotSequence: 0,
			Snapshot:         primitive,
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if testEntity.Value != 987 {
		t.Fatalf("testEntity.Value should be 0 but was not: %+v", testEntity)
	}
}

func TestEventSourcedServerHandlesCommandAndEvents(t *testing.T) {
	resetTestEntity()
	handler := newHandler(t)
	initHandler(handler, t)
	incrementedTo := int64(7)
	incrCmdValue, err := marshal(&IncrementByCommand{Amount: incrementedTo}, t)
	incrCommand := protocol.Command{
		EntityId: "entity-0",
		Id:       1,
		Name:     "IncrementByCommand",
		Payload: &any.Any{
			TypeUrl: "type.googleapis.com/IncrementByCommand",
			Value:   incrCmdValue,
		},
	}
	err = handler.handleCommand(&incrCommand, TestEventSourcedHandleServer{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if testEntity.Value != incrementedTo {
		t.Fatalf("testEntity.Value: (%v) != incrementedTo: (%v)", testEntity.Value, incrementedTo)
	}

	decrCmdValue, err := proto.Marshal(&DecrementByCommand{Amount: incrementedTo})
	if err != nil {
		t.Fatalf("%v", err)
	}
	decrCommand := protocol.Command{
		EntityId: "entity-0",
		Id:       1,
		Name:     "DecrementByCommand",
		Payload: &any.Any{
			TypeUrl: "type.googleapis.com/DecrementByCommand",
			Value:   decrCmdValue,
		},
	}
	err = handler.handleCommand(&decrCommand, TestEventSourcedHandleServer{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if testEntity.Value != 0 {
		t.Fatalf("testEntity.Value != 0")
	}
}
