package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	page := in.Page
	limit := in.Limit
	reply := &v1.GetAllGIDsReply{}
	gids, err := s.uc.GetAllGIDs(int(page), int(limit))
	if err != nil {
		return nil, err
	}

	for _, gid := range gids {
		initialFunc, err := s.uc.GetInitialFunc(gid)
		if err != nil {
			return nil, err
		}
		reply.Body = append(reply.Body, &v1.GetAllGIDsReply_Body{
			Gid:         gid,
			InitialFunc: initialFunc,
		})
	}

	total, err := s.uc.GetTotalGIDs()
	if err != nil {
		return nil, err
	}
	reply.Total = int32(total)

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

			// 处理函数名，避免出现方括号导致前端无法加载
			parentNameSafe := sanitizeMermaidText(removeParentheses(stack[len(stack)-1].Name))
			childNameSafe := sanitizeMermaidText(removeParentheses(trace.Name))

			edge := fmt.Sprintf("    %s[\"%s\"] --> %s[\"%s\"];\n",
				parentName, parentNameSafe,
				childName, childNameSafe)

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

// sanitizeMermaidText 处理可能导致 Mermaid 语法错误的特殊字符
func sanitizeMermaidText(text string) string {
	// 替换方括号，这些在 Mermaid 中有特殊含义
	text = strings.ReplaceAll(text, "[", "(")
	text = strings.ReplaceAll(text, "]", ")")
	// 处理其他可能导致问题的字符
	text = strings.ReplaceAll(text, "\"", "'")
	return text
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
	reply := &v1.GetGidsByFunctionNameReply{}
	for _, gid := range gids {
		gidUint, err := strconv.ParseUint(gid, 10, 64)
		if err != nil {
			return nil, err
		}
		initialFunc, err := s.uc.GetInitialFunc(gidUint)
		if err != nil {
			return nil, err
		}
		reply.Body = append(reply.Body, &v1.GetGidsByFunctionNameReply_Body{
			Gid:         gidUint,
			InitialFunc: initialFunc,
		})
	}
	return reply, nil
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

func (s *AnalysisService) VerifyProjectPath(ctx context.Context, in *v1.VerifyProjectPathReq) (*v1.VerifyProjectPathReply, error) {
	verified := s.uc.VerifyProjectPath(in.Path)
	return &v1.VerifyProjectPathReply{Verified: verified}, nil
}

func (s *AnalysisService) CheckDatabase(ctx context.Context, req *v1.CheckDatabaseRequest) (*v1.CheckDatabaseResponse, error) {
	exists := s.uc.CheckDatabase()
	return &v1.CheckDatabaseResponse{
		Exists: exists,
	}, nil
}

type DbFile struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	CreateTime time.Time `json:"createTime"`
}

type AnalysisResult struct {
	TotalFunctions int `json:"totalFunctions"`
	TotalCalls     int `json:"totalCalls"`
	TotalPackages  int `json:"totalPackages"`
	// 可以添加更多分析结果字段
}

// GetDbFiles 获取数据库文件列表
func (s *AnalysisService) GetDbFiles() ([]DbFile, error) {
	dbPath := s.uc.GetStaticDBPath()
	files, err := os.ReadDir(dbPath)
	if err != nil {
		return nil, err
	}

	var dbFiles []DbFile
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".db" {
			info, err := file.Info()
			if err != nil {
				continue
			}

			dbFiles = append(dbFiles, DbFile{
				Path:       filepath.Join(dbPath, file.Name()),
				Name:       file.Name(),
				Size:       info.Size(),
				CreateTime: info.ModTime(),
			})
		}
	}

	return dbFiles, nil
}

// AnalyzeDb 分析指定的数据库文件
func (s *AnalysisService) AnalyzeDb(dbPath string) (*AnalysisResult, error) {
	// 验证文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		return nil, err
	}

	// TODO: 实现实际的数据库分析逻辑
	// 这里需要根据您的具体需求实现分析逻辑

	// 示例返回
	return &AnalysisResult{
		TotalFunctions: 100,
		TotalCalls:     500,
		TotalPackages:  20,
	}, nil
}

// 定义图形数据的结构
type GraphNode struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CallCount int    `json:"callCount"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

func (s *AnalysisService) GetTraceGraph(ctx context.Context, in *v1.GetTraceGraphReq) (*v1.GetTraceGraphReply, error) {
	traces, err := s.uc.GetTracesByGID(in.Gid)
	if err != nil {
		return nil, err
	}

	// 构建图形数据
	graphData := buildGraphFromTraces(traces)

	return &v1.GetTraceGraphReply{
		Nodes: convertToProtoNodes(graphData.Nodes),
		Edges: convertToProtoEdges(graphData.Edges),
	}, nil
}

func buildGraphFromTraces(traces []entity.TraceData) *GraphData {
	graphData := &GraphData{
		Nodes: []GraphNode{},
		Edges: []GraphEdge{},
	}

	nodeMap := make(map[string]bool)
	callCounts := make(map[string]int)

	// 使用栈来跟踪调用关系
	stack := make([]entity.TraceData, 0)

	for _, trace := range traces {
		// 根据缩进级别调整栈
		for len(stack) > 0 && stack[len(stack)-1].Indent >= trace.Indent {
			stack = stack[:len(stack)-1] // 弹出栈顶元素
		}

		// 添加节点
		nodeID := fmt.Sprintf("n%d", trace.ID)
		if !nodeMap[nodeID] {
			nodeMap[nodeID] = true
			callCounts[trace.Name]++
		}

		// 如果有父节点，添加边
		if len(stack) > 0 {
			parentID := fmt.Sprintf("n%d", stack[len(stack)-1].ID)

			// 添加边
			graphData.Edges = append(graphData.Edges, GraphEdge{
				Source: parentID,
				Target: nodeID,
				Label:  trace.TimeCost,
			})
		}

		// 将当前节点压入栈
		stack = append(stack, trace)
	}

	// 添加所有节点
	for id, _ := range nodeMap {
		for _, trace := range traces {
			if fmt.Sprintf("n%d", trace.ID) == id {
				graphData.Nodes = append(graphData.Nodes, GraphNode{
					ID:        id,
					Name:      trace.Name,
					CallCount: callCounts[trace.Name],
				})
				break
			}
		}
	}

	return graphData
}

// 转换为protobuf类型的辅助函数
func convertToProtoNodes(nodes []GraphNode) []*v1.GraphNode {
	protoNodes := make([]*v1.GraphNode, len(nodes))
	for i, node := range nodes {
		protoNodes[i] = &v1.GraphNode{
			Id:        node.ID,
			Name:      node.Name,
			CallCount: int32(node.CallCount),
		}
	}
	return protoNodes
}

func convertToProtoEdges(edges []GraphEdge) []*v1.GraphEdge {
	protoEdges := make([]*v1.GraphEdge, len(edges))
	for i, edge := range edges {
		protoEdges[i] = &v1.GraphEdge{
			Source: edge.Source,
			Target: edge.Target,
			Label:  edge.Label,
		}
	}
	return protoEdges
}
