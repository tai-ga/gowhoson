// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: pkg/whoson/sync.proto

package whoson

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

// SyncClient is the client API for Sync service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SyncClient interface {
	Set(ctx context.Context, in *WSRequest, opts ...grpc.CallOption) (*WSResponse, error)
	Del(ctx context.Context, in *WSRequest, opts ...grpc.CallOption) (*WSResponse, error)
	Dump(ctx context.Context, in *WSDumpRequest, opts ...grpc.CallOption) (*WSDumpResponse, error)
}

type syncClient struct {
	cc grpc.ClientConnInterface
}

func NewSyncClient(cc grpc.ClientConnInterface) SyncClient {
	return &syncClient{cc}
}

func (c *syncClient) Set(ctx context.Context, in *WSRequest, opts ...grpc.CallOption) (*WSResponse, error) {
	out := new(WSResponse)
	err := c.cc.Invoke(ctx, "/whoson.sync/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syncClient) Del(ctx context.Context, in *WSRequest, opts ...grpc.CallOption) (*WSResponse, error) {
	out := new(WSResponse)
	err := c.cc.Invoke(ctx, "/whoson.sync/Del", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syncClient) Dump(ctx context.Context, in *WSDumpRequest, opts ...grpc.CallOption) (*WSDumpResponse, error) {
	out := new(WSDumpResponse)
	err := c.cc.Invoke(ctx, "/whoson.sync/Dump", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SyncServer is the server API for Sync service.
// All implementations must embed UnimplementedSyncServer
// for forward compatibility
type SyncServer interface {
	Set(context.Context, *WSRequest) (*WSResponse, error)
	Del(context.Context, *WSRequest) (*WSResponse, error)
	Dump(context.Context, *WSDumpRequest) (*WSDumpResponse, error)
	mustEmbedUnimplementedSyncServer()
}

// UnimplementedSyncServer must be embedded to have forward compatible implementations.
type UnimplementedSyncServer struct {
}

func (UnimplementedSyncServer) Set(context.Context, *WSRequest) (*WSResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (UnimplementedSyncServer) Del(context.Context, *WSRequest) (*WSResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Del not implemented")
}
func (UnimplementedSyncServer) Dump(context.Context, *WSDumpRequest) (*WSDumpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Dump not implemented")
}
func (UnimplementedSyncServer) mustEmbedUnimplementedSyncServer() {}

// UnsafeSyncServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SyncServer will
// result in compilation errors.
type UnsafeSyncServer interface {
	mustEmbedUnimplementedSyncServer()
}

func RegisterSyncServer(s grpc.ServiceRegistrar, srv SyncServer) {
	s.RegisterService(&Sync_ServiceDesc, srv)
}

func _Sync_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WSRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyncServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whoson.sync/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyncServer).Set(ctx, req.(*WSRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Sync_Del_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WSRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyncServer).Del(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whoson.sync/Del",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyncServer).Del(ctx, req.(*WSRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Sync_Dump_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WSDumpRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyncServer).Dump(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whoson.sync/Dump",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyncServer).Dump(ctx, req.(*WSDumpRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Sync_ServiceDesc is the grpc.ServiceDesc for Sync service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Sync_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "whoson.sync",
	HandlerType: (*SyncServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Set",
			Handler:    _Sync_Set_Handler,
		},
		{
			MethodName: "Del",
			Handler:    _Sync_Del_Handler,
		},
		{
			MethodName: "Dump",
			Handler:    _Sync_Dump_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/whoson/sync.proto",
}
