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

//
// == Cloudstate TCK model test for event-sourced entities ==
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.21.0
// 	protoc        v3.11.2
// source: eventsourced.proto

package eventsourced

import (
	context "context"
	_ "github.com/cloudstateio/go-support/cloudstate"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

//
// A `Request` message contains any actions that the entity should process.
// Actions must be processed in order. Any actions after a `Fail` may be ignored.
//
type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string           `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Actions []*RequestAction `protobuf:"bytes,2,rep,name=actions,proto3" json:"actions,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventsourced_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_eventsourced_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_eventsourced_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Request) GetActions() []*RequestAction {
	if x != nil {
		return x.Actions
	}
	return nil
}

//
// Each `RequestAction` is one of:
//
// - Emit: emit an event, with a given value.
// - Forward: forward to another service, in place of replying with a Response.
// - Effect: add a side effect to another service to the reply.
// - Fail: fail the current `Process` command.
//
type RequestAction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Action:
	//	*RequestAction_Emit
	//	*RequestAction_Forward
	//	*RequestAction_Effect
	//	*RequestAction_Fail
	Action isRequestAction_Action `protobuf_oneof:"action"`
}

func (x *RequestAction) Reset() {
	*x = RequestAction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventsourced_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestAction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestAction) ProtoMessage() {}

func (x *RequestAction) ProtoReflect() protoreflect.Message {
	mi := &file_eventsourced_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestAction.ProtoReflect.Descriptor instead.
func (*RequestAction) Descriptor() ([]byte, []int) {
	return file_eventsourced_proto_rawDescGZIP(), []int{1}
}

func (m *RequestAction) GetAction() isRequestAction_Action {
	if m != nil {
		return m.Action
	}
	return nil
}

func (x *RequestAction) GetEmit() *Emit {
	if x, ok := x.GetAction().(*RequestAction_Emit); ok {
		return x.Emit
	}
	return nil
}

func (x *RequestAction) GetForward() *Forward {
	if x, ok := x.GetAction().(*RequestAction_Forward); ok {
		return x.Forward
	}
	return nil
}

func (x *RequestAction) GetEffect() *Effect {
	if x, ok := x.GetAction().(*RequestAction_Effect); ok {
		return x.Effect
	}
	return nil
}

func (x *RequestAction) GetFail() *Fail {
	if x, ok := x.GetAction().(*RequestAction_Fail); ok {
		return x.Fail
	}
	return nil
}

type isRequestAction_Action interface {
	isRequestAction_Action()
}

type RequestAction_Emit struct {
	Emit *Emit `protobuf:"bytes,1,opt,name=emit,proto3,oneof"`
}

type RequestAction_Forward struct {
	Forward *Forward `protobuf:"bytes,2,opt,name=forward,proto3,oneof"`
}

type RequestAction_Effect struct {
	Effect *Effect `protobuf:"bytes,3,opt,name=effect,proto3,oneof"`
}

type RequestAction_Fail struct {
	Fail *Fail `protobuf:"bytes,4,opt,name=fail,proto3,oneof"`
}

func (*RequestAction_Emit) isRequestAction_Action() {}

func (*RequestAction_Forward) isRequestAction_Action() {}

func (*RequestAction_Effect) isRequestAction_Action() {}

func (*RequestAction_Fail) isRequestAction_Action() {}

//
// Emit an event, with the event value in a `Persisted` message.
//
type Emit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Emit) Reset() {
	*x = Emit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventsourced_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Emit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Emit) ProtoMessage() {}

func (x *Emit) ProtoReflect() protoreflect.Message {
	mi := &file_eventsourced_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Emit.ProtoReflect.Descriptor instead.
func (*Emit) Descriptor() ([]byte, []int) {
	return file_eventsourced_proto_rawDescGZIP(), []int{2}
}

func (x *Emit) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

//
// Replace the response with a forward to `cloudstate.tck.model.EventSourcedTwo/Call`.
// The payload must be a `Request` message with the given `id`.
//
type Forward struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Forward) Reset() {
	*x = Forward{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventsourced_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Forward) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Forward) ProtoMessage() {}

