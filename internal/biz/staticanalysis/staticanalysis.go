package staticanalysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/sourcegraph/conc/pool"
	v1 "github.com/toheart/goanalysis/api/staticanalysis/v1"
	"github.com/toheart/goanalysis/internal/biz/callgraph"
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
	"github.com/toheart/goanalysis/internal/biz/chanMgr"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/repo"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
)

// StaticAnalysisBiz 静态分析业务逻辑
type StaticAnalysisBiz struct {
	sync.RWMutex
	conf *conf.Biz
	data *data.Data
	log  *log.Helper

	AnalysisTaskChan   chan *entity.AnalysisTask
	globalChan         *chanMgr.ChannelManager
	analysisTaskStatus map[string]entity.AnalysisTaskStatus
}

// NewStaticAnalysisBiz 创建静态分析业务逻辑实例
func NewStaticAnalysisBiz(conf *conf.Biz, data *data.Data, mgr *chanMgr.ChannelManager, logger log.Logger) *StaticAnalysisBiz {
	return &StaticAnalysisBiz{
		conf:               conf,
		data:               data,
		log:                log.NewHelper(logger),
		AnalysisTaskChan:   make(chan *entity.AnalysisTask, 10),
		analysisTaskStatus: make(map[string]entity.AnalysisTaskStatus),
		globalChan:         mgr,
	}
}

// processAnalysisTasks 处理分析任务
func (s *StaticAnalysisBiz) ProcessAnalysisTasks() {
	s.log.Info("start process analysis tasks")
	p := pool.New().WithMaxGoroutines(3)
	for task := range s.AnalysisTaskChan {
		pTask := task
		p.Go(func() {
			s.log.Infof("start process analysis task: %s, project path: %s", task.ID, task.ProjectPath)
			s.SetTaskStatus(task.ID, entity.AnalysisTaskStatus{
				Status:   entity.TaskStatusProcessing,
				Progress: 0,
				Message:  "Processing...",
			})
			// 执行分析
			err := s.runCallgraphAnalysis(pTask)
			if err != nil {
				s.log.Errorf("analysis task failed: %s, error: %v", pTask.ID, err)
				s.SetTaskStatus(pTask.ID, entity.AnalysisTaskStatus{
					Status:   entity.TaskStatusFailed,
					Progress: 0,
					Message:  "Failed...",
				})
				return
			}
		})
	}
}

func (s *StaticAnalysisBiz) AnalyzeProjectPath(projectPath string, DbPath string) string {
	task := entity.AnalysisTask{
		ID:          uuid.New().String(),
		ProjectPath: projectPath,
		Filename:    DbPath,
	}
	s.SetTaskStatus(task.ID, entity.AnalysisTaskStatus{
		Status:   entity.TaskStatusStarting,
		Progress: 0,
		Message:  "Starting...",
	})
	s.AnalysisTaskChan <- &task
	return task.ID
}

// AnalyzeProjectPathWithOptions 使用指定选项分析项目路径
func (s *StaticAnalysisBiz) AnalyzeProjectPathWithOptions(projectPath string, DbPath string, options *entity.AnalysisOptions) string {
	task := entity.AnalysisTask{
		ID:          uuid.New().String(),
		ProjectPath: projectPath,
		Filename:    DbPath,
		Options:     options,
	}
	s.SetTaskStatus(task.ID, entity.AnalysisTaskStatus{
		Status:   entity.TaskStatusStarting,
		Progress: 0,
		Message:  "Starting...",
	})
	s.AnalysisTaskChan <- &task
	return task.ID
}

