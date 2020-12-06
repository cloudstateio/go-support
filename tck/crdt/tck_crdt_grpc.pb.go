// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package crdt

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// TckCrdtClient is the client API for TckCrdt service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TckCrdtClient interface {
	ProcessGCounter(ctx context.Context, in *GCounterRequest, opts ...grpc.CallOption) (*GCounterResponse, error)
	ProcessGCounterStreamed(ctx context.Context, in *GCounterRequest, opts ...grpc.CallOption) (TckCrdt_ProcessGCounterStreamedClient, error)
	ProcessPNCounter(ctx context.Context, in *PNCounterRequest, opts ...grpc.CallOption) (*PNCounterResponse, error)
	ProcessGSet(ctx context.Context, in *GSetRequest, opts ...grpc.CallOption) (*GSetResponse, error)
	ProcessORSet(ctx context.Context, in *ORSetRequest, opts ...grpc.CallOption) (*ORSetResponse, error)
	ProcessFlag(ctx context.Context, in *FlagRequest, opts ...grpc.CallOption) (*FlagResponse, error)
	ProcessLWWRegister(ctx context.Context, in *LWWRegisterRequest, opts ...grpc.CallOption) (*LWWRegisterResponse, error)
	ProcessORMap(ctx context.Context, in *ORMapRequest, opts ...grpc.CallOption) (*ORMapResponse, error)
	ProcessVote(ctx context.Context, in *VoteRequest, opts ...grpc.CallOption) (*VoteResponse, error)
}

type tckCrdtClient struct {
	cc grpc.ClientConnInterface
}

func NewTckCrdtClient(cc grpc.ClientConnInterface) TckCrdtClient {
	return &tckCrdtClient{cc}
}

