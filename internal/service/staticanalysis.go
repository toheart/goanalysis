package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/toheart/goanalysis/api/staticanalysis/v1"
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/repo"
	"github.com/toheart/goanalysis/internal/biz/staticanalysis"
	"google.golang.org/grpc"
)

// StaticAnalysisService 静态分析服务
type StaticAnalysisService struct {
	v1.UnimplementedStaticAnalysisServer

	uc  *staticanalysis.StaticAnalysisBiz
	log *log.Helper
}

func (s *StaticAnalysisService) RegisterGrpc(svr *grpc.Server) {
	v1.RegisterStaticAnalysisServer(svr, s)
}

func (s *StaticAnalysisService) RegisterHttp(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return v1.RegisterStaticAnalysisHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
}

// NewStaticAnalysisService 创建静态分析服务
func NewStaticAnalysisService(uc *staticanalysis.StaticAnalysisBiz, logger log.Logger) *StaticAnalysisService {
	srv := &StaticAnalysisService{uc: uc, log: log.NewHelper(logger)}

	// 启动分析任务处理协程
	go uc.ProcessAnalysisTasks()

	return srv
}

// GetStaticDbFiles 获取静态分析数据库文件列表
func (s *StaticAnalysisService) GetStaticDbFiles(ctx context.Context, req *v1.GetStaticDbFilesRequest) (*v1.GetStaticDbFilesResponse, error) {
	s.log.Info("get static db files")

	dbFiles, err := s.getDbFiles()
	if err != nil {
		s.log.Errorf("get static db files failed: %v", err)
		return nil, err
	}

	var files []*v1.DbFileInfo
	for _, file := range dbFiles {
		files = append(files, &v1.DbFileInfo{
			Path:       file.Path,
			Name:       file.Name,
			Size:       file.Size,
			CreateTime: file.CreateTime.Format(time.RFC3339),
		})
	}

	s.log.Infof("found %d static db files", len(files))
	return &v1.GetStaticDbFilesResponse{
		Files: files,
	}, nil
}

// AnalyzeProjectPath 分析指定路径的项目并生成callgraph
func (s *StaticAnalysisService) AnalyzeProjectPath(ctx context.Context, req *v1.AnalyzeProjectPathRequest) (*v1.AnalyzeProjectPathResponse, error) {
	s.log.Infof("analyze project path: %s", req.Path)

	// 验证项目路径
	if !s.verifyProjectPath(req.Path) {
		return &v1.AnalyzeProjectPathResponse{
			Success: false,
			Message: "Invalid project path",
		}, nil
	}

	// 生成数据库路径，使用req.Path的最后一个路径和当前时间
	basePath := filepath.Base(req.Path)
	dbPath := fmt.Sprintf("%s_%s.db", basePath, time.Now().Format("20060102_150405"))

	// 创建分析选项
	options := &entity.AnalysisOptions{
		Algo:         req.Algo,
		IgnoreMethod: req.IgnoreMethod,
	}

	// 启动分析任务
	taskID := s.uc.AnalyzeProjectPathWithOptions(req.Path, dbPath, options)

	return &v1.AnalyzeProjectPathResponse{
		Success: true,
		Message: "Analysis task started",
		TaskId:  taskID,
	}, nil
}

// 验证项目路径是否存在
func (s *StaticAnalysisService) verifyProjectPath(path string) bool {
	if path == "" {
		s.log.Error("path can't be empty")
		return false
	}

	s.log.Infof("verify project path: %s", path)

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.log.Errorf("project path not found: %s", path)
		}
		s.log.Errorf("verify project path failed: %v", err)
		return false
	}

	if !info.IsDir() {
		s.log.Errorf("the specified path is not a directory: %s", path)
		return false
	}

	// 检查是否是Go项目（查找go.mod文件）
	goModPath := filepath.Join(path, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		if os.IsNotExist(err) {
			s.log.Warnf("go.mod file not found, maybe not a go module project: %s", path)
			// 不返回false，因为有些项目可能不使用Go模块
		}
	} else {
		s.log.Infof("found go.mod file, confirmed as go module project: %s", path)
	}

	return true
}

// AnalyzeDbFile 分析指定数据库文件
func (s *StaticAnalysisService) AnalyzeDbFile(ctx context.Context, req *v1.AnalyzeDbFileRequest) (*v1.AnalyzeDbFileResponse, error) {
	// 验证请求参数
	if req.DbPath == "" {
		s.log.Error("db path can't be empty")
		return nil, fmt.Errorf("db path can't be empty")
	}

	s.log.Infof("Start analyzing database file: %s", req.DbPath)

	result, err := s.analyzeDbReal(req.DbPath)
	if err != nil {
		s.log.Errorf("analyze db failed: %v", err)
		return nil, err
	}

	// 转换为API响应格式
	var packageDeps []*v1.PackageDependency
	if result.PackageDependencies != nil {
		for _, dep := range result.PackageDependencies {
			packageDeps = append(packageDeps, &v1.PackageDependency{
				Source: dep.Source,
				Target: dep.Target,
				Count:  int32(dep.Count),
			})
		}
	}

	var hotFuncs []*v1.HotFunction
	if result.HotFunctions != nil {
		for _, fn := range result.HotFunctions {
			hotFuncs = append(hotFuncs, &v1.HotFunction{
				Name:      fn.Name,
				CallCount: int32(fn.CallCount),
			})
		}
	}

	s.log.Infof("Database analysis completed, found %d functions, %d calls, %d packages",
		result.TotalFunctions, result.TotalCalls, result.TotalPackages)

	return &v1.AnalyzeDbFileResponse{
		TotalFunctions:      int32(result.TotalFunctions),
		TotalCalls:          int32(result.TotalCalls),
		TotalPackages:       int32(result.TotalPackages),
		PackageDependencies: packageDeps,
		HotFunctions:        hotFuncs,
	}, nil
}

