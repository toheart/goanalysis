package sqlite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/klauspost/compress/zstd"
	"github.com/toheart/goanalysis/internal/biz/analysis/dos"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen/goroutinetrace"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen/paramstoredata"
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
func (d *TraceEntDB) GetTracesByGID(gid uint64, depth int, createTime string) ([]dos.TraceData, error) {
	ctx := context.Background()

	// 查询跟踪数据
	traces, err := d.client.TraceData.
		Query().
		Where(tracedata.Gid(gid),
			tracedata.IndentLT(depth),
			tracedata.CreatedAtGTE(createTime)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query trace data failed: %w", err)
	}

	// 转换为业务实体
	var result []dos.TraceData
	for _, trace := range traces {
		createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt failed: %w", err)
		}

		// 创建业务实体
		traceData := dos.TraceData{
			ID:         int64(trace.ID),
			Name:       trace.Name,
			GID:        uint64(trace.Gid),
			Indent:     trace.Indent,
			ParamCount: trace.ParamsCount,
			TimeCost:   trace.TimeCost,
			ParentId:   uint64(trace.ParentId),
			CreatedAt:  createdAt.Format(time.RFC3339Nano),
			Seq:        trace.Seq,
		}

		result = append(result, traceData)
	}

	return result, nil
}

// GetTraceByID 根据 ID 获取跟踪数据
func (d *TraceEntDB) GetTraceByID(id int) (*dos.TraceData, error) {
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
	traceData := &dos.TraceData{
		ID:         int64(trace.ID),
		Name:       trace.Name,
		GID:        trace.Gid,
		Indent:     trace.Indent,
		ParamCount: trace.ParamsCount,
		TimeCost:   trace.TimeCost,
		ParentId:   uint64(trace.ParentId),
		CreatedAt:  createdAt.Format(time.RFC3339Nano),
		Seq:        trace.Seq,
	}

	return traceData, nil
}