// VerifyProjectPath 验证项目路径是否存在
func (s *StaticAnalysisBiz) VerifyProjectPath(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetStaticDBPath 获取静态分析数据库路径
func (s *StaticAnalysisBiz) GetStaticDBPath() string {
	return entity.GetFileStoragePath(s.conf.FileStoragePath, false)
}

// GetFuncNodeDB 获取函数节点数据库
func (s *StaticAnalysisBiz) GetFuncNodeDB(dbPath string) (repo.StaticDBStore, error) {
	s.log.Infof("Getting function node database: %s", dbPath)
	return s.data.GetFuncNodeDB(dbPath)
}

// GetStatusChan 获取指定类型的状态通道
func (s *StaticAnalysisBiz) GetStatusChan(taskId string) (chan []byte, error) {
	s.log.Infof("Getting status channel: %s", taskId)

	ch, err := s.globalChan.Get(taskId)
	if err != nil {
		s.log.Errorf("Failed to get status channel: %v", err)
		return nil, err
	}

	s.log.Infof("Successfully got status channel: %s", taskId)
	return ch, nil
}

// GetTaskStatusChan 获取指定任务ID的状态通道
func (s *StaticAnalysisBiz) GetTaskStatusChan(taskID string) (chan []byte, error) {
	s.log.Infof("Getting task status channel for task: %s", taskID)

	ch, err := s.globalChan.Get(taskID)
	if err != nil {
		s.log.Errorf("Failed to get task status channel: %v", err)
		return nil, err
	}

	s.log.Infof("Successfully got task status channel for task: %s", taskID)
	return ch, nil
}

// 运行callgraph分析
func (s *StaticAnalysisBiz) runCallgraphAnalysis(task *entity.AnalysisTask) error {
	s.log.Infof("start callgraph analysis for project %s, db path: %s", task.ProjectPath, task.Filename)

	// 设置通道
	statusChan := make(chan []byte, 100)
	s.globalChan.Set(task.ID, statusChan)
	defer s.globalChan.Close(task.ID)
	// 发送初始状态消息
	statusChan <- []byte(fmt.Sprintf("Starting analysis for project: %s", task.ProjectPath))

	// 查找对应的任务
	dbPath := filepath.Join(s.GetStaticDBPath(), task.Filename)
	s.log.Infof("Database path: %s", dbPath)

	funcNodeDB, err := s.data.GetFuncNodeDB(dbPath)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get database: %v", err)
		statusChan <- []byte(errMsg)
		s.log.Error(errMsg)
		return err
	}

	s.log.Infof("start to generate call graph for %s...", task.ProjectPath)

	// 创建分析选项
	options := []callgraph.ProgramOption{}

	// 如果有分析选项，则使用它们
	if task.Options != nil {
		// 设置算法
		if task.Options.Algo != "" {
			options = append(options, callgraph.WithAlgo(task.Options.Algo))
			statusChan <- []byte(fmt.Sprintf("Using algorithm: %s", task.Options.Algo))
		} else {
			options = append(options, callgraph.WithAlgo(callgraph.CallGraphTypeVta))
			statusChan <- []byte(fmt.Sprintf("Using default algorithm: %s", callgraph.CallGraphTypeVta))
		}

		// 设置缓存路径
		if task.Options.IgnoreMethod != "" {
			options = append(options, callgraph.WithIgnorePaths(task.Options.IgnoreMethod))
			statusChan <- []byte(fmt.Sprintf("Ignore method: %s", task.Options.IgnoreMethod))
		}
	} else {
		// 使用默认选项
		options = append(options, callgraph.WithAlgo(callgraph.CallGraphTypeVta))
		statusChan <- []byte("Using default analysis options")
	}

	// 创建程序分析实例
	c := callgraph.NewProgramAnalysis(task.ProjectPath, log.NewHelper(log.With(s.log.Logger(), "module", "callgraph", "task", task.ID)), funcNodeDB, options...)

	// 使用 WaitGroup 等待所有任务完成
	var wg sync.WaitGroup
	// 设置完成标志
	var completed bool

	// 启动调用图分析（使用新的 Execute 方法）
	wg.Add(1)
	go func() {
		defer wg.Done()
		statusChan <- []byte("Starting call graph analysis...")
		if err := c.Execute(context.Background(), statusChan); err != nil {
			errMsg := fmt.Sprintf("Call graph analysis failed: %v", err)
			statusChan <- []byte(errMsg)
			s.log.Error(errMsg)
			return
		}
		// 标记为完成
		statusChan <- []byte("Analysis task completed")
		completed = true
		s.log.Info("Analysis task completed")
	}()

	// 启动进度监控
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 3)
		defer ticker.Stop()
		for range ticker.C {
			if completed {
				s.SetTaskStatus(task.ID, entity.AnalysisTaskStatus{
					Status:   entity.TaskStatusCompleted,
					Progress: 1.0,
					Message:  "Completed...",
				})
				return
			}
			s.SetTaskStatus(task.ID, entity.AnalysisTaskStatus{
				Status:   entity.TaskStatusProcessing,
				Progress: c.GetProgress(),
				Message:  "Processing...",
			})
		}
	}()

	// 等待任务完成
	wg.Wait()
	s.log.Infof("callgraph analysis for %s completed", task.ProjectPath)
	statusChan <- []byte("EOF")

	return nil
}

// GetAllTasks 获取所有任务ID
func (s *StaticAnalysisBiz) GetAllTasks() ([]string, error) {
	s.log.Info("Getting all tasks")

	// 获取所有通道
	channels := s.globalChan.GetAll()

	// 过滤出任务ID
	var taskIDs []string
	for key := range channels {
		taskIDs = append(taskIDs, key)
	}

	s.log.Infof("Found %d tasks", len(taskIDs))
	return taskIDs, nil
}

func (s *StaticAnalysisBiz) GetTaskStatus(taskID string) (entity.AnalysisTaskStatus, error) {
	s.RLock()
	defer s.RUnlock()
	status, ok := s.analysisTaskStatus[taskID]
	if !ok {
		return entity.AnalysisTaskStatus{
			Status:   entity.TaskStatusFailed,
			Progress: 0,
			Message:  "Failed...",
		}, fmt.Errorf("task %s not found", taskID)
	}
	return status, nil
}

func (s *StaticAnalysisBiz) SetTaskStatus(taskID string, status entity.AnalysisTaskStatus) {
	s.Lock()
	defer s.Unlock()
	s.analysisTaskStatus[taskID] = status
}

// GetTaskProgress 获取任务进度
func (s *StaticAnalysisBiz) GetTaskProgress(taskID string) (float64, error) {
	s.log.Infof("Getting task progress for task: %s", taskID)

	// 获取当前任务的状态
	s.RLock()
	status, ok := s.analysisTaskStatus[taskID]
	s.RUnlock()

	if !ok {
		return 0, fmt.Errorf("task %s not found", taskID)
	}

	// 如果任务不在处理中，返回相应的进度
	if status.Status != entity.TaskStatusProcessing {
		if status.Status == entity.TaskStatusCompleted {
			return 1.0, nil // 已完成，进度为100%
		}
		return 0, nil // 其他状态，进度为0
	}

	return status.Progress, nil
}

// ================ 可视化分析方法 ================