// GetHotFunctions 分页获取热点函数
func (s *StaticAnalysisService) GetHotFunctions(ctx context.Context, req *v1.GetHotFunctionsRequest) (*v1.GetHotFunctionsResponse, error) {
	s.log.Infof("Getting hot functions for db: %s, page: %d, pageSize: %d", req.DbPath, req.Page, req.PageSize)

	// 验证文件是否存在
	if _, err := os.Stat(req.DbPath); err != nil {
		s.log.Errorf("Database file not found: %s", req.DbPath)
		return nil, fmt.Errorf("database file not found: %s", req.DbPath)
	}

	// 获取数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(req.DbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	// 获取所有函数节点
	nodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		s.log.Errorf("Failed to get function nodes: %v", err)
		return nil, fmt.Errorf("Failed to get function nodes: %v", err)
	}

	// 获取所有函数调用边
	edges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		s.log.Errorf("Failed to get function edges: %v", err)
		return nil, fmt.Errorf("Failed to get function edges: %v", err)
	}

	// 统计热点函数（被调用次数最多的函数）
	funcCallCounts := make(map[string]int)
	for _, edge := range edges {
		funcCallCounts[edge.CalleeKey]++
	}

	// 创建一个函数键到函数节点的映射
	funcNodeMap := make(map[string]*dos.FuncNode)
	for _, node := range nodes {
		funcNodeMap[node.Key] = node
	}

	// 将函数调用次数转换为热点函数列表
	var hotFunctions []*v1.HotFunction
	for key, count := range funcCallCounts {
		if node, ok := funcNodeMap[key]; ok {
			hotFunctions = append(hotFunctions, &v1.HotFunction{
				Key:       key,
				Name:      node.Name,
				Package:   node.Pkg,
				CallCount: int32(count),
			})
		}
	}

	// 按调用次数排序
	sort.Slice(hotFunctions, func(i, j int) bool {
		return hotFunctions[i].CallCount > hotFunctions[j].CallCount
	})

	// 计算总数
	total := len(hotFunctions)

	// 计算分页
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	// 计算起始和结束索引
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= total {
		startIndex = 0
		endIndex = 0
	}

	if endIndex > total {
		endIndex = total
	}

	// 获取当前页的数据
	var pagedFunctions []*v1.HotFunction
	if startIndex < endIndex {
		pagedFunctions = hotFunctions[startIndex:endIndex]
	}

	s.log.Infof("Returning %d hot functions (total: %d)", len(pagedFunctions), total)

	return &v1.GetHotFunctionsResponse{
		Functions: pagedFunctions,
		Total:     int32(total),
		Page:      int32(page),
		PageSize:  int32(pageSize),
		PageCount: int32((total + pageSize - 1) / pageSize),
	}, nil
}

// GetFunctionAnalysis 获取函数调用关系分析
func (s *StaticAnalysisService) GetFunctionAnalysis(ctx context.Context, req *v1.GetFunctionAnalysisReq) (*v1.GetFunctionAnalysisReply, error) {
	if req.FunctionName == "" {
		s.log.Error("Function name cannot be empty")
		return nil, fmt.Errorf("Function name cannot be empty")
	}

	// 验证查询类型
	if req.Type != "" && req.Type != "caller" && req.Type != "callee" {
		s.log.Errorf("Invalid query type: %s, should be 'caller' or 'callee'", req.Type)
		return nil, fmt.Errorf("Invalid query type: %s, should be 'caller' or 'callee'", req.Type)
	}

	// 如果未指定类型，默认为 "callee"
	queryType := req.Type
	if queryType == "" {
		queryType = "callee"
	}

	s.log.Infof("Analyzing %s relationships for function %s", queryType, req.FunctionName)
	functionNodes, err := s.uc.GetFunctionAnalysis(req.FunctionName, queryType, req.Path)
	if err != nil {
		s.log.Errorf("Failed to get function relationship analysis: %v", err)
		return nil, err
	}

	return &v1.GetFunctionAnalysisReply{
		CallData: s.convertToProtoFunctionNodes(functionNodes),
	}, nil
}

