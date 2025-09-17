package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/toheart/goanalysis/api/staticanalysis/v1"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/session"
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

// getDBPathFromContext 从context中获取数据库路径的辅助函数
func (s *StaticAnalysisService) getDBPathFromContext(ctx context.Context) (string, error) {
	dbPath, exists := session.GetDBPathFromGRPCContext(ctx)
	if !exists {
		return "", errors.New("database path not found in session, please set analysis path first")
	}
	return dbPath, nil
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

	s.log.Infof("Database analysis completed, found %d functions, %d calls, %d packages",
		result.TotalFunctions, result.TotalCalls, result.TotalPackages)

	// 注意：数据库路径的session设置需要通过HTTP层处理，这里只是分析数据库内容
	s.log.Infof("Analysis completed for database: %s", req.DbPath)

	return &v1.AnalyzeDbFileResponse{
		TotalFunctions: int32(result.TotalFunctions),
		TotalCalls:     int32(result.TotalCalls),
		TotalPackages:  int32(result.TotalPackages),
	}, nil
}

// 数据库文件结构
type staticDbFile struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	CreateTime time.Time `json:"createTime"`
}

// 分析结果结构（简化版：仅包含基本统计信息）
type staticAnalysisResult struct {
	TotalFunctions int `json:"totalFunctions"`
	TotalCalls     int `json:"totalCalls"`
	TotalPackages  int `json:"totalPackages"`
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

	var dbFiles []staticDbFile
	err := filepath.Walk(dbPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.log.Warnf("Failed to access path %s: %v", path, err)
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == ".db" {
			dbFiles = append(dbFiles, staticDbFile{
				Path:       path,
				Name:       info.Name(),
				Size:       info.Size(),
				CreateTime: info.ModTime(),
			})
		}
		return nil
	})

	if err != nil {
		s.log.Errorf("Failed to walk database directory: %v", err)
		return nil, fmt.Errorf("Failed to walk database directory: %v", err)
	}

	// 按创建时间降序排序，最新的文件排在前面
	sort.Slice(dbFiles, func(i, j int) bool {
		return dbFiles[i].CreateTime.After(dbFiles[j].CreateTime)
	})

	s.log.Infof("Found %d database files", len(dbFiles))
	return dbFiles, nil
}

// 从数据库中读取实际的调用图数据（优化后：仅统计基本信息，移除包依赖和热点函数分析）
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

	// 获取所有函数节点（只为统计数量和包数量）
	nodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		s.log.Errorf("Failed to get function nodes: %v", err)
		return nil, fmt.Errorf("Failed to get function nodes: %v", err)
	}

	// 获取所有函数调用边（只为统计数量）
	edges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		s.log.Errorf("Failed to get function edges: %v", err)
		return nil, fmt.Errorf("Failed to get function edges: %v", err)
	}

	// 统计包的数量（优化：使用map去重）
	packages := make(map[string]bool)
	for _, node := range nodes {
		if node.Pkg != "" {
			packages[node.Pkg] = true
		}
	}

	result := &staticAnalysisResult{
		TotalFunctions: len(nodes),
		TotalCalls:     len(edges),
		TotalPackages:  len(packages),
	}

	s.log.Infof("Database analysis completed, found %d functions, %d calls, %d packages",
		result.TotalFunctions, result.TotalCalls, result.TotalPackages)
	return result, nil
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

