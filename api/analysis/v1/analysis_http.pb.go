// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.8.3
// - protoc             v6.30.0--rc1
// source: analysis/v1/analysis.proto

package v1

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationAnalysisGenerateImage = "/analysis.v1.Analysis/GenerateImage"
const OperationAnalysisGetAllFunctionName = "/analysis.v1.Analysis/GetAllFunctionName"
const OperationAnalysisGetAllGIDs = "/analysis.v1.Analysis/GetAllGIDs"
const OperationAnalysisGetAnalysis = "/analysis.v1.Analysis/GetAnalysis"
const OperationAnalysisGetAnalysisByGID = "/analysis.v1.Analysis/GetAnalysisByGID"
const OperationAnalysisGetGidsByFunctionName = "/analysis.v1.Analysis/GetGidsByFunctionName"
const OperationAnalysisGetParamsByID = "/analysis.v1.Analysis/GetParamsByID"

type AnalysisHTTPServer interface {
	GenerateImage(context.Context, *GenerateImageReq) (*GenerateImageReply, error)
	GetAllFunctionName(context.Context, *GetAllFunctionNameReq) (*GetAllFunctionNameReply, error)
	GetAllGIDs(context.Context, *GetAllGIDsReq) (*GetAllGIDsReply, error)
	// GetAnalysis Sends a greeting
	GetAnalysis(context.Context, *AnalysisRequest) (*AnalysisReply, error)
	GetAnalysisByGID(context.Context, *AnalysisByGIDRequest) (*AnalysisByGIDReply, error)
	GetGidsByFunctionName(context.Context, *GetGidsByFunctionNameReq) (*GetGidsByFunctionNameReply, error)
	GetParamsByID(context.Context, *GetParamsByIDReq) (*GetParamsByIDReply, error)
}

func RegisterAnalysisHTTPServer(s *http.Server, srv AnalysisHTTPServer) {
	r := s.Route("/")
	r.GET("/analysis/{name}", _Analysis_GetAnalysis0_HTTP_Handler(srv))
	r.GET("/api/traces/{gid}", _Analysis_GetAnalysisByGID0_HTTP_Handler(srv))
	r.GET("/api/gids", _Analysis_GetAllGIDs0_HTTP_Handler(srv))
	r.GET("/api/params/{id}", _Analysis_GetParamsByID0_HTTP_Handler(srv))
	r.GET("/api/traces/{gid}/mermaid", _Analysis_GenerateImage0_HTTP_Handler(srv))
	r.GET("/api/functions", _Analysis_GetAllFunctionName0_HTTP_Handler(srv))
	r.POST("/api/gids/function", _Analysis_GetGidsByFunctionName0_HTTP_Handler(srv))
}

func _Analysis_GetAnalysis0_HTTP_Handler(srv AnalysisHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in AnalysisRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAnalysisGetAnalysis)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetAnalysis(ctx, req.(*AnalysisRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*AnalysisReply)
		return ctx.Result(200, reply)
	}
}

func _Analysis_GetAnalysisByGID0_HTTP_Handler(srv AnalysisHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in AnalysisByGIDRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAnalysisGetAnalysisByGID)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetAnalysisByGID(ctx, req.(*AnalysisByGIDRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*AnalysisByGIDReply)
		return ctx.Result(200, reply)
	}
}

func _Analysis_GetAllGIDs0_HTTP_Handler(srv AnalysisHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetAllGIDsReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAnalysisGetAllGIDs)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetAllGIDs(ctx, req.(*GetAllGIDsReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetAllGIDsReply)
		return ctx.Result(200, reply)
	}
}

func _Analysis_GetParamsByID0_HTTP_Handler(srv AnalysisHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetParamsByIDReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAnalysisGetParamsByID)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetParamsByID(ctx, req.(*GetParamsByIDReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetParamsByIDReply)
		return ctx.Result(200, reply)
	}
}

func _Analysis_GenerateImage0_HTTP_Handler(srv AnalysisHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GenerateImageReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAnalysisGenerateImage)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GenerateImage(ctx, req.(*GenerateImageReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GenerateImageReply)
		return ctx.Result(200, reply)
	}
}

