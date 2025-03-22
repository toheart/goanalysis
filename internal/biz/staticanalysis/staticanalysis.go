package staticanalysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/sourcegraph/conc/pool"
	"github.com/toheart/goanalysis/internal/biz/callgraph"
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

// GetHotFunctions 获取热点函数
func (s *StaticAnalysisBiz) GetHotFunctions(sortBy string) ([]entity.Function, error) {
	// 这里实现获取热点函数的逻辑
	// 示例数据
	return []entity.Function{
		{
			Name:      "main.main",
			Package:   "main",
			CallCount: 1,
			TotalTime: "10ms",
			AvgTime:   "10ms",
		},
		{
			Name:      "internal/biz.Process",
			Package:   "internal/biz",
			CallCount: 45,
			TotalTime: "500ms",
			AvgTime:   "11.1ms",
		},
		{
			Name:      "internal/data.Query",
			Package:   "internal/data",
			CallCount: 78,
			TotalTime: "800ms",
			AvgTime:   "10.3ms",
		},
	}, nil
}

// GetFunctionAnalysis 获取函数调用关系分析
func (s *StaticAnalysisBiz) GetFunctionAnalysis(functionName, queryType, path string) ([]entity.FunctionNode, error) {
	// 这里实现获取函数调用关系分析的逻辑
	// 示例数据
	return []entity.FunctionNode{
		{
			ID:        "1",
			Name:      functionName,
			Package:   "main",
			CallCount: 10,
			AvgTime:   "5ms",
			Children:  []entity.FunctionNode{},
		},
	}, nil
}

// GetFunctionCallGraph 获取函数调用关系图
func (s *StaticAnalysisBiz) GetFunctionCallGraph(functionName string, depth int, direction string) ([]entity.FunctionGraphNode, []entity.FunctionGraphEdge, error) {
	// 这里实现获取函数调用关系图的逻辑
	// 示例数据
	nodes := []entity.FunctionGraphNode{
		{
			ID:        "1",
			Name:      functionName,
			Package:   "main",
			CallCount: 1,
			AvgTime:   "10ms",
			NodeType:  "root",
		},
		{
			ID:        "2",
			Name:      "subFunc",
			Package:   "main",
			CallCount: 5,
			AvgTime:   "2ms",
			NodeType:  "callee",
		},
	}

	edges := []entity.FunctionGraphEdge{
		{
			Source:   "1",
			Target:   "2",
			Label:    "calls",
			EdgeType: "root_to_callee",
		},
	}

	return nodes, edges, nil
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
	wg.Add(1)
	// 启动调用图生成
	go func() {
		defer wg.Done()
		if err := c.SetTree(statusChan); err != nil {
			errMsg := fmt.Sprintf("Call graph generation failed: %v", err)
			statusChan <- []byte(errMsg)
			s.log.Error(errMsg)
			return
		}
		// 标记为完成
		statusChan <- []byte("Analysis task completed")
		completed = true
		s.log.Info("Analysis task completed")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 3)
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
	// 启动保存数据
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 发送状态消息
		statusChan <- []byte("Starting to build call graph...")
		// 保存数据
		statusChan <- []byte("Starting to save data to database...")
		if err := c.SaveData(context.Background(), statusChan); err != nil {
			errMsg := fmt.Sprintf("Failed to save data: %v", err)
			statusChan <- []byte(errMsg)
			s.log.Error(errMsg)
			return
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

// GetTreeGraph 获取静态分析的树状图数据
func (s *StaticAnalysisBiz) GetTreeGraph(functionName string, dbPath string) (*entity.TreeGraph, error) {
	s.log.Infof("get tree graph, function: %s, dbpath: %s", functionName, dbPath)

	// 检查数据库路径是否有效
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file not found: %s", dbPath)
	}

	// 构建树状图结构
	root := &entity.TreeNode{
		Name: functionName,
	}

	// 获取函数调用图
	nodes, edges, err := s.GetFunctionCallGraph(functionName, 3, "outgoing") // 获取向外的调用，深度3
	if err != nil {
		s.log.Errorf("get function call graph failed: %v", err)
		// 返回只有根节点的树
		return &entity.TreeGraph{Root: root}, nil
	}

	// 创建节点映射，用于快速查找
	nodeMap := make(map[string]*entity.TreeNode)
	nodeMap[functionName] = root

	// 建立节点ID到节点的映射
	funcNodes := make(map[string]entity.FunctionGraphNode)
	for _, node := range nodes {
		funcNodes[node.ID] = node
	}

	// 从调用图构建树状图
	for _, edge := range edges {
		sourceName := funcNodes[edge.Source].Name
		targetName := funcNodes[edge.Target].Name

		// 只处理从当前节点出发的边，避免形成环
		if sourceName == functionName || nodeMap[sourceName] != nil {
			sourceNode := nodeMap[sourceName]

			// 创建或获取目标节点
			targetNode, exists := nodeMap[targetName]
			if !exists {
				targetNode = &entity.TreeNode{
					Name:  targetName,
					Value: int64(funcNodes[edge.Target].CallCount),
				}
				nodeMap[targetName] = targetNode
			}

			// 添加子节点
			sourceNode.Children = append(sourceNode.Children, targetNode)
		}
	}

	return &entity.TreeGraph{Root: root}, nil
}
