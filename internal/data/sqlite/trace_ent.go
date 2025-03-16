package sqlite

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/toheart/functrace"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen/goroutinetrace"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen/tracedata"
)

// TraceEntDB 使用 Ent 框架的跟踪数据库
type TraceEntDB struct {
	client *gen.Client
}

// NewTraceEntDB 创建跟踪数据库（使用 Ent 框架）
func NewTraceEntDB(dbPath string) (*TraceEntDB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("trace db file not found: %w", err)
	}

	// 创建 Ent 客户端
	client, err := gen.Open(dialect.SQLite, ParseDBPath(dbPath))
	if err != nil {
		return nil, fmt.Errorf("create ent client failed: %w", err)
	}

	return &TraceEntDB{client: client}, nil
}

// GetTracesByGID 根据 GID 获取跟踪数据
func (d *TraceEntDB) GetTracesByGID(gid uint64) ([]entity.TraceData, error) {
	ctx := context.Background()

	// 查询跟踪数据
	traces, err := d.client.TraceData.
		Query().
		Where(tracedata.Gid(gid)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query trace data failed: %w", err)
	}

	// 转换为业务实体
	var result []entity.TraceData
	for _, trace := range traces {
		createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt failed: %w", err)
		}
		// 创建业务实体
		traceData := entity.TraceData{
			ID:        int64(trace.ID),
			Name:      trace.Name,
			GID:       uint64(trace.Gid),
			Indent:    trace.Indent,
			Params:    trace.Params,
			TimeCost:  trace.TimeCost,
			ParentId:  uint64(trace.ParentId),
			CreatedAt: createdAt.Format(time.RFC3339Nano),
			Seq:       trace.Seq,
		}

		result = append(result, traceData)
	}

	return result, nil
}

// GetTraceByID 根据 ID 获取跟踪数据
func (d *TraceEntDB) GetTraceByID(id int) (*entity.TraceData, error) {
	ctx := context.Background()

	// 查询跟踪数据
	trace, err := d.client.TraceData.Get(ctx, id)
	if err != nil {
		if gen.IsNotFound(err) {
			return nil, nil // 跟踪数据不存在
		}
		return nil, fmt.Errorf("查询跟踪数据失败: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse createdAt failed: %w", err)
	}

	// 创建业务实体
	traceData := &entity.TraceData{
		ID:        int64(trace.ID),
		Name:      trace.Name,
		GID:       trace.Gid,
		Indent:    trace.Indent,
		Params:    trace.Params,
		TimeCost:  trace.TimeCost,
		ParentId:  uint64(trace.ParentId),
		CreatedAt: createdAt.Format(time.RFC3339Nano),
		Seq:       trace.Seq,
	}

	return traceData, nil
}

// GetTraceChildren 获取跟踪数据的子节点
func (d *TraceEntDB) GetTraceChildren(parentID int64) ([]entity.TraceData, error) {
	ctx := context.Background()

	// 查询子节点
	traces, err := d.client.TraceData.
		Query().
		Where(tracedata.ParentId(parentID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询子节点失败: %w", err)
	}

	// 转换为业务实体
	var result []entity.TraceData
	for _, trace := range traces {
		createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt failed: %w", err)
		}
		// 创建业务实体
		traceData := entity.TraceData{
			ID:        int64(trace.ID),
			Name:      trace.Name,
			GID:       trace.Gid,
			Indent:    trace.Indent,
			Params:    trace.Params,
			TimeCost:  trace.TimeCost,
			ParentId:  uint64(trace.ParentId),
			CreatedAt: createdAt.Format(time.RFC3339Nano),
			Seq:       trace.Seq,
		}

		result = append(result, traceData)
	}

	return result, nil
}

// Close 关闭数据库连接
func (d *TraceEntDB) Close() error {
	return d.client.Close()
}