func (c *tckCrdtClient) ProcessGCounter(ctx context.Context, in *GCounterRequest, opts ...grpc.CallOption) (*GCounterResponse, error) {
	out := new(GCounterResponse)
	err := c.cc.Invoke(ctx, "/crdt.TckCrdt/ProcessGCounter", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tckCrdtClient) ProcessGCounterStreamed(ctx context.Context, in *GCounterRequest, opts ...grpc.CallOption) (TckCrdt_ProcessGCounterStreamedClient, error) {
	stream, err := c.cc.NewStream(ctx, &_TckCrdt_serviceDesc.Streams[0], "/crdt.TckCrdt/ProcessGCounterStreamed", opts...)
	if err != nil {
		return nil, err
	}
	x := &tckCrdtProcessGCounterStreamedClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type TckCrdt_ProcessGCounterStreamedClient interface {
	Recv() (*GCounterResponse, error)
	grpc.ClientStream
}

type tckCrdtProcessGCounterStreamedClient struct {
	grpc.ClientStream
}

func (x *tckCrdtProcessGCounterStreamedClient) Recv() (*GCounterResponse, error) {
	m := new(GCounterResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *tckCrdtClient) ProcessPNCounter(ctx context.Context, in *PNCounterRequest, opts ...grpc.CallOption) (*PNCounterResponse, error) {
	out := new(PNCounterResponse)
	err := c.cc.Invoke(ctx, "/crdt.TckCrdt/ProcessPNCounter", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tckCrdtClient) ProcessGSet(ctx context.Context, in *GSetRequest, opts ...grpc.CallOption) (*GSetResponse, error) {
	out := new(GSetResponse)
	err := c.cc.Invoke(ctx, "/crdt.TckCrdt/ProcessGSet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tckCrdtClient) ProcessORSet(ctx context.Context, in *ORSetRequest, opts ...grpc.CallOption) (*ORSetResponse, error) {
	out := new(ORSetResponse)
	err := c.cc.Invoke(ctx, "/crdt.TckCrdt/ProcessORSet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tckCrdtClient) ProcessFlag(ctx context.Context, in *FlagRequest, opts ...grpc.CallOption) (*FlagResponse, error) {
	out := new(FlagResponse)
	err := c.cc.Invoke(ctx, "/crdt.TckCrdt/ProcessFlag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tckCrdtClient) ProcessLWWRegister(ctx context.Context, in *LWWRegisterRequest, opts ...grpc.CallOption) (*LWWRegisterResponse, error) {
	out := new(LWWRegisterResponse)
	err := c.cc.Invoke(ctx, "/crdt.TckCrdt/ProcessLWWRegister", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tckCrdtClient) ProcessORMap(ctx context.Context, in *ORMapRequest, opts ...grpc.CallOption) (*ORMapResponse, error) {
	out := new(ORMapResponse)
	err := c.cc.Invoke(ctx, "/crdt.TckCrdt/ProcessORMap", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tckCrdtClient) ProcessVote(ctx context.Context, in *VoteRequest, opts ...grpc.CallOption) (*VoteResponse, error) {
	out := new(VoteResponse)
	err := c.cc.Invoke(ctx, "/crdt.TckCrdt/ProcessVote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TckCrdtServer is the server API for TckCrdt service.
// All implementations must embed UnimplementedTckCrdtServer
// for forward compatibility
type TckCrdtServer interface {
	ProcessGCounter(context.Context, *GCounterRequest) (*GCounterResponse, error)
	ProcessGCounterStreamed(*GCounterRequest, TckCrdt_ProcessGCounterStreamedServer) error
	ProcessPNCounter(context.Context, *PNCounterRequest) (*PNCounterResponse, error)
	ProcessGSet(context.Context, *GSetRequest) (*GSetResponse, error)
	ProcessORSet(context.Context, *ORSetRequest) (*ORSetResponse, error)
	ProcessFlag(context.Context, *FlagRequest) (*FlagResponse, error)
	ProcessLWWRegister(context.Context, *LWWRegisterRequest) (*LWWRegisterResponse, error)
	ProcessORMap(context.Context, *ORMapRequest) (*ORMapResponse, error)
	ProcessVote(context.Context, *VoteRequest) (*VoteResponse, error)
	mustEmbedUnimplementedTckCrdtServer()
}

// UnimplementedTckCrdtServer must be embedded to have forward compatible implementations.
type UnimplementedTckCrdtServer struct {
}

func (UnimplementedTckCrdtServer) ProcessGCounter(context.Context, *GCounterRequest) (*GCounterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessGCounter not implemented")
}
func (UnimplementedTckCrdtServer) ProcessGCounterStreamed(*GCounterRequest, TckCrdt_ProcessGCounterStreamedServer) error {
	return status.Errorf(codes.Unimplemented, "method ProcessGCounterStreamed not implemented")
}
func (UnimplementedTckCrdtServer) ProcessPNCounter(context.Context, *PNCounterRequest) (*PNCounterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessPNCounter not implemented")
}
func (UnimplementedTckCrdtServer) ProcessGSet(context.Context, *GSetRequest) (*GSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessGSet not implemented")
}
func (UnimplementedTckCrdtServer) ProcessORSet(context.Context, *ORSetRequest) (*ORSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessORSet not implemented")
}
func (UnimplementedTckCrdtServer) ProcessFlag(context.Context, *FlagRequest) (*FlagResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessFlag not implemented")
}
func (UnimplementedTckCrdtServer) ProcessLWWRegister(context.Context, *LWWRegisterRequest) (*LWWRegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessLWWRegister not implemented")
}
func (UnimplementedTckCrdtServer) ProcessORMap(context.Context, *ORMapRequest) (*ORMapResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessORMap not implemented")
}
func (UnimplementedTckCrdtServer) ProcessVote(context.Context, *VoteRequest) (*VoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessVote not implemented")
}
func (UnimplementedTckCrdtServer) mustEmbedUnimplementedTckCrdtServer() {}

// UnsafeTckCrdtServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TckCrdtServer will
// result in compilation errors.
type UnsafeTckCrdtServer interface {
	mustEmbedUnimplementedTckCrdtServer()
}

func RegisterTckCrdtServer(s grpc.ServiceRegistrar, srv TckCrdtServer) {
	s.RegisterService(&_TckCrdt_serviceDesc, srv)
}

func _TckCrdt_ProcessGCounter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GCounterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TckCrdtServer).ProcessGCounter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crdt.TckCrdt/ProcessGCounter",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TckCrdtServer).ProcessGCounter(ctx, req.(*GCounterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TckCrdt_ProcessGCounterStreamed_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GCounterRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TckCrdtServer).ProcessGCounterStreamed(m, &tckCrdtProcessGCounterStreamedServer{stream})
}

type TckCrdt_ProcessGCounterStreamedServer interface {
	Send(*GCounterResponse) error
	grpc.ServerStream
}

type tckCrdtProcessGCounterStreamedServer struct {
	grpc.ServerStream
}

func (x *tckCrdtProcessGCounterStreamedServer) Send(m *GCounterResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _TckCrdt_ProcessPNCounter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PNCounterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TckCrdtServer).ProcessPNCounter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crdt.TckCrdt/ProcessPNCounter",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TckCrdtServer).ProcessPNCounter(ctx, req.(*PNCounterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TckCrdt_ProcessGSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TckCrdtServer).ProcessGSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crdt.TckCrdt/ProcessGSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TckCrdtServer).ProcessGSet(ctx, req.(*GSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TckCrdt_ProcessORSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ORSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TckCrdtServer).ProcessORSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crdt.TckCrdt/ProcessORSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TckCrdtServer).ProcessORSet(ctx, req.(*ORSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TckCrdt_ProcessFlag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FlagRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TckCrdtServer).ProcessFlag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crdt.TckCrdt/ProcessFlag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TckCrdtServer).ProcessFlag(ctx, req.(*FlagRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TckCrdt_ProcessLWWRegister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LWWRegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TckCrdtServer).ProcessLWWRegister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crdt.TckCrdt/ProcessLWWRegister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TckCrdtServer).ProcessLWWRegister(ctx, req.(*LWWRegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TckCrdt_ProcessORMap_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ORMapRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TckCrdtServer).ProcessORMap(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crdt.TckCrdt/ProcessORMap",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TckCrdtServer).ProcessORMap(ctx, req.(*ORMapRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TckCrdt_ProcessVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TckCrdtServer).ProcessVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crdt.TckCrdt/ProcessVote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TckCrdtServer).ProcessVote(ctx, req.(*VoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TckCrdt_serviceDesc = grpc.ServiceDesc{
	ServiceName: "crdt.TckCrdt",
	HandlerType: (*TckCrdtServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessGCounter",
			Handler:    _TckCrdt_ProcessGCounter_Handler,
		},
		{
			MethodName: "ProcessPNCounter",
			Handler:    _TckCrdt_ProcessPNCounter_Handler,
		},
		{
			MethodName: "ProcessGSet",
			Handler:    _TckCrdt_ProcessGSet_Handler,
		},
		{
			MethodName: "ProcessORSet",
			Handler:    _TckCrdt_ProcessORSet_Handler,
		},
		{
			MethodName: "ProcessFlag",
			Handler:    _TckCrdt_ProcessFlag_Handler,
		},
		{
			MethodName: "ProcessLWWRegister",
			Handler:    _TckCrdt_ProcessLWWRegister_Handler,
		},
		{
			MethodName: "ProcessORMap",
			Handler:    _TckCrdt_ProcessORMap_Handler,
		},
		{
			MethodName: "ProcessVote",
			Handler:    _TckCrdt_ProcessVote_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ProcessGCounterStreamed",
			Handler:       _TckCrdt_ProcessGCounterStreamed_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "tck_crdt.proto",
}
