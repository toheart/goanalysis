package service

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/toheart/goanalysis/api/analysis/v1"
	"github.com/toheart/goanalysis/internal/biz/analysis"
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
			Seq:        trace.Seq, // 添加seq字段
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

func (a *AnalysisService) GetGidsByFunctionName(ctx context.Context, in *v1.GetGidsByFunctionNameReq) (*v1.GetGidsByFunctionNameReply, error) {
	// 获取函数名称
	functionName := in.FunctionName
	includeMetrics := in.IncludeMetrics
	dbpath := in.Path // 使用 Path 字段作为 dbpath

	a.log.Infof("find %s in which goroutines, dbpath: %s", functionName, dbpath)

	// 调用业务逻辑层获取GID列表（已经处理了去重逻辑）
	functions, err := a.uc.GetGidsByFunctionName(dbpath, functionName)
	if err != nil {
		a.log.Errorf("found %s gids failed: %v", functionName, err)
		return nil, err
	}

	// 构建响应
	reply := &v1.GetGidsByFunctionNameReply{}

	for _, function := range functions {
		// 获取Goroutine的基本信息
		goroutine, err := a.uc.GetGoroutineByGID(dbpath, function.GID)
		if err != nil {
			a.log.Warnf("get goroutine %d info failed: %v", function.GID, err)
			continue
		}

		isfinished := (goroutine.IsFinished == 1)
		body := &v1.GetGidsByFunctionNameReply_Body{
			Gid:         function.GID,
			InitialFunc: goroutine.InitFuncName,
			IsFinished:  isfinished,
			FunctionId:  function.ID,
		}

		// 如果需要包含调用深度和执行时间
		if includeMetrics {
			// 获取调用深度
			depth, err := a.uc.GetGoroutineCallDepth(dbpath, function.GID)
			if err != nil {
				// 如果获取失败，使用默认值
				body.Depth = 0
			} else {
				body.Depth = int32(depth)
			}

			// 获取执行时间
			execTime, err := a.uc.GetGoroutineExecutionTime(dbpath, *goroutine)
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
	reply.Total = int32(len(reply.Body))

	a.log.Infof("found %s in %d goroutines", functionName, len(reply.Body))

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
			Seq:        function.Seq, // 添加seq字段
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
			Name:    f.Name,
			Package: f.Package,
		})
	}

	return reply, nil
}

// GetFunctionInfoInGoroutine 获取函数在指定Goroutine中的信息
func (a *AnalysisService) GetFunctionInfoInGoroutine(ctx context.Context, in *v1.GetFunctionInfoInGoroutineReq) (*v1.GetFunctionInfoInGoroutineReply, error) {
	a.log.Infof("get function info in goroutine, gid: %d, functionId: %d from db: %s", in.Gid, in.FunctionId, in.Dbpath)

	functionInfo, err := a.uc.GetFunctionInfoInGoroutine(in.Dbpath, in.Gid, in.FunctionId)
	if err != nil {
		return nil, err
	}
	// 转换ParentInfo为proto格式
	var parentInfos []*v1.ParentInfo
	for _, parentInfo := range functionInfo.ParentIds {
		parentInfos = append(parentInfos, &v1.ParentInfo{
			ParentId: parentInfo.ParentId,
			Depth:    int32(parentInfo.Depth),
			Name:     parentInfo.Name,
		})
	}

	reply := &v1.GetFunctionInfoInGoroutineReply{
		FunctionInfo: &v1.GetFunctionInfoInGoroutineReply_FunctionInfo{
			Id:        functionInfo.ID,
			Name:      functionInfo.Name,
			Depth:     int32(functionInfo.Indent),
			ParentIds: parentInfos,
		},
	}

	return reply, nil
}

func (a *AnalysisService) GetModuleNames(ctx context.Context, in *v1.GetModuleNamesReq) (*v1.GetModuleNamesReply, error) {
	// 设置默认采样数量
	maxSamples := in.MaxSamples
	if maxSamples <= 0 {
		maxSamples = 5000 // 默认采样5000条记录
	}

	moduleNames, err := a.uc.GetModuleNames(in.Dbpath, maxSamples)
	if err != nil {
		return nil, err
	}

	return &v1.GetModuleNamesReply{
		ModuleNames: moduleNames,
	}, nil
}