// GetFunctionCallGraph 获取函数调用关系图
func (s *StaticAnalysisService) GetFunctionCallGraph(ctx context.Context, req *v1.GetFunctionCallGraphReq) (*v1.GetFunctionCallGraphReply, error) {
	// 验证函数名称
	if req.FunctionName == "" {
		s.log.Error("Function name cannot be empty")
		return nil, fmt.Errorf("Function name cannot be empty")
	}

	// 设置默认值
	depth := int(req.Depth)
	if depth <= 0 {
		depth = 2 // 默认深度为2
	}

	direction := req.Direction
	if direction == "" {
		direction = "both" // 默认双向
	} else if direction != "caller" && direction != "callee" && direction != "both" {
		s.log.Errorf("Invalid direction: %s, should be 'caller', 'callee' or 'both'", direction)
		return nil, fmt.Errorf("Invalid direction: %s, should be 'caller', 'callee' or 'both'", direction)
	}

	s.log.Infof("Getting call graph for function %s, depth: %d, direction: %s", req.FunctionName, depth, direction)
	nodes, edges, err := s.uc.GetFunctionCallGraph(req.FunctionName, depth, direction)
	if err != nil {
		s.log.Errorf("Failed to get function call graph: %v", err)
		return nil, err
	}

	var protoNodes []*v1.GetFunctionCallGraphReply_GraphNode
	for _, node := range nodes {
		protoNodes = append(protoNodes, &v1.GetFunctionCallGraphReply_GraphNode{
			Id:        node.ID,
			Name:      node.Name,
			Package:   node.Package,
			CallCount: int32(node.CallCount),
			AvgTime:   node.AvgTime,
			NodeType:  node.NodeType,
		})
	}

	var protoEdges []*v1.GetFunctionCallGraphReply_GraphEdge
	for _, edge := range edges {
		protoEdges = append(protoEdges, &v1.GetFunctionCallGraphReply_GraphEdge{
			Source:   edge.Source,
			Target:   edge.Target,
			Label:    edge.Label,
			EdgeType: edge.EdgeType,
		})
	}

	return &v1.GetFunctionCallGraphReply{
		Nodes: protoNodes,
		Edges: protoEdges,
	}, nil
}

// 数据库文件结构
type staticDbFile struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	CreateTime time.Time `json:"createTime"`
}

// 分析结果结构
type staticAnalysisResult struct {
	TotalFunctions      int `json:"totalFunctions"`
	TotalCalls          int `json:"totalCalls"`
	TotalPackages       int `json:"totalPackages"`
	PackageDependencies []struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Count  int    `json:"count"`
	} `json:"packageDependencies"`
	HotFunctions []struct {
		Name      string `json:"name"`
		CallCount int    `json:"callCount"`
	} `json:"hotFunctions"`
}

// 获取数据库文件列表
func (s *StaticAnalysisService) getDbFiles() ([]staticDbFile, error) {
	dbPath := s.uc.GetStaticDBPath()

	// 确保目录存在
	if err := os.MkdirAll(dbPath, 0o755); err != nil {
		s.log.Errorf("Error ensuring database directory exists: %v", err)
		return nil, fmt.Errorf("Error ensuring database directory exists: %v", err)
	}

	s.log.Infof("Getting database files from directory %s", dbPath)

	files, err := os.ReadDir(dbPath)
	if err != nil {
		s.log.Errorf("Failed to read database directory: %v", err)
		return nil, fmt.Errorf("Failed to read database directory: %v", err)
	}

	var dbFiles []staticDbFile
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".db" {
			info, err := file.Info()
			if err != nil {
				s.log.Warnf("Failed to get file info: %s, error: %v", file.Name(), err)
				continue
			}

			dbFiles = append(dbFiles, staticDbFile{
				Path:       filepath.Join(dbPath, file.Name()),
				Name:       file.Name(),
				Size:       info.Size(),
				CreateTime: info.ModTime(),
			})
		}
	}

	// 按创建时间降序排序，最新的文件排在前面
	sort.Slice(dbFiles, func(i, j int) bool {
		return dbFiles[i].CreateTime.After(dbFiles[j].CreateTime)
	})

	s.log.Infof("Found %d database files", len(dbFiles))
	return dbFiles, nil
}