// GetGlobalStatistics 获取全局统计信息
func (s *StaticAnalysisBiz) GetGlobalStatistics(ctx context.Context, funcNodeDB repo.StaticDBStore) (*v1.GetGlobalStatisticsResponse, error) {
	s.log.Info("Starting global statistics calculation...")

	// 获取所有函数节点
	s.log.Info("Loading function nodes...")
	funcNodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		return nil, fmt.Errorf("get all func nodes failed: %w", err)
	}
	s.log.Infof("Loaded %d function nodes", len(funcNodes))

	// 获取所有调用边
	s.log.Info("Loading function edges...")
	funcEdges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		return nil, fmt.Errorf("get all func edges failed: %w", err)
	}
	s.log.Infof("Loaded %d function edges", len(funcEdges))

	// 统计包信息
	s.log.Info("Calculating package statistics...")
	packageStats := s.calculatePackageStatistics(funcNodes, funcEdges, funcNodeDB)
	s.log.Infof("Calculated statistics for %d packages", len(packageStats))

	// 计算热点函数（减少数量以提高性能）
	s.log.Info("Calculating hot functions...")
	hotFunctions := s.calculateHotFunctions(funcNodes, funcEdges, 3) // 减少到3个
	s.log.Infof("Calculated %d hot functions", len(hotFunctions))

	// 计算平均调用深度
	s.log.Info("Calculating call depth...")
	avgDepth, maxDepth := s.calculateCallDepth(funcNodes, funcEdges)
	s.log.Infof("Call depth calculation completed: avg=%.2f, max=%d", avgDepth, maxDepth)

	s.log.Info("Global statistics calculation completed successfully")
	return &v1.GetGlobalStatisticsResponse{
		TotalFunctions: int32(len(funcNodes)),
		TotalPackages:  int32(len(packageStats)),
		TotalEdges:     int32(len(funcEdges)),
		AvgCallDepth:   float32(avgDepth),
		MaxCallDepth:   int32(maxDepth),
		HotFunctions:   hotFunctions,
		PackageStats:   packageStats,
	}, nil
}

// GetCallGraph 获取调用图数据
func (s *StaticAnalysisBiz) GetCallGraph(ctx context.Context, req *v1.GetCallGraphRequest, funcNodeDB repo.StaticDBStore) (*v1.GetCallGraphResponse, error) {
	// 获取所有函数节点
	allNodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		return nil, fmt.Errorf("get all func nodes failed: %w", err)
	}

	// 获取所有调用边
	allEdges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		return nil, fmt.Errorf("get all func edges failed: %w", err)
	}

	// 构建调用图
	nodes, edges := s.buildCallGraph(allNodes, allEdges, req)

	return &v1.GetCallGraphResponse{
		Nodes:      nodes,
		Edges:      edges,
		TotalNodes: int32(len(nodes)),
		TotalEdges: int32(len(edges)),
		RootNode:   req.RootFunction,
	}, nil
}

// GetPackageDependencies 获取包依赖关系
func (s *StaticAnalysisBiz) GetPackageDependencies(ctx context.Context, funcNodeDB repo.StaticDBStore) (*v1.GetPackageDependenciesResponse, error) {
	s.log.Info("Starting package dependencies calculation...")

	// 获取所有函数节点
	funcNodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		return nil, fmt.Errorf("get all func nodes failed: %w", err)
	}
	s.log.Infof("Loaded %d function nodes", len(funcNodes))

	// 获取所有调用边
	funcEdges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		return nil, fmt.Errorf("get all func edges failed: %w", err)
	}
	s.log.Infof("Loaded %d function edges", len(funcEdges))

	packages, dependencies := s.buildPackageDependencies(funcNodes, funcEdges, funcNodeDB)
	s.log.Infof("Built %d packages with %d dependencies", len(packages), len(dependencies))

	s.log.Info("Package dependencies calculation completed successfully")
	return &v1.GetPackageDependenciesResponse{
		Packages:     packages,
		Dependencies: dependencies,
	}, nil
}

// GetHotFunctions 获取热点函数分析
func (s *StaticAnalysisBiz) GetHotFunctions(ctx context.Context, req *v1.GetHotFunctionsRequest, funcNodeDB repo.StaticDBStore) (*v1.GetHotFunctionsResponse, error) {
	// 获取所有函数节点
	funcNodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		return nil, fmt.Errorf("get all func nodes failed: %w", err)
	}

	// 获取所有调用边
	funcEdges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		return nil, fmt.Errorf("get all func edges failed: %w", err)
	}

	// 获取详细热点函数信息
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	detailedHotFunctions := s.getDetailedHotFunctions(funcNodes, funcEdges, limit, req.SortBy)

	return &v1.GetHotFunctionsResponse{
		Functions: detailedHotFunctions,
	}, nil
}

// GetFunctionDetails 获取函数详细信息
func (s *StaticAnalysisBiz) GetFunctionDetails(ctx context.Context, functionKey string, funcNodeDB repo.StaticDBStore) (*v1.GetFunctionDetailsResponse, error) {
	// 获取函数节点
	funcNode, err := funcNodeDB.GetFuncNodeByKey(functionKey)
	if err != nil {
		return nil, fmt.Errorf("get func node failed: %w", err)
	}
	if funcNode == nil {
		return nil, fmt.Errorf("function not found: %s", functionKey)
	}

	// 获取调用者
	callers, err := funcNodeDB.GetCallerEdges(functionKey)
	if err != nil {
		return nil, fmt.Errorf("get caller edges failed: %w", err)
	}

	// 获取被调用者
	callees, err := funcNodeDB.GetCalleeEdges(functionKey)
	if err != nil {
		return nil, fmt.Errorf("get callee edges failed: %w", err)
	}

	// 转换为响应格式
	function := s.convertToFunctionDetail(funcNode)
	callerDetails := make([]*v1.FunctionDetail, 0, len(callers))
	for _, caller := range callers {
		callerDetails = append(callerDetails, s.convertToFunctionDetail(caller))
	}

	calleeDetails := make([]*v1.FunctionDetail, 0, len(callees))
	for _, callee := range callees {
		calleeDetails = append(calleeDetails, s.convertToFunctionDetail(callee))
	}

	return &v1.GetFunctionDetailsResponse{
		Function: function,
		Callers:  callerDetails,
		Callees:  calleeDetails,
	}, nil
}

