// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// LocationClient is the client API for Location service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LocationClient interface {
	Set(ctx context.Context, in *RequestLocation, opts ...grpc.CallOption) (*empty.Empty, error)
	Nearest(ctx context.Context, in *RequestGeo, opts ...grpc.CallOption) (*ResponseGeo, error)
}

type locationClient struct {
	cc grpc.ClientConnInterface
}

func NewLocationClient(cc grpc.ClientConnInterface) LocationClient {
	return &locationClient{cc}
}

func (c *locationClient) Set(ctx context.Context, in *RequestLocation, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/pb.Location/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locationClient) Nearest(ctx context.Context, in *RequestGeo, opts ...grpc.CallOption) (*ResponseGeo, error) {
	out := new(ResponseGeo)
	err := c.cc.Invoke(ctx, "/pb.Location/Nearest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LocationServer is the server API for Location service.
// All implementations must embed UnimplementedLocationServer
// for forward compatibility
type LocationServer interface {
	Set(context.Context, *RequestLocation) (*empty.Empty, error)
	Nearest(context.Context, *RequestGeo) (*ResponseGeo, error)
	mustEmbedUnimplementedLocationServer()
}

// UnimplementedLocationServer must be embedded to have forward compatible implementations.
type UnimplementedLocationServer struct {
}

func (UnimplementedLocationServer) Set(context.Context, *RequestLocation) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (UnimplementedLocationServer) Nearest(context.Context, *RequestGeo) (*ResponseGeo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Nearest not implemented")
}
func (UnimplementedLocationServer) mustEmbedUnimplementedLocationServer() {}

// UnsafeLocationServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LocationServer will
// result in compilation errors.
type UnsafeLocationServer interface {
	mustEmbedUnimplementedLocationServer()
}

func RegisterLocationServer(s grpc.ServiceRegistrar, srv LocationServer) {
	s.RegisterService(&Location_ServiceDesc, srv)
}

func _Location_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestLocation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocationServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Location/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocationServer).Set(ctx, req.(*RequestLocation))
	}
	return interceptor(ctx, in, info, handler)
}

func _Location_Nearest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestGeo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocationServer).Nearest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Location/Nearest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocationServer).Nearest(ctx, req.(*RequestGeo))
	}
	return interceptor(ctx, in, info, handler)
}

// Location_ServiceDesc is the grpc.ServiceDesc for Location service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Location_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Location",
	HandlerType: (*LocationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Set",
			Handler:    _Location_Set_Handler,
		},
		{
			MethodName: "Nearest",
			Handler:    _Location_Nearest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/location.proto",
}