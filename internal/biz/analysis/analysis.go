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
	v1 "github.com/toheart/goanalysis/api/analysis/v1"
	"github.com/toheart/goanalysis/internal/biz/analysis/dos"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
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

func NewAnalysisBiz(conf *conf.Biz, data *data.Data, logger log.Logger) *AnalysisBiz {
	return &AnalysisBiz{conf: conf, data: data, log: log.NewHelper(logger)}
}

func (a *AnalysisBiz) GetTracesByGID(req *v1.AnalysisByGIDRequest) ([]dos.TraceData, error) {
	a.log.Infof("get traces by gid: %s from db: %s", req.Gid, req.Dbpath)
	traceDB, err := a.data.GetTraceDB(req.Dbpath)
	if err != nil {
		return nil, err
	}

	return traceDB.GetTracesByGID(req.Gid, int(req.Depth), req.CreateTime)
}

func (a *AnalysisBiz) GetAllGIDs(dbpath string, page int, limit int) ([]dos.GoroutineTrace, error) {
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

func (a *AnalysisBiz) GetParamsByID(dbpath string, id int32) ([]dos.TraceParams, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetParamsByID(id)
}

func (a *AnalysisBiz) GetGidsByFunctionName(dbpath string, functionName string) ([]*dos.TraceData, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	a.log.Infof("get gids by function name: %s from db: %s", functionName, dbpath)
	traces, err := traceDB.GetTracesByFuncName(functionName)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool)
	var result []*dos.TraceData
	for _, trace := range traces {
		if _, ok := set[fmt.Sprintf("%d:%d", trace.GID, trace.ParentId)]; !ok {
			set[fmt.Sprintf("%d:%d", trace.GID, trace.ParentId)] = true
			result = append(result, &dos.TraceData{
				ID:       int64(trace.ID),
				Name:     trace.Name,
				GID:      trace.GID,
				ParentId: trace.ParentId,
			})
		}
	}

	return result, nil
}

// GetTracesByFunctionName 根据函数名获取所有跟踪数据
func (a *AnalysisBiz) GetTracesByFunctionName(dbpath string, functionName string) ([]dos.TraceData, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetTracesByFuncName(functionName)
}