// GetTraceChildren 获取跟踪数据的子节点
func (d *TraceEntDB) GetTraceChildren(parentID int64) ([]dos.TraceData, error) {
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
	var result []dos.TraceData
	for _, trace := range traces {
		createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt failed: %w", err)
		}
		// 创建业务实体
		traceData := dos.TraceData{
			ID:         int64(trace.ID),
			Name:       trace.Name,
			GID:        trace.Gid,
			Indent:     trace.Indent,
			ParamCount: trace.ParamsCount,
			TimeCost:   trace.TimeCost,
			ParentId:   uint64(trace.ParentId),
			CreatedAt:  createdAt.Format(time.RFC3339Nano),
			Seq:        trace.Seq,
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
func (d *TraceEntDB) GetAllGIDs(page int, limit int) ([]dos.GoroutineTrace, error) {
	ctx := context.Background()
	offset := (page - 1) * limit // 计算偏移量

	// 查询所有不同的 GID
	var gids []dos.GoroutineTrace
	result, err := d.client.GoroutineTrace.Query().Order(gen.Asc(goroutinetrace.FieldIsFinished)).Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all gids failed: %w", err)
	}

	// 转换为 uint64 类型
	for _, gid := range result {
		gids = append(gids, dos.GoroutineTrace{
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
func (d *TraceEntDB) GetParamsByID(id int32) ([]dos.TraceParams, error) {
	ctx := context.Background()

	// 查询跟踪数据
	params, err := d.client.ParamStoreData.Query().Where(paramstoredata.TraceId(int64(id))).All(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return nil, nil // 跟踪数据不存在
		}
		return nil, fmt.Errorf("found params failed: %w", err)
	}
	var result []dos.TraceParams
	// 获取receiver参数
	for _, param := range params {
		item := dos.TraceParams{
			ID:         param.ID,
			TraceID:    param.TraceId,
			Position:   param.Position,
			IsReceiver: param.IsReceiver,
			BaseID:     *param.BaseId,
		}
		data := decompress(param.Data)
		if param.IsReceiver && param.BaseId != nil && *param.BaseId != 0 {
			// 通过BaseId 获取数据
			parentParam, err := d.client.ParamStoreData.Query().Where(paramstoredata.ID(int64(*param.BaseId))).First(ctx)
			if err != nil {
				return nil, fmt.Errorf("found parent param failed: %w", err)
			}
			parentParamData := decompress(parentParam.Data)
			// 使用jsonPath 恢复数据
			deData, err := jsonpatch.MergePatch([]byte(parentParamData), []byte(data))
			if err != nil {
				return nil, fmt.Errorf("create merge patch failed: %w", err)
			}
			data = string(deData)
		}
		item.Data = data
		result = append(result, item)
	}

	return result, nil
}

// GetGidsByFunctionName 根据函数名获取 GID 列表
func (d *TraceEntDB) GetGidsByFunctionName(functionName string) ([]string, error) {
	ctx := context.Background()

	// 查询具有指定函数名的所有跟踪数据，不按GID分组
	traces, err := d.client.TraceData.
		Query().
		Where(tracedata.Name(functionName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询函数名对应的跟踪数据失败: %w", err)
	}

	// 用于去重的map，key为"gid:parentId"
	processedGIDParentID := make(map[string]bool)
	var result []string

	for _, trace := range traces {
		// 创建唯一键：gid:parentId
		uniqueKey := fmt.Sprintf("%d:%d", trace.Gid, trace.ParentId)

		// 如果已经处理过相同的gid:parentId组合，跳过
		if processedGIDParentID[uniqueKey] {
			continue
		}

		// 标记为已处理
		processedGIDParentID[uniqueKey] = true

		// 将GID添加到结果中
		result = append(result, fmt.Sprintf("%d", trace.Gid))
	}

	return result, nil
}

// GetGoroutineByGID 根据 GID 获取单个 Goroutine 信息
func (d *TraceEntDB) GetGoroutineByGID(gid int64) (*dos.GoroutineTrace, error) {
	ctx := context.Background()

	// 查询指定 GID 的 Goroutine 信息
	goroutine, err := d.client.GoroutineTrace.
		Query().
		Where(goroutinetrace.ID(gid)).
		First(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return nil, fmt.Errorf("Goroutine with GID %d not found", gid)
		}
		return nil, fmt.Errorf("查询 Goroutine 信息失败: %w", err)
	}

	// 转换为业务实体
	result := &dos.GoroutineTrace{
		ID:           int64(goroutine.ID),
		GID:          goroutine.OriginGid,
		TimeCost:     goroutine.TimeCost,
		CreateTime:   goroutine.CreateTime,
		IsFinished:   goroutine.IsFinished,
		InitFuncName: goroutine.InitFuncName,
	}

	return result, nil
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
func (d *TraceEntDB) GetTracesByParentId(parentId int64) ([]dos.TraceData, error) {
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
	var result []dos.TraceData
	for _, trace := range traces {
		createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt failed: %w", err)
		}
		// 创建业务实体
		traceData := dos.TraceData{
			ID:         int64(trace.ID),
			Name:       trace.Name,
			GID:        uint64(trace.Gid),
			Indent:     trace.Indent,
			ParamCount: trace.ParamsCount,
			TimeCost:   trace.TimeCost,
			ParentId:   uint64(trace.ParentId),
			CreatedAt:  createdAt.Format(time.RFC3339Nano),
			Seq:        trace.Seq,
		}

		result = append(result, traceData)
	}

	return result, nil
}

// GetAllParentIds 获取所有的父函数 ID
func (d *TraceEntDB) GetAllParentFunctions(functionName string) ([]*dos.Function, error) {
	ctx := context.Background()

	// 查询所有不为空的父 ID
	parentIds, err := d.client.TraceData.
		Query().
		Where(
			tracedata.ParentIdNEQ(0),
			tracedata.Name(functionName),
		).
		GroupBy(tracedata.FieldParentId).
		Ints(ctx)
	if err != nil {
		return nil, fmt.Errorf("find parent function id failed: %w", err)
	}
	parentIdInt64 := make([]int64, len(parentIds))
	for i, id := range parentIds {
		parentIdInt64[i] = int64(id)
	}

	// 查询所有父函数
	parentFunctions, err := d.client.TraceData.
		Query().
		Where(
			tracedata.ParentIdIn(parentIdInt64...),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find parent function failed: %w", err)
	}

	// 转换为 int64 类型
	var result []*dos.Function
	for _, item := range parentFunctions {
		f := dos.NewFunction(int64(item.ID), item.Name, 0, "0ms", "0ms")
		f.SetPackage()
		result = append(result, f)
	}

	return result, nil
}

// GetChildFunctions 获取函数的子函数
func (d *TraceEntDB) GetChildFunctions(parentId int64) ([]*dos.Function, error) {
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
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询子函数失败: %w", err)
	}
	result := make([]*dos.Function, 0)
	for _, item := range childFunctions {
		f := dos.NewFunction(int64(item.ID), item.Name, 0, item.TimeCost, "0ms")
		f.ParamCount = item.ParamsCount
		f.Depth = item.Indent + 1
		f.Seq = item.Seq // 添加seq字段
		result = append(result, f)
	}

	return result, nil
}

// GetHotFunctions 获取热点函数分析数据
func (d *TraceEntDB) GetHotFunctions(sortBy string) ([]dos.Function, error) {
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
			timeCost = parseTimeString(trace.TimeCost)
		}

		// 更新函数统计
		stats := funcStats[trace.Name]
		stats.CallCount++
		stats.TotalTime += timeCost
		funcStats[trace.Name] = stats
	}

	// 转换为热点函数列表
	var hotFunctions []dos.Function
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

		hotFunctions = append(hotFunctions, dos.Function{
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

// GetActiveGoroutineCount 获取活跃Goroutine数量
func (d *TraceEntDB) GetActiveGoroutineCount() (int, error) {
	ctx := context.Background()

	// 使用GoroutineTrace表获取活跃Goroutine数量
	gids, err := d.client.GoroutineTrace.
		Query().
		Where(goroutinetrace.IsFinished(0)).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("found active goroutines failed: %w", err)
	}

	return len(gids), nil
}

// GetMaxCallDepth 获取最大调用深度
func (d *TraceEntDB) GetMaxCallDepth() (int, error) {
	ctx := context.Background()

	// 通过TraceData表获取最大调用深度
	maxDepth, err := d.client.TraceData.
		Query().
		Aggregate(gen.Max(tracedata.FieldIndent)).
		Int(ctx)
	if err != nil {
		return 0, fmt.Errorf("found max call depth failed: %w", err)
	}

	return maxDepth + 1, nil // 调用深度为最大缩进级别 + 1
}

// GetAllTraceTimeCosts 获取所有跟踪数据的时间消耗
func (d *TraceEntDB) GetAllTraceTimeCosts() ([]string, error) {
	ctx := context.Background()

	// 查询所有跟踪数据的时间消耗
	traces, err := d.client.TraceData.
		Query().
		Select(tracedata.FieldTimeCost).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("found traces failed: %w", err)
	}

	// 提取时间消耗字符串
	var timeCosts []string
	for _, trace := range traces {
		if trace.TimeCost != "" {
			timeCosts = append(timeCosts, trace.TimeCost)
		}
	}

	return timeCosts, nil
}

// GetFunctionAnalysis 获取函数调用关系分析
func (d *TraceEntDB) GetFunctionAnalysis(functionName string, queryType string) ([]dos.FunctionNode, error) {
	// 不需要ctx变量
	var result []dos.FunctionNode

	// 生成唯一ID
	nextID := 1
	generateID := func() string {
		id := fmt.Sprintf("node_%d", nextID)
		nextID++
		return id
	}

	// 创建根节点
	rootNode := dos.FunctionNode{
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
			callerNode := dos.FunctionNode{
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
			calleeNode := dos.FunctionNode{
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
func (d *TraceEntDB) GetTracesByFuncName(functionName string) ([]dos.TraceData, error) {
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
	var result []dos.TraceData
	for _, trace := range traces {
		createdAt, err := time.Parse(time.RFC3339Nano, trace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt failed: %w", err)
		}
		// 创建业务实体
		traceData := dos.TraceData{
			ID:         int64(trace.ID),
			Name:       trace.Name,
			GID:        trace.Gid,
			Indent:     trace.Indent,
			ParamCount: trace.ParamsCount,
			TimeCost:   trace.TimeCost,
			ParentId:   uint64(trace.ParentId),
			CreatedAt:  createdAt.Format(time.RFC3339Nano),
			Seq:        trace.Seq,
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

func (d *TraceEntDB) GetLastFunction() (*dos.TraceData, error) {
	ctx := context.Background()
	trace, err := d.client.TraceData.
		Query().
		Order(gen.Desc(tracedata.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("get last function failed: %w", err)
	}

	return &dos.TraceData{
		ID:         int64(trace.ID),
		Name:       trace.Name,
		ParamCount: trace.ParamsCount,
		TimeCost:   trace.TimeCost,
		ParentId:   uint64(trace.ParentId),
		CreatedAt:  trace.CreatedAt,
		Seq:        trace.Seq,
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

func (d *TraceEntDB) SearchFunctions(ctx context.Context, dbPath string, query string, limit int32) ([]*dos.Function, int32, error) {
	traces, err := d.client.TraceData.Query().
		Where(tracedata.NameContains(query)).
		Limit(int(limit)).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("search functions failed: %w", err)
	}
	var functions []*dos.Function
	for _, trace := range traces {
		stats, err := d.getFunctionStats(trace.Name)
		if err != nil {
			return nil, 0, fmt.Errorf("get function stats failed: %w", err)
		}
		f := dos.NewFunction(int64(trace.ID), trace.Name, stats.CallCount, stats.AvgTime, stats.AvgTime)
		f.SetPackage()
		functions = append(functions, f)
	}
	return functions, int32(len(functions)), nil
}

// GetFunctionInfoInGoroutine 获取函数在指定Goroutine中的信息
func (d *TraceEntDB) GetFunctionInfoInGoroutine(gid uint64, targetFunctionId int64) (*dos.FunctionInfo, error) {
	ctx := context.Background()

	// 1. 先查询目标函数信息
	targetTrace, err := d.client.TraceData.Query().Where(
		tracedata.ID(int(targetFunctionId)),
		tracedata.Gid(uint64(gid)),
	).First(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return nil, fmt.Errorf("function not found")
		}
		return nil, fmt.Errorf("query target function failed: %w", err)
	}

	functionInfo := &dos.FunctionInfo{
		ID:     targetFunctionId,
		Name:   targetTrace.Name,
		Indent: targetTrace.Indent,
	}

	// 2. 获取从当前函数到深度为0的所有父函数
	parentInfos, err := d.getParentFunctionInfos(targetTrace.Gid, targetFunctionId)
	if err != nil {
		return nil, fmt.Errorf("get parent function infos failed: %w", err)
	}
	functionInfo.ParentIds = parentInfos

	return functionInfo, nil
}

// getParentFunctionInfos 获取从当前函数到深度为0的所有父函数信息列表（去重）
func (d *TraceEntDB) getParentFunctionInfos(gid uint64, targetFunctionId int64) ([]dos.ParentInfo, error) {
	ctx := context.Background()

	// 查询目标函数的所有父函数
	query := d.client.TraceData.Query().
		Where(tracedata.Gid(gid)).
		Where(tracedata.IDLTE(int(targetFunctionId)))

	traces, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query parent functions failed: %w", err)
	}

	// 构建父函数信息
	parentMap := make(map[int64]dos.ParentInfo)
	for _, trace := range traces {
		if trace.ID < int(targetFunctionId) {
			parentInfo := dos.ParentInfo{
				ParentId: int64(trace.ID),
				Name:     trace.Name,
				Depth:    trace.Indent,
			}
			parentMap[int64(trace.ID)] = parentInfo
		}
	}

	// 转换为切片
	var result []dos.ParentInfo
	for _, parent := range parentMap {
		result = append(result, parent)
	}

	return result, nil
}

// GetSampledFunctionNames 获取采样后的函数名称列表
func (d *TraceEntDB) GetSampledFunctionNames(maxSamples int32) ([]string, error) {
	ctx := context.Background()

	// 首先获取总记录数
	totalCount, err := d.client.TraceData.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("get total count failed: %w", err)
	}

	var functionNames []string

	// 如果总记录数小于等于maxSamples，全量查询
	if int32(totalCount) <= maxSamples {
		traces, err := d.client.TraceData.Query().
			Select(tracedata.FieldName).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("query all function names failed: %w", err)
		}

		for _, trace := range traces {
			functionNames = append(functionNames, trace.Name)
		}
	} else {
		// 如果总记录数大于maxSamples，使用分页采样
		// 计算采样间隔
		sampleInterval := totalCount / int(maxSamples)
		if sampleInterval < 1 {
			sampleInterval = 1
		}

		// 分页查询进行采样
		for offset := 0; offset < totalCount && len(functionNames) < int(maxSamples); offset += sampleInterval {
			traces, err := d.client.TraceData.Query().
				Select(tracedata.FieldName).
				Limit(1).
				Offset(offset).
				All(ctx)
			if err != nil {
				return nil, fmt.Errorf("query sampled function names failed: %w", err)
			}

			if len(traces) > 0 {
				functionNames = append(functionNames, traces[0].Name)
			}
		}
	}

	return functionNames, nil
}

var (
	zstdEncoder, _ = zstd.NewWriter(nil)
	zstdDecoder, _ = zstd.NewReader(nil)
	magicNumber    = []byte{'F', 'T', 'Z', '$'} // Fun-Trace-Zstd Magic Number
)

// compress uses zstd to compress a string and returns the compressed data prefixed with a magic number.
func compress(s string) []byte {
	compressed := zstdEncoder.EncodeAll([]byte(s), nil)
	return append(magicNumber, compressed...)
}

// decompress uses zstd to decompress data. It checks for a magic number to identify
// compressed data and includes a fallback for uncompressed or corrupted data.
func decompress(data []byte) string {
	if len(data) < len(magicNumber) || !bytes.Equal(data[:len(magicNumber)], magicNumber) {
		// Data doesn't have the magic number, so it's legacy uncompressed data.
		return string(data)
	}

	// Data has the magic number, so it should be decompressed.
	decompressed, err := zstdDecoder.DecodeAll(data[len(magicNumber):], nil)
	if err != nil {
		// Data is corrupted. The calling function will fail to unmarshal it,
		// which will be logged as a high-level error.
		return string(data)
	}
	return string(decompressed)
}