func _Analysis_GetAllFunctionName0_HTTP_Handler(srv AnalysisHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetAllFunctionNameReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAnalysisGetAllFunctionName)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetAllFunctionName(ctx, req.(*GetAllFunctionNameReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetAllFunctionNameReply)
		return ctx.Result(200, reply)
	}
}

func _Analysis_GetGidsByFunctionName0_HTTP_Handler(srv AnalysisHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetGidsByFunctionNameReq
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAnalysisGetGidsByFunctionName)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetGidsByFunctionName(ctx, req.(*GetGidsByFunctionNameReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetGidsByFunctionNameReply)
		return ctx.Result(200, reply)
	}
}

type AnalysisHTTPClient interface {
	GenerateImage(ctx context.Context, req *GenerateImageReq, opts ...http.CallOption) (rsp *GenerateImageReply, err error)
	GetAllFunctionName(ctx context.Context, req *GetAllFunctionNameReq, opts ...http.CallOption) (rsp *GetAllFunctionNameReply, err error)
	GetAllGIDs(ctx context.Context, req *GetAllGIDsReq, opts ...http.CallOption) (rsp *GetAllGIDsReply, err error)
	GetAnalysis(ctx context.Context, req *AnalysisRequest, opts ...http.CallOption) (rsp *AnalysisReply, err error)
	GetAnalysisByGID(ctx context.Context, req *AnalysisByGIDRequest, opts ...http.CallOption) (rsp *AnalysisByGIDReply, err error)
	GetGidsByFunctionName(ctx context.Context, req *GetGidsByFunctionNameReq, opts ...http.CallOption) (rsp *GetGidsByFunctionNameReply, err error)
	GetParamsByID(ctx context.Context, req *GetParamsByIDReq, opts ...http.CallOption) (rsp *GetParamsByIDReply, err error)
}

type AnalysisHTTPClientImpl struct {
	cc *http.Client
}

func NewAnalysisHTTPClient(client *http.Client) AnalysisHTTPClient {
	return &AnalysisHTTPClientImpl{client}
}

func (c *AnalysisHTTPClientImpl) GenerateImage(ctx context.Context, in *GenerateImageReq, opts ...http.CallOption) (*GenerateImageReply, error) {
	var out GenerateImageReply
	pattern := "/api/traces/{gid}/mermaid"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationAnalysisGenerateImage))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AnalysisHTTPClientImpl) GetAllFunctionName(ctx context.Context, in *GetAllFunctionNameReq, opts ...http.CallOption) (*GetAllFunctionNameReply, error) {
	var out GetAllFunctionNameReply
	pattern := "/api/functions"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationAnalysisGetAllFunctionName))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AnalysisHTTPClientImpl) GetAllGIDs(ctx context.Context, in *GetAllGIDsReq, opts ...http.CallOption) (*GetAllGIDsReply, error) {
	var out GetAllGIDsReply
	pattern := "/api/gids"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationAnalysisGetAllGIDs))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AnalysisHTTPClientImpl) GetAnalysis(ctx context.Context, in *AnalysisRequest, opts ...http.CallOption) (*AnalysisReply, error) {
	var out AnalysisReply
	pattern := "/analysis/{name}"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationAnalysisGetAnalysis))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AnalysisHTTPClientImpl) GetAnalysisByGID(ctx context.Context, in *AnalysisByGIDRequest, opts ...http.CallOption) (*AnalysisByGIDReply, error) {
	var out AnalysisByGIDReply
	pattern := "/api/traces/{gid}"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationAnalysisGetAnalysisByGID))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AnalysisHTTPClientImpl) GetGidsByFunctionName(ctx context.Context, in *GetGidsByFunctionNameReq, opts ...http.CallOption) (*GetGidsByFunctionNameReply, error) {
	var out GetGidsByFunctionNameReply
	pattern := "/api/gids/function"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationAnalysisGetGidsByFunctionName))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AnalysisHTTPClientImpl) GetParamsByID(ctx context.Context, in *GetParamsByIDReq, opts ...http.CallOption) (*GetParamsByIDReply, error) {
	var out GetParamsByIDReply
	pattern := "/api/params/{id}"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationAnalysisGetParamsByID))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
