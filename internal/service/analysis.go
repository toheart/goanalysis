package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/toheart/goanalysis/api/analysis/v1"
	"github.com/toheart/goanalysis/internal/biz/analysis"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"google.golang.org/grpc"
)

// GreeterService is a greeter service.
type AnalysisService struct {
	v1.UnimplementedAnalysisServer

	uc  *analysis.AnalysisBiz
	log *log.Helper
}

// NewGreeterService new a greeter service.
func NewAnalysisService(uc *analysis.AnalysisBiz, logger log.Logger) *AnalysisService {
	return &AnalysisService{uc: uc, log: log.NewHelper(logger)}
}

func (a *AnalysisService) RegisterGrpc(svr *grpc.Server) {
	v1.RegisterAnalysisServer(svr, a)
}

func (a *AnalysisService) RegisterHttp(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return v1.RegisterAnalysisHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
}

// SayHello implements helloworld.GreeterServer.
func (a *AnalysisService) GetAnalysis(ctx context.Context, in *v1.AnalysisRequest) (*v1.AnalysisReply, error) {
	return &v1.AnalysisReply{Message: "Hello " + in.Name}, nil
}

func (a *AnalysisService) GetAnalysisByGID(ctx context.Context, in *v1.AnalysisByGIDRequest) (*v1.AnalysisByGIDReply, error) {
	traces, err := a.uc.GetTracesByGID(in.Gid)
	if err != nil {
		return nil, err
	}

	reply := &v1.AnalysisByGIDReply{}
	for _, trace := range traces {
		traceData := &v1.AnalysisByGIDReply_TraceData{
			Id:             int32(trace.ID),
			Name:           trace.Name,
			Gid:            int32(trace.GID),
			Indent:         int32(trace.Indent),
			ParamCount:     int32(len(trace.Params)),
			TimeCost:       trace.TimeCost,
			ParentFuncname: trace.ParentFuncname,
		}

		reply.TraceData = append(reply.TraceData, traceData)
	}
	return reply, nil
}

func (a *AnalysisService) GetAllGIDs(ctx context.Context, in *v1.GetAllGIDsReq) (*v1.GetAllGIDsReply, error) {
	page := in.Page
	limit := in.Limit
	includeMetrics := in.IncludeMetrics
	reply := &v1.GetAllGIDsReply{}
	gids, err := a.uc.GetAllGIDs(int(page), int(limit))
	if err != nil {
		return nil, err
	}

	for _, gid := range gids {
		initialFunc, err := a.uc.GetInitialFunc(gid)
		if err != nil {
			return nil, err
		}

		body := &v1.GetAllGIDsReply_Body{
			Gid:         gid,
			InitialFunc: initialFunc,
		}

		// 如果需要包含调用深度和执行时间
		if includeMetrics {
			// 获取调用深度
			depth, err := a.uc.GetGoroutineCallDepth(gid)
			if err != nil {
				// 如果获取失败，使用默认值
				body.Depth = 0
			} else {
				body.Depth = int32(depth)
			}

			// 获取执行时间
			execTime, err := a.uc.GetGoroutineExecutionTime(gid)
			if err != nil {
				// 如果获取失败，使用默认值
				body.ExecutionTime = "N/A"
			} else {
				body.ExecutionTime = execTime
			}
		}

		reply.Body = append(reply.Body, body)
	}

	// 获取总数
	total, err := a.uc.GetTotalGIDs()
	if err != nil {
		return nil, err
	}
	reply.Total = int32(total)

	return reply, nil
}

func (a *AnalysisService) GetParamsByID(ctx context.Context, in *v1.GetParamsByIDReq) (*v1.GetParamsByIDReply, error) {
	params, err := a.uc.GetParamsByID(in.Id)
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

func (a *AnalysisService) GenerateImage(ctx context.Context, in *v1.GenerateImageReq) (*v1.GenerateImageReply, error) {
	traces, err := a.uc.GetTracesByGID(in.Gid)
	if err != nil {
		return nil, err
	}

	// 构建Mermaid图表
	var mermaidText strings.Builder
	mermaidText.WriteString("graph TD\n")

	// 使用栈来跟踪调用关系
	stack := make([]entity.TraceData, 0)

	for _, trace := range traces {
		// 根据缩进级别调整栈
		for len(stack) > 0 && stack[len(stack)-1].Indent >= trace.Indent {
			stack = stack[:len(stack)-1] // 弹出栈顶元素
		}

		// 添加节点
		nodeID := fmt.Sprintf("n%d", trace.ID)
		nodeName := sanitizeMermaidText(trace.Name)
		mermaidText.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeID, nodeName))

		// 如果有父节点，添加边
		if len(stack) > 0 {
			parentID := fmt.Sprintf("n%d", stack[len(stack)-1].ID)
			mermaidText.WriteString(fmt.Sprintf("    %s --> |%s| %s\n", parentID, trace.TimeCost, nodeID))
		}

		// 将当前节点压入栈
		stack = append(stack, trace)
	}

	return &v1.GenerateImageReply{
		Image: mermaidText.String(),
	}, nil
}