// GetFunctionCallPaths 获取函数调用路径
func (s *StaticAnalysisBiz) GetFunctionCallPaths(ctx context.Context, req *v1.GetFunctionCallPathsRequest, funcNodeDB repo.StaticDBStore) (*v1.GetFunctionCallPathsResponse, error) {
	// 获取所有调用边来构建图
	allEdges, err := funcNodeDB.GetAllFuncEdges()
	if err != nil {
		return nil, fmt.Errorf("get all func edges failed: %w", err)
	}

	// 使用DFS或BFS查找路径
	paths := s.findCallPaths(allEdges, req.FromFunction, req.ToFunction, int(req.MaxPaths), int(req.MaxDepth))

	return &v1.GetFunctionCallPathsResponse{
		Paths:   paths,
		HasMore: len(paths) >= int(req.MaxPaths),
	}, nil
}

// GetInitFunctions 获取Init函数列表
func (s *StaticAnalysisBiz) GetInitFunctions(ctx context.Context, req *v1.GetInitFunctionsRequest, funcNodeDB repo.StaticDBStore) (*v1.GetInitFunctionsResponse, error) {
	s.log.Info("Starting init functions retrieval...")

	// 获取所有函数节点
	s.log.Info("Loading function nodes...")
	funcNodes, err := funcNodeDB.GetAllFuncNodes()
	if err != nil {
		return nil, fmt.Errorf("get all func nodes failed: %w", err)
	}
	s.log.Infof("Loaded %d function nodes", len(funcNodes))

	// 过滤出 init 函数
	var initFunctions []*v1.InitFunction
	totalInitCount := 0

	for _, node := range funcNodes {
		// 检查是否是 init 函数
		if s.isInitFunction(node) {
			totalInitCount++

			// 应用包过滤器
			if req.PackageFilter != "" && !strings.Contains(node.Pkg, req.PackageFilter) {
				continue
			}

			// 检查是否达到限制
			if req.Limit > 0 && int32(len(initFunctions)) >= req.Limit {
				continue
			}

			initFunc := &v1.InitFunction{
				Key:       node.Key,
				Name:      node.Name,
				Package:   node.Pkg,
				FullName:  node.FullName,
				Signature: node.Signature,
				Doc:       node.SourceCode,
				Position: &v1.NodePosition{
					Filename:  node.Position.Filename,
					StartLine: int32(node.Position.Line),
					EndLine:   int32(node.Position.EndLine),
				},
			}

			initFunctions = append(initFunctions, initFunc)
		}
	}

	s.log.Infof("Found %d init functions (total: %d)", len(initFunctions), totalInitCount)

	return &v1.GetInitFunctionsResponse{
		InitFunctions: initFunctions,
		TotalCount:    int32(totalInitCount),
	}, nil
}

// isInitFunction 判断是否是 Go 包级 init 初始化函数
func (s *StaticAnalysisBiz) isInitFunction(node *dos.FuncNode) bool {
	if node == nil {
		return false
	}

	// 严格匹配：函数名为 "init"
	if node.Name == "init" {
		return true
	}

	// 兼容 SSA 可能生成的包装函数（如 init#1、init$1）
	if strings.HasPrefix(node.Name, "init#") || strings.HasPrefix(node.Name, "init$") {
		return true
	}

	// 某些情况下仅在 FullName 中体现（如 "pkg/path.init" 或 "pkg/path.init#1"）
	if strings.HasSuffix(node.FullName, ".init") ||
		strings.Contains(node.FullName, ".init#") ||
		strings.Contains(node.FullName, ".init$") {
		return true
	}

	return false
}

// ================ 私有辅助方法 ================

// calculatePackageStatistics 计算包统计信息
func (s *StaticAnalysisBiz) calculatePackageStatistics(funcNodes []*dos.FuncNode, funcEdges []*dos.FuncEdge, funcNodeDB repo.StaticDBStore) []*v1.PackageStatistic {
	packageMap := make(map[string]*v1.PackageStatistic)

	// 创建函数key到包名的映射，避免重复数据库查询
	funcKeyToPkg := make(map[string]string)
	for _, node := range funcNodes {
		pkg := node.Pkg
		if pkg == "" {
			pkg = "unknown"
		}
		funcKeyToPkg[node.Key] = pkg
	}

	// 统计每个包的函数数量
	for _, node := range funcNodes {
		pkg := node.Pkg
		if pkg == "" {
			pkg = "unknown"
		}

		if _, exists := packageMap[pkg]; !exists {
			packageMap[pkg] = &v1.PackageStatistic{
				PackageName:     pkg,
				FunctionCount:   0,
				InternalCalls:   0,
				ExternalCalls:   0,
				ComplexityScore: 0,
			}
		}
		packageMap[pkg].FunctionCount++
	}

	// 统计调用关系 - 使用内存映射而不是数据库查询
	for _, edge := range funcEdges {
		callerPkg, callerExists := funcKeyToPkg[edge.CallerKey]
		calleePkg, calleeExists := funcKeyToPkg[edge.CalleeKey]

		if callerExists && calleeExists {
			if callerPkg == "" {
				callerPkg = "unknown"
			}
			if calleePkg == "" {
				calleePkg = "unknown"
			}

			if stat, exists := packageMap[callerPkg]; exists {
				if callerPkg == calleePkg {
					stat.InternalCalls++
				} else {
					stat.ExternalCalls++
				}
			}
		}
	}

	// 计算复杂度评分
	for _, stat := range packageMap {
		// 简单的复杂度计算：基于函数数量和外部调用
		stat.ComplexityScore = float32(stat.FunctionCount)*0.3 + float32(stat.ExternalCalls)*0.7
	}

	// 转换为切片并排序
	stats := make([]*v1.PackageStatistic, 0, len(packageMap))
	for _, stat := range packageMap {
		stats = append(stats, stat)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].ComplexityScore > stats[j].ComplexityScore
	})

	return stats
}