// 从数据库中读取实际的调用图数据
func (s *StaticAnalysisService) analyzeDbReal(dbPath string) (*staticAnalysisResult, error) {
	// 验证文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		return nil, err
	}

	s.log.Infof("Start analyzing database: %s", dbPath)

	// 获取数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	// 获取所有函数节点
	nodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		s.log.Errorf("Failed to get function nodes: %v", err)
		return nil, fmt.Errorf("Failed to get function nodes: %v", err)
	}

	// 获取所有函数调用边
	edges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		s.log.Errorf("Failed to get function edges: %v", err)
		return nil, fmt.Errorf("Failed to get function edges: %v", err)
	}

	// 统计包依赖关系
	packageDeps := make(map[string]map[string]int)
	for _, edge := range edges {
		// 获取调用方和被调用方的节点
		caller, err := funcNodeDB.GetFuncNodeByKey(edge.CallerKey)
		if err != nil || caller == nil {
			continue
		}

		callee, err := funcNodeDB.GetFuncNodeByKey(edge.CalleeKey)
		if err != nil || callee == nil {
			continue
		}

		// 如果调用方和被调用方的包不同，则记录包依赖关系
		if caller.Pkg != callee.Pkg {
			if _, ok := packageDeps[caller.Pkg]; !ok {
				packageDeps[caller.Pkg] = make(map[string]int)
			}
			packageDeps[caller.Pkg][callee.Pkg]++
		}
	}

	// 统计热点函数（被调用次数最多的函数）
	funcCallCounts := make(map[string]int)
	for _, edge := range edges {
		funcCallCounts[edge.CalleeKey]++
	}

	// 转换包依赖关系为结果格式
	var packageDependencies []struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Count  int    `json:"count"`
	}

	for source, targets := range packageDeps {
		for target, count := range targets {
			packageDependencies = append(packageDependencies, struct {
				Source string `json:"source"`
				Target string `json:"target"`
				Count  int    `json:"count"`
			}{
				Source: source,
				Target: target,
				Count:  count,
			})
		}
	}

	// 按依赖关系数量排序
	sort.Slice(packageDependencies, func(i, j int) bool {
		return packageDependencies[i].Count > packageDependencies[j].Count
	})

	// 限制返回的包依赖关系数量
	if len(packageDependencies) > 20 {
		packageDependencies = packageDependencies[:20]
	}

	// 转换热点函数为结果格式
	var hotFunctions []struct {
		Name      string `json:"name"`
		CallCount int    `json:"callCount"`
	}

	// 创建一个函数键到函数节点的映射
	funcNodeMap := make(map[string]*dos.FuncNode)
	for _, node := range nodes {
		funcNodeMap[node.Key] = node
	}

	// 将函数调用次数转换为热点函数列表
	for key, count := range funcCallCounts {
		if node, ok := funcNodeMap[key]; ok {
			hotFunctions = append(hotFunctions, struct {
				Name      string `json:"name"`
				CallCount int    `json:"callCount"`
			}{
				Name:      fmt.Sprintf("%s.%s", node.Pkg, node.Name),
				CallCount: count,
			})
		}
	}

	// 按调用次数排序
	sort.Slice(hotFunctions, func(i, j int) bool {
		return hotFunctions[i].CallCount > hotFunctions[j].CallCount
	})

	// 限制返回的热点函数数量
	if len(hotFunctions) > 20 {
		hotFunctions = hotFunctions[:20]
	}

	// 统计包的数量
	packages := make(map[string]bool)
	for _, node := range nodes {
		packages[node.Pkg] = true
	}

	result := &staticAnalysisResult{
		TotalFunctions:      len(nodes),
		TotalCalls:          len(edges),
		TotalPackages:       len(packages),
		PackageDependencies: packageDependencies,
		HotFunctions:        hotFunctions,
	}

	s.log.Infof("Database analysis completed, found %d functions, %d calls, %d packages",
		result.TotalFunctions, result.TotalCalls, result.TotalPackages)
	return result, nil
}

// 转换为proto函数节点
func (s *StaticAnalysisService) convertToProtoFunctionNodes(nodes []entity.FunctionNode) []*v1.GetFunctionAnalysisReply_FunctionNode {
	var protoNodes []*v1.GetFunctionAnalysisReply_FunctionNode
	for _, node := range nodes {
		protoNodes = append(protoNodes, s.convertToProtoFunctionNode(node))
	}
	return protoNodes
}

// 转换为proto函数节点
func (s *StaticAnalysisService) convertToProtoFunctionNode(node entity.FunctionNode) *v1.GetFunctionAnalysisReply_FunctionNode {
	protoNode := &v1.GetFunctionAnalysisReply_FunctionNode{
		Id:        node.ID,
		Name:      node.Name,
		Package:   node.Package,
		CallCount: int32(node.CallCount),
		AvgTime:   node.AvgTime,
	}

	if len(node.Children) > 0 {
		for _, child := range node.Children {
			protoNode.Children = append(protoNode.Children, s.convertToProtoFunctionNode(child))
		}
	}

	return protoNode
}

// GetAnalysisTaskStatus 获取分析任务状态
func (s *StaticAnalysisService) GetAnalysisTaskStatus(ctx context.Context, req *v1.GetAnalysisTaskStatusRequest) (*v1.GetAnalysisTaskStatusResponse, error) {
	s.log.Infof("get analysis task status: %s", req.TaskId)
	status, err := s.uc.GetTaskStatus(req.TaskId)
	if err != nil {
		s.log.Errorf("failed to get task status: %v", err)
		return nil, fmt.Errorf("failed to get task status: %v", err)
	}

	// 获取任务进度
	progress := 0.0
	if status.Status == entity.TaskStatusProcessing {
		// 获取分析进度
		progress, err = s.uc.GetTaskProgress(req.TaskId)
		if err != nil {
			s.log.Warnf("failed to get task progress: %v", err)
			// 进度获取失败不影响状态返回
		}
	} else if status.Status == entity.TaskStatusCompleted {
		progress = 1.0
	}

	resp := &v1.GetAnalysisTaskStatusResponse{
		Status:   int32(status.Status),
		Progress: float32(progress * 100), // 转换为百分比
	}

	return resp, nil
}