// GetAllGIDs 获取所有的 GID，支持分页
func (d *TraceEntDB) GetAllGIDs(page int, limit int) ([]entity.GoroutineTrace, error) {
	ctx := context.Background()
	offset := (page - 1) * limit // 计算偏移量

	// 查询所有不同的 GID
	var gids []entity.GoroutineTrace
	result, err := d.client.GoroutineTrace.Query().Order(gen.Desc(goroutinetrace.FieldIsFinished)).Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all gids failed: %w", err)
	}

	// 转换为 uint64 类型
	for _, gid := range result {
		gids = append(gids, entity.GoroutineTrace{
			ID:           int64(gid.ID),
			GID:          gid.OriginGid,
			TimeCost:     gid.TimeCost,
			CreateTime:   gid.CreateTime,
			IsFinished:   gid.IsFinished,
			InitFuncName: gid.InitFuncName,
		})
	}

	return gids, nil
}

// GetAllFunctionName 获取所有函数名
func (d *TraceEntDB) GetAllFunctionName() ([]string, error) {
	ctx := context.Background()

	// 查询所有不同的函数名
	functionNames, err := d.client.TraceData.
		Query().
		GroupBy(tracedata.FieldName).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询所有函数名失败: %w", err)
	}

	return functionNames, nil
}

// GetParamsByID 根据 ID 获取参数
func (d *TraceEntDB) GetParamsByID(id int32) ([]functrace.TraceParams, error) {
	ctx := context.Background()

	// 查询跟踪数据
	trace, err := d.client.TraceData.Get(ctx, int(id))
	if err != nil {
		if gen.IsNotFound(err) {
			return nil, nil // 跟踪数据不存在
		}
		return nil, fmt.Errorf("查询跟踪数据失败: %w", err)
	}

	return trace.Params, nil
}

// GetGidsByFunctionName 根据函数名获取 GID 列表
func (d *TraceEntDB) GetGidsByFunctionName(functionName string) ([]string, error) {
	ctx := context.Background()

	// 查询具有指定函数名的所有 GID
	gids, err := d.client.TraceData.
		Query().
		Where(tracedata.Name(functionName)).
		GroupBy(tracedata.FieldGid).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询函数名对应的 GID 失败: %w", err)
	}

	return gids, nil
}

// GetTotalGIDs 获取 GID 总数
func (d *TraceEntDB) GetTotalGIDs() (int, error) {
	ctx := context.Background()

	// 查询不同 GID 的数量
	result, err := d.client.TraceData.
		Query().
		GroupBy(tracedata.FieldGid).
		Strings(ctx)
	if err != nil {
		return 0, fmt.Errorf("查询 GID 总数失败: %w", err)
	}

	return len(result), nil
}

// GetInitialFunc 获取初始函数
func (d *TraceEntDB) GetInitialFunc(gid uint64) (string, error) {
	ctx := context.Background()

	// 查询指定 GID 的第一条记录
	trace, err := d.client.TraceData.
		Query().
		Where(tracedata.Gid(gid)).
		Order(gen.Asc(tracedata.FieldID)).
		First(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return "", fmt.Errorf("未找到 GID 为 %d 的记录", gid)
		}
		return "", fmt.Errorf("查询初始函数失败: %w", err)
	}

	return trace.Name, nil
}

// GetTracesByParentId 根据父函数 ID 查询函数调用
func (d *TraceEntDB) GetTracesByParentId(parentId int64) ([]entity.TraceData, error) {
	ctx := context.Background()

	// 查询具有指定父 ID 的所有跟踪数据
	traces, err := d.client.TraceData.
		Query().
		Where(
			func(s *sql.Selector) {
				// 使用原生SQL查询，避免类型转换问题
				s.Where(sql.EQ(tracedata.FieldParentId, parentId))
			},
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询子函数调用失败: %w", err)
	}

	// 转换为业务实体
	var result []entity.TraceData
	for _, trace := range traces {
		createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt failed: %w", err)
		}
		// 创建业务实体
		traceData := entity.TraceData{
			ID:        int64(trace.ID),
			Name:      trace.Name,
			GID:       uint64(trace.Gid),
			Indent:    trace.Indent,
			Params:    trace.Params,
			TimeCost:  trace.TimeCost,
			ParentId:  uint64(trace.ParentId),
			CreatedAt: createdAt.Format(time.RFC3339Nano),
			Seq:       trace.Seq,
		}

		result = append(result, traceData)
	}

	return result, nil
}