// calculateHotFunctions 计算热点函数
func (s *StaticAnalysisBiz) calculateHotFunctions(funcNodes []*dos.FuncNode, funcEdges []*dos.FuncEdge, limit int) []*v1.HotFunction {
	// 统计每个函数的调用情况
	callerCount := make(map[string]int)
	calleeCount := make(map[string]int)

	for _, edge := range funcEdges {
		callerCount[edge.CalleeKey]++ // 被调用次数
		calleeCount[edge.CallerKey]++ // 调用其他函数次数
	}

	// 创建热点函数列表
	hotFunctions := make([]*v1.HotFunction, 0)
	for _, node := range funcNodes {
		callers := callerCount[node.Key]
		callees := calleeCount[node.Key]

		// 计算重要性评分：被调用次数权重更高
		importance := float32(callers)*0.7 + float32(callees)*0.3

		if callers > 0 || callees > 0 {
			hotFunctions = append(hotFunctions, &v1.HotFunction{
				Key:             node.Key,
				Name:            node.Name,
				Package:         node.Pkg,
				CallerCount:     int32(callers),
				CalleeCount:     int32(callees),
				ImportanceScore: importance,
			})
		}
	}

	// 按重要性排序
	sort.Slice(hotFunctions, func(i, j int) bool {
		return hotFunctions[i].ImportanceScore > hotFunctions[j].ImportanceScore
	})

	// 限制返回数量
	if len(hotFunctions) > limit {
		hotFunctions = hotFunctions[:limit]
	}

	return hotFunctions
}

// calculateCallDepth 计算调用深度（优化版本：限制搜索深度和入口点数量）
func (s *StaticAnalysisBiz) calculateCallDepth(funcNodes []*dos.FuncNode, funcEdges []*dos.FuncEdge) (float64, int) {
	// 对于大型项目，限制处理的数据量以提高性能
	maxEntryPoints := 50 // 最多处理50个入口点
	maxSearchDepth := 20 // 最大搜索深度限制为20层

	// 构建邻接表
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, node := range funcNodes {
		graph[node.Key] = make([]string, 0)
		inDegree[node.Key] = 0
	}

	for _, edge := range funcEdges {
		graph[edge.CallerKey] = append(graph[edge.CallerKey], edge.CalleeKey)
		inDegree[edge.CalleeKey]++
	}

	// 找到入口节点（入度为0的节点）
	entryPoints := make([]string, 0)
	for key, degree := range inDegree {
		if degree == 0 {
			entryPoints = append(entryPoints, key)
			// 限制入口点数量以提高性能
			if len(entryPoints) >= maxEntryPoints {
				break
			}
		}
	}

	// 如果没有入口点，随机选择一些节点作为起点
	if len(entryPoints) == 0 && len(funcNodes) > 0 {
		count := 10
		if len(funcNodes) < count {
			count = len(funcNodes)
		}
		for i := 0; i < count; i++ {
			entryPoints = append(entryPoints, funcNodes[i].Key)
		}
	}

	// 计算最大深度
	maxDepth := 0
	totalDepth := 0
	nodeCount := 0

	for _, entry := range entryPoints {
		depth := s.dfsMaxDepthLimited(graph, entry, make(map[string]bool), 0, maxSearchDepth)
		if depth > maxDepth {
			maxDepth = depth
		}
		totalDepth += depth
		nodeCount++
	}

	avgDepth := 0.0
	if nodeCount > 0 {
		avgDepth = float64(totalDepth) / float64(nodeCount)
	}

	return avgDepth, maxDepth
}

// dfsMaxDepth 使用DFS计算最大深度
func (s *StaticAnalysisBiz) dfsMaxDepth(graph map[string][]string, node string, visited map[string]bool) int {
	if visited[node] {
		return 0 // 避免循环
	}

	visited[node] = true
	maxChildDepth := 0

	for _, child := range graph[node] {
		depth := s.dfsMaxDepth(graph, child, visited)
		if depth > maxChildDepth {
			maxChildDepth = depth
		}
	}

	visited[node] = false // 回溯
	return maxChildDepth + 1
}

// dfsMaxDepthLimited 使用DFS计算最大深度（限制版本）
func (s *StaticAnalysisBiz) dfsMaxDepthLimited(graph map[string][]string, node string, visited map[string]bool, currentDepth, maxDepth int) int {
	if visited[node] || currentDepth >= maxDepth {
		return currentDepth // 避免循环或达到最大深度限制
	}

	visited[node] = true
	maxChildDepth := currentDepth

	for _, child := range graph[node] {
		depth := s.dfsMaxDepthLimited(graph, child, visited, currentDepth+1, maxDepth)
		if depth > maxChildDepth {
			maxChildDepth = depth
		}
	}

	visited[node] = false // 回溯
	return maxChildDepth
}

