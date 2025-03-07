package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/toheart/goanalysis/api/staticanalysis/v1"
	"github.com/toheart/goanalysis/internal/biz/entity"
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
	return &StaticAnalysisService{uc: uc, log: log.NewHelper(logger)}
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
	// 验证请求参数
	if req.Path == "" {
		s.log.Error("project path can't be empty")
		return &v1.AnalyzeProjectPathResponse{
			Success: false,
			Message: "project path can't be empty",
		}, nil
	}

	s.log.Infof("开始分析项目路径: %s", req.Path)

	// 1. 验证路径是否存在
	if !s.verifyProjectPath(req.Path) {
		s.log.Errorf("project path not found: %s", req.Path)
		return &v1.AnalyzeProjectPathResponse{
			Success: false,
			Message: "project path not found",
		}, nil
	}

	// 2. 获取路径的最后一个目录名
	dir := filepath.Base(req.Path)
	if dir == "." || dir == "/" {
		// 处理路径末尾有斜杠的情况
		dir = filepath.Base(filepath.Dir(req.Path))
	}

	// 3. 生成数据库路径，添加时间戳
	timestamp := time.Now().Format("20060102_150405")
	dbName := fmt.Sprintf("%s_%s.db", dir, timestamp)
	dbPath := filepath.Join(s.uc.GetStaticDBPath(), dbName)

	s.log.Infof("will use db path: %s", dbPath)

	// 4. 调用callgraph功能分析项目
	s.log.Info("start callgraph analysis...")
	err := s.runCallgraphAnalysis(req.Path, dbPath)
	if err != nil {
		s.log.Errorf("analysis failed: %v", err)
		return &v1.AnalyzeProjectPathResponse{
			Success: false,
			Message: fmt.Sprintf("analysis failed: %v", err),
		}, nil
	}

	s.log.Infof("project analysis completed, db path: %s", dbPath)
	return &v1.AnalyzeProjectPathResponse{
		Success: true,
		Message: "analysis completed",
		DbPath:  dbPath,
	}, nil
}

// 验证项目路径是否存在
func (s *StaticAnalysisService) verifyProjectPath(path string) bool {
	if path == "" {
		s.log.Error("项目路径为空")
		return false
	}

	s.log.Infof("验证项目路径: %s", path)

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.log.Errorf("project path not found: %s", path)
		} else {
			s.log.Errorf("verify project path failed: %v", err)
		}
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

	s.log.Infof("开始分析数据库文件: %s", req.DbPath)

	// 验证文件是否存在
	if _, err := os.Stat(req.DbPath); err != nil {
		if os.IsNotExist(err) {
			s.log.Errorf("db file not found: %s", req.DbPath)
			return nil, fmt.Errorf("db file not found: %s", req.DbPath)
		}
		s.log.Errorf("check db file failed: %v", err)
		return nil, fmt.Errorf("check db file failed: %v", err)
	}

	result, err := s.analyzeDb(req.DbPath)
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

	s.log.Infof("数据库分析完成，共 %d 个函数，%d 个调用，%d 个包",
		result.TotalFunctions, result.TotalCalls, result.TotalPackages)

	return &v1.AnalyzeDbFileResponse{
		TotalFunctions:      int32(result.TotalFunctions),
		TotalCalls:          int32(result.TotalCalls),
		TotalPackages:       int32(result.TotalPackages),
		PackageDependencies: packageDeps,
		HotFunctions:        hotFuncs,
	}, nil
}

// GetHotFunctions 获取热点函数分析数据
func (s *StaticAnalysisService) GetHotFunctions(ctx context.Context, req *v1.GetHotFunctionsReq) (*v1.GetHotFunctionsReply, error) {
	hotFunctions, err := s.uc.GetHotFunctions(req.SortBy)
	if err != nil {
		s.log.Errorf("获取热点函数失败: %v", err)
		return nil, err
	}

	var functions []*v1.GetHotFunctionsReply_HotFunction
	for _, fn := range hotFunctions {
		functions = append(functions, &v1.GetHotFunctionsReply_HotFunction{
			Name:      fn.Name,
			Package:   fn.Package,
			CallCount: int32(fn.CallCount),
			TotalTime: fn.TotalTime,
			AvgTime:   fn.AvgTime,
		})
	}

	return &v1.GetHotFunctionsReply{
		Functions: functions,
	}, nil
}