// SearchFunctions 模糊搜索函数
func (s *StaticAnalysisService) SearchFunctions(ctx context.Context, req *v1.SearchFunctionsRequest) (*v1.SearchFunctionsResponse, error) {
	// 从gRPC metadata中获取数据库路径
	dbPath, err := s.getDBPathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Infof("Searching functions in db: %s, query: %s", dbPath, req.Query)

	// 验证文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("Database file not found: %s", dbPath)
		return nil, fmt.Errorf("Database file not found: %s", dbPath)
	}

	// 获取数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	// 设置查询限制
	maxResults := 50

	// 使用数据库模糊查询
	nodes, err := funcNodeDB.SearchFuncNodes(req.Query, maxResults)
	if err != nil {
		s.log.Errorf("Failed to search function nodes: %v", err)
		return nil, fmt.Errorf("Failed to search function nodes: %v", err)
	}

	// 转换为API响应格式
	var matchedFunctions []*v1.FunctionInfo
	for _, node := range nodes {
		matchedFunctions = append(matchedFunctions, &v1.FunctionInfo{
			Key:       node.Key,
			Name:      node.Name,
			Package:   node.Pkg,
			CallCount: 0, // 不计算调用次数，提高性能
		})
	}

	s.log.Infof("Found %d matching functions for query: %s", len(matchedFunctions), req.Query)

	// 打印前几个匹配结果的详细信息
	if len(matchedFunctions) > 0 {
		for i := 0; i < min(5, len(matchedFunctions)); i++ {
			s.log.Infof("Match %d: Key=%s, Name=%s, Package=%s",
				i+1,
				matchedFunctions[i].Key,
				matchedFunctions[i].Name,
				matchedFunctions[i].Package)
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

// ================ 可视化分析API实现 ================

// GetGlobalStatistics 获取全局统计信息
func (s *StaticAnalysisService) GetGlobalStatistics(ctx context.Context, req *v1.GetGlobalStatisticsRequest) (*v1.GetGlobalStatisticsResponse, error) {
	// 从gRPC metadata中获取数据库路径
	dbPath, err := s.getDBPathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Infof("get global statistics for db: %s", dbPath)

	// 验证数据库文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("database file not found: %s", dbPath)
		return nil, fmt.Errorf("database file not found: %s", dbPath)
	}

	// 设置数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("failed to get database connection: %v", err)
		return nil, fmt.Errorf("failed to get database connection: %v", err)
	}

	// 添加超时控制，为大型数据库设置更长的超时时间
	// 使用 channel 来实现超时控制
	resultChan := make(chan *v1.GetGlobalStatisticsResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		response, err := s.uc.GetGlobalStatistics(ctx, funcNodeDB)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- response
		}
	}()

	// 等待结果或超时
	select {
	case response := <-resultChan:
		s.log.Infof("Retrieved global statistics: %d functions, %d packages, %d edges",
			response.TotalFunctions, response.TotalPackages, response.TotalEdges)
		return response, nil
	case err := <-errorChan:
		s.log.Errorf("failed to get global statistics: %v", err)
		return nil, err
	case <-ctx.Done():
		s.log.Errorf("global statistics calculation timed out or was cancelled")
		return nil, fmt.Errorf("global statistics calculation timed out")
	}
}

// GetCallGraph 获取调用图数据
func (s *StaticAnalysisService) GetCallGraph(ctx context.Context, req *v1.GetCallGraphRequest) (*v1.GetCallGraphResponse, error) {
	// 从gRPC metadata中获取数据库路径
	dbPath, err := s.getDBPathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Infof("get call graph for db: %s, root: %s, max_depth: %d",
		dbPath, req.RootFunction, req.MaxDepth)

	// 验证数据库文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("Database file not found: %s", dbPath)
		return nil, fmt.Errorf("Database file not found: %s", dbPath)
	}

	// 设置默认值
	if req.MaxDepth <= 0 {
		req.MaxDepth = 3
	}

	// 设置数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	response, err := s.uc.GetCallGraph(ctx, req, funcNodeDB)
	if err != nil {
		s.log.Errorf("Failed to get call graph: %v", err)
		return nil, err
	}

	s.log.Infof("Retrieved call graph: %d nodes, %d edges",
		response.TotalNodes, response.TotalEdges)

	return response, nil
}

// GetPackageDependencies 获取包依赖关系
func (s *StaticAnalysisService) GetPackageDependencies(ctx context.Context, req *v1.GetPackageDependenciesRequest) (*v1.GetPackageDependenciesResponse, error) {
	// 从gRPC metadata中获取数据库路径
	dbPath, err := s.getDBPathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Infof("get package dependencies for db: %s", dbPath)

	// 验证数据库文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("database file not found: %s", dbPath)
		return nil, fmt.Errorf("database file not found: %s", dbPath)
	}

	// 设置数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("failed to get database connection: %v", err)
		return nil, fmt.Errorf("failed to get database connection: %v", err)
	}

	// 添加超时控制，为大型数据库设置更长的超时时间
	resultChan := make(chan *v1.GetPackageDependenciesResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		response, err := s.uc.GetPackageDependencies(ctx, funcNodeDB)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- response
		}
	}()

	// 等待结果或超时
	select {
	case response := <-resultChan:
		s.log.Infof("Retrieved package dependencies: %d packages, %d dependencies",
			len(response.Packages), len(response.Dependencies))
		return response, nil
	case err := <-errorChan:
		s.log.Errorf("failed to get package dependencies: %v", err)
		return nil, err
	case <-ctx.Done():
		s.log.Errorf("package dependencies calculation timed out or was cancelled")
		return nil, fmt.Errorf("package dependencies calculation timed out")
	}
}