// buildCallGraph 构建调用图
func (s *StaticAnalysisBiz) buildCallGraph(funcNodes []*dos.FuncNode, funcEdges []*dos.FuncEdge, req *v1.GetCallGraphRequest) ([]*v1.GraphNode, []*v1.GraphEdge) {
	// 如果指定了根函数，则从根函数开始构建子图
	if req.RootFunction != "" {
		return s.buildSubGraph(funcNodes, funcEdges, req.RootFunction, int(req.MaxDepth))
	}

	// 否则构建完整图
	nodes := make([]*v1.GraphNode, 0, len(funcNodes))
	edges := make([]*v1.GraphEdge, 0, len(funcEdges))

	// 创建节点映射
	nodeMap := make(map[string]*dos.FuncNode)
	for _, node := range funcNodes {
		nodeMap[node.Key] = node
	}

	// 统计调用情况用于节点度量
	callerCount := make(map[string]int)
	calleeCount := make(map[string]int)
	for _, edge := range funcEdges {
		callerCount[edge.CalleeKey]++
		calleeCount[edge.CallerKey]++
	}

	// 转换节点
	for _, node := range funcNodes {
		// 过滤外部包（如果设置）
		if !req.IncludeExternal && s.isExternalPackage(node.Pkg) {
			continue
		}

		graphNode := &v1.GraphNode{
			Id:        node.Key,
			Label:     node.Name,
			Package:   node.Pkg,
			FullName:  node.FullName,
			Signature: node.Signature,
			Doc:       node.SourceCode,
			Position: &v1.NodePosition{
				Filename:  node.Position.Filename,
				StartLine: int32(node.Position.Line),
				EndLine:   int32(node.Position.EndLine),
			},
			Metrics: &v1.NodeMetrics{
				CallerCount: int32(callerCount[node.Key]),
				CalleeCount: int32(calleeCount[node.Key]),
				DepthLevel:  0, // 需要通过图分析计算
				Centrality:  0, // 简化，可以用度数中心性
			},
			Type: s.determineNodeType(node, callerCount[node.Key], calleeCount[node.Key]),
		}

		nodes = append(nodes, graphNode)
	}

	// 转换边
	for _, edge := range funcEdges {
		// 检查源和目标节点是否都在节点列表中
		sourceExists := false
		targetExists := false
		for _, node := range nodes {
			if node.Id == edge.CallerKey {
				sourceExists = true
			}
			if node.Id == edge.CalleeKey {
				targetExists = true
			}
		}

		if sourceExists && targetExists {
			graphEdge := &v1.GraphEdge{
				Id:       fmt.Sprintf("%s->%s", edge.CallerKey, edge.CalleeKey),
				Source:   edge.CallerKey,
				Target:   edge.CalleeKey,
				CallType: edge.CallType,
				CallSites: []*v1.CallSite{
					{
						File: edge.CallSiteFile,
						Line: int32(edge.CallSiteLine),
					},
				},
			}
			edges = append(edges, graphEdge)
		}
	}

	return nodes, edges
}

// buildSubGraph 构建以指定函数为根的子图
func (s *StaticAnalysisBiz) buildSubGraph(funcNodes []*dos.FuncNode, funcEdges []*dos.FuncEdge, rootKey string, maxDepth int) ([]*v1.GraphNode, []*v1.GraphEdge) {
	// 使用BFS构建子图
	visited := make(map[string]bool)
	nodeQueue := []string{rootKey}
	depthMap := map[string]int{rootKey: 0}

	// 构建邻接表
	graph := make(map[string][]string)
	edgeMap := make(map[string]*dos.FuncEdge)

	for _, edge := range funcEdges {
		graph[edge.CallerKey] = append(graph[edge.CallerKey], edge.CalleeKey)
		edgeMap[fmt.Sprintf("%s->%s", edge.CallerKey, edge.CalleeKey)] = edge
	}

	// BFS遍历
	validNodes := make(map[string]bool)
	for len(nodeQueue) > 0 {
		current := nodeQueue[0]
		nodeQueue = nodeQueue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true
		validNodes[current] = true

		currentDepth := depthMap[current]
		if currentDepth < maxDepth {
			for _, neighbor := range graph[current] {
				if !visited[neighbor] {
					nodeQueue = append(nodeQueue, neighbor)
					depthMap[neighbor] = currentDepth + 1
				}
			}
		}
	}

	// 构建节点和边
	nodes := make([]*v1.GraphNode, 0)
	edges := make([]*v1.GraphEdge, 0)

	nodeMap := make(map[string]*dos.FuncNode)
	for _, node := range funcNodes {
		nodeMap[node.Key] = node
	}

	// 统计调用情况
	callerCount := make(map[string]int)
	calleeCount := make(map[string]int)
	for _, edge := range funcEdges {
		if validNodes[edge.CallerKey] && validNodes[edge.CalleeKey] {
			callerCount[edge.CalleeKey]++
			calleeCount[edge.CallerKey]++
		}
	}

	// 添加有效节点
	for nodeKey := range validNodes {
		if node, exists := nodeMap[nodeKey]; exists {
			graphNode := &v1.GraphNode{
				Id:        node.Key,
				Label:     node.Name,
				Package:   node.Pkg,
				FullName:  node.FullName,
				Signature: node.Signature,
				Doc:       node.SourceCode,
				Position: &v1.NodePosition{
					Filename:  node.Position.Filename,
					StartLine: int32(node.Position.Line),
					EndLine:   int32(node.Position.EndLine),
				},
				Metrics: &v1.NodeMetrics{
					CallerCount: int32(callerCount[node.Key]),
					CalleeCount: int32(calleeCount[node.Key]),
					DepthLevel:  int32(depthMap[node.Key]),
					Centrality:  0,
				},
				Type: s.determineNodeType(node, callerCount[node.Key], calleeCount[node.Key]),
			}
			nodes = append(nodes, graphNode)
		}
	}

	// 添加有效边
	for _, edge := range funcEdges {
		if validNodes[edge.CallerKey] && validNodes[edge.CalleeKey] {
			graphEdge := &v1.GraphEdge{
				Id:       fmt.Sprintf("%s->%s", edge.CallerKey, edge.CalleeKey),
				Source:   edge.CallerKey,
				Target:   edge.CalleeKey,
				CallType: edge.CallType,
				CallSites: []*v1.CallSite{
					{
						File: edge.CallSiteFile,
						Line: int32(edge.CallSiteLine),
					},
				},
			}
			edges = append(edges, graphEdge)
		}
	}

	return nodes, edges
}