// GetPackageDependencies 分页获取包依赖关系
func (s *StaticAnalysisService) GetPackageDependencies(ctx context.Context, req *v1.GetPackageDependenciesRequest) (*v1.GetPackageDependenciesResponse, error) {
	s.log.Infof("Getting package dependencies for db: %s, page: %d, pageSize: %d", req.DbPath, req.Page, req.PageSize)

	// 验证文件是否存在
	if _, err := os.Stat(req.DbPath); err != nil {
		s.log.Errorf("Database file not found: %s", req.DbPath)
		return nil, fmt.Errorf("database file not found: %s", req.DbPath)
	}

	// 获取数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(req.DbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	// 获取所有函数调用边
	edges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		s.log.Errorf("Failed to get function edges: %v", err)
		return nil, fmt.Errorf("Failed to get function edges: %v", err)
	}

	// 统计包依赖关系
	packageDeps := make(map[string]map[string]int)
	for _, edge := range edges {
		// 获取调用方和被调用方的节点
		caller, err := funcNodeDB.GetFuncNodeByKey(edge.CallerKey)
		if err != nil || caller == nil {
			continue
		}

		callee, err := funcNodeDB.GetFuncNodeByKey(edge.CalleeKey)
		if err != nil || callee == nil {
			continue
		}

		// 如果调用方和被调用方的包不同，则记录包依赖关系
		if caller.Pkg != callee.Pkg {
			if _, ok := packageDeps[caller.Pkg]; !ok {
				packageDeps[caller.Pkg] = make(map[string]int)
			}
			packageDeps[caller.Pkg][callee.Pkg]++
		}
	}

	// 转换包依赖关系为结果格式
	var packageDependencies []*v1.PackageDependency

	for source, targets := range packageDeps {
		for target, count := range targets {
			packageDependencies = append(packageDependencies, &v1.PackageDependency{
				Source: source,
				Target: target,
				Count:  int32(count),
			})
		}
	}

	// 按依赖关系数量排序
	sort.Slice(packageDependencies, func(i, j int) bool {
		return packageDependencies[i].Count > packageDependencies[j].Count
	})

	// 计算总数
	total := len(packageDependencies)

	// 计算分页
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	// 计算起始和结束索引
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= total {
		startIndex = 0
		endIndex = 0
	}

	if endIndex > total {
		endIndex = total
	}

	// 获取当前页的数据
	var pagedDependencies []*v1.PackageDependency
	if startIndex < endIndex {
		pagedDependencies = packageDependencies[startIndex:endIndex]
	}

	s.log.Infof("Returning %d package dependencies (total: %d)", len(pagedDependencies), total)

	return &v1.GetPackageDependenciesResponse{
		Dependencies: pagedDependencies,
		Total:        int32(total),
		Page:         int32(page),
		PageSize:     int32(pageSize),
		PageCount:    int32((total + pageSize - 1) / pageSize),
	}, nil
}

// SearchFunctions 模糊搜索函数
func (s *StaticAnalysisService) SearchFunctions(ctx context.Context, req *v1.SearchFunctionsRequest) (*v1.SearchFunctionsResponse, error) {
	s.log.Infof("Searching functions in db: %s, query: %s", req.DbPath, req.Query)

	// 验证文件是否存在
	if _, err := os.Stat(req.DbPath); err != nil {
		s.log.Errorf("Database file not found: %s", req.DbPath)
		return nil, fmt.Errorf("Database file not found: %s", req.DbPath)
	}

	// 获取数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(req.DbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	// 获取所有函数节点
	nodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		s.log.Errorf("Failed to get function nodes: %v", err)
		return nil, fmt.Errorf("Failed to get function nodes: %v", err)
	}
	s.log.Infof("Total function nodes: %d", len(nodes))

	// 获取所有函数调用边
	edges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		s.log.Errorf("Failed to get function edges: %v", err)
		return nil, fmt.Errorf("Failed to get function edges: %v", err)
	}
	s.log.Infof("Total function edges: %d", len(edges))

	// 统计函数被调用次数
	funcCallCounts := make(map[string]int)
	for _, edge := range edges {
		funcCallCounts[edge.CalleeKey]++
	}

	// 模糊搜索函数
	query := strings.ToLower(req.Query)
	var matchedFunctions []*v1.FunctionInfo

	for _, node := range nodes {
		// 检查函数名或包名是否包含查询字符串
		if strings.Contains(strings.ToLower(node.Name), query) ||
			strings.Contains(strings.ToLower(node.Pkg), query) {
			matchedFunctions = append(matchedFunctions, &v1.FunctionInfo{
				Key:       node.Key,
				Name:      node.Name,
				Package:   node.Pkg,
				CallCount: int32(funcCallCounts[node.Key]),
			})
		}
	}

	// 按调用次数排序
	sort.Slice(matchedFunctions, func(i, j int) bool {
		return matchedFunctions[i].CallCount > matchedFunctions[j].CallCount
	})

	// 限制返回的结果数量
	maxResults := 50
	if len(matchedFunctions) > maxResults {
		matchedFunctions = matchedFunctions[:maxResults]
	}

	s.log.Infof("Found %d matching functions for query: %s", len(matchedFunctions), req.Query)

	// 打印前几个匹配结果的详细信息
	if len(matchedFunctions) > 0 {
		for i := 0; i < min(5, len(matchedFunctions)); i++ {
			s.log.Infof("Match %d: Key=%s, Name=%s, Package=%s, CallCount=%d",
				i+1,
				matchedFunctions[i].Key,
				matchedFunctions[i].Name,
				matchedFunctions[i].Package,
				matchedFunctions[i].CallCount)
		}
	}

	return &v1.SearchFunctionsResponse{
		Functions: matchedFunctions,
	}, nil
}

