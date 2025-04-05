package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/toheart/goanalysis/api/analysis/v1"
	"github.com/toheart/goanalysis/internal/biz/analysis"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/rewrite"
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
	traces, err := a.uc.GetTracesByGID(in)
	if err != nil {
		return nil, err
	}

	reply := &v1.AnalysisByGIDReply{}
	for _, trace := range traces {
		traceData := &v1.AnalysisByGIDReply_TraceData{
			Id:         int32(trace.ID),
			Name:       trace.Name,
			Gid:        trace.GID,
			Indent:     int32(trace.Indent),
			ParamCount: int32(trace.ParamCount),
			TimeCost:   trace.TimeCost,
			ParentId:   int64(trace.ParentId),
		}

		reply.TraceData = append(reply.TraceData, traceData)
	}
	return reply, nil
}

func (a *AnalysisService) GetAllGIDs(ctx context.Context, in *v1.GetAllGIDsReq) (*v1.GetAllGIDsReply, error) {
	page := in.Page
	limit := in.Limit
	includeMetrics := in.IncludeMetrics
	dbpath := in.Dbpath
	reply := &v1.GetAllGIDsReply{}
	groutines, err := a.uc.GetAllGIDs(dbpath, int(page), int(limit))
	if err != nil {
		return nil, err
	}

	for _, g := range groutines {
		isFinished := (g.IsFinished == 1)
		body := &v1.GetAllGIDsReply_Body{
			Gid:         uint64(g.ID),
			InitialFunc: g.InitFuncName,
			IsFinished:  isFinished,
		}

		// 如果需要包含调用深度和执行时间
		if includeMetrics {
			// 获取调用深度
			depth, err := a.uc.GetGoroutineCallDepth(dbpath, uint64(g.ID))
			if err != nil {
				// 如果获取失败，使用默认值
				body.Depth = 0
			} else {
				body.Depth = int32(depth)
			}
			execTime, err := a.uc.GetGoroutineExecutionTime(dbpath, g)
			if err != nil {
				// 如果获取失败，使用默认值
				a.log.Errorf("get goroutine execution time for gid: %d from db: %s failed: %v", g.ID, dbpath, err)
				body.ExecutionTime = "N/A"
			} else {
				body.ExecutionTime = execTime
			}
		}
		reply.Body = append(reply.Body, body)
	}

	// 获取总数
	total, err := a.uc.GetTotalGIDs(dbpath)
	if err != nil {
		return nil, err
	}
	reply.Total = int32(total)

	return reply, nil
}

func (a *AnalysisService) GetParamsByID(ctx context.Context, in *v1.GetParamsByIDReq) (*v1.GetParamsByIDReply, error) {
	params, err := a.uc.GetParamsByID(in.Dbpath, in.Id)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetParamsByIDReply{}
	for _, param := range params {
		reply.Params = append(reply.Params, &v1.TraceParams{
			Pos:   int32(param.Position),
			Param: param.Data,
		})
	}
	return reply, nil
}

func (a *AnalysisService) GetAllFunctionName(ctx context.Context, in *v1.GetAllFunctionNameReq) (*v1.GetAllFunctionNameReply, error) {
	a.log.Infof("get all function name from db: %s", in.Dbpath)
	functionNames, err := a.uc.GetAllFunctionName(in.Dbpath)
	if err != nil {
		return nil, err
	}
	return &v1.GetAllFunctionNameReply{FunctionNames: functionNames}, nil
}