// 其他辅助方法的实现...
func (s *StaticAnalysisBiz) isExternalPackage(pkg string) bool {
	// 简单判断：如果包名包含外部域名，则认为是外部包
	return strings.Contains(pkg, "github.com") ||
		strings.Contains(pkg, "golang.org") ||
		strings.Contains(pkg, "google.golang.org")
}

func (s *StaticAnalysisBiz) determineNodeType(node *dos.FuncNode, callerCount, calleeCount int) v1.NodeType {
	if strings.Contains(node.Name, "main") {
		return v1.NodeType_ENTRY_POINT
	}
	if callerCount > 5 { // 热点函数阈值
		return v1.NodeType_HOT_SPOT
	}
	if s.isExternalPackage(node.Pkg) {
		return v1.NodeType_EXTERNAL
	}
	return v1.NodeType_NORMAL
}

func (s *StaticAnalysisBiz) convertToFunctionDetail(node *dos.FuncNode) *v1.FunctionDetail {
	return &v1.FunctionDetail{
		Key:       node.Key,
		Name:      node.Name,
		Package:   node.Pkg,
		FullName:  node.FullName,
		Signature: node.Signature,
		Doc:       node.SourceCode,
		Position: &v1.NodePosition{
			Filename:  node.Position.Filename,
			StartLine: int32(node.Position.Line),
			EndLine:   int32(node.Position.EndLine),
		},
		Metrics: &v1.NodeMetrics{
			CallerCount: 0, // 需要查询
			CalleeCount: 0, // 需要查询
			DepthLevel:  0,
			Centrality:  0,
		},
	}
}

// buildPackageDependencies 构建包依赖关系
func (s *StaticAnalysisBiz) buildPackageDependencies(funcNodes []*dos.FuncNode, funcEdges []*dos.FuncEdge, funcNodeDB repo.StaticDBStore) ([]*v1.PackageNode, []*v1.PackageDependency) {
	// 统计包信息
	packageStats := make(map[string]*v1.PackageNode)
	dependencyMap := make(map[string]*v1.PackageDependency)

	// 创建函数key到包名的映射，避免重复数据库查询
	funcKeyToPkg := make(map[string]string)
	for _, node := range funcNodes {
		pkg := node.Pkg
		if pkg == "" {
			pkg = "unknown"
		}
		funcKeyToPkg[node.Key] = pkg
	}

	// 初始化包节点
	for _, node := range funcNodes {
		pkg := node.Pkg
		if pkg == "" {
			pkg = "unknown"
		}

		if _, exists := packageStats[pkg]; !exists {
			packageStats[pkg] = &v1.PackageNode{
				Name:              pkg,
				FunctionCount:     0,
				InternalCallCount: 0,
				ExternalCallCount: 0,
				Complexity:        0,
				SubPackages:       []string{},
			}
		}
		packageStats[pkg].FunctionCount++
	}

	// 构建依赖关系 - 使用内存映射而不是数据库查询
	for _, edge := range funcEdges {
		callerPkg, callerExists := funcKeyToPkg[edge.CallerKey]
		calleePkg, calleeExists := funcKeyToPkg[edge.CalleeKey]

		if callerExists && calleeExists {
			if callerPkg == "" {
				callerPkg = "unknown"
			}
			if calleePkg == "" {
				calleePkg = "unknown"
			}

			if callerPkg == calleePkg {
				// 包内调用
				if stat, exists := packageStats[callerPkg]; exists {
					stat.InternalCallCount++
				}
			} else {
				// 跨包调用
				if stat, exists := packageStats[callerPkg]; exists {
					stat.ExternalCallCount++
				}

				// 添加依赖关系
				depKey := fmt.Sprintf("%s->%s", callerPkg, calleePkg)
				if dep, exists := dependencyMap[depKey]; exists {
					dep.CallCount++
				} else {
					dependencyMap[depKey] = &v1.PackageDependency{
						SourcePackage:      callerPkg,
						TargetPackage:      calleePkg,
						CallCount:          1,
						DependencyStrength: 0, // 稍后计算
						CallFunctions:      []string{edge.CallerKey},
					}
				}
			}
		}
	}

	// 计算复杂度和依赖强度
	for _, pkg := range packageStats {
		pkg.Complexity = float32(pkg.FunctionCount)*0.3 + float32(pkg.ExternalCallCount)*0.7
	}

	for _, dep := range dependencyMap {
		// 依赖强度基于调用次数和源包大小
		sourceSize := float32(1)
		if sourcePkg, exists := packageStats[dep.SourcePackage]; exists {
			sourceSize = float32(sourcePkg.FunctionCount)
		}
		dep.DependencyStrength = float32(dep.CallCount) / sourceSize
	}

	// 转换为切片
	packages := make([]*v1.PackageNode, 0, len(packageStats))
	for _, pkg := range packageStats {
		packages = append(packages, pkg)
	}

	dependencies := make([]*v1.PackageDependency, 0, len(dependencyMap))
	for _, dep := range dependencyMap {
		dependencies = append(dependencies, dep)
	}

	return packages, dependencies
}