// min returns the smaller of x or y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// GetFunctionUpstream 获取函数上游调用关系
func (s *StaticAnalysisService) GetFunctionUpstream(ctx context.Context, req *v1.GetFunctionUpstreamRequest) (*v1.GetFunctionUpstreamResponse, error) {
	s.log.Infof("Getting function upstream for db: %s, functionKey: %s", req.DbPath, req.FunctionKey)

	// 验证文件是否存在
	if _, err := os.Stat(req.DbPath); err != nil {
		s.log.Errorf("Database file not found: %s", req.DbPath)
		return nil, fmt.Errorf("database file not found: %s", req.DbPath)
	}

	// 获取数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(req.DbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	// 获取目标函数节点
	targetNode, err := funcNodeDB.GetFuncNodeByKey(req.FunctionKey)
	if err != nil || targetNode == nil {
		s.log.Errorf("Function not found: %s", req.FunctionKey)
		return nil, fmt.Errorf("Function not found: %s", req.FunctionKey)
	}

	// 获取所有函数调用边
	allEdges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		s.log.Errorf("Failed to get function edges: %v", err)
		return nil, fmt.Errorf("Failed to get function edges: %v", err)
	}

	// 统计函数被调用次数
	funcCallCounts := make(map[string]int)
	for _, edge := range allEdges {
		funcCallCounts[edge.CalleeKey]++
	}

	// 创建调用方到被调用方的映射
	callerToCallees := make(map[string][]string)
	calleeToCallers := make(map[string][]string)

	for _, edge := range allEdges {
		callerToCallees[edge.CallerKey] = append(callerToCallees[edge.CallerKey], edge.CalleeKey)
		calleeToCallers[edge.CalleeKey] = append(calleeToCallers[edge.CalleeKey], edge.CallerKey)
	}

	// 使用递归DFS算法获取所有上游调用节点
	visited := make(map[string]bool)
	var graphNodes []*v1.GraphNode
	var graphEdges []*v1.GraphEdge

	// 添加目标节点
	graphNodes = append(graphNodes, &v1.GraphNode{
		Id:        targetNode.Key,
		Name:      targetNode.Name,
		Package:   targetNode.Pkg,
		CallCount: int32(funcCallCounts[targetNode.Key]),
	})

	// 记录已访问的节点
	visited[targetNode.Key] = true

	// 递归查找所有上游调用
	s.findAllUpstreamCalls(targetNode, funcNodeDB, calleeToCallers, funcCallCounts, visited, &graphNodes, &graphEdges)

	// 查找最顶层调用函数（没有被其他函数调用的函数）
	topLevelNodes := s.findTopLevelCallers(graphNodes, graphEdges)
	s.log.Infof("Found %d top level callers", len(topLevelNodes))

	s.log.Infof("Found %d nodes and %d edges in the upstream graph", len(graphNodes), len(graphEdges))

	return &v1.GetFunctionUpstreamResponse{
		Nodes: graphNodes,
		Edges: graphEdges,
	}, nil
}

// findAllUpstreamCalls 递归查找所有上游调用
func (s *StaticAnalysisService) findAllUpstreamCalls(
	currentNode *dos.FuncNode,
	funcNodeDB repo.StaticDBStore,
	calleeToCallers map[string][]string,
	funcCallCounts map[string]int,
	visited map[string]bool,
	graphNodes *[]*v1.GraphNode,
	graphEdges *[]*v1.GraphEdge,
) {
	// 获取调用当前节点的所有函数
	callers, ok := calleeToCallers[currentNode.Key]
	if !ok {
		return
	}

	for _, callerKey := range callers {
		// 如果已经访问过，则跳过
		if visited[callerKey] {
			continue
		}

		// 获取调用方节点
		callerNode, err := funcNodeDB.GetFuncNodeByKey(callerKey)
		if err != nil || callerNode == nil {
			continue
		}

		// 标记为已访问
		visited[callerKey] = true

		// 添加节点
		*graphNodes = append(*graphNodes, &v1.GraphNode{
			Id:        callerNode.Key,
			Name:      callerNode.Name,
			Package:   callerNode.Pkg,
			CallCount: int32(funcCallCounts[callerNode.Key]),
		})

		// 添加边
		*graphEdges = append(*graphEdges, &v1.GraphEdge{
			Source: callerNode.Key,
			Target: currentNode.Key,
			Value:  1,
		})

		// 递归查找上游调用
		s.findAllUpstreamCalls(callerNode, funcNodeDB, calleeToCallers, funcCallCounts, visited, graphNodes, graphEdges)
	}
}

// findTopLevelCallers 查找最顶层调用函数（没有被其他函数调用的函数）
func (s *StaticAnalysisService) findTopLevelCallers(nodes []*v1.GraphNode, edges []*v1.GraphEdge) []*v1.GraphNode {
	// 创建一个映射，记录每个节点是否被其他节点调用
	hasCallers := make(map[string]bool)

	// 遍历所有边，记录被调用的节点
	for _, edge := range edges {
		hasCallers[edge.Target] = true
	}

	// 找出没有被调用的节点（最顶层调用函数）
	var topLevelNodes []*v1.GraphNode
	for _, node := range nodes {
		if !hasCallers[node.Id] && node.Id != "" {
			topLevelNodes = append(topLevelNodes, node)
		}
	}

	return topLevelNodes
}

