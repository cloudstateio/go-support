// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package entity

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// EventSourcedClient is the client API for EventSourced service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EventSourcedClient interface {
	// The stream. One stream will be established per active entity.
	// Once established, the first message sent will be Init, which contains the entity ID, and,
	// if the entity has previously persisted a snapshot, it will contain that snapshot. It will
	// then send zero to many event messages, one for each event previously persisted. The entity
	// is expected to apply these to its state in a deterministic fashion. Once all the events
	// are sent, one to many commands are sent, with new commands being sent as new requests for
	// the entity come in. The entity is expected to reply to each command with exactly one reply
	// message. The entity should reply in order, and any events that the entity requests to be
	// persisted the entity should handle itself, applying them to its own state, as if they had
	// arrived as events when the event stream was being replayed on load.
	Handle(ctx context.Context, opts ...grpc.CallOption) (EventSourced_HandleClient, error)
}

type eventSourcedClient struct {
	cc grpc.ClientConnInterface
}

func NewEventSourcedClient(cc grpc.ClientConnInterface) EventSourcedClient {
	return &eventSourcedClient{cc}
}

func (c *eventSourcedClient) Handle(ctx context.Context, opts ...grpc.CallOption) (EventSourced_HandleClient, error) {
	stream, err := c.cc.NewStream(ctx, &_EventSourced_serviceDesc.Streams[0], "/cloudstate.eventsourced.EventSourced/handle", opts...)
	if err != nil {
		return nil, err
	}
	x := &eventSourcedHandleClient{stream}
	return x, nil
}

type EventSourced_HandleClient interface {
	Send(*EventSourcedStreamIn) error
	Recv() (*EventSourcedStreamOut, error)
	grpc.ClientStream
}

type eventSourcedHandleClient struct {
	grpc.ClientStream
}

func (x *eventSourcedHandleClient) Send(m *EventSourcedStreamIn) error {
	return x.ClientStream.SendMsg(m)
}

func (x *eventSourcedHandleClient) Recv() (*EventSourcedStreamOut, error) {
	m := new(EventSourcedStreamOut)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EventSourcedServer is the server API for EventSourced service.
// All implementations must embed UnimplementedEventSourcedServer
// for forward compatibility
type EventSourcedServer interface {
	// The stream. One stream will be established per active entity.
	// Once established, the first message sent will be Init, which contains the entity ID, and,
	// if the entity has previously persisted a snapshot, it will contain that snapshot. It will
	// then send zero to many event messages, one for each event previously persisted. The entity
	// is expected to apply these to its state in a deterministic fashion. Once all the events
	// are sent, one to many commands are sent, with new commands being sent as new requests for
	// the entity come in. The entity is expected to reply to each command with exactly one reply
	// message. The entity should reply in order, and any events that the entity requests to be
	// persisted the entity should handle itself, applying them to its own state, as if they had
	// arrived as events when the event stream was being replayed on load.
	Handle(EventSourced_HandleServer) error
	mustEmbedUnimplementedEventSourcedServer()
}

// UnimplementedEventSourcedServer must be embedded to have forward compatible implementations.
type UnimplementedEventSourcedServer struct {
}

func (UnimplementedEventSourcedServer) Handle(EventSourced_HandleServer) error {
	return status.Errorf(codes.Unimplemented, "method Handle not implemented")
}
func (UnimplementedEventSourcedServer) mustEmbedUnimplementedEventSourcedServer() {}

// UnsafeEventSourcedServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EventSourcedServer will
// result in compilation errors.
type UnsafeEventSourcedServer interface {
	mustEmbedUnimplementedEventSourcedServer()
}

func RegisterEventSourcedServer(s grpc.ServiceRegistrar, srv EventSourcedServer) {
	s.RegisterService(&_EventSourced_serviceDesc, srv)
}

func _EventSourced_Handle_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EventSourcedServer).Handle(&eventSourcedHandleServer{stream})
}

type EventSourced_HandleServer interface {
	Send(*EventSourcedStreamOut) error
	Recv() (*EventSourcedStreamIn, error)
	grpc.ServerStream
}

type eventSourcedHandleServer struct {
	grpc.ServerStream
}

func (x *eventSourcedHandleServer) Send(m *EventSourcedStreamOut) error {
	return x.ServerStream.SendMsg(m)
}

func (x *eventSourcedHandleServer) Recv() (*EventSourcedStreamIn, error) {
	m := new(EventSourcedStreamIn)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _EventSourced_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cloudstate.eventsourced.EventSourced",
	HandlerType: (*EventSourcedServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "handle",
			Handler:       _EventSourced_Handle_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "event_sourced.proto",
}