func (a *AnalysisBiz) GetGoroutineByGID(dbpath string, gid uint64) (*dos.GoroutineTrace, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetGoroutineByGID(int64(gid))
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
func (a *AnalysisBiz) GetTracesByParentFunc(dbpath string, parentId int64) ([]dos.TraceData, error) {
	a.log.Infof("get traces by parent id: %d from db: %s", parentId, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetTracesByParentId(parentId)
}

// GetAllParentIds 获取所有的父函数ID
func (a *AnalysisBiz) GetParentFunctions(dbpath string, functionName string) ([]*dos.Function, error) {
	a.log.Infof("get all parent ids from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllParentFunctions(functionName)
}

// GetChildFunctions 获取函数的子函数
func (a *AnalysisBiz) GetChildFunctions(dbpath string, parentId int64) ([]*dos.Function, error) {
	a.log.Infof("get child functions of parent id: %d from db: %s", parentId, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetChildFunctions(parentId)
}

// GetHotFunctions 获取热点函数分析数据
func (a *AnalysisBiz) GetHotFunctions(dbpath string, sortBy string) ([]dos.Function, error) {
	a.log.Infof("get hot functions, sort by: %s from db: %s", sortBy, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetHotFunctions(sortBy)
}

// GetGoroutineStats 获取Goroutine统计信息
func (a *AnalysisBiz) GetGoroutineStats(dbpath string) (*v1.GetGoroutineStatsReply, error) {
	a.log.Infof("get goroutine stats from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}

	// 获取活跃Goroutine数量
	activeCount, err := traceDB.GetActiveGoroutineCount()
	if err != nil {
		return nil, err
	}

	// 获取最大调用深度
	maxDepth, err := traceDB.GetMaxCallDepth()
	if err != nil {
		return nil, err
	}

	// 获取所有时间消耗并计算平均值
	timeCosts, err := traceDB.GetAllTraceTimeCosts()
	if err != nil {
		return nil, err
	}

	// 计算平均执行时间
	var totalTime float64
	var count int
	for _, timeCost := range timeCosts {
		if timeCost != "" {
			timeCostMs := parseTimeCost(timeCost)
			if timeCostMs > 0 {
				totalTime += timeCostMs
				count++
			}
		}
	}

	// 计算平均时间
	var avgTime float64
	if count > 0 {
		avgTime = totalTime / float64(count)
	}

	// 格式化平均时间
	avgTimeStr := formatTime(avgTime)

	return &v1.GetGoroutineStatsReply{
		Active:   int32(activeCount),
		AvgTime:  avgTimeStr,
		MaxDepth: int32(maxDepth),
	}, nil
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
func (a *AnalysisBiz) GetGoroutineExecutionTime(dbpath string, groutine dos.GoroutineTrace) (string, error) {
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

func (a *AnalysisBiz) GetAllGoroutineTrace(dbpath string, includeMetrics bool) ([]dos.GoroutineTrace, error) {
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

// GetFunctionCallStats 获取函数调用统计分析
func (a *AnalysisBiz) GetFunctionCallStats(dbpath string, functionName string) ([]dos.FunctionCallStats, error) {
	a.log.Infof("获取函数调用统计，函数：%s，数据库：%s", functionName, dbpath)

	// 获取追踪数据库
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}

	var stats []dos.FunctionCallStats

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
		stat := dos.FunctionCallStats{
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

// GetFunctionInfoInGoroutine 获取函数在指定Goroutine中的信息
func (a *AnalysisBiz) GetFunctionInfoInGoroutine(dbpath string, gid uint64, targetFunctionId int64) (*dos.FunctionInfo, error) {
	a.log.Infof("get function info in goroutine, gid: %d, functionId: %d from db: %s", gid, targetFunctionId, dbpath)

	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}

	return traceDB.GetFunctionInfoInGoroutine(gid, targetFunctionId)
}

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
func (a *AnalysisBiz) SearchFunctions(ctx context.Context, dbPath string, query string, limit int32) ([]*dos.Function, int32, error) {
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
	// 名字去重
	set := make(map[string]bool)
	var result []*dos.Function
	for _, function := range functions {
		if _, ok := set[function.Name]; !ok {
			set[function.Name] = true
			result = append(result, function)
		}
	}

	return result, total, nil
}

// GetModuleNames 获取数据库中的模块名称列表
func (a *AnalysisBiz) GetModuleNames(dbpath string, maxSamples int32) ([]string, error) {
	a.log.Infof("get module names from db: %s with max samples: %d", dbpath, maxSamples)

	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}

	// 获取采样后的函数名称列表
	functionNames, err := traceDB.GetSampledFunctionNames(maxSamples)
	if err != nil {
		return nil, err
	}

	// 提取模块名称并统计频率
	moduleCount := make(map[string]int)

	for _, functionName := range functionNames {
		// 提取包路径前缀（最后一个.之前的部分）
		lastDotIndex := strings.LastIndex(functionName, ".")
		if lastDotIndex > 0 {
			packagePath := functionName[:lastDotIndex]

			// 按/分割，生成所有层级的前缀
			parts := strings.Split(packagePath, "/")
			for i := 1; i <= len(parts); i++ {
				prefix := strings.Join(parts[:i], "/")
				moduleCount[prefix]++
			}
		} else {
			// 如果没有.，说明是标准库函数，直接使用函数名
			moduleCount[functionName]++
		}
	}

	var stats []dos.ModuleStat
	for name, count := range moduleCount {
		stats = append(stats, dos.ModuleStat{Name: name, Count: count})
	}

	// 按频率降序排序
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	// 取前5个
	result := make([]string, 0, 5)
	for i := 0; i < len(stats) && i < 5; i++ {
		result = append(result, stats[i].Name)
	}

	return result, nil
}
