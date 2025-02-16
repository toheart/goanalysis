package service

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/toheart/goanalysis/api/analysis/v1"
	"github.com/toheart/goanalysis/internal/biz/analysis"
	"github.com/toheart/goanalysis/internal/biz/entity"
)

// GreeterService is a greeter service.
type AnalysisService struct {
	v1.UnimplementedAnalysisServer

	uc *analysis.AnalysisBiz
}

// NewGreeterService new a greeter service.
func NewAnalysisService(uc *analysis.AnalysisBiz) *AnalysisService {
	return &AnalysisService{uc: uc}
}

// SayHello implements helloworld.GreeterServer.
func (s *AnalysisService) GetAnalysis(ctx context.Context, in *v1.AnalysisRequest) (*v1.AnalysisReply, error) {
	return &v1.AnalysisReply{Message: "Hello " + in.Name}, nil
}

func (s *AnalysisService) GetAnalysisByGID(ctx context.Context, in *v1.AnalysisByGIDRequest) (*v1.AnalysisByGIDReply, error) {
	traces, err := s.uc.GetTracesByGID(in.Gid)
	if err != nil {
		return nil, err
	}

	reply := &v1.AnalysisByGIDReply{}
	for _, trace := range traces {
		reply.TraceData = append(reply.TraceData, &v1.AnalysisByGIDReply_TraceData{
			Id:         int32(trace.ID),
			Name:       trace.Name,
			Gid:        int32(trace.GID),
			Indent:     int32(trace.Indent),
			ParamCount: int32(len(trace.Params)),
			TimeCost:   trace.TimeCost,
		})
	}
	return reply, nil
}

func (s *AnalysisService) GetAllGIDs(ctx context.Context, in *v1.GetAllGIDsReq) (*v1.GetAllGIDsReply, error) {
	gids, err := s.uc.GetAllGIDs()
	if err != nil {
		return nil, err
	}

	reply := &v1.GetAllGIDsReply{
		Gids: gids,
	}
	return reply, nil
}

func (s *AnalysisService) GetParamsByID(ctx context.Context, in *v1.GetParamsByIDReq) (*v1.GetParamsByIDReply, error) {
	params, err := s.uc.GetParamsByID(in.Id)
	if err != nil {
		return nil, err
	}
	reply := &v1.GetParamsByIDReply{}
	for _, param := range params {
		reply.Params = append(reply.Params, &v1.TraceParams{
			Pos:   int32(param.Pos),
			Param: param.Param,
		})
	}
	return reply, nil
}

func (s *AnalysisService) GenerateImage(ctx context.Context, in *v1.GenerateImageReq) (*v1.GenerateImageReply, error) {
	traces, err := s.uc.GetTracesByGID(in.Gid)
	if err != nil {
		return nil, err
	}
	mermaid := ""
	stack := make([]*entity.TraceData, 0) // 用于存储父节点
	existd := make(map[string]bool)
	for _, trace := range traces {
		// 根据 indent 判断父子关系
		for len(stack) > 0 && stack[len(stack)-1].Indent >= trace.Indent {
			stack = stack[:len(stack)-1] // 弹出栈顶元素
		}

		if len(stack) > 0 {
			// 当前 trace 是子节点
			parentName := getLastSegment(stack[len(stack)-1].Name)
			childName := getLastSegment(trace.Name)
			edge := fmt.Sprintf("    %s[%s] --> %s[%s];\n", parentName, removeParentheses(stack[len(stack)-1].Name), childName, removeParentheses(trace.Name))
			if !existd[edge] {
				existd[edge] = true
				mermaid += edge
			}
		}

		// 将当前 trace 压入栈中
		stack = append(stack, &entity.TraceData{
			ID:       trace.ID,
			Name:     trace.Name,
			GID:      trace.GID,
			Indent:   trace.Indent,
			TimeCost: trace.TimeCost,
		})
	}

	mermaid = "graph TD;\n" + mermaid

	return &v1.GenerateImageReply{Image: mermaid}, nil
}

func (s *AnalysisService) GetAllFunctionName(ctx context.Context, in *v1.GetAllFunctionNameReq) (*v1.GetAllFunctionNameReply, error) {
	functionNames, err := s.uc.GetAllFunctionName()
	if err != nil {
		return nil, err
	}
	return &v1.GetAllFunctionNameReply{FunctionNames: functionNames}, nil
}

func (s *AnalysisService) GetGidsByFunctionName(ctx context.Context, in *v1.GetGidsByFunctionNameReq) (*v1.GetGidsByFunctionNameReply, error) {
	gids, err := s.uc.GetGidsByFunctionName(in.FunctionName)
	if err != nil {
		return nil, err
	}
	return &v1.GetGidsByFunctionNameReply{Gids: gids}, nil
}

func getLastSegment(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func removeParentheses(name string) string {
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	return name
}
