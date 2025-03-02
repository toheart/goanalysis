// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: analysis/v1/analysis.proto

package v1

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

const (
	Analysis_GetAnalysis_FullMethodName           = "/analysis.v1.Analysis/GetAnalysis"
	Analysis_GetAnalysisByGID_FullMethodName      = "/analysis.v1.Analysis/GetAnalysisByGID"
	Analysis_GetAllGIDs_FullMethodName            = "/analysis.v1.Analysis/GetAllGIDs"
	Analysis_GetParamsByID_FullMethodName         = "/analysis.v1.Analysis/GetParamsByID"
	Analysis_GenerateImage_FullMethodName         = "/analysis.v1.Analysis/GenerateImage"
	Analysis_GetAllFunctionName_FullMethodName    = "/analysis.v1.Analysis/GetAllFunctionName"
	Analysis_GetGidsByFunctionName_FullMethodName = "/analysis.v1.Analysis/GetGidsByFunctionName"
	Analysis_VerifyProjectPath_FullMethodName     = "/analysis.v1.Analysis/VerifyProjectPath"
	Analysis_CheckDatabase_FullMethodName         = "/analysis.v1.Analysis/CheckDatabase"
	Analysis_GetTraceGraph_FullMethodName         = "/analysis.v1.Analysis/GetTraceGraph"
)

// AnalysisClient is the client API for Analysis service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AnalysisClient interface {
	// Sends a greeting
	GetAnalysis(ctx context.Context, in *AnalysisRequest, opts ...grpc.CallOption) (*AnalysisReply, error)
	GetAnalysisByGID(ctx context.Context, in *AnalysisByGIDRequest, opts ...grpc.CallOption) (*AnalysisByGIDReply, error)
	GetAllGIDs(ctx context.Context, in *GetAllGIDsReq, opts ...grpc.CallOption) (*GetAllGIDsReply, error)
	GetParamsByID(ctx context.Context, in *GetParamsByIDReq, opts ...grpc.CallOption) (*GetParamsByIDReply, error)
	GenerateImage(ctx context.Context, in *GenerateImageReq, opts ...grpc.CallOption) (*GenerateImageReply, error)
	GetAllFunctionName(ctx context.Context, in *GetAllFunctionNameReq, opts ...grpc.CallOption) (*GetAllFunctionNameReply, error)
	GetGidsByFunctionName(ctx context.Context, in *GetGidsByFunctionNameReq, opts ...grpc.CallOption) (*GetGidsByFunctionNameReply, error)
	VerifyProjectPath(ctx context.Context, in *VerifyProjectPathReq, opts ...grpc.CallOption) (*VerifyProjectPathReply, error)
	// CheckDatabase checks if the trace database exists
	CheckDatabase(ctx context.Context, in *CheckDatabaseRequest, opts ...grpc.CallOption) (*CheckDatabaseResponse, error)
	GetTraceGraph(ctx context.Context, in *GetTraceGraphReq, opts ...grpc.CallOption) (*GetTraceGraphReply, error)
}

type analysisClient struct {
	cc grpc.ClientConnInterface
}

func NewAnalysisClient(cc grpc.ClientConnInterface) AnalysisClient {
	return &analysisClient{cc}
}