// GetFunctionDownstream 获取函数下游调用关系
func (s *StaticAnalysisService) GetFunctionDownstream(ctx context.Context, req *v1.GetFunctionDownstreamRequest) (*v1.GetFunctionDownstreamResponse, error) {
	s.log.Infof("Getting function downstream for db: %s, functionKey: %s", req.DbPath, req.FunctionKey)

	// 验证文件是否存在
	if _, err := os.Stat(req.DbPath); err != nil {
		s.log.Errorf("Database file not found: %s", req.DbPath)
		return nil, fmt.Errorf("database file not found: %s", req.DbPath)
	}

	// 获取数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(req.DbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	// 获取目标函数节点
	targetNode, err := funcNodeDB.GetFuncNodeByKey(req.FunctionKey)
	if err != nil || targetNode == nil {
		s.log.Errorf("Function not found: %s", req.FunctionKey)
		return nil, fmt.Errorf("Function not found: %s", req.FunctionKey)
	}

	// 获取所有函数调用边
	allEdges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		s.log.Errorf("Failed to get function edges: %v", err)
		return nil, fmt.Errorf("Failed to get function edges: %v", err)
	}

	// 统计函数被调用次数
	funcCallCounts := make(map[string]int)
	for _, edge := range allEdges {
		funcCallCounts[edge.CalleeKey]++
	}

	// 创建调用方到被调用方的映射
	callerToCallees := make(map[string][]string)
	calleeToCallers := make(map[string][]string)

	for _, edge := range allEdges {
		callerToCallees[edge.CallerKey] = append(callerToCallees[edge.CallerKey], edge.CalleeKey)
		calleeToCallers[edge.CalleeKey] = append(calleeToCallers[edge.CalleeKey], edge.CallerKey)
	}

	// 使用递归DFS算法获取所有下游调用节点
	visited := make(map[string]bool)
	var graphNodes []*v1.GraphNode
	var graphEdges []*v1.GraphEdge

	// 添加目标节点
	graphNodes = append(graphNodes, &v1.GraphNode{
		Id:        targetNode.Key,
		Name:      targetNode.Name,
		Package:   targetNode.Pkg,
		CallCount: int32(funcCallCounts[targetNode.Key]),
	})

	// 记录已访问的节点
	visited[targetNode.Key] = true

	// 递归查找所有下游调用
	s.findAllDownstreamCalls(targetNode, funcNodeDB, callerToCallees, funcCallCounts, visited, &graphNodes, &graphEdges)

	// 查找最底层被调用函数（不调用其他函数的函数）
	leafNodes := s.findLeafNodes(graphNodes, graphEdges)
	s.log.Infof("Found %d leaf nodes", len(leafNodes))

	s.log.Infof("Found %d nodes and %d edges in the downstream graph", len(graphNodes), len(graphEdges))

	return &v1.GetFunctionDownstreamResponse{
		Nodes: graphNodes,
		Edges: graphEdges,
	}, nil
}

// findAllDownstreamCalls 递归查找所有下游调用
func (s *StaticAnalysisService) findAllDownstreamCalls(
	currentNode *dos.FuncNode,
	funcNodeDB repo.StaticDBStore,
	callerToCallees map[string][]string,
	funcCallCounts map[string]int,
	visited map[string]bool,
	graphNodes *[]*v1.GraphNode,
	graphEdges *[]*v1.GraphEdge,
) {
	// 获取当前节点调用的所有函数
	callees, ok := callerToCallees[currentNode.Key]
	if !ok {
		return
	}

	for _, calleeKey := range callees {
		// 如果已经访问过，则跳过
		if visited[calleeKey] {
			continue
		}

		// 获取被调用方节点
		calleeNode, err := funcNodeDB.GetFuncNodeByKey(calleeKey)
		if err != nil || calleeNode == nil {
			continue
		}

		// 标记为已访问
		visited[calleeKey] = true

		// 添加节点
		*graphNodes = append(*graphNodes, &v1.GraphNode{
			Id:        calleeNode.Key,
			Name:      calleeNode.Name,
			Package:   calleeNode.Pkg,
			CallCount: int32(funcCallCounts[calleeNode.Key]),
		})

		// 添加边
		*graphEdges = append(*graphEdges, &v1.GraphEdge{
			Source: currentNode.Key,
			Target: calleeNode.Key,
			Value:  1,
		})

		// 递归查找下游调用
		s.findAllDownstreamCalls(calleeNode, funcNodeDB, callerToCallees, funcCallCounts, visited, graphNodes, graphEdges)
	}
}