// GetHotFunctions 获取热点函数分析
func (s *StaticAnalysisService) GetHotFunctions(ctx context.Context, req *v1.GetHotFunctionsRequest) (*v1.GetHotFunctionsResponse, error) {
	// 从gRPC metadata中获取数据库路径
	dbPath, err := s.getDBPathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Infof("get hot functions for db: %s, limit: %d, sort_by: %s",
		dbPath, req.Limit, req.SortBy)

	// 验证数据库文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("Database file not found: %s", dbPath)
		return nil, fmt.Errorf("Database file not found: %s", dbPath)
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.SortBy == "" {
		req.SortBy = "importance"
	}

	// 设置数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	response, err := s.uc.GetHotFunctions(ctx, req, funcNodeDB)
	if err != nil {
		s.log.Errorf("Failed to get hot functions: %v", err)
		return nil, err
	}

	s.log.Infof("Retrieved hot functions: %d functions", len(response.Functions))

	return response, nil
}

// GetFunctionDetails 获取函数详细信息
func (s *StaticAnalysisService) GetFunctionDetails(ctx context.Context, req *v1.GetFunctionDetailsRequest) (*v1.GetFunctionDetailsResponse, error) {
	// 从gRPC metadata中获取数据库路径
	dbPath, err := s.getDBPathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Infof("get function details for db: %s, function: %s",
		dbPath, req.FunctionKey)

	// 验证数据库文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("Database file not found: %s", dbPath)
		return nil, fmt.Errorf("Database file not found: %s", dbPath)
	}

	// 设置数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	response, err := s.uc.GetFunctionDetails(ctx, req.FunctionKey, funcNodeDB)
	if err != nil {
		s.log.Errorf("Failed to get function details: %v", err)
		return nil, err
	}

	s.log.Infof("Retrieved function details: %s, callers: %d, callees: %d",
		response.Function.Name, len(response.Callers), len(response.Callees))

	return response, nil
}

// GetFunctionCallPaths 获取函数调用路径
func (s *StaticAnalysisService) GetFunctionCallPaths(ctx context.Context, req *v1.GetFunctionCallPathsRequest) (*v1.GetFunctionCallPathsResponse, error) {
	// 从gRPC metadata中获取数据库路径
	dbPath, err := s.getDBPathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Infof("get function call paths for db: %s, from: %s, to: %s",
		dbPath, req.FromFunction, req.ToFunction)

	// 验证数据库文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("Database file not found: %s", dbPath)
		return nil, fmt.Errorf("Database file not found: %s", dbPath)
	}

	// 设置默认值
	if req.MaxPaths <= 0 {
		req.MaxPaths = 5
	}
	if req.MaxDepth <= 0 {
		req.MaxDepth = 10
	}

	// 设置数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("Failed to get database connection: %v", err)
	}

	response, err := s.uc.GetFunctionCallPaths(ctx, req, funcNodeDB)
	if err != nil {
		s.log.Errorf("Failed to get function call paths: %v", err)
		return nil, err
	}

	s.log.Infof("Retrieved function call paths: %d paths", len(response.Paths))

	return response, nil
}

// GetInitFunctions 获取Init函数列表
func (s *StaticAnalysisService) GetInitFunctions(ctx context.Context, req *v1.GetInitFunctionsRequest) (*v1.GetInitFunctionsResponse, error) {
	// 从gRPC metadata中获取数据库路径
	dbPath, err := s.getDBPathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Infof("get init functions for db: %s, limit: %d, package_filter: %s",
		dbPath, req.Limit, req.PackageFilter)

	// 验证数据库文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("database file not found: %s", dbPath)
		return nil, fmt.Errorf("database file not found: %s", dbPath)
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 50 // 默认返回50个
	}

	// 设置数据库连接
	funcNodeDB, err := s.uc.GetFuncNodeDB(dbPath)
	if err != nil {
		s.log.Errorf("failed to get database connection: %v", err)
		return nil, fmt.Errorf("failed to get database connection: %v", err)
	}

	// 添加超时控制
	resultChan := make(chan *v1.GetInitFunctionsResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		response, err := s.uc.GetInitFunctions(ctx, req, funcNodeDB)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- response
		}
	}()

	// 等待结果或超时
	select {
	case response := <-resultChan:
		s.log.Infof("Retrieved init functions: %d functions (total: %d)",
			len(response.InitFunctions), response.TotalCount)
		return response, nil
	case err := <-errorChan:
		s.log.Errorf("failed to get init functions: %v", err)
		return nil, err
	case <-ctx.Done():
		s.log.Errorf("init functions retrieval timed out or was cancelled")
		return nil, fmt.Errorf("init functions retrieval timed out")
	}
}