func (x *Forward) ProtoReflect() protoreflect.Message {
	mi := &file_eventsourced_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Forward.ProtoReflect.Descriptor instead.
func (*Forward) Descriptor() ([]byte, []int) {
	return file_eventsourced_proto_rawDescGZIP(), []int{3}
}

func (x *Forward) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

//
// Add a side effect to the reply, to `cloudstate.tck.model.EventSourcedTwo/Call`.
// The payload must be a `Request` message with the given `id`.
// The side effect should be marked synchronous based on the given `synchronous` value.
//
type Effect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Synchronous bool   `protobuf:"varint,2,opt,name=synchronous,proto3" json:"synchronous,omitempty"`
}

func (x *Effect) Reset() {
	*x = Effect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventsourced_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Effect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Effect) ProtoMessage() {}

func (x *Effect) ProtoReflect() protoreflect.Message {
	mi := &file_eventsourced_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Effect.ProtoReflect.Descriptor instead.
func (*Effect) Descriptor() ([]byte, []int) {
	return file_eventsourced_proto_rawDescGZIP(), []int{4}
}

func (x *Effect) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Effect) GetSynchronous() bool {
	if x != nil {
		return x.Synchronous
	}
	return false
}

//
// Fail the current command with the given description `message`.
//
type Fail struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Fail) Reset() {
	*x = Fail{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventsourced_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Fail) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fail) ProtoMessage() {}

func (x *Fail) ProtoReflect() protoreflect.Message {
	mi := &file_eventsourced_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fail.ProtoReflect.Descriptor instead.
func (*Fail) Descriptor() ([]byte, []int) {
	return file_eventsourced_proto_rawDescGZIP(), []int{5}
}

func (x *Fail) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

//
// The `Response` message for the `Process` must contain the current state (after processing actions).
//
type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventsourced_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_eventsourced_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_eventsourced_proto_rawDescGZIP(), []int{6}
}

func (x *Response) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

//
// The `Persisted` message wraps both snapshot and event values.
//
type Persisted struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Persisted) Reset() {
	*x = Persisted{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventsourced_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Persisted) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Persisted) ProtoMessage() {}

