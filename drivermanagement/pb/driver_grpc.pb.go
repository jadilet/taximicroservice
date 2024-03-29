// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DriverClient is the client API for Driver service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DriverClient interface {
	Send(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type driverClient struct {
	cc grpc.ClientConnInterface
}

func NewDriverClient(cc grpc.ClientConnInterface) DriverClient {
	return &driverClient{cc}
}

func (c *driverClient) Send(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/pb.Driver/Send", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DriverServer is the server API for Driver service.
// All implementations must embed UnimplementedDriverServer
// for forward compatibility
type DriverServer interface {
	Send(context.Context, *Request) (*Response, error)
	mustEmbedUnimplementedDriverServer()
}

// UnimplementedDriverServer must be embedded to have forward compatible implementations.
type UnimplementedDriverServer struct {
}

func (UnimplementedDriverServer) Send(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedDriverServer) mustEmbedUnimplementedDriverServer() {}

// UnsafeDriverServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DriverServer will
// result in compilation errors.
type UnsafeDriverServer interface {
	mustEmbedUnimplementedDriverServer()
}

func RegisterDriverServer(s grpc.ServiceRegistrar, srv DriverServer) {
	s.RegisterService(&Driver_ServiceDesc, srv)
}

func _Driver_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DriverServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Driver/Send",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DriverServer).Send(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

// Driver_ServiceDesc is the grpc.ServiceDesc for Driver service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Driver_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Driver",
	HandlerType: (*DriverServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _Driver_Send_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/driver.proto",
}