// GetAllParentIds 获取所有的父函数 ID
func (d *TraceEntDB) GetAllParentIds() ([]int64, error) {
	ctx := context.Background()

	// 查询所有不为空的父 ID
	parentIds, err := d.client.TraceData.
		Query().
		Where(
			tracedata.ParentIdNotNil(),
			tracedata.ParentIdNEQ(0),
		).
		GroupBy(tracedata.FieldParentId).
		Ints(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询所有父函数 ID 失败: %w", err)
	}

	// 转换为 int64 类型
	var result []int64
	for _, id := range parentIds {
		result = append(result, int64(id))
	}

	return result, nil
}

// GetChildFunctions 获取函数的子函数
func (d *TraceEntDB) GetChildFunctions(parentId int64) ([]string, error) {
	ctx := context.Background()

	// 查询具有指定父 ID 的所有不同函数名
	childFunctions, err := d.client.TraceData.
		Query().
		Where(
			func(s *sql.Selector) {
				// 使用原生SQL查询，避免类型转换问题
				s.Where(sql.EQ(tracedata.FieldParentId, parentId))
			},
		).
		GroupBy(tracedata.FieldName).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询子函数失败: %w", err)
	}

	return childFunctions, nil
}

// GetHotFunctions 获取热点函数分析数据
func (d *TraceEntDB) GetHotFunctions(sortBy string) ([]entity.HotFunction, error) {
	ctx := context.Background()

	// 查询所有跟踪数据
	traces, err := d.client.TraceData.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询跟踪数据失败: %w", err)
	}

	// 按函数名分组统计
	funcStats := make(map[string]struct {
		CallCount int
		TotalTime float64
	})

	for _, trace := range traces {
		// 解析时间消耗
		var timeCost float64
		if trace.TimeCost != "" {
			// 处理时间格式，将 ms 和 s 转换为毫秒值
			if strings.HasSuffix(trace.TimeCost, "ms") {
				trace.TimeCost = strings.TrimSuffix(trace.TimeCost, "ms")
				timeCost, _ = strconv.ParseFloat(trace.TimeCost, 64)
			} else if strings.HasSuffix(trace.TimeCost, "s") {
				trace.TimeCost = strings.TrimSuffix(trace.TimeCost, "s")
				timeCostVal, _ := strconv.ParseFloat(trace.TimeCost, 64)
				timeCost = timeCostVal * 1000 // 转换为毫秒
			}
		}

		// 更新函数统计
		stats := funcStats[trace.Name]
		stats.CallCount++
		stats.TotalTime += timeCost
		funcStats[trace.Name] = stats
	}

	// 转换为热点函数列表
	var hotFunctions []entity.HotFunction
	for name, stats := range funcStats {
		// 提取包名
		parts := strings.Split(name, ".")
		pkg := "main"
		if len(parts) > 1 {
			pkg = strings.Join(parts[:len(parts)-1], ".")
		}

		// 格式化总时间
		totalTimeStr := ""
		if stats.TotalTime > 1000 {
			totalTimeStr = fmt.Sprintf("%.2fs", stats.TotalTime/1000)
		} else {
			totalTimeStr = fmt.Sprintf("%.2fms", stats.TotalTime)
		}

		// 计算平均时间
		avgTime := 0.0
		if stats.CallCount > 0 {
			avgTime = stats.TotalTime / float64(stats.CallCount)
		}

		// 格式化平均时间
		avgTimeStr := ""
		if avgTime > 1000 {
			avgTimeStr = fmt.Sprintf("%.2fs", avgTime/1000)
		} else {
			avgTimeStr = fmt.Sprintf("%.2fms", avgTime)
		}

		hotFunctions = append(hotFunctions, entity.HotFunction{
			Name:      name,
			Package:   pkg,
			CallCount: stats.CallCount,
			TotalTime: totalTimeStr,
			AvgTime:   avgTimeStr,
		})
	}

	// 排序
	if sortBy == "time" {
		sort.Slice(hotFunctions, func(i, j int) bool {
			// 解析时间字符串进行比较
			timeI := parseTimeString(hotFunctions[i].TotalTime)
			timeJ := parseTimeString(hotFunctions[j].TotalTime)
			return timeI > timeJ
		})
	} else {
		// 默认按调用次数排序
		sort.Slice(hotFunctions, func(i, j int) bool {
			return hotFunctions[i].CallCount > hotFunctions[j].CallCount
		})
	}

	// 限制返回数量
	if len(hotFunctions) > 50 {
		hotFunctions = hotFunctions[:50]
	}

	return hotFunctions, nil
}

// 解析时间字符串为毫秒值
func parseTimeString(timeStr string) float64 {
	var value float64
	if strings.HasSuffix(timeStr, "ms") {
		value, _ = strconv.ParseFloat(strings.TrimSuffix(timeStr, "ms"), 64)
	} else if strings.HasSuffix(timeStr, "s") {
		seconds, _ := strconv.ParseFloat(strings.TrimSuffix(timeStr, "s"), 64)
		value = seconds * 1000
	}
	return value
}

// GetGoroutineStats 获取 Goroutine 统计信息
func (d *TraceEntDB) GetGoroutineStats() (*entity.GoroutineStats, error) {
	ctx := context.Background()
	stats := &entity.GoroutineStats{}

	// 使用GoroutineTrace表获取活跃Goroutine数量
	gids, err := d.client.GoroutineTrace.
		Query().
		Where(goroutinetrace.IsFinished(0)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("found active goroutines failed: %w", err)
	}
	stats.Active = len(gids)

	// 通过TraceData表获取最大调用深度
	maxDepth, err := d.client.TraceData.
		Query().
		Aggregate(gen.Max(tracedata.FieldIndent)).
		Int(ctx)
	if err != nil {
		return nil, fmt.Errorf("found max call depth failed: %w", err)
	}
	stats.MaxDepth = maxDepth + 1 // 调用深度为最大缩进级别 + 1

	// 计算平均执行时间
	traces, err := d.client.TraceData.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("found traces failed: %w", err)
	}

	var totalTime float64
	var count int
	for _, trace := range traces {
		if trace.TimeCost != "" {
			timeCost := parseTimeString(trace.TimeCost)
			if timeCost > 0 {
				totalTime += timeCost
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
	if avgTime > 1000 {
		stats.AvgTime = fmt.Sprintf("%.2fs", avgTime/1000)
	} else {
		stats.AvgTime = fmt.Sprintf("%.2fms", avgTime)
	}

	return stats, nil
}

// GetFunctionAnalysis 获取函数调用关系分析
func (d *TraceEntDB) GetFunctionAnalysis(functionName string, queryType string) ([]entity.FunctionNode, error) {
	// 不需要ctx变量
	var result []entity.FunctionNode

	// 生成唯一ID
	nextID := 1
	generateID := func() string {
		id := fmt.Sprintf("node_%d", nextID)
		nextID++
		return id
	}

	// 创建根节点
	rootNode := entity.FunctionNode{
		ID:        generateID(),
		Name:      functionName,
		CallCount: 0,
		AvgTime:   "0ms",
	}

	// 提取包名
	parts := strings.Split(functionName, ".")
	if len(parts) > 1 {
		rootNode.Package = strings.Join(parts[:len(parts)-1], ".")
	} else {
		rootNode.Package = "main"
	}

	// 获取根节点的调用次数和平均时间
	stats, err := d.getFunctionStats(functionName)
	if err == nil && stats != nil {
		rootNode.CallCount = stats.CallCount
		rootNode.AvgTime = stats.AvgTime
	}

	// 根据查询类型选择不同的查询方式
	if queryType == "caller" {
		// 查询调用指定函数的函数
		callers, err := d.GetCallerFunctions(functionName)
		if err != nil {
			return nil, fmt.Errorf("获取调用方函数失败: %w", err)
		}

		// 添加调用者作为子节点
		for _, caller := range callers {
			callerNode := entity.FunctionNode{
				ID:   generateID(),
				Name: caller,
			}

			// 提取包名
			parts := strings.Split(caller, ".")
			if len(parts) > 1 {
				callerNode.Package = strings.Join(parts[:len(parts)-1], ".")
			} else {
				callerNode.Package = "main"
			}

			// 获取调用者的调用次数和平均时间
			stats, err := d.getFunctionStats(caller)
			if err == nil && stats != nil {
				callerNode.CallCount = stats.CallCount
				callerNode.AvgTime = stats.AvgTime
			}

			rootNode.Children = append(rootNode.Children, callerNode)
		}
	} else {
		// 查询被指定函数调用的函数
		callees, err := d.GetCalleeFunctions(functionName)
		if err != nil {
			return nil, fmt.Errorf("获取被调用方函数失败: %w", err)
		}

		// 添加被调用者作为子节点
		for _, callee := range callees {
			calleeNode := entity.FunctionNode{
				ID:   generateID(),
				Name: callee,
			}

			// 提取包名
			parts := strings.Split(callee, ".")
			if len(parts) > 1 {
				calleeNode.Package = strings.Join(parts[:len(parts)-1], ".")
			} else {
				calleeNode.Package = "main"
			}

			// 获取被调用者的调用次数和平均时间
			stats, err := d.getFunctionStats(callee)
			if err == nil && stats != nil {
				calleeNode.CallCount = stats.CallCount
				calleeNode.AvgTime = stats.AvgTime
			}

			rootNode.Children = append(rootNode.Children, calleeNode)
		}
	}

	result = append(result, rootNode)
	return result, nil
}

// GetCallerFunctions 获取调用指定函数的函数列表
func (d *TraceEntDB) GetCallerFunctions(functionName string) ([]string, error) {
	ctx := context.Background()

	// 首先获取具有指定函数名的跟踪数据的 ID
	funcTraces, err := d.client.TraceData.
		Query().
		Where(tracedata.Name(functionName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get caller functions failed: %w", err)
	}

	// 收集所有函数 ID
	var funcIDs []int
	for _, trace := range funcTraces {
		funcIDs = append(funcIDs, trace.ID)
	}

	if len(funcIDs) == 0 {
		return []string{}, nil
	}

	// 查询调用了这些函数的父函数
	callers, err := d.client.TraceData.
		Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.InInts(tracedata.FieldParentId, funcIDs...))
		}).
		GroupBy(tracedata.FieldName).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get caller functions failed: %w", err)
	}

	return callers, nil
}

// GetCalleeFunctions 获取被指定函数调用的函数列表
func (d *TraceEntDB) GetCalleeFunctions(functionName string) ([]string, error) {
	ctx := context.Background()

	// 首先获取具有指定函数名的第一条跟踪数据
	funcTrace, err := d.client.TraceData.
		Query().
		Where(tracedata.Name(functionName)).
		First(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("get callee functions failed: %w", err)
	}

	// 查询以该函数为父函数的所有函数
	callees, err := d.client.TraceData.
		Query().
		Where(
			func(s *sql.Selector) {
				// 使用原生SQL查询，避免类型转换问题
				s.Where(sql.EQ(tracedata.FieldParentId, funcTrace.ID))
			},
		).
		GroupBy(tracedata.FieldName).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get callee functions failed: %w", err)
	}

	return callees, nil
}

// GetTracesByFuncName 获取指定函数名的所有跟踪数据
func (d *TraceEntDB) GetTracesByFuncName(functionName string) ([]entity.TraceData, error) {
	ctx := context.Background()

	// 查询跟踪数据
	traces, err := d.client.TraceData.
		Query().
		Where(tracedata.Name(functionName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query traces by function name failed: %w", err)
	}

	// 转换为业务实体
	var result []entity.TraceData
	for _, trace := range traces {
		createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt failed: %w", err)
		}
		// 创建业务实体
		traceData := entity.TraceData{
			ID:        int64(trace.ID),
			Name:      trace.Name,
			GID:       trace.Gid,
			Indent:    trace.Indent,
			Params:    trace.Params,
			TimeCost:  trace.TimeCost,
			ParentId:  uint64(trace.ParentId),
			CreatedAt: createdAt.Format(time.RFC3339Nano),
			Seq:       trace.Seq,
		}

		result = append(result, traceData)
	}

	return result, nil
}

