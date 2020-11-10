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
	"context"
	"os"
	"testing"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type TestEntity struct {
	Value int64
}

func (e *TestEntity) HandleCommand(ctx *Context, name string, msg proto.Message) (reply proto.Message, err error) {
	switch cmd := msg.(type) {
	case *IncrementByCommand:
		return e.IncrementByCommand(ctx, cmd)
	case *DecrementByCommand:
		return e.DecrementByCommand(ctx, cmd)
	}
	return
}

func (e TestEntity) String() string {
	return proto.CompactTextString(e)
}

func (e TestEntity) ProtoMessage() {
}

func (e TestEntity) Reset() {
}

func (e *TestEntity) Snapshot(*Context) (snapshot interface{}, err error) {
	return encoding.MarshalPrimitive(e.Value)
}

func (e *TestEntity) HandleSnapshot(_ *Context, snapshot interface{}) error {
	switch v := snapshot.(type) {
	case int64:
		e.Value = v
	}
	return nil
}

func (e *TestEntity) IncrementBy(n int64) (int64, error) {
	e.Value += n
	return e.Value, nil
}

func (e *TestEntity) DecrementBy(n int64) (int64, error) {
	e.Value -= n
	return e.Value, nil
}

// Initialize value to <0 let us check whether an initCommand works.
var testEntity = &TestEntity{
	Value: -1,
}

func resetTestEntity() {
	testEntity = &TestEntity{
		Value: -1,
	}
}

// IncrementByCommand with value receiver.
func (e *TestEntity) IncrementByCommand(ctx *Context, ibc *IncrementByCommand) (*empty.Empty, error) {
	ctx.Emit(&IncrementByEvent{
		Value: ibc.Amount,
	})
	return &empty.Empty{}, nil
}

// DecrementByCommand with pointer receiver.
func (e *TestEntity) DecrementByCommand(ctx *Context, ibc *DecrementByCommand) (*empty.Empty, error) {
	ctx.Emit(&DecrementByEvent{
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

func (e *TestEntity) DecrementByEvent(d *DecrementByEvent) error {
	_, err := e.DecrementBy(d.Value)
	return err
}

func (e *TestEntity) HandleEvent(_ *Context, event interface{}) error {
	switch evt := event.(type) {
	case *IncrementByEvent:
		_, err := e.IncrementBy(evt.Value)
		return err
	case *DecrementByEvent:
		_, err := e.DecrementBy(evt.Value)
		return err
	default:
		return nil
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

// TestEventSourcedHandleServer is a grpc.ServerStream mock.
type TestEventSourcedHandleServer struct {
	grpc.ServerStream
}

func (t TestEventSourcedHandleServer) Context() context.Context {
	return context.Background()
}

func (t TestEventSourcedHandleServer) Send(out *entity.EventSourcedStreamOut) error {
	return nil
}
func (t TestEventSourcedHandleServer) Recv() (*entity.EventSourcedStreamIn, error) {
	return nil, nil
}

func newHandler(t *testing.T) *Server {
	handler := NewServer()
	entity := Entity{
		EntityFunc: func(id EntityID) EntityHandler {
			resetTestEntity()
			testEntity.Value = 0
			return testEntity
		},
		ServiceName:   "TestEventSourcedServer-Service",
		SnapshotEvery: 0,
	}
	if err := handler.Register(&entity); err != nil {
		t.Errorf("%v", err)
	}
	return handler
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

func TestSnapshot(t *testing.T) {
	resetTestEntity()
	handler := newHandler(t)
	primitive, err := encoding.MarshalPrimitive(int64(987))
	if err != nil {
		t.Fatalf("%v", err)
	}
	r := &runner{stream: TestEventSourcedHandleServer{}}
	err = handler.handleInit(&entity.EventSourcedInit{
		ServiceName: "TestEventSourcedServer-Service",
		EntityId:    "entity-0",
		Snapshot: &entity.EventSourcedSnapshot{
			SnapshotSequence: 0,
			Snapshot:         primitive,
		},
	}, r)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if testEntity.Value != 987 {
		t.Fatalf("testEntity.Value should be 0 but was not: %+v", testEntity)
	}
}