// getDetailedHotFunctions 获取详细热点函数信息
func (s *StaticAnalysisBiz) getDetailedHotFunctions(funcNodes []*dos.FuncNode, funcEdges []*dos.FuncEdge, limit int, sortBy string) []*v1.DetailedHotFunction {
	// 统计调用关系
	callerCount := make(map[string]int)
	calleeCount := make(map[string]int)
	callerMap := make(map[string][]string)
	calleeMap := make(map[string][]string)

	for _, edge := range funcEdges {
		callerCount[edge.CalleeKey]++
		calleeCount[edge.CallerKey]++
		callerMap[edge.CalleeKey] = append(callerMap[edge.CalleeKey], edge.CallerKey)
		calleeMap[edge.CallerKey] = append(calleeMap[edge.CallerKey], edge.CalleeKey)
	}

	// 创建详细热点函数列表
	hotFunctions := make([]*v1.DetailedHotFunction, 0)
	for _, node := range funcNodes {
		callers := callerCount[node.Key]
		callees := calleeCount[node.Key]

		if callers > 0 || callees > 0 {
			// 获取主要调用者和被调用者
			topCallers := s.getTopFunctions(callerMap[node.Key], 5)
			topCallees := s.getTopFunctions(calleeMap[node.Key], 5)

			importance := float32(callers)*0.7 + float32(callees)*0.3

			hotFunction := &v1.DetailedHotFunction{
				Key:       node.Key,
				Name:      node.Name,
				Package:   node.Pkg,
				Signature: node.Signature,
				Doc:       node.SourceCode,
				Position: &v1.NodePosition{
					Filename:  node.Position.Filename,
					StartLine: int32(node.Position.Line),
					EndLine:   int32(node.Position.EndLine),
				},
				CallerCount:     int32(callers),
				CalleeCount:     int32(callees),
				ImportanceScore: importance,
				TopCallers:      topCallers,
				TopCallees:      topCallees,
			}

			hotFunctions = append(hotFunctions, hotFunction)
		}
	}

	// 排序
	sort.Slice(hotFunctions, func(i, j int) bool {
		switch sortBy {
		case "caller_count":
			return hotFunctions[i].CallerCount > hotFunctions[j].CallerCount
		case "callee_count":
			return hotFunctions[i].CalleeCount > hotFunctions[j].CalleeCount
		default: // "importance"
			return hotFunctions[i].ImportanceScore > hotFunctions[j].ImportanceScore
		}
	})

	// 限制返回数量
	if len(hotFunctions) > limit {
		hotFunctions = hotFunctions[:limit]
	}

	return hotFunctions
}

// getTopFunctions 获取前N个函数
func (s *StaticAnalysisBiz) getTopFunctions(functions []string, limit int) []string {
	if len(functions) <= limit {
		return functions
	}
	return functions[:limit]
}

// findCallPaths 查找函数调用路径
func (s *StaticAnalysisBiz) findCallPaths(edges []*dos.FuncEdge, from, to string, maxPaths, maxDepth int) []*v1.CallPath {
	// 构建邻接表
	graph := make(map[string][]string)
	for _, edge := range edges {
		graph[edge.CallerKey] = append(graph[edge.CallerKey], edge.CalleeKey)
	}

	// 使用DFS查找路径
	paths := make([]*v1.CallPath, 0)
	visited := make(map[string]bool)
	currentPath := []string{from}

	s.dfsPath(graph, from, to, currentPath, visited, &paths, maxPaths, maxDepth, 0)

	return paths
}

// dfsPath DFS查找路径
func (s *StaticAnalysisBiz) dfsPath(graph map[string][]string, current, target string,
	currentPath []string, visited map[string]bool, paths *[]*v1.CallPath,
	maxPaths, maxDepth, currentDepth int) {

	if len(*paths) >= maxPaths || currentDepth >= maxDepth {
		return
	}

	if current == target {
		// 找到路径
		pathCopy := make([]string, len(currentPath))
		copy(pathCopy, currentPath)

		path := &v1.CallPath{
			FunctionKeys: pathCopy,
			PathLength:   int32(len(pathCopy)),
			Confidence:   1.0 / float32(len(pathCopy)), // 简单的可信度计算
		}
		*paths = append(*paths, path)
		return
	}

	if visited[current] {
		return // 避免循环
	}

	visited[current] = true

	for _, neighbor := range graph[current] {
		currentPath = append(currentPath, neighbor)
		s.dfsPath(graph, neighbor, target, currentPath, visited, paths, maxPaths, maxDepth, currentDepth+1)
		currentPath = currentPath[:len(currentPath)-1] // 回溯
	}

	visited[current] = false
}
