// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.7
// source: proto/go-failure.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PhiAccrualClient is the client API for PhiAccrual service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PhiAccrualClient interface {
	Heartbeat(ctx context.Context, in *Beat, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type phiAccrualClient struct {
	cc grpc.ClientConnInterface
}

func NewPhiAccrualClient(cc grpc.ClientConnInterface) PhiAccrualClient {
	return &phiAccrualClient{cc}
}

func (c *phiAccrualClient) Heartbeat(ctx context.Context, in *Beat, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/failure.PhiAccrual/Heartbeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PhiAccrualServer is the server API for PhiAccrual service.
// All implementations must embed UnimplementedPhiAccrualServer
// for forward compatibility
type PhiAccrualServer interface {
	Heartbeat(context.Context, *Beat) (*emptypb.Empty, error)
	mustEmbedUnimplementedPhiAccrualServer()
}

// UnimplementedPhiAccrualServer must be embedded to have forward compatible implementations.
type UnimplementedPhiAccrualServer struct {
}

func (UnimplementedPhiAccrualServer) Heartbeat(context.Context, *Beat) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedPhiAccrualServer) mustEmbedUnimplementedPhiAccrualServer() {}

// UnsafePhiAccrualServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PhiAccrualServer will
// result in compilation errors.
type UnsafePhiAccrualServer interface {
	mustEmbedUnimplementedPhiAccrualServer()
}

func RegisterPhiAccrualServer(s grpc.ServiceRegistrar, srv PhiAccrualServer) {
	s.RegisterService(&PhiAccrual_ServiceDesc, srv)
}

func _PhiAccrual_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Beat)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PhiAccrualServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/failure.PhiAccrual/Heartbeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PhiAccrualServer).Heartbeat(ctx, req.(*Beat))
	}
	return interceptor(ctx, in, info, handler)
}

// PhiAccrual_ServiceDesc is the grpc.ServiceDesc for PhiAccrual service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PhiAccrual_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "failure.PhiAccrual",
	HandlerType: (*PhiAccrualServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Heartbeat",
			Handler:    _PhiAccrual_Heartbeat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/go-failure.proto",
}