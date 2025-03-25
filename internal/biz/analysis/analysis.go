package analysis

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/toheart/functrace"
	v1 "github.com/toheart/goanalysis/api/analysis/v1"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
	sqlite "github.com/toheart/goanalysis/internal/data/sqlite"
)

type AnalysisBiz struct {
	conf *conf.Biz
	data *data.Data
	log  *log.Helper
}

func (a *AnalysisBiz) GetTotalGIDs(dbpath string) (int, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return 0, err
	}
	return traceDB.GetTotalGIDs()
}

func (a *AnalysisBiz) GetAllFunctionName(dbpath string) ([]string, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllFunctionName()
}

func NewAnalysisBiz(conf *conf.Biz, data *data.Data, logger log.Logger) *AnalysisBiz {
	return &AnalysisBiz{conf: conf, data: data, log: log.NewHelper(logger)}
}

func (a *AnalysisBiz) GetTracesByGID(req *v1.AnalysisByGIDRequest) ([]entity.TraceData, error) {
	a.log.Infof("get traces by gid: %s from db: %s", req.Gid, req.Dbpath)
	traceDB, err := a.data.GetTraceDB(req.Dbpath)
	if err != nil {
		return nil, err
	}

	return traceDB.GetTracesByGID(req.Gid, int(req.Depth), req.CreateTime)
}

func (a *AnalysisBiz) GetAllGIDs(dbpath string, page int, limit int) ([]entity.GoroutineTrace, error) {
	a.log.Infof("get all gids from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllGIDs(page, limit)
}

func (a *AnalysisBiz) GetInitialFunc(dbpath string, gid uint64) (string, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return "", err
	}
	return traceDB.GetInitialFunc(gid)
}

func (a *AnalysisBiz) GetParamsByID(dbpath string, id int32) ([]functrace.TraceParams, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetParamsByID(id)
}

func (a *AnalysisBiz) GetGidsByFunctionName(dbpath string, functionName string) ([]string, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetGidsByFunctionName(functionName)
}

func (a *AnalysisBiz) VerifyProjectPath(path string) bool {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	a.log.Infof("verified db path: %s", path)
	return true
}

// GetTracesByParentFunc 根据父函数ID获取函数调用
func (a *AnalysisBiz) GetTracesByParentFunc(dbpath string, parentId int64) ([]entity.TraceData, error) {
	a.log.Infof("get traces by parent id: %d from db: %s", parentId, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetTracesByParentId(parentId)
}

// GetAllParentIds 获取所有的父函数ID
func (a *AnalysisBiz) GetParentFunctions(dbpath string, functionName string) ([]*entity.Function, error) {
	a.log.Infof("get all parent ids from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllParentFunctions(functionName)
}

// GetChildFunctions 获取函数的子函数
func (a *AnalysisBiz) GetChildFunctions(dbpath string, parentId int64) ([]*entity.Function, error) {
	a.log.Infof("get child functions of parent id: %d from db: %s", parentId, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetChildFunctions(parentId)
}

// GetHotFunctions 获取热点函数分析数据
func (a *AnalysisBiz) GetHotFunctions(dbpath string, sortBy string) ([]entity.Function, error) {
	a.log.Infof("get hot functions, sort by: %s from db: %s", sortBy, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetHotFunctions(sortBy)
}

// GetGoroutineStats 获取Goroutine统计信息
func (a *AnalysisBiz) GetGoroutineStats(dbpath string) (*entity.GoroutineStats, error) {
	a.log.Infof("get goroutine stats from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetGoroutineStats()
}

// GetFunctionAnalysis 获取函数调用关系分析
func (a *AnalysisBiz) GetFunctionAnalysis(dbpath string, functionName string, queryType string) ([]entity.FunctionNode, error) {
	a.log.Infof("get function analysis, function: %s, type: %s from db: %s", functionName, queryType, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetFunctionAnalysis(functionName, queryType)
}

// GetGoroutineCallDepth 获取指定 Goroutine 的最大调用深度
func (a *AnalysisBiz) GetGoroutineCallDepth(dbpath string, gid uint64) (int, error) {
	a.log.Infof("get goroutine call depth for gid: %d from db: %s", gid, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return 0, err
	}
	return traceDB.GetGoroutineCallDepth(gid)
}

// GetGoroutineExecutionTime 获取指定 Goroutine 的总执行时间
func (a *AnalysisBiz) GetGoroutineExecutionTime(dbpath string, groutine entity.GoroutineTrace) (string, error) {
	a.log.Infof("get goroutine execution time for gid: %d from db: %s", groutine.ID, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return "", err
	}
	if groutine.TimeCost != "" {
		return groutine.TimeCost, nil
	}
	// 获取最后一个函数的执行时间
	trace, err := traceDB.GetLastFunction()
	if err != nil {
		return "", err
	}
	createTime, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
	if err != nil {
		return "", err
	}
	gcreateTime, err := time.Parse(time.RFC3339Nano, groutine.CreateTime)
	if err != nil {
		return "", err
	}
	duration := createTime.Sub(gcreateTime)
	executionTime := duration.String()

	return executionTime, nil
}

func (a *AnalysisBiz) GetAllGoroutineTrace(dbpath string, includeMetrics bool) ([]entity.GoroutineTrace, error) {
	a.log.Infof("get all goroutine trace from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	// 获取所有goroutine trace
	groutines, err := traceDB.GetAllGIDs(0, 1000)
	if err != nil {
		return nil, err
	}
	if includeMetrics {
		// 获取所有goroutine trace
		for _, g := range groutines {
			depth, err := a.GetGoroutineCallDepth(dbpath, uint64(g.ID))
			if err != nil {
				return nil, err
			}
			g.Depth = depth
			execTime, err := a.GetGoroutineExecutionTime(dbpath, g)
			if err != nil {
				return nil, err
			}
			g.TimeCost = execTime
		}
	}

	return groutines, nil
}

// IsGoroutineFinished 检查指定的goroutine是否已完成
func (a *AnalysisBiz) IsGoroutineFinished(dbpath string, gid uint64) (bool, error) {
	a.log.Infof("check if goroutine is finished, gid: %d from db: %s", gid, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return false, err
	}
	return traceDB.IsGoroutineFinished(gid)
}

// GetUnfinishedFunctions 获取未完成的函数列表
func (a *AnalysisBiz) GetUnfinishedFunctions(dbpath string, threshold int64) ([]entity.AllUnfinishedFunction, error) {
	a.log.Infof("get unfinished functions with threshold: %d ms from db: %s", threshold, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	functions, err := traceDB.GetAllUnfinishedFunctions(threshold)
	if err != nil {
		return nil, err
	}

	return functions, nil
}

func (a *AnalysisBiz) GetUpstreamTreeGraph(dbpath string, functionName string, depth int) ([]*entity.TreeNode, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}

	// 获取指定函数的所有跟踪数据
	traces, err := traceDB.GetTracesByFuncName(functionName)
	if err != nil {
		return nil, err
	}

	// 存储所有生成的树节点
	var trees []*entity.TreeNode

	// 为每个跟踪数据创建一个树
	for _, trace := range traces {
		// 创建根节点
		rootNode := &entity.TreeNode{
			Name:  functionName,
			Value: trace.ID,
		}

		// 递归构建上游调用树
		err = a.buildUpstreamTree(traceDB, rootNode, trace.ParentId, depth-1)
		if err != nil {
			return nil, err
		}

		trees = append(trees, rootNode)
	}

	return trees, nil
}

// 递归构建上游调用树
func (a *AnalysisBiz) buildUpstreamTree(traceDB *sqlite.TraceEntDB, currentNode *entity.TreeNode, parentId uint64, remainingDepth int) error {
	// 如果父ID为0或者已达到最大深度，则停止递归
	if parentId == 0 || remainingDepth <= 0 {
		return nil
	}

	// 获取父函数的跟踪数据
	parentTrace, err := traceDB.GetTraceByID(int(parentId))
	if err != nil {
		return err
	}

	if parentTrace == nil {
		return nil // 父跟踪不存在，可能是根调用
	}

	// 创建父节点
	parentNode := &entity.TreeNode{
		Name:  parentTrace.Name,
		Value: parentTrace.ID,
	}

	// 将当前节点添加为父节点的子节点
	if parentNode.Children == nil {
		parentNode.Children = []*entity.TreeNode{}
	}
	parentNode.Children = append(parentNode.Children, currentNode)

	// 继续向上递归构建树
	return a.buildUpstreamTree(traceDB, parentNode, parentTrace.ParentId, remainingDepth-1)
}

func (a *AnalysisBiz) GetDownstreamTreeGraph(dbpath string, functionName string, depth int) (*entity.TreeNode, error) {
	a.log.Infof("get downstream tree graph, function: %s, depth: %d, dbpath: %s", functionName, depth, dbpath)

	// 获取追踪数据库
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, fmt.Errorf("get trace db failed: %w", err)
	}

	// 获取函数的所有调用记录
	traces, err := traceDB.GetTracesByFuncName(functionName)
	if err != nil {
		return nil, fmt.Errorf("get traces by func name failed: %w", err)
	}

	// 如果没有找到调用记录，返回空数组
	if len(traces) == 0 {
		return nil, nil
	}

	// 为每个trace创建一个根节点
	root := &entity.TreeNode{
		Name:     functionName,
		Value:    traces[0].ID,
		Children: []*entity.TreeNode{},
	}
	parentIds := []int64{}

	for _, trace := range traces {
		parentIds = append(parentIds, trace.ID)
	}

	// 递归构建下游调用树
	err = a.buildDownstreamTree(traceDB, root, parentIds, depth)
	if err != nil {
		return nil, fmt.Errorf("build downstream tree failed: %w", err)
	}

	return root, nil
}

// 递归构建下游调用树
func (a *AnalysisBiz) buildDownstreamTree(traceDB *sqlite.TraceEntDB, currentNode *entity.TreeNode, parentIds []int64, remainingDepth int) error {
	// 如果已达到最大深度，则停止递归
	if remainingDepth <= 0 || currentNode == nil {
		return nil
	}

	for _, parentId := range parentIds {
		// 获取当前函数调用的子函数
		childFunctions, err := traceDB.GetTracesByParentId(parentId)
		if err != nil {
			return err
		}

		// 为每个子函数创建节点
		for _, childFunc := range childFunctions {
			// 获取子函数的统计信息

			// 创建子节点
			childNode := &entity.TreeNode{
				Name:     childFunc.Name,
				Children: []*entity.TreeNode{},
			}

			// 递归构建子节点的下游调用
			err = a.buildDownstreamTree(traceDB, childNode, []int64{childFunc.ID}, remainingDepth-1)
			if err != nil {
				return err
			}
			currentNode.Children = append(currentNode.Children, childNode)
		}
	}

	return nil
}

// GetTreeGraph 获取运行时的树状图数据
func (a *AnalysisBiz) GetTreeGraph(dbpath string, functionName string, chainType string, depth int) ([]*entity.TreeNode, error) {
	a.log.Infof("get tree graph, function: %s, chain type: %s, dbpath: %s", functionName, chainType, dbpath)

	// 根据链路类型选择不同的查询方向
	var trees []*entity.TreeNode
	var err error
	switch chainType {
	case "upstream":
		trees, err = a.GetUpstreamTreeGraph(dbpath, functionName, depth)
		if err != nil {
			return nil, err
		}
	case "downstream":
		tree, err := a.GetDownstreamTreeGraph(dbpath, functionName, depth)
		if err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	case "full":
		// 全链路需要分别获取上游和下游调用，然后合并
		upstreamGraph, err := a.GetUpstreamTreeGraph(dbpath, functionName, depth)
		if err != nil {
			a.log.Warnf("get upstream call graph failed: %v", err)
		}

		downstreamGraph, err := a.GetDownstreamTreeGraph(dbpath, functionName, depth)
		if err != nil {
			a.log.Warnf("get downstream call graph failed: %v", err)
		}

		// 合并处理上游和下游调用图
		trees, err := a.buildFullChainTreeNodes(functionName, upstreamGraph, downstreamGraph)
		if err != nil {
			return nil, err
		}
		return trees, nil
	default:
		return nil, fmt.Errorf("invalid chain type: %s", chainType)
	}

	return trees, nil
}

func (a *AnalysisBiz) buildFullChainTreeNodes(functionName string, upstreamGraph []*entity.TreeNode, downstreamGraph *entity.TreeNode) ([]*entity.TreeNode, error) {
	// 如果上游图为空，直接返回下游图
	if len(upstreamGraph) == 0 {
		return []*entity.TreeNode{downstreamGraph}, nil
	}

	for _, node := range upstreamGraph {
		// 递归查找叶子节点
		var findLeafNodes func(node *entity.TreeNode) *entity.TreeNode
		findLeafNodes = func(node *entity.TreeNode) *entity.TreeNode {
			// 如果节点没有子节点，则为叶子节点
			if len(node.Children) == 0 {
				return node
			}

			// 递归处理所有子节点
			for _, child := range node.Children {
				leafNode := findLeafNodes(child)
				if leafNode != nil {
					return leafNode
				}
			}
			return nil
		}

		// 对当前节点执行递归查找
		leafNode := findLeafNodes(node)
		if leafNode != nil {
			leafNode.Children = append(leafNode.Children, downstreamGraph)
		}
	}

	return upstreamGraph, nil
}

// 根据节点ID获取节点名称
func getNodeNameById(nodes []entity.FunctionGraphNode, id string) string {
	for _, node := range nodes {
		if node.ID == id {
			return node.Name
		}
	}
	return id // 如果找不到，返回ID本身
}

// 从边标签解析调用次数
func parseCallCount(label string) (int, error) {
	var count int
	_, err := fmt.Sscanf(label, "调用次数: %d", &count)
	if err != nil {
		return 1, err // 默认返回1
	}
	return count, nil
}

// GetTreeGraphByGID 根据GID获取多棵树状图数据
func (a *AnalysisBiz) GetTreeGraphByGID(dbpath string, gid uint64) ([]*entity.TreeNode, error) {
	a.log.Infof("get tree graph by gid: %d, dbpath: %s", gid, dbpath)

	// 获取追踪数据库
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}

	// 获取指定GID的所有调用跟踪数据
	traces, err := traceDB.GetTracesByGID(gid, 3, "")
	if err != nil {
		return nil, err
	}

	// 如果没有找到相关调用数据
	if len(traces) == 0 {
		return []*entity.TreeNode{}, nil
	}

	// 按照调用层级构建多棵树
	// 首先找到所有根调用，即没有父函数的调用
	var rootTrees []*entity.TreeNode

	// 根据ParentId对调用进行分组
	tracesByParentId := make(map[uint64][]entity.TraceData)
	traceById := make(map[int64]entity.TraceData)

	for _, trace := range traces {
		tracesByParentId[trace.ParentId] = append(tracesByParentId[trace.ParentId], trace)
		traceById[trace.ID] = trace
	}

	// 找出所有根调用（ParentId为0或不存在父调用的）
	rootTraces := tracesByParentId[0]

	// 如果没有明确的根调用，则使用缩进级别为0的调用作为根
	if len(rootTraces) == 0 {
		for _, trace := range traces {
			if trace.Indent == 0 {
				rootTraces = append(rootTraces, trace)
			}
		}
	}

	// 如果仍然没有根调用，则使用所有调用中缩进级别最小的作为根
	if len(rootTraces) == 0 && len(traces) > 0 {
		minIndent := traces[0].Indent
		for _, trace := range traces {
			if trace.Indent < minIndent {
				minIndent = trace.Indent
			}
		}

		for _, trace := range traces {
			if trace.Indent == minIndent {
				rootTraces = append(rootTraces, trace)
			}
		}
	}

	// 为每个根调用构建树
	for _, rootTrace := range rootTraces {
		rootNode := &entity.TreeNode{
			Name:  rootTrace.Name,
			Value: int64(rootTrace.ID),
		}

		// 递归构建子树
		buildTreeFromTrace(rootNode, rootTrace.ID, tracesByParentId, 0, 3) // 限制最大深度为5层

		rootTrees = append(rootTrees, rootNode)
	}

	return rootTrees, nil
}

// buildTreeFromTrace 递归构建调用树
func buildTreeFromTrace(node *entity.TreeNode, traceId int64, tracesByParentId map[uint64][]entity.TraceData, currentDepth, maxDepth int) {
	// 防止过深递归
	if currentDepth >= maxDepth {
		return
	}

	// 获取当前调用的所有子调用
	childTraces := tracesByParentId[uint64(traceId)]

	// 为每个子调用创建节点
	for _, childTrace := range childTraces {
		childNode := &entity.TreeNode{
			Name:  childTrace.Name,
			Value: childTrace.ID,
		}
		// 添加子节点
		node.Children = append(node.Children, childNode)

		// 递归处理子调用
		buildTreeFromTrace(childNode, childTrace.ID, tracesByParentId, currentDepth+1, maxDepth)
	}
}

// GetFunctionHotPaths 获取函数热点路径分析
func (a *AnalysisBiz) GetFunctionHotPaths(dbpath string, functionName string, limit int) ([]entity.HotPathInfo, error) {
	// TODO 没想法怎么做
	return nil, nil
}

// calculatePathStats 计算路径的调用统计信息
func (a *AnalysisBiz) calculatePathStats(traceDB *sqlite.TraceEntDB, functionNames []string) (int, string, string) {
	// 这里应该根据实际情况从数据库中查询这条路径的统计信息
	// 简单实现：假设调用次数是叶子节点函数的调用次数
	if len(functionNames) == 0 {
		return 0, "0ms", "0ms"
	}

	leafFunction := functionNames[len(functionNames)-1]

	// 获取叶子函数的所有调用数据
	traces, err := traceDB.GetTracesByFuncName(leafFunction)
	if err != nil {
		a.log.Warnf("获取函数%s的调用数据失败: %v", leafFunction, err)
		return 0, "0ms", "0ms"
	}

	// 统计调用次数和总时间
	callCount := len(traces)

	// 解析时间
	var totalTimeMs float64
	for _, trace := range traces {
		timeCostMs := parseTimeCost(trace.TimeCost)
		totalTimeMs += timeCostMs
	}

	// 计算平均时间
	var avgTimeMs float64
	if callCount > 0 {
		avgTimeMs = totalTimeMs / float64(callCount)
	}

	// 格式化时间
	totalTime := formatTime(totalTimeMs)
	avgTime := formatTime(avgTimeMs)

	return callCount, totalTime, avgTime
}

// sortHotPathsByCallCount 按调用次数对热点路径进行排序
func (a *AnalysisBiz) sortHotPathsByCallCount(paths []entity.HotPathInfo) {
	// 按调用次数降序排序
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].CallCount > paths[j].CallCount
	})
}

// GetFunctionCallStats 获取函数调用统计分析
func (a *AnalysisBiz) GetFunctionCallStats(dbpath string, functionName string) ([]entity.FunctionCallStats, error) {
	a.log.Infof("获取函数调用统计，函数：%s，数据库：%s", functionName, dbpath)

	// 获取追踪数据库
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}

	var stats []entity.FunctionCallStats

	// 如果指定了函数名，则只获取该函数的统计信息
	if functionName != "" {
		// 获取函数的所有调用记录
		traces, err := traceDB.GetTracesByFuncName(functionName)
		if err != nil {
			a.log.Errorf("get function call stats failed: %v", err)
			return nil, err
		}

		// 获取调用该函数的函数集合（调用方）
		callers, err := traceDB.GetCallerFunctions(functionName)
		if err != nil {
			a.log.Warnf("get function call stats failed: %v", err)
			callers = []string{}
		}

		// 获取该函数调用的函数集合（被调用方）
		callees, err := traceDB.GetCalleeFunctions(functionName)
		if err != nil {
			a.log.Warnf("get function call stats failed: %v", err)
			callees = []string{}
		}

		// 计算执行时间统计信息
		timeCosts := make([]float64, 0, len(traces))
		var totalTime float64
		var maxTime float64
		var minTime float64 = -1 // 初始化为-1表示未设置

		for _, trace := range traces {
			timeCostMs := parseTimeCost(trace.TimeCost)

			timeCosts = append(timeCosts, timeCostMs)
			totalTime += timeCostMs

			if timeCostMs > maxTime {
				maxTime = timeCostMs
			}

			if minTime < 0 || timeCostMs < minTime {
				minTime = timeCostMs
			}
		}

		// 计算平均时间和标准差
		var avgTime float64
		var stdDev float64

		if len(traces) > 0 {
			avgTime = totalTime / float64(len(traces))

			// 计算标准差
			var sumSquares float64
			for _, t := range timeCosts {
				diff := t - avgTime
				sumSquares += diff * diff
			}

			if len(timeCosts) > 1 {
				stdDev = math.Sqrt(sumSquares / float64(len(timeCosts)-1))
			}
		}

		// 如果没有调用记录，设置最小时间为0
		if minTime < 0 {
			minTime = 0
		}

		// 创建函数调用统计信息
		stat := entity.FunctionCallStats{
			Name:        functionName,
			Package:     extractPackage(functionName),
			CallCount:   len(traces),
			CallerCount: len(callers),
			CalleeCount: len(callees),
			AvgTime:     formatTime(avgTime),
			MaxTime:     formatTime(maxTime),
			MinTime:     formatTime(minTime),
			TimeStdDev:  stdDev,
		}

		stats = append(stats, stat)
	} else {
		// 获取所有函数名称
		funcNames, err := traceDB.GetAllFunctionName()
		if err != nil {
			a.log.Errorf("获取所有函数名称失败: %v", err)
			return nil, err
		}

		// 对每个函数计算统计信息
		for _, funcName := range funcNames {
			// 递归调用自身以获取单个函数的统计信息
			funcStats, err := a.GetFunctionCallStats(dbpath, funcName)
			if err != nil {
				a.log.Warnf("获取函数%s的统计信息失败: %v", funcName, err)
				continue
			}

			if len(funcStats) > 0 {
				stats = append(stats, funcStats[0])
			}
		}
	}

	return stats, nil
}

// GetPerformanceAnomalies 获取性能异常检测结果
func (a *AnalysisBiz) GetPerformanceAnomalies(dbpath string, functionName string, threshold float64) ([]entity.PerformanceAnomaly, error) {
	a.log.Infof("获取性能异常检测，函数：%s，数据库：%s，阈值：%f", functionName, dbpath, threshold)

	// 获取追踪数据库
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}

	// 如果阈值未设置，使用默认值
	if threshold <= 0 {
		threshold = 2.0 // 默认为2个标准差
	}

	var anomalies []entity.PerformanceAnomaly

	// 如果指定了函数名，则只检测该函数
	if functionName != "" {
		// 获取函数的统计信息
		stats, err := a.GetFunctionCallStats(dbpath, functionName)
		if err != nil {
			a.log.Errorf("获取函数%s的统计信息失败: %v", functionName, err)
			return nil, err
		}

		if len(stats) > 0 {
			// 检测时间波动异常
			if a.detectTimeVarianceAnomaly(&stats[0], threshold) {
				anomalies = append(anomalies, createTimeVarianceAnomaly(&stats[0], threshold))
			}

			// 检测调用深度异常
			if a.detectDepthAnomaly(traceDB, functionName, threshold) {
				anomalies = append(anomalies, createDepthAnomaly(functionName, extractPackage(functionName), threshold))
			}

			// 检测调用频率异常
			if a.detectFrequencyAnomaly(traceDB, functionName, threshold) {
				anomalies = append(anomalies, createFrequencyAnomaly(functionName, extractPackage(functionName), threshold))
			}
		}
	} else {
		// 获取所有函数的统计信息
		allStats, err := a.GetFunctionCallStats(dbpath, "")
		if err != nil {
			a.log.Errorf("获取所有函数统计信息失败: %v", err)
			return nil, err
		}

		// 计算所有函数的标准差的均值，用于比较
		var totalStdDev float64
		for _, stat := range allStats {
			totalStdDev += stat.TimeStdDev
		}
		avgStdDev := 0.0
		if len(allStats) > 0 {
			avgStdDev = totalStdDev / float64(len(allStats))
		}

		// 对每个函数检测异常
		for _, stat := range allStats {
			// 检测时间波动异常，使用相对于平均标准差的比例
			if stat.TimeStdDev > avgStdDev*threshold {
				anomalies = append(anomalies, createTimeVarianceAnomaly(&stat, threshold))
			}

			// 检测调用深度异常
			if a.detectDepthAnomaly(traceDB, stat.Name, threshold) {
				anomalies = append(anomalies, createDepthAnomaly(stat.Name, stat.Package, threshold))
			}

			// 检测调用频率异常
			if a.detectFrequencyAnomaly(traceDB, stat.Name, threshold) {
				anomalies = append(anomalies, createFrequencyAnomaly(stat.Name, stat.Package, threshold))
			}
		}
	}

	// 按严重程度排序
	sort.Slice(anomalies, func(i, j int) bool {
		return anomalies[i].Severity > anomalies[j].Severity
	})

	return anomalies, nil
}