// GetFunctionAnalysis 获取函数调用关系分析
func (s *StaticAnalysisService) GetFunctionAnalysis(ctx context.Context, req *v1.GetFunctionAnalysisReq) (*v1.GetFunctionAnalysisReply, error) {
	if req.FunctionName == "" {
		s.log.Error("函数名称不能为空")
		return nil, fmt.Errorf("函数名称不能为空")
	}

	// 验证查询类型
	if req.Type != "" && req.Type != "caller" && req.Type != "callee" {
		s.log.Errorf("无效的查询类型: %s, 应为 'caller' 或 'callee'", req.Type)
		return nil, fmt.Errorf("无效的查询类型: %s, 应为 'caller' 或 'callee'", req.Type)
	}

	// 如果未指定类型，默认为 "callee"
	queryType := req.Type
	if queryType == "" {
		queryType = "callee"
	}

	s.log.Infof("分析函数 %s 的%s关系", req.FunctionName, queryType)
	functionNodes, err := s.uc.GetFunctionAnalysis(req.FunctionName, queryType, req.Path)
	if err != nil {
		s.log.Errorf("获取函数调用关系分析失败: %v", err)
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
		s.log.Error("函数名称不能为空")
		return nil, fmt.Errorf("函数名称不能为空")
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
		s.log.Errorf("无效的方向: %s, 应为 'caller', 'callee' 或 'both'", direction)
		return nil, fmt.Errorf("无效的方向: %s, 应为 'caller', 'callee' 或 'both'", direction)
	}

	s.log.Infof("获取函数 %s 的调用关系图，深度: %d, 方向: %s", req.FunctionName, depth, direction)
	nodes, edges, err := s.uc.GetFunctionCallGraph(req.FunctionName, depth, direction)
	if err != nil {
		s.log.Errorf("获取函数调用关系图失败: %v", err)
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
		s.log.Errorf("确保数据库目录存在时出错: %v", err)
		return nil, fmt.Errorf("确保数据库目录存在时出错: %v", err)
	}

	s.log.Infof("从目录 %s 获取数据库文件", dbPath)

	files, err := os.ReadDir(dbPath)
	if err != nil {
		s.log.Errorf("读取数据库目录失败: %v", err)
		return nil, fmt.Errorf("读取数据库目录失败: %v", err)
	}

	var dbFiles []staticDbFile
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".db" {
			info, err := file.Info()
			if err != nil {
				s.log.Warnf("获取文件信息失败: %s, 错误: %v", file.Name(), err)
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

	s.log.Infof("找到 %d 个数据库文件", len(dbFiles))
	return dbFiles, nil
}

// 分析数据库文件
func (s *StaticAnalysisService) analyzeDb(dbPath string) (*staticAnalysisResult, error) {
	// 验证文件是否存在
	if _, err := os.Stat(dbPath); err != nil {
		return nil, err
	}

	s.log.Infof("开始分析数据库: %s", dbPath)

	// TODO: 实现实际的数据库分析逻辑
	// 这里应该连接SQLite数据库，查询函数调用关系等信息
	// 以下是示例实现，实际项目中应该替换为真实的数据库查询

	// 模拟数据库分析延迟
	time.Sleep(500 * time.Millisecond)

	// 示例返回
	result := &staticAnalysisResult{
		TotalFunctions: 100,
		TotalCalls:     500,
		TotalPackages:  20,
		PackageDependencies: []struct {
			Source string `json:"source"`
			Target string `json:"target"`
			Count  int    `json:"count"`
		}{
			{Source: "main", Target: "internal/biz", Count: 15},
			{Source: "internal/biz", Target: "internal/data", Count: 25},
			{Source: "internal/service", Target: "internal/biz", Count: 30},
			{Source: "internal/data", Target: "database/sql", Count: 18},
			{Source: "internal/server", Target: "internal/service", Count: 12},
		},
		HotFunctions: []struct {
			Name      string `json:"name"`
			CallCount int    `json:"callCount"`
		}{
			{Name: "main.main", CallCount: 1},
			{Name: "internal/biz.Process", CallCount: 45},
			{Name: "internal/data.Query", CallCount: 78},
			{Name: "internal/service.Handle", CallCount: 56},
			{Name: "internal/data.Execute", CallCount: 62},
		},
	}

	s.log.Info("数据库分析完成")
	return result, nil
}

// 运行callgraph分析
func (s *StaticAnalysisService) runCallgraphAnalysis(projectPath, dbPath string) error {
	s.log.Infof("开始对项目 %s 进行callgraph分析，数据库路径: %s", projectPath, dbPath)

	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		s.log.Errorf("创建数据库目录失败: %v", err)
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	// 这里实现调用callgraph功能的逻辑
	// 可以使用exec.Command执行命令行工具，或者直接调用相关函数

	// 示例：使用exec.Command执行callgraph命令
	cmd := exec.Command("goanalysis", "callgraph", "-d", projectPath)

	// 设置环境变量，指定数据库路径
	cmd.Env = append(os.Environ(), fmt.Sprintf("DB_PATH=%s", dbPath))

	// 设置标准输出和标准错误
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	s.log.Info("执行命令: goanalysis callgraph -d " + projectPath)

	// 执行命令
	err := cmd.Run()
	if err != nil {
		s.log.Errorf("执行callgraph分析失败: %v\n标准输出: %s\n标准错误: %s",
			err, stdout.String(), stderr.String())
		return fmt.Errorf("执行callgraph分析失败: %v", err)
	}

	s.log.Infof("callgraph分析完成\n输出: %s", stdout.String())

	// 验证数据库文件是否生成
	if _, err := os.Stat(dbPath); err != nil {
		s.log.Errorf("分析完成但数据库文件未生成: %v", err)
		return fmt.Errorf("分析完成但数据库文件未生成: %v", err)
	}

	return nil
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