func (c *analysisClient) GetAnalysis(ctx context.Context, in *AnalysisRequest, opts ...grpc.CallOption) (*AnalysisReply, error) {
	out := new(AnalysisReply)
	err := c.cc.Invoke(ctx, Analysis_GetAnalysis_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) GetAnalysisByGID(ctx context.Context, in *AnalysisByGIDRequest, opts ...grpc.CallOption) (*AnalysisByGIDReply, error) {
	out := new(AnalysisByGIDReply)
	err := c.cc.Invoke(ctx, Analysis_GetAnalysisByGID_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) GetAllGIDs(ctx context.Context, in *GetAllGIDsReq, opts ...grpc.CallOption) (*GetAllGIDsReply, error) {
	out := new(GetAllGIDsReply)
	err := c.cc.Invoke(ctx, Analysis_GetAllGIDs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) GetParamsByID(ctx context.Context, in *GetParamsByIDReq, opts ...grpc.CallOption) (*GetParamsByIDReply, error) {
	out := new(GetParamsByIDReply)
	err := c.cc.Invoke(ctx, Analysis_GetParamsByID_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) GenerateImage(ctx context.Context, in *GenerateImageReq, opts ...grpc.CallOption) (*GenerateImageReply, error) {
	out := new(GenerateImageReply)
	err := c.cc.Invoke(ctx, Analysis_GenerateImage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) GetAllFunctionName(ctx context.Context, in *GetAllFunctionNameReq, opts ...grpc.CallOption) (*GetAllFunctionNameReply, error) {
	out := new(GetAllFunctionNameReply)
	err := c.cc.Invoke(ctx, Analysis_GetAllFunctionName_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) GetGidsByFunctionName(ctx context.Context, in *GetGidsByFunctionNameReq, opts ...grpc.CallOption) (*GetGidsByFunctionNameReply, error) {
	out := new(GetGidsByFunctionNameReply)
	err := c.cc.Invoke(ctx, Analysis_GetGidsByFunctionName_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) VerifyProjectPath(ctx context.Context, in *VerifyProjectPathReq, opts ...grpc.CallOption) (*VerifyProjectPathReply, error) {
	out := new(VerifyProjectPathReply)
	err := c.cc.Invoke(ctx, Analysis_VerifyProjectPath_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) CheckDatabase(ctx context.Context, in *CheckDatabaseRequest, opts ...grpc.CallOption) (*CheckDatabaseResponse, error) {
	out := new(CheckDatabaseResponse)
	err := c.cc.Invoke(ctx, Analysis_CheckDatabase_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analysisClient) GetTraceGraph(ctx context.Context, in *GetTraceGraphReq, opts ...grpc.CallOption) (*GetTraceGraphReply, error) {
	out := new(GetTraceGraphReply)
	err := c.cc.Invoke(ctx, Analysis_GetTraceGraph_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AnalysisServer is the server API for Analysis service.
// All implementations must embed UnimplementedAnalysisServer
// for forward compatibility
type AnalysisServer interface {
	// Sends a greeting
	GetAnalysis(context.Context, *AnalysisRequest) (*AnalysisReply, error)
	GetAnalysisByGID(context.Context, *AnalysisByGIDRequest) (*AnalysisByGIDReply, error)
	GetAllGIDs(context.Context, *GetAllGIDsReq) (*GetAllGIDsReply, error)
	GetParamsByID(context.Context, *GetParamsByIDReq) (*GetParamsByIDReply, error)
	GenerateImage(context.Context, *GenerateImageReq) (*GenerateImageReply, error)
	GetAllFunctionName(context.Context, *GetAllFunctionNameReq) (*GetAllFunctionNameReply, error)
	GetGidsByFunctionName(context.Context, *GetGidsByFunctionNameReq) (*GetGidsByFunctionNameReply, error)
	VerifyProjectPath(context.Context, *VerifyProjectPathReq) (*VerifyProjectPathReply, error)
	// CheckDatabase checks if the trace database exists
	CheckDatabase(context.Context, *CheckDatabaseRequest) (*CheckDatabaseResponse, error)
	GetTraceGraph(context.Context, *GetTraceGraphReq) (*GetTraceGraphReply, error)
	mustEmbedUnimplementedAnalysisServer()
}

// UnimplementedAnalysisServer must be embedded to have forward compatible implementations.
type UnimplementedAnalysisServer struct {
}

func (UnimplementedAnalysisServer) GetAnalysis(context.Context, *AnalysisRequest) (*AnalysisReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAnalysis not implemented")
}
func (UnimplementedAnalysisServer) GetAnalysisByGID(context.Context, *AnalysisByGIDRequest) (*AnalysisByGIDReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAnalysisByGID not implemented")
}
func (UnimplementedAnalysisServer) GetAllGIDs(context.Context, *GetAllGIDsReq) (*GetAllGIDsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllGIDs not implemented")
}
func (UnimplementedAnalysisServer) GetParamsByID(context.Context, *GetParamsByIDReq) (*GetParamsByIDReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetParamsByID not implemented")
}
func (UnimplementedAnalysisServer) GenerateImage(context.Context, *GenerateImageReq) (*GenerateImageReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateImage not implemented")
}
func (UnimplementedAnalysisServer) GetAllFunctionName(context.Context, *GetAllFunctionNameReq) (*GetAllFunctionNameReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllFunctionName not implemented")
}
func (UnimplementedAnalysisServer) GetGidsByFunctionName(context.Context, *GetGidsByFunctionNameReq) (*GetGidsByFunctionNameReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGidsByFunctionName not implemented")
}
func (UnimplementedAnalysisServer) VerifyProjectPath(context.Context, *VerifyProjectPathReq) (*VerifyProjectPathReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyProjectPath not implemented")
}
func (UnimplementedAnalysisServer) CheckDatabase(context.Context, *CheckDatabaseRequest) (*CheckDatabaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckDatabase not implemented")
}
func (UnimplementedAnalysisServer) GetTraceGraph(context.Context, *GetTraceGraphReq) (*GetTraceGraphReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTraceGraph not implemented")
}
func (UnimplementedAnalysisServer) mustEmbedUnimplementedAnalysisServer() {}

// UnsafeAnalysisServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AnalysisServer will
// result in compilation errors.
type UnsafeAnalysisServer interface {
	mustEmbedUnimplementedAnalysisServer()
}

func RegisterAnalysisServer(s grpc.ServiceRegistrar, srv AnalysisServer) {
	s.RegisterService(&Analysis_ServiceDesc, srv)
}

func _Analysis_GetAnalysis_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnalysisRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).GetAnalysis(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_GetAnalysis_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).GetAnalysis(ctx, req.(*AnalysisRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_GetAnalysisByGID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnalysisByGIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).GetAnalysisByGID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_GetAnalysisByGID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).GetAnalysisByGID(ctx, req.(*AnalysisByGIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_GetAllGIDs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllGIDsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).GetAllGIDs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_GetAllGIDs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).GetAllGIDs(ctx, req.(*GetAllGIDsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_GetParamsByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetParamsByIDReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).GetParamsByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_GetParamsByID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).GetParamsByID(ctx, req.(*GetParamsByIDReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_GenerateImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateImageReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).GenerateImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_GenerateImage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).GenerateImage(ctx, req.(*GenerateImageReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_GetAllFunctionName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllFunctionNameReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).GetAllFunctionName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_GetAllFunctionName_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).GetAllFunctionName(ctx, req.(*GetAllFunctionNameReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_GetGidsByFunctionName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGidsByFunctionNameReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).GetGidsByFunctionName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_GetGidsByFunctionName_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).GetGidsByFunctionName(ctx, req.(*GetGidsByFunctionNameReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_VerifyProjectPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyProjectPathReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).VerifyProjectPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_VerifyProjectPath_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).VerifyProjectPath(ctx, req.(*VerifyProjectPathReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_CheckDatabase_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckDatabaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).CheckDatabase(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_CheckDatabase_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).CheckDatabase(ctx, req.(*CheckDatabaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analysis_GetTraceGraph_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTraceGraphReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalysisServer).GetTraceGraph(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Analysis_GetTraceGraph_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalysisServer).GetTraceGraph(ctx, req.(*GetTraceGraphReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Analysis_ServiceDesc is the grpc.ServiceDesc for Analysis service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Analysis_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "analysis.v1.Analysis",
	HandlerType: (*AnalysisServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAnalysis",
			Handler:    _Analysis_GetAnalysis_Handler,
		},
		{
			MethodName: "GetAnalysisByGID",
			Handler:    _Analysis_GetAnalysisByGID_Handler,
		},
		{
			MethodName: "GetAllGIDs",
			Handler:    _Analysis_GetAllGIDs_Handler,
		},
		{
			MethodName: "GetParamsByID",
			Handler:    _Analysis_GetParamsByID_Handler,
		},
		{
			MethodName: "GenerateImage",
			Handler:    _Analysis_GenerateImage_Handler,
		},
		{
			MethodName: "GetAllFunctionName",
			Handler:    _Analysis_GetAllFunctionName_Handler,
		},
		{
			MethodName: "GetGidsByFunctionName",
			Handler:    _Analysis_GetGidsByFunctionName_Handler,
		},
		{
			MethodName: "VerifyProjectPath",
			Handler:    _Analysis_VerifyProjectPath_Handler,
		},
		{
			MethodName: "CheckDatabase",
			Handler:    _Analysis_CheckDatabase_Handler,
		},
		{
			MethodName: "GetTraceGraph",
			Handler:    _Analysis_GetTraceGraph_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "analysis/v1/analysis.proto",
}