// 辅助函数

// 检测时间波动异常
func (a *AnalysisBiz) detectTimeVarianceAnomaly(stat *entity.FunctionCallStats, threshold float64) bool {
	// 如果函数只被调用一次，无法判断波动
	if stat.CallCount <= 1 {
		return false
	}

	// 计算变异系数(CV)：标准差/平均值，用于衡量相对波动程度
	avgTimeMs := parseTimeCost(stat.AvgTime)
	if avgTimeMs <= 0 {
		return false
	}

	cv := stat.TimeStdDev / avgTimeMs

	// 如果变异系数大于阈值，则认为存在时间波动异常
	return cv > threshold/10 // 调整阈值，使其更符合CV的比例
}

// 检测调用深度异常
func (a *AnalysisBiz) detectDepthAnomaly(traceDB *sqlite.TraceEntDB, functionName string, threshold float64) bool {
	// 获取该函数在所有Goroutine中的调用深度
	depths, err := traceDB.GetFunctionCallDepths(functionName)
	if err != nil {
		a.log.Warnf("获取函数%s的调用深度失败: %v", functionName, err)
		return false
	}

	// 如果样本太少，无法判断
	if len(depths) <= 1 {
		return false
	}

	// 计算深度的平均值和标准差
	var totalDepth int
	for _, d := range depths {
		totalDepth += d
	}
	avgDepth := float64(totalDepth) / float64(len(depths))

	var sumSquares float64
	for _, d := range depths {
		diff := float64(d) - avgDepth
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(depths)-1))

	// 检查是否有深度异常的调用
	for _, d := range depths {
		if math.Abs(float64(d)-avgDepth) > stdDev*threshold {
			return true
		}
	}

	return false
}