// GetFunctionCallDepths 获取指定函数在各种调用中的调用深度
func (d *TraceEntDB) GetFunctionCallDepths(functionName string) ([]int, error) {
	ctx := context.Background()

	// 查询指定函数的所有调用
	traces, err := d.client.TraceData.
		Query().
		Where(tracedata.Name(functionName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query traces by function name failed: %w", err)
	}

	// 收集所有调用深度
	depths := make([]int, 0, len(traces))
	for _, trace := range traces {
		depths = append(depths, trace.Indent)
	}

	return depths, nil
}

// 函数统计信息
type functionStats struct {
	CallCount int
	AvgTime   string
}

// 获取函数的调用次数和平均时间
func (d *TraceEntDB) getFunctionStats(functionName string) (*functionStats, error) {
	ctx := context.Background()

	// 查询具有指定函数名的所有跟踪数据
	traces, err := d.client.TraceData.
		Query().
		Where(tracedata.Name(functionName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get function stats failed: %w", err)
	}

	// 计算调用次数和总时间
	callCount := len(traces)
	var totalTime float64

	for _, trace := range traces {
		if trace.TimeCost != "" {
			timeCost := parseTimeString(trace.TimeCost)
			totalTime += timeCost
		}
	}

	// 计算平均时间
	var avgTime float64
	if callCount > 0 {
		avgTime = totalTime / float64(callCount)
	}

	// 格式化平均时间
	avgTimeStr := ""
	if avgTime > 1000 {
		avgTimeStr = fmt.Sprintf("%.2fs", avgTime/1000)
	} else {
		avgTimeStr = fmt.Sprintf("%.2fms", avgTime)
	}

	return &functionStats{
		CallCount: callCount,
		AvgTime:   avgTimeStr,
	}, nil
}

// GetFunctionCallGraph 获取函数调用关系图
func (d *TraceEntDB) GetFunctionCallGraph(functionName string, depth int, direction string) (*entity.FunctionCallGraph, error) {
	ctx := context.Background()

	// 初始化结果
	graph := &entity.FunctionCallGraph{
		Nodes: []entity.FunctionGraphNode{},
		Edges: []entity.FunctionGraphEdge{},
	}

	// 生成唯一ID
	nextID := 1
	generateID := func() string {
		id := fmt.Sprintf("node_%d", nextID)
		nextID++
		return id
	}

	// 节点ID映射，避免重复添加节点
	nodeMap := make(map[string]string) // 函数名 -> 节点ID

	// 添加根节点
	rootID := generateID()
	rootNode := entity.FunctionGraphNode{
		ID:       rootID,
		Name:     functionName,
		NodeType: "root",
	}

	// 提取包名
	parts := strings.Split(functionName, ".")
	if len(parts) > 1 {
		rootNode.Package = strings.Join(parts[:len(parts)-1], ".")
	} else {
		rootNode.Package = "main"
	}

	// 获取根节点的调用次数和平均时间
	stats, err := d.getFunctionStats(functionName)
	if err == nil && stats != nil {
		rootNode.CallCount = stats.CallCount
		rootNode.AvgTime = stats.AvgTime
	}

	// 添加根节点到图中
	graph.Nodes = append(graph.Nodes, rootNode)
	nodeMap[functionName] = rootID

	// 处理调用者（向上查询）
	if direction == "caller" || direction == "both" {
		err := d.addCallerNodes(ctx, graph, nodeMap, functionName, rootID, depth, generateID)
		if err != nil {
			return nil, fmt.Errorf("add caller nodes failed: %w", err)
		}
	}

	// 处理被调用者（向下查询）
	if direction == "callee" || direction == "both" {
		err := d.addCalleeNodes(ctx, graph, nodeMap, functionName, rootID, depth, generateID)
		if err != nil {
			return nil, fmt.Errorf("add callee nodes failed: %w", err)
		}
	}

	return graph, nil
}

// 添加调用者节点（向上查询）
func (d *TraceEntDB) addCallerNodes(ctx context.Context, graph *entity.FunctionCallGraph, nodeMap map[string]string, funcName string, parentID string, depth int, generateID func() string) error {
	if depth <= 0 {
		return nil
	}

	// 获取调用者列表
	callers, err := d.GetCallerFunctions(funcName)
	if err != nil {
		return err
	}

	for _, caller := range callers {
		var nodeID string

		// 检查节点是否已存在
		if id, exists := nodeMap[caller]; exists {
			nodeID = id
		} else {
			// 创建新节点
			nodeID = generateID()
			callerNode := entity.FunctionGraphNode{
				ID:       nodeID,
				Name:     caller,
				NodeType: "caller",
			}

			// 提取包名
			parts := strings.Split(caller, ".")
			if len(parts) > 1 {
				callerNode.Package = strings.Join(parts[:len(parts)-1], ".")
			} else {
				callerNode.Package = "main"
			}

			// 获取调用者的调用次数和平均时间
			stats, err := d.getFunctionStats(caller)
			if err == nil && stats != nil {
				callerNode.CallCount = stats.CallCount
				callerNode.AvgTime = stats.AvgTime
			}

			// 添加节点到图中
			graph.Nodes = append(graph.Nodes, callerNode)
			nodeMap[caller] = nodeID
		}

		// 添加边
		edge := entity.FunctionGraphEdge{
			Source:   nodeID,
			Target:   parentID,
			Label:    "调用",
			EdgeType: "caller_to_root",
		}
		graph.Edges = append(graph.Edges, edge)

		// 递归处理上一级调用者
		if depth > 1 {
			err := d.addCallerNodes(ctx, graph, nodeMap, caller, nodeID, depth-1, generateID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 添加被调用者节点（向下查询）
func (d *TraceEntDB) addCalleeNodes(ctx context.Context, graph *entity.FunctionCallGraph, nodeMap map[string]string, funcName string, parentID string, depth int, generateID func() string) error {
	if depth <= 0 {
		return nil
	}

	// 获取被调用者列表
	callees, err := d.GetCalleeFunctions(funcName)
	if err != nil {
		return err
	}

	for _, callee := range callees {
		var nodeID string

		// 检查节点是否已存在
		if id, exists := nodeMap[callee]; exists {
			nodeID = id
		} else {
			// 创建新节点
			nodeID = generateID()
			calleeNode := entity.FunctionGraphNode{
				ID:       nodeID,
				Name:     callee,
				NodeType: "callee",
			}

			// 提取包名
			parts := strings.Split(callee, ".")
			if len(parts) > 1 {
				calleeNode.Package = strings.Join(parts[:len(parts)-1], ".")
			} else {
				calleeNode.Package = "main"
			}

			// 获取被调用者的调用次数和平均时间
			stats, err := d.getFunctionStats(callee)
			if err == nil && stats != nil {
				calleeNode.CallCount = stats.CallCount
				calleeNode.AvgTime = stats.AvgTime
			}

			// 添加节点到图中
			graph.Nodes = append(graph.Nodes, calleeNode)
			nodeMap[callee] = nodeID
		}

		// 添加边
		edge := entity.FunctionGraphEdge{
			Source:   parentID,
			Target:   nodeID,
			Label:    "调用",
			EdgeType: "root_to_callee",
		}
		graph.Edges = append(graph.Edges, edge)

		// 递归处理下一级被调用者
		if depth > 1 {
			err := d.addCalleeNodes(ctx, graph, nodeMap, callee, nodeID, depth-1, generateID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetGoroutineCallDepth 获取指定 Goroutine 的最大调用深度
func (d *TraceEntDB) GetGoroutineCallDepth(gid uint64) (int, error) {
	ctx := context.Background()

	// 查询指定 GID 的最大缩进级别
	maxIndent, err := d.client.TraceData.
		Query().
		Where(tracedata.Gid(gid)).
		Aggregate(gen.Max(tracedata.FieldIndent)).
		Int(ctx)
	if err != nil {
		return 0, fmt.Errorf("get max call depth failed: %w", err)
	}
	return maxIndent + 1, nil
}

func (d *TraceEntDB) GetLastFunction() (*entity.TraceData, error) {
	ctx := context.Background()
	trace, err := d.client.TraceData.
		Query().
		Order(gen.Desc(tracedata.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("get last function failed: %w", err)
	}

	return &entity.TraceData{
		ID:        int64(trace.ID),
		Name:      trace.Name,
		Params:    trace.Params,
		TimeCost:  trace.TimeCost,
		ParentId:  uint64(trace.ParentId),
		CreatedAt: trace.CreatedAt,
		Seq:       trace.Seq,
	}, nil
}

// GetGoroutineExecutionTime 获取指定 Goroutine 的总执行时间
func (d *TraceEntDB) GetGoroutineExecutionTime(gid uint64) (string, error) {
	ctx := context.Background()

	// 从GoroutineTrace表查询指定GID的执行时间
	goroutine, err := d.client.GoroutineTrace.
		Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ("gid", gid))
		}).
		First(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return "0ms", nil
		}
		return "", fmt.Errorf("get total execution time failed: %w", err)
	}

	if goroutine.TimeCost == "" {
		return "0ms", nil
	}

	return goroutine.TimeCost, nil
}

// IsGoroutineFinished 检查指定的goroutine是否已完成
func (d *TraceEntDB) IsGoroutineFinished(gid uint64) (bool, error) {
	ctx := context.Background()

	// 从GoroutineTrace表查询指定GID的完成状态
	goroutine, err := d.client.GoroutineTrace.
		Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ("gid", gid))
		}).
		First(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("check goroutine finished status failed: %w", err)
	}

	// 检查is_finished字段
	return goroutine.IsFinished == 1, nil
}

// GetAllUnfinishedFunctions 获取所有未完成的函数
func (d *TraceEntDB) GetAllUnfinishedFunctions(threshold int64) ([]entity.AllUnfinishedFunction, error) {
	ctx := context.Background()

	// 获取最后一条trace
	trace, err := d.GetLastFunction()
	if err != nil {
		return nil, fmt.Errorf("get last function failed: %w", err)
	}
	lastTraceTime, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse last trace time failed: %w", err)
	}

	// 查询未完成的跟踪数据
	var unfinishedFunctions []entity.AllUnfinishedFunction
	traces, err := d.client.TraceData.Query().
		Where(tracedata.TimeCostIsNil()).
		Order(gen.Desc(tracedata.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find unfinished functions failed: %w", err)
	}

	// 过滤超过阈值的函数
	for _, trace := range traces {
		// 解析创建时间
		createTime, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			continue
		}

		// 计算已经过去的毫秒数
		elapsedMS := lastTraceTime.Sub(createTime).Milliseconds()

		// 只保留超过阈值的函数
		if elapsedMS > threshold {
			unfinishedFunction := entity.AllUnfinishedFunction{
				Name:        trace.Name,
				GID:         trace.Gid,
				RunningTime: lastTraceTime.Sub(createTime).String(),
				IsBlocking:  true,
				FunctionID:  int64(trace.ID),
			}
			unfinishedFunctions = append(unfinishedFunctions, unfinishedFunction)
		}
	}

	return unfinishedFunctions, nil
}

func formatTime(totalTime int64) string {
	return fmt.Sprintf("%d ms", totalTime) // 返回总时间的字符串格式
}