func (a *AnalysisService) GetGidsByFunctionName(ctx context.Context, in *v1.GetGidsByFunctionNameReq) (*v1.GetGidsByFunctionNameReply, error) {
	// 获取函数名称
	functionName := in.FunctionName
	includeMetrics := in.IncludeMetrics
	dbpath := in.Path // 使用 Path 字段作为 dbpath

	// 获取所有GID
	groutines, err := a.uc.GetAllGIDs(dbpath, 0, 1000) // 假设最多1000个GID
	if err != nil {
		return nil, err
	}

	// 过滤包含指定函数的GID
	var matchingG []entity.GoroutineTrace
	for _, g := range groutines {
		traces, err := a.uc.GetTracesByGID(&v1.AnalysisByGIDRequest{
			Dbpath:     dbpath,
			Gid:        uint64(g.ID),
			Depth:      3,
			CreateTime: "",
		})
		if err != nil {
			continue
		}

		// 检查是否包含指定函数
		for _, trace := range traces {
			// 简化函数名称进行比较
			simplifiedName := getLastSegment(removeParentheses(trace.Name))
			if simplifiedName == functionName {
				matchingG = append(matchingG, g)
				break
			}
		}
	}

	// 构建响应
	reply := &v1.GetGidsByFunctionNameReply{}
	for _, g := range matchingG {

		isfinished := (g.IsFinished == 1)
		body := &v1.GetGidsByFunctionNameReply_Body{
			Gid:         uint64(g.ID),
			InitialFunc: g.InitFuncName,
			IsFinished:  isfinished,
		}

		// 如果需要包含调用深度和执行时间
		if includeMetrics {
			// 获取调用深度
			depth, err := a.uc.GetGoroutineCallDepth(dbpath, uint64(g.ID))
			if err != nil {
				// 如果获取失败，使用默认值
				body.Depth = 0
			} else {
				body.Depth = int32(depth)
			}

			// 获取执行时间
			execTime, err := a.uc.GetGoroutineExecutionTime(dbpath, g)
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
	reply.Total = int32(len(matchingG))

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

// GetTracesByParentFunc 根据父函数ID获取函数调用
func (a *AnalysisService) GetTracesByParentFunc(ctx context.Context, in *v1.GetTracesByParentFuncReq) (*v1.GetTracesByParentFuncReply, error) {
	traces, err := a.uc.GetTracesByParentFunc(in.Dbpath, in.ParentId)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetTracesByParentFuncReply{}
	for _, trace := range traces {
		traceData := &v1.GetTracesByParentFuncReply_TraceData{
			Id:         int32(trace.ID),
			Name:       trace.Name,
			Gid:        int32(trace.GID),
			Indent:     int32(trace.Indent),
			ParamCount: int32(trace.ParamCount),
			TimeCost:   trace.TimeCost,
			ParentId:   int64(trace.ParentId),
		}

		reply.TraceData = append(reply.TraceData, traceData)
	}
	return reply, nil
}

// GetAllParentIds 获取所有的父函数ID
func (a *AnalysisService) GetParentFunctions(ctx context.Context, in *v1.GetParentFunctionsReq) (*v1.GetParentFunctionsReply, error) {
	parentFunctions, err := a.uc.GetParentFunctions(in.Dbpath, in.FunctionName)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetParentFunctionsReply{}
	for _, function := range parentFunctions {
		reply.Functions = append(reply.Functions, &v1.FunctionNode{
			Id:        function.Id,
			Name:      function.Name,
			Package:   function.Package,
			CallCount: int32(function.CallCount),
			AvgTime:   function.AvgTime,
		})
	}
	return reply, nil
}

// GetChildFunctions 获取函数的子函数
func (a *AnalysisService) GetChildFunctions(ctx context.Context, in *v1.GetChildFunctionsReq) (*v1.GetChildFunctionsReply, error) {
	childFunctions, err := a.uc.GetChildFunctions(in.Dbpath, in.ParentId)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetChildFunctionsReply{}
	for _, function := range childFunctions {
		reply.Functions = append(reply.Functions, &v1.FunctionNode{
			Id:         function.Id,
			Name:       function.Name,
			Package:    function.Package,
			CallCount:  int32(function.CallCount),
			AvgTime:    function.AvgTime,
			TimeCost:   function.TotalTime,
			ParamCount: int32(function.ParamCount),
			Depth:      int32(function.Depth),
		})
	}
	return reply, nil
}

// GetHotFunctions 获取热点函数分析数据
func (a *AnalysisService) GetHotFunctions(ctx context.Context, in *v1.GetHotFunctionsReq) (*v1.GetHotFunctionsReply, error) {
	hotFunctions, err := a.uc.GetHotFunctions(in.Dbpath, in.SortBy)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetHotFunctionsReply{}
	for _, hf := range hotFunctions {
		hotFunc := &v1.GetHotFunctionsReply_HotFunction{
			Name:      hf.Name,
			Package:   hf.Package,
			CallCount: int32(hf.CallCount),
			TotalTime: hf.TotalTime,
			AvgTime:   hf.AvgTime,
		}
		reply.Functions = append(reply.Functions, hotFunc)
	}
	return reply, nil
}

// GetGoroutineStats 获取Goroutine统计信息
func (a *AnalysisService) GetGoroutineStats(ctx context.Context, in *v1.GetGoroutineStatsReq) (*v1.GetGoroutineStatsReply, error) {
	stats, err := a.uc.GetGoroutineStats(in.Dbpath)
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
func (a *AnalysisService) GetFunctionAnalysis(ctx context.Context, in *v1.GetFunctionAnalysisReq) (*v1.GetFunctionAnalysisReply, error) {
	functionNodes, err := a.uc.GetFunctionAnalysis(in.Dbpath, in.FunctionName, in.Type)
	if err != nil {
		return nil, err
	}

	reply := &v1.GetFunctionAnalysisReply{}
	reply.CallData = convertToProtoFunctionNodes(functionNodes)
	return reply, nil
}

// 转换为protobuf类型的辅助函数
func convertToProtoFunctionNodes(nodes []entity.FunctionNode) []*v1.GetFunctionAnalysisReply_FunctionNode {
	protoNodes := make([]*v1.GetFunctionAnalysisReply_FunctionNode, len(nodes))
	for i, node := range nodes {
		protoNode := &v1.GetFunctionAnalysisReply_FunctionNode{
			Id:        node.ID,
			Name:      node.Name,
			Package:   node.Package,
			CallCount: int32(node.CallCount),
			AvgTime:   node.AvgTime,
		}

		if len(node.Children) > 0 {
			protoNode.Children = convertToProtoFunctionNodes(node.Children)
		}

		protoNodes[i] = protoNode
	}
	return protoNodes
}

// InstrumentProject 对项目进行插桩
func (a *AnalysisService) InstrumentProject(ctx context.Context, in *v1.InstrumentProjectReq) (*v1.InstrumentProjectReply, error) {
	a.log.Infof("Instrumenting project at path: %s", in.Path)

	if in.Path == "" {
		return &v1.InstrumentProjectReply{
			Success: false,
			Message: "项目路径不能为空",
		}, nil
	}

	// 调用rewrite包进行插桩
	defer func() {
		if r := recover(); r != nil {
			a.log.Errorf("Panic during instrumentation: %v", r)
		}
	}()

	// 执行插桩操作
	rewrite.RewriteDir(in.Path)
	// 暂时返回成功
	return &v1.InstrumentProjectReply{
		Success: true,
		Message: "项目插桩成功",
	}, nil
}

// GetUnfinishedFunctions 获取未完成的函数列表
func (a *AnalysisService) GetUnfinishedFunctions(ctx context.Context, in *v1.GetUnfinishedFunctionsReq) (*v1.GetUnfinishedFunctionsReply, error) {
	a.log.WithContext(ctx).Infof("GetUnfinishedFunctions request received: threshold=%d, dbpath=%s, page=%d, limit=%d",
		in.Threshold, in.Dbpath, in.Page, in.Limit)

	// 参数验证
	if in.Dbpath == "" {
		a.log.WithContext(ctx).Errorf("GetUnfinishedFunctions failed: dbpath is empty")
		return nil, errors.New("dbpath is required")
	}

	// 获取未完成函数列表
	functions, err := a.uc.GetUnfinishedFunctions(in.Dbpath, in.Threshold)
	if err != nil {
		a.log.WithContext(ctx).Errorf("Failed to get unfinished functions: %v", err)
		return nil, fmt.Errorf("failed to get unfinished functions: %w", err)
	}

	// 计算分页
	totalCount := len(functions)
	page := int(in.Page)
	limit := int(in.Limit)

	// 默认值处理
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}

	// 计算起始和结束索引
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	// 边界检查
	if startIndex >= totalCount {
		startIndex = 0
		endIndex = 0
	} else if endIndex > totalCount {
		endIndex = totalCount
	}

	// 获取当前页的数据
	var pagedFunctions []entity.AllUnfinishedFunction
	if startIndex < endIndex {
		pagedFunctions = functions[startIndex:endIndex]
	} else {
		pagedFunctions = []entity.AllUnfinishedFunction{}
	}

	// 构建响应
	reply := &v1.GetUnfinishedFunctionsReply{
		Total: int32(totalCount),
	}

	for _, function := range pagedFunctions {
		reply.Functions = append(reply.Functions, &v1.GetUnfinishedFunctionsReply_UnfinishedFunction{
			Name:        function.Name,
			Gid:         function.GID,
			RunningTime: function.RunningTime,
			IsBlocking:  function.IsBlocking,
			FunctionId:  int64(function.FunctionID),
		})
	}

	a.log.WithContext(ctx).Infof("Found %d unfinished functions, returning page %d with %d items",
		totalCount, page, len(pagedFunctions))
	return reply, nil
}

// GetTreeGraph 获取树状图
func (a *AnalysisService) GetTreeGraph(ctx context.Context, req *v1.GetTreeGraphReq) (*v1.GetTreeGraphReply, error) {
	a.log.Infof("get tree graph, function: %s, dbpath: %s, chain_type: %s, depth: %d", req.FunctionName, req.DbPath, req.ChainType, req.Depth)

	// 调用业务逻辑获取树状图数据
	trees, err := a.uc.GetTreeGraph(req.DbPath, req.FunctionName, req.ChainType, int(req.Depth))
	if err != nil {
		a.log.Errorf("get tree graph failed: %v", err)
		return nil, err
	}

	// 转换为API响应格式
	reply := &v1.GetTreeGraphReply{
		Trees: make([]*v1.TreeNode, 0, len(trees)),
	}

	// 转换每棵树节点
	for _, tree := range trees {
		protoTree := a.convertTreeNodeToProto(tree)
		reply.Trees = append(reply.Trees, protoTree)
	}

	return reply, nil
}

// 将实体TreeNode转换为proto TreeNode
func (a *AnalysisService) convertTreeNodeToProto(node *entity.TreeNode) *v1.TreeNode {
	if node == nil {
		return nil
	}

	protoNode := &v1.TreeNode{
		Name:      node.Name,
		Value:     node.Value,
		Collapsed: true,
	}

	// 递归转换子节点
	if len(node.Children) > 0 {
		protoNode.Children = make([]*v1.TreeNode, 0, len(node.Children))
		for _, child := range node.Children {
			protoNode.Children = append(protoNode.Children, a.convertTreeNodeToProto(child))
		}
	}

	return protoNode
}

// GetTreeGraphByGID 根据GID获取多棵树状图数据
func (a *AnalysisService) GetTreeGraphByGID(ctx context.Context, req *v1.GetTreeGraphByGIDReq) (*v1.GetTreeGraphByGIDReply, error) {
	a.log.Infof("get tree graph by gid: %d, dbpath: %s", req.Gid, req.DbPath)

	// 调用业务逻辑获取树状图数据
	treeTrees, err := a.uc.GetTreeGraphByGID(req.DbPath, req.Gid)
	if err != nil {
		a.log.Errorf("get tree graph by gid failed: %v", err)
		return nil, err
	}

	// 转换为API响应格式
	reply := &v1.GetTreeGraphByGIDReply{
		Trees: make([]*v1.TreeNode, 0, len(treeTrees)),
	}

	// 转换每棵树
	for _, tree := range treeTrees {
		protoTree := a.convertTreeNodeToProto(tree)
		reply.Trees = append(reply.Trees, protoTree)
	}

	return reply, nil
}

// GetFunctionCallStats 获取函数调用统计分析
func (a *AnalysisService) GetFunctionCallStats(ctx context.Context, req *v1.GetFunctionCallStatsReq) (*v1.GetFunctionCallStatsReply, error) {
	a.log.Infof("get function call stats, function: %s, dbpath: %s", req.FunctionName, req.DbPath)

	// 调用业务逻辑获取函数调用统计数据
	stats, err := a.uc.GetFunctionCallStats(req.DbPath, req.FunctionName)
	if err != nil {
		a.log.Errorf("get function call stats failed: %v", err)
		return nil, err
	}

	// 转换为API响应格式
	reply := &v1.GetFunctionCallStatsReply{
		Stats: make([]*v1.FunctionCallStats, 0, len(stats)),
	}

	for _, stat := range stats {
		functionStat := &v1.FunctionCallStats{
			Name:        stat.Name,
			Package:     stat.Package,
			CallCount:   int32(stat.CallCount),
			CallerCount: int32(stat.CallerCount),
			CalleeCount: int32(stat.CalleeCount),
			AvgTime:     stat.AvgTime,
			MaxTime:     stat.MaxTime,
			MinTime:     stat.MinTime,
			TimeStdDev:  stat.TimeStdDev,
		}
		reply.Stats = append(reply.Stats, functionStat)
	}

	return reply, nil
}

// GetPerformanceAnomalies 获取性能异常检测结果
func (a *AnalysisService) GetPerformanceAnomalies(ctx context.Context, req *v1.GetPerformanceAnomaliesReq) (*v1.GetPerformanceAnomaliesReply, error) {
	a.log.Infof("获取性能异常检测, 函数: %s, dbpath: %s, 阈值: %f", req.FunctionName, req.DbPath, req.Threshold)

	// 调用业务逻辑获取性能异常数据
	anomalies, err := a.uc.GetPerformanceAnomalies(req.DbPath, req.FunctionName, req.Threshold)
	if err != nil {
		a.log.Errorf("获取性能异常检测失败: %v", err)
		return nil, err
	}

	// 转换为API响应格式
	reply := &v1.GetPerformanceAnomaliesReply{
		Anomalies: make([]*v1.PerformanceAnomaly, 0, len(anomalies)),
	}

	for _, anomaly := range anomalies {
		performanceAnomaly := &v1.PerformanceAnomaly{
			Name:        anomaly.Name,
			Package:     anomaly.Package,
			AnomalyType: anomaly.AnomalyType,
			Description: anomaly.Description,
			Severity:    anomaly.Severity,
			Details:     anomaly.Details,
		}
		reply.Anomalies = append(reply.Anomalies, performanceAnomaly)
	}

	return reply, nil
}

// SearchFunctions 实现函数搜索服务
func (s *AnalysisService) SearchFunctions(ctx context.Context, req *v1.SearchFunctionsReq) (*v1.SearchFunctionsReply, error) {
	functions, total, err := s.uc.SearchFunctions(ctx, req.Dbpath, req.Query, req.Limit)
	if err != nil {
		return nil, err
	}

	reply := &v1.SearchFunctionsReply{
		Functions: make([]*v1.SearchFunctionsReply_FunctionInfo, 0, len(functions)),
		Total:     total,
	}

	for _, f := range functions {
		reply.Functions = append(reply.Functions, &v1.SearchFunctionsReply_FunctionInfo{
			Name:      f.Name,
			Package:   f.Package,
			CallCount: int32(f.CallCount),
			AvgTime:   f.AvgTime,
		})
	}

	return reply, nil
}