// 检测调用频率异常
func (a *AnalysisBiz) detectFrequencyAnomaly(traceDB *sqlite.TraceEntDB, functionName string, threshold float64) bool {
	// 获取该函数的所有调用记录
	traces, err := traceDB.GetTracesByFuncName(functionName)
	if err != nil {
		a.log.Warnf("获取函数%s的调用记录失败: %v", functionName, err)
		return false
	}

	// 如果样本太少，无法判断
	if len(traces) <= 2 {
		return false
	}

	// 按时间戳排序，计算调用间隔
	sort.Slice(traces, func(i, j int) bool {
		return traces[i].CreatedAt < traces[j].CreatedAt
	})

	// 计算调用间隔
	intervals := make([]float64, len(traces)-1)
	for i := 1; i < len(traces); i++ {
		// 解析时间戳，计算间隔（毫秒）
		t1, err1 := time.Parse(time.RFC3339, traces[i-1].CreatedAt)
		t2, err2 := time.Parse(time.RFC3339, traces[i].CreatedAt)

		if err1 != nil || err2 != nil {
			continue
		}

		interval := t2.Sub(t1).Milliseconds()
		intervals[i-1] = float64(interval)
	}

	// 计算间隔的平均值和标准差
	var totalInterval float64
	for _, interval := range intervals {
		totalInterval += interval
	}
	avgInterval := totalInterval / float64(len(intervals))

	var sumSquares float64
	for _, interval := range intervals {
		diff := interval - avgInterval
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(intervals)-1))

	// 计算变异系数
	cv := stdDev / avgInterval

	// 如果变异系数大于阈值，则认为存在频率异常
	return cv > threshold/5 // 调整阈值，使其更符合调用频率CV的比例
}