// findLeafNodes 查找叶子节点（不调用其他函数的函数）
func (s *StaticAnalysisService) findLeafNodes(nodes []*v1.GraphNode, edges []*v1.GraphEdge) []*v1.GraphNode {
	// 创建一个映射，记录每个节点是否调用其他节点
	hasCallees := make(map[string]bool)

	// 遍历所有边，记录调用其他节点的节点
	for _, edge := range edges {
		hasCallees[edge.Source] = true
	}

	// 找出没有调用其他节点的节点（叶子节点）
	var leafNodes []*v1.GraphNode
	for _, node := range nodes {
		if !hasCallees[node.Id] {
			leafNodes = append(leafNodes, node)
		}
	}

	return leafNodes
}

// GetFunctionFullChain 获取函数全链路调用关系
func (s *StaticAnalysisService) GetFunctionFullChain(ctx context.Context, req *v1.GetFunctionFullChainRequest) (*v1.GetFunctionFullChainResponse, error) {
	s.log.Infof("Getting function full chain for db: %s, functionKey: %s", req.DbPath, req.FunctionKey)

	// 验证文件是否存在
	if _, err := os.Stat(req.DbPath); err != nil {
		s.log.Errorf("Database file not found: %s", req.DbPath)
		return nil, fmt.Errorf("database file not found: %s", req.DbPath)
	}

	// 获取数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(req.DbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	// 获取目标函数节点
	targetNode, err := funcNodeDB.GetFuncNodeByKey(req.FunctionKey)
	if err != nil || targetNode == nil {
		s.log.Errorf("Function not found: %s", req.FunctionKey)
		return nil, fmt.Errorf("Function not found: %s", req.FunctionKey)
	}

	// 获取所有函数调用边
	allEdges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		s.log.Errorf("Failed to get function edges: %v", err)
		return nil, fmt.Errorf("Failed to get function edges: %v", err)
	}

	// 统计函数被调用次数
	funcCallCounts := make(map[string]int)
	for _, edge := range allEdges {
		funcCallCounts[edge.CalleeKey]++
	}

	// 创建调用方到被调用方的映射
	callerToCallees := make(map[string][]string)
	calleeToCallers := make(map[string][]string)

	for _, edge := range allEdges {
		callerToCallees[edge.CallerKey] = append(callerToCallees[edge.CallerKey], edge.CalleeKey)
		calleeToCallers[edge.CalleeKey] = append(calleeToCallers[edge.CalleeKey], edge.CallerKey)
	}

	// 使用两个独立的DFS算法获取上游和下游调用节点
	upstreamVisited := make(map[string]bool)
	downstreamVisited := make(map[string]bool)
	var graphNodes []*v1.GraphNode
	var graphEdges []*v1.GraphEdge

	// 添加目标节点
	graphNodes = append(graphNodes, &v1.GraphNode{
		Id:        targetNode.Key,
		Name:      targetNode.Name,
		Package:   targetNode.Pkg,
		CallCount: int32(funcCallCounts[targetNode.Key]),
	})

	// 记录已访问的节点
	upstreamVisited[targetNode.Key] = true
	downstreamVisited[targetNode.Key] = true

	// 递归查找所有上游调用
	s.findAllUpstreamCalls(targetNode, funcNodeDB, calleeToCallers, funcCallCounts, upstreamVisited, &graphNodes, &graphEdges)

	// 递归查找所有下游调用
	s.findAllDownstreamCalls(targetNode, funcNodeDB, callerToCallees, funcCallCounts, downstreamVisited, &graphNodes, &graphEdges)

	// 合并节点（去重）
	mergedNodes := s.mergeNodes(graphNodes)

	s.log.Infof("Found %d nodes and %d edges in the full chain graph", len(mergedNodes), len(graphEdges))

	return &v1.GetFunctionFullChainResponse{
		Nodes: mergedNodes,
		Edges: graphEdges,
	}, nil
}

// mergeNodes 合并节点（去重）
func (s *StaticAnalysisService) mergeNodes(nodes []*v1.GraphNode) []*v1.GraphNode {
	nodeMap := make(map[string]*v1.GraphNode)

	for _, node := range nodes {
		if _, exists := nodeMap[node.Id]; !exists {
			nodeMap[node.Id] = node
		}
	}

	var mergedNodes []*v1.GraphNode
	for _, node := range nodeMap {
		mergedNodes = append(mergedNodes, node)
	}

	return mergedNodes
}

// GetTreeGraph 获取静态分析树状图数据
func (s *StaticAnalysisService) GetTreeGraph(ctx context.Context, req *v1.GetTreeGraphReq) (*v1.GetTreeGraphReply, error) {
	s.log.Infof("get tree graph, function: %s, dbpath: %s", req.FunctionName, req.DbPath)

	// 调用业务逻辑获取树状图数据
	treeGraph, err := s.uc.GetTreeGraph(req.FunctionName, req.DbPath)
	if err != nil {
		s.log.Errorf("get tree graph failed: %v", err)
		return nil, err
	}

	// 转换为API响应格式
	reply := &v1.GetTreeGraphReply{
		Root: s.convertTreeNodeToProto(treeGraph.Root),
	}

	return reply, nil
}

// 将实体TreeNode转换为proto TreeNode
func (s *StaticAnalysisService) convertTreeNodeToProto(node *entity.TreeNode) *v1.TreeNode {
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
			protoNode.Children = append(protoNode.Children, s.convertTreeNodeToProto(child))
		}
	}

	return protoNode
}