func (x *Persisted) ProtoReflect() protoreflect.Message {
	mi := &file_eventsourced_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Persisted.ProtoReflect.Descriptor instead.
func (*Persisted) Descriptor() ([]byte, []int) {
	return file_eventsourced_proto_rawDescGZIP(), []int{7}
}

func (x *Persisted) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_eventsourced_proto protoreflect.FileDescriptor

var file_eventsourced_proto_rawDesc = []byte{
	0x0a, 0x12, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x1a, 0x1b, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x6b, 0x65,
	0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5e, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x14, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04,
	0x90, 0xb5, 0x18, 0x01, 0x52, 0x02, 0x69, 0x64, 0x12, 0x3d, 0x0a, 0x07, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x07,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0xf0, 0x01, 0x0a, 0x0d, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x30, 0x0a, 0x04, 0x65, 0x6d, 0x69,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x45,
	0x6d, 0x69, 0x74, 0x48, 0x00, 0x52, 0x04, 0x65, 0x6d, 0x69, 0x74, 0x12, 0x39, 0x0a, 0x07, 0x66,
	0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x2e, 0x46, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x48, 0x00, 0x52, 0x07, 0x66,
	0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x12, 0x36, 0x0a, 0x06, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x45, 0x66,
	0x66, 0x65, 0x63, 0x74, 0x48, 0x00, 0x52, 0x06, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x12, 0x30,
	0x0a, 0x04, 0x66, 0x61, 0x69, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x2e, 0x46, 0x61, 0x69, 0x6c, 0x48, 0x00, 0x52, 0x04, 0x66, 0x61, 0x69, 0x6c,
	0x42, 0x08, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x1c, 0x0a, 0x04, 0x45, 0x6d,
	0x69, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x19, 0x0a, 0x07, 0x46, 0x6f, 0x72, 0x77,
	0x61, 0x72, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x22, 0x3a, 0x0a, 0x06, 0x45, 0x66, 0x66, 0x65, 0x63, 0x74, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x20, 0x0a,
	0x0b, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x6f, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0b, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x6f, 0x75, 0x73, 0x22,
	0x20, 0x0a, 0x04, 0x46, 0x61, 0x69, 0x6c, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x22, 0x24, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x21, 0x0a, 0x09, 0x50, 0x65, 0x72, 0x73, 0x69,
	0x73, 0x74, 0x65, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x60, 0x0a, 0x14, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x64, 0x54, 0x63, 0x6b, 0x4d, 0x6f, 0x64,
	0x65, 0x6c, 0x12, 0x48, 0x0a, 0x07, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x12, 0x1d, 0x2e,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x58, 0x0a, 0x0f,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x64, 0x54, 0x77, 0x6f, 0x12,
	0x45, 0x0a, 0x04, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x1d, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x5b, 0x0a, 0x17, 0x69, 0x6f, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x74, 0x63, 0x6b, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x73, 0x74, 0x61, 0x74, 0x65, 0x69, 0x6f, 0x2f, 0x67, 0x6f, 0x2d, 0x73, 0x75,
	0x70, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x74, 0x63, 0x6b, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x64, 0x3b, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_eventsourced_proto_rawDescOnce sync.Once
	file_eventsourced_proto_rawDescData = file_eventsourced_proto_rawDesc
)

func file_eventsourced_proto_rawDescGZIP() []byte {
	file_eventsourced_proto_rawDescOnce.Do(func() {
		file_eventsourced_proto_rawDescData = protoimpl.X.CompressGZIP(file_eventsourced_proto_rawDescData)
	})
	return file_eventsourced_proto_rawDescData
}

var file_eventsourced_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_eventsourced_proto_goTypes = []interface{}{
	(*Request)(nil),       // 0: cloudstate.tck.model.Request
	(*RequestAction)(nil), // 1: cloudstate.tck.model.RequestAction
	(*Emit)(nil),          // 2: cloudstate.tck.model.Emit
	(*Forward)(nil),       // 3: cloudstate.tck.model.Forward
	(*Effect)(nil),        // 4: cloudstate.tck.model.Effect
	(*Fail)(nil),          // 5: cloudstate.tck.model.Fail
	(*Response)(nil),      // 6: cloudstate.tck.model.Response
	(*Persisted)(nil),     // 7: cloudstate.tck.model.Persisted
}
var file_eventsourced_proto_depIdxs = []int32{
	1, // 0: cloudstate.tck.model.Request.actions:type_name -> cloudstate.tck.model.RequestAction
	2, // 1: cloudstate.tck.model.RequestAction.emit:type_name -> cloudstate.tck.model.Emit
	3, // 2: cloudstate.tck.model.RequestAction.forward:type_name -> cloudstate.tck.model.Forward
	4, // 3: cloudstate.tck.model.RequestAction.effect:type_name -> cloudstate.tck.model.Effect
	5, // 4: cloudstate.tck.model.RequestAction.fail:type_name -> cloudstate.tck.model.Fail
	0, // 5: cloudstate.tck.model.EventSourcedTckModel.Process:input_type -> cloudstate.tck.model.Request
	0, // 6: cloudstate.tck.model.EventSourcedTwo.Call:input_type -> cloudstate.tck.model.Request
	6, // 7: cloudstate.tck.model.EventSourcedTckModel.Process:output_type -> cloudstate.tck.model.Response
	6, // 8: cloudstate.tck.model.EventSourcedTwo.Call:output_type -> cloudstate.tck.model.Response
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_eventsourced_proto_init() }
func file_eventsourced_proto_init() {
	if File_eventsourced_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_eventsourced_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventsourced_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestAction); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventsourced_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Emit); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventsourced_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Forward); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventsourced_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Effect); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventsourced_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Fail); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventsourced_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventsourced_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Persisted); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_eventsourced_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*RequestAction_Emit)(nil),
		(*RequestAction_Forward)(nil),
		(*RequestAction_Effect)(nil),
		(*RequestAction_Fail)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_eventsourced_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_eventsourced_proto_goTypes,
		DependencyIndexes: file_eventsourced_proto_depIdxs,
		MessageInfos:      file_eventsourced_proto_msgTypes,
	}.Build()
	File_eventsourced_proto = out.File
	file_eventsourced_proto_rawDesc = nil
	file_eventsourced_proto_goTypes = nil
	file_eventsourced_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// EventSourcedTckModelClient is the client API for EventSourcedTckModel service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type EventSourcedTckModelClient interface {
	Process(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type eventSourcedTckModelClient struct {
	cc grpc.ClientConnInterface
}

func NewEventSourcedTckModelClient(cc grpc.ClientConnInterface) EventSourcedTckModelClient {
	return &eventSourcedTckModelClient{cc}
}

func (c *eventSourcedTckModelClient) Process(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/cloudstate.tck.model.EventSourcedTckModel/Process", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EventSourcedTckModelServer is the server API for EventSourcedTckModel service.
type EventSourcedTckModelServer interface {
	Process(context.Context, *Request) (*Response, error)
}

// UnimplementedEventSourcedTckModelServer can be embedded to have forward compatible implementations.
type UnimplementedEventSourcedTckModelServer struct {
}

func (*UnimplementedEventSourcedTckModelServer) Process(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Process not implemented")
}

func RegisterEventSourcedTckModelServer(s *grpc.Server, srv EventSourcedTckModelServer) {
	s.RegisterService(&_EventSourcedTckModel_serviceDesc, srv)
}

func _EventSourcedTckModel_Process_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventSourcedTckModelServer).Process(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cloudstate.tck.model.EventSourcedTckModel/Process",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EventSourcedTckModelServer).Process(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _EventSourcedTckModel_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cloudstate.tck.model.EventSourcedTckModel",
	HandlerType: (*EventSourcedTckModelServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Process",
			Handler:    _EventSourcedTckModel_Process_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "eventsourced.proto",
}

// EventSourcedTwoClient is the client API for EventSourcedTwo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type EventSourcedTwoClient interface {
	Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type eventSourcedTwoClient struct {
	cc grpc.ClientConnInterface
}

func NewEventSourcedTwoClient(cc grpc.ClientConnInterface) EventSourcedTwoClient {
	return &eventSourcedTwoClient{cc}
}

func (c *eventSourcedTwoClient) Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/cloudstate.tck.model.EventSourcedTwo/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EventSourcedTwoServer is the server API for EventSourcedTwo service.
type EventSourcedTwoServer interface {
	Call(context.Context, *Request) (*Response, error)
}

// UnimplementedEventSourcedTwoServer can be embedded to have forward compatible implementations.
type UnimplementedEventSourcedTwoServer struct {
}

func (*UnimplementedEventSourcedTwoServer) Call(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}

func RegisterEventSourcedTwoServer(s *grpc.Server, srv EventSourcedTwoServer) {
	s.RegisterService(&_EventSourcedTwo_serviceDesc, srv)
}

func _EventSourcedTwo_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventSourcedTwoServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cloudstate.tck.model.EventSourcedTwo/Call",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EventSourcedTwoServer).Call(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _EventSourcedTwo_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cloudstate.tck.model.EventSourcedTwo",
	HandlerType: (*EventSourcedTwoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler:    _EventSourcedTwo_Call_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "eventsourced.proto",
}