// 创建时间波动异常
func createTimeVarianceAnomaly(stat *entity.FunctionCallStats, threshold float64) entity.PerformanceAnomaly {
	// 计算变异系数
	avgTimeMs := parseTimeCost(stat.AvgTime)
	cv := stat.TimeStdDev / avgTimeMs

	// 计算严重程度 (0.0-1.0)
	severity := math.Min(cv/(threshold/5), 1.0)

	// 创建详细信息
	details := map[string]string{
		"avg_time":     stat.AvgTime,
		"max_time":     stat.MaxTime,
		"min_time":     stat.MinTime,
		"std_dev":      fmt.Sprintf("%.2f ms", stat.TimeStdDev),
		"call_count":   fmt.Sprintf("%d", stat.CallCount),
		"cv":           fmt.Sprintf("%.2f", cv),
		"threshold_cv": fmt.Sprintf("%.2f", threshold/10),
	}

	return entity.PerformanceAnomaly{
		Name:        stat.Name,
		Package:     stat.Package,
		AnomalyType: "time_variance",
		Description: fmt.Sprintf("函数 %s 的执行时间波动较大，变异系数为 %.2f (阈值: %.2f)", stat.Name, cv, threshold/10),
		Severity:    severity,
		Details:     details,
	}
}

// 创建深度异常
func createDepthAnomaly(funcName string, pkg string, threshold float64) entity.PerformanceAnomaly {
	// 固定严重程度为0.7，因为深度异常通常表示潜在的调用栈问题
	severity := 0.7

	details := map[string]string{
		"threshold": fmt.Sprintf("%.1f 个标准差", threshold),
	}

	return entity.PerformanceAnomaly{
		Name:        funcName,
		Package:     pkg,
		AnomalyType: "depth_anomaly",
		Description: fmt.Sprintf("函数 %s 在不同调用中的调用深度存在异常波动", funcName),
		Severity:    severity,
		Details:     details,
	}
}