func sanitizeMermaidText(text string) string {
	// 替换可能导致Mermaid语法问题的字符
	text = strings.ReplaceAll(text, "\"", "'")
	return text
}

func (a *AnalysisService) GetAllFunctionName(ctx context.Context, in *v1.GetAllFunctionNameReq) (*v1.GetAllFunctionNameReply, error) {
	functionNames, err := a.uc.GetAllFunctionName()
	if err != nil {
		return nil, err
	}
	return &v1.GetAllFunctionNameReply{FunctionNames: functionNames}, nil
}

func (a *AnalysisService) GetGidsByFunctionName(ctx context.Context, in *v1.GetGidsByFunctionNameReq) (*v1.GetGidsByFunctionNameReply, error) {
	// 获取函数名称
	functionName := in.FunctionName
	includeMetrics := in.IncludeMetrics

	// 获取所有GID
	allGids, err := a.uc.GetAllGIDs(0, 1000) // 假设最多1000个GID
	if err != nil {
		return nil, err
	}

	// 过滤包含指定函数的GID
	var matchingGids []uint64
	for _, gid := range allGids {
		traces, err := a.uc.GetTracesByGID(strconv.FormatUint(gid, 10))
		if err != nil {
			continue
		}

		// 检查是否包含指定函数
		for _, trace := range traces {
			// 简化函数名称进行比较
			simplifiedName := getLastSegment(removeParentheses(trace.Name))
			if simplifiedName == functionName {
				matchingGids = append(matchingGids, gid)
				break
			}
		}
	}

	// 构建响应
	reply := &v1.GetGidsByFunctionNameReply{}
	for _, gid := range matchingGids {
		initialFunc, err := a.uc.GetInitialFunc(gid)
		if err != nil {
			continue
		}

		body := &v1.GetGidsByFunctionNameReply_Body{
			Gid:         gid,
			InitialFunc: initialFunc,
		}

		// 如果需要包含调用深度和执行时间
		if includeMetrics {
			// 获取调用深度
			depth, err := a.uc.GetGoroutineCallDepth(gid)
			if err != nil {
				// 如果获取失败，使用默认值
				body.Depth = 0
			} else {
				body.Depth = int32(depth)
			}

			// 获取执行时间
			execTime, err := a.uc.GetGoroutineExecutionTime(gid)
			if err != nil {
				// 如果获取失败，使用默认值
				body.ExecutionTime = "N/A"
			} else {
				body.ExecutionTime = execTime
			}
		}

		reply.Body = append(reply.Body, body)
	}

	// 获取总数
	reply.Total = int32(len(matchingGids))

	return reply, nil
}