// 创建频率异常
func createFrequencyAnomaly(funcName string, pkg string, threshold float64) entity.PerformanceAnomaly {
	// 固定严重程度为0.5，因为频率异常可能是正常的业务逻辑导致的
	severity := 0.5

	details := map[string]string{
		"threshold": fmt.Sprintf("%.1f 个标准差", threshold/5),
	}

	return entity.PerformanceAnomaly{
		Name:        funcName,
		Package:     pkg,
		AnomalyType: "frequency_anomaly",
		Description: fmt.Sprintf("函数 %s 的调用频率存在异常波动", funcName),
		Severity:    severity,
		Details:     details,
	}
}

// 提取包名
func extractPackage(functionName string) string {
	lastDot := strings.LastIndex(functionName, ".")
	if lastDot <= 0 {
		return ""
	}
	return functionName[:lastDot]
}

// 解析时间字符串为毫秒数
func parseTimeCost(timeCost string) float64 {
	// 默认返回0
	if timeCost == "" {
		return 0
	}

	var value float64

	// 处理"xxx ms"格式
	if strings.HasSuffix(timeCost, "ms") {
		ms, err := strconv.ParseFloat(strings.TrimSuffix(timeCost, "ms"), 64)
		if err == nil {
			value = ms
		}
	} else if strings.HasSuffix(timeCost, "s") {
		// 处理"xxx s"格式，转换为毫秒
		s, err := strconv.ParseFloat(strings.TrimSuffix(timeCost, "s"), 64)
		if err == nil {
			value = s * 1000
		}
	} else if strings.HasSuffix(timeCost, "µs") {
		// 处理"xxx µs"格式，转换为毫秒
		us, err := strconv.ParseFloat(strings.TrimSuffix(timeCost, "µs"), 64)
		if err == nil {
			value = us / 1000
		}
	} else {
		// 尝试直接解析为数字
		v, err := strconv.ParseFloat(timeCost, 64)
		if err == nil {
			value = v
		}
	}

	return value
}

// 格式化毫秒数为时间字符串
func formatTime(timeMs float64) string {
	if timeMs < 1 {
		return fmt.Sprintf("%.3f µs", timeMs*1000)
	} else if timeMs < 1000 {
		return fmt.Sprintf("%.2f ms", timeMs)
	} else {
		return fmt.Sprintf("%.3f s", timeMs/1000)
	}
}

// SearchFunctions 搜索函数
func (a *AnalysisBiz) SearchFunctions(ctx context.Context, dbPath string, query string, limit int32) ([]*entity.Function, int32, error) {
	// 验证参数
	if dbPath == "" {
		return nil, 0, errors.New("database path is required")
	}
	if query == "" {
		return nil, 0, errors.New("search query is required")
	}
	if limit <= 0 {
		limit = 10 // 默认限制返回10条结果
	}
	a.log.Infof("search functions, dbpath: %s, query: %s, limit: %d", dbPath, query, limit)
	traceDB, err := a.data.GetTraceDB(dbPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get trace db: %v", err)
	}

	// 调用数据层执行搜索
	functions, total, err := traceDB.SearchFunctions(ctx, dbPath, query, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search functions: %v", err)
	}

	return functions, total, nil
}