func getLastSegment(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func removeParentheses(name string) string {
	idx := strings.Index(name, "(")
	if idx > 0 {
		return name[:idx]
	}
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

// 以下是需要保留的动态分析相关方法，但需要移除静态分析相关的方法
// 保留的方法：
// - GetTraceGraph
// - GetTracesByParentFunc
// - GetAllParentFuncNames
// - GetChildFunctions
// - GetHotFunctions
// - GetGoroutineStats
// - GetFunctionAnalysis
// - GetFunctionCallGraph

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

func (a *AnalysisService) GetTraceGraph(ctx context.Context, in *v1.GetTraceGraphReq) (*v1.GetTraceGraphReply, error) {
	traces, err := a.uc.GetTracesByGID(in.Gid)
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
	for id := range nodeMap {
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

// GetTracesByParentFunc 根据父函数名称获取函数调用
func (a *AnalysisService) GetTracesByParentFunc(ctx context.Context, in *v1.GetTracesByParentFuncReq) (*v1.GetTracesByParentFuncReply, error) {
	traces, err := a.uc.GetTracesByParentFunc(in.ParentFunc)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetTracesByParentFuncReply{}
	for _, trace := range traces {
		traceData := &v1.GetTracesByParentFuncReply_TraceData{
			Id:             int32(trace.ID),
			Name:           trace.Name,
			Gid:            int32(trace.GID),
			Indent:         int32(trace.Indent),
			ParamCount:     int32(len(trace.Params)),
			TimeCost:       trace.TimeCost,
			ParentFuncname: trace.ParentFuncname,
		}

		reply.TraceData = append(reply.TraceData, traceData)
	}
	return reply, nil
}

// GetAllParentFuncNames 获取所有的父函数名称
func (a *AnalysisService) GetAllParentFuncNames(ctx context.Context, in *v1.GetAllParentFuncNamesReq) (*v1.GetAllParentFuncNamesReply, error) {
	parentFuncNames, err := a.uc.GetAllParentFuncNames()
	if err != nil {
		return nil, err
	}

	reply := &v1.GetAllParentFuncNamesReply{
		ParentFuncNames: parentFuncNames,
	}
	return reply, nil
}

// GetChildFunctions 获取函数的子函数
func (a *AnalysisService) GetChildFunctions(ctx context.Context, in *v1.GetChildFunctionsReq) (*v1.GetChildFunctionsReply, error) {
	childFunctions, err := a.uc.GetChildFunctions(in.FuncName)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetChildFunctionsReply{
		ChildFunctions: childFunctions,
	}
	return reply, nil
}

// GetHotFunctions 获取热点函数分析数据
func (a *AnalysisService) GetHotFunctions(ctx context.Context, req *v1.GetHotFunctionsReq) (*v1.GetHotFunctionsReply, error) {
	hotFunctions, err := a.uc.GetHotFunctions(req.SortBy)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetHotFunctionsReply{
		Functions: make([]*v1.GetHotFunctionsReply_HotFunction, 0, len(hotFunctions)),
	}

	for _, hf := range hotFunctions {
		reply.Functions = append(reply.Functions, &v1.GetHotFunctionsReply_HotFunction{
			Name:      hf.Name,
			Package:   hf.Package,
			CallCount: int32(hf.CallCount),
			TotalTime: hf.TotalTime,
			AvgTime:   hf.AvgTime,
		})
	}

	return reply, nil
}

// GetGoroutineStats 获取Goroutine统计信息
func (a *AnalysisService) GetGoroutineStats(ctx context.Context, req *v1.GetGoroutineStatsReq) (*v1.GetGoroutineStatsReply, error) {
	stats, err := a.uc.GetGoroutineStats()
	if err != nil {
		return nil, err
	}

	return &v1.GetGoroutineStatsReply{
		Active:   int32(stats.Active),
		AvgTime:  stats.AvgTime,
		MaxDepth: int32(stats.MaxDepth),
	}, nil
}

// GetFunctionAnalysis 获取函数调用关系分析
func (a *AnalysisService) GetFunctionAnalysis(ctx context.Context, req *v1.GetFunctionAnalysisReq) (*v1.GetFunctionAnalysisReply, error) {
	// 验证项目路径
	if req.Path != "" {
		a.uc.SetCurrDB(req.Path)
	}

	// 获取函数调用关系
	nodes, err := a.uc.GetFunctionAnalysis(req.FunctionName, req.Type)
	if err != nil {
		return nil, err
	}

	// 转换为API响应格式
	reply := &v1.GetFunctionAnalysisReply{
		CallData: convertToProtoFunctionNodes(nodes),
	}

	return reply, nil
}

// 将实体FunctionNode转换为proto FunctionNode
func convertToProtoFunctionNodes(nodes []entity.FunctionNode) []*v1.GetFunctionAnalysisReply_FunctionNode {
	var result []*v1.GetFunctionAnalysisReply_FunctionNode

	for _, node := range nodes {
		protoNode := convertToProtoFunctionNode(node)
		result = append(result, protoNode)
	}

	return result
}

// 将单个实体FunctionNode转换为proto FunctionNode
func convertToProtoFunctionNode(node entity.FunctionNode) *v1.GetFunctionAnalysisReply_FunctionNode {
	protoNode := &v1.GetFunctionAnalysisReply_FunctionNode{
		Id:        node.ID,
		Name:      node.Name,
		Package:   node.Package,
		CallCount: int32(node.CallCount),
		AvgTime:   node.AvgTime,
		Children:  make([]*v1.GetFunctionAnalysisReply_FunctionNode, 0, len(node.Children)),
	}

	// 递归转换子节点
	for _, child := range node.Children {
		childNode := convertToProtoFunctionNode(child)
		protoNode.Children = append(protoNode.Children, childNode)
	}

	return protoNode
}

// GetFunctionCallGraph 获取函数调用关系图
func (a *AnalysisService) GetFunctionCallGraph(ctx context.Context, req *v1.GetFunctionCallGraphReq) (*v1.GetFunctionCallGraphReply, error) {
	// 设置默认值
	depth := int(req.Depth)
	if depth <= 0 {
		depth = 2 // 默认深度为2
	}

	direction := req.Direction
	if direction == "" {
		direction = "both" // 默认双向
	}

	// 获取函数调用关系图
	graph, err := a.uc.GetFunctionCallGraph(req.FunctionName, depth, direction)
	if err != nil {
		return nil, err
	}

	// 转换为API响应格式
	reply := &v1.GetFunctionCallGraphReply{
		Nodes: make([]*v1.GetFunctionCallGraphReply_GraphNode, 0, len(graph.Nodes)),
		Edges: make([]*v1.GetFunctionCallGraphReply_GraphEdge, 0, len(graph.Edges)),
	}

	// 转换节点
	for _, node := range graph.Nodes {
		protoNode := &v1.GetFunctionCallGraphReply_GraphNode{
			Id:        node.ID,
			Name:      node.Name,
			Package:   node.Package,
			CallCount: int32(node.CallCount),
			AvgTime:   node.AvgTime,
			NodeType:  node.NodeType,
		}
		reply.Nodes = append(reply.Nodes, protoNode)
	}

	// 转换边
	for _, edge := range graph.Edges {
		protoEdge := &v1.GetFunctionCallGraphReply_GraphEdge{
			Source:   edge.Source,
			Target:   edge.Target,
			Label:    edge.Label,
			EdgeType: edge.EdgeType,
		}
		reply.Edges = append(reply.Edges, protoEdge)
	}

	return reply, nil
}
