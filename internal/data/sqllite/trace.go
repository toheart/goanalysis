package sqllite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite" // 引入 sqlite3 驱动

	"github.com/toheart/functrace"
	"github.com/toheart/goanalysis/internal/biz/entity"
)

type TraceDB struct {
	db *sql.DB
}

func NewTraceDB(dbPath string) (*TraceDB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("trace db file not found: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open trace db failed: %w", err)
	}

	return &TraceDB{db: db}, nil
}

func (d *TraceDB) GetTracesByGID(gid string) ([]entity.TraceData, error) {
	var traces []entity.TraceData
	rows, err := d.db.Query("SELECT id, name, gid, indent, params, timeCost, parentId, createdAt, seq FROM TraceData WHERE gid = ?", gid) // 使用 sqlite3 查询
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var trace entity.TraceData
		var paramsJSON string
		var timeCost sql.NullString  // 使用 sql.NullString 处理可能的 NULL 值
		var parentId sql.NullInt64   // 使用 sql.NullInt64 处理可能的 NULL 值
		var createdAt sql.NullString // 使用 sql.NullString 处理可能的 NULL 值
		var seq sql.NullString       // 使用 sql.NullString 处理可能的 NULL 值
		if err := rows.Scan(&trace.ID, &trace.Name, &trace.GID, &trace.Indent, &paramsJSON, &timeCost, &parentId, &createdAt, &seq); err != nil {
			return nil, err
		}

		// 将 JSON 字符串解析为列表
		var params []functrace.TraceParams
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, err
		}
		trace.Params = params // 假设 TraceData 结构体中有 Params 字段

		// 处理 timeCost 的值
		if timeCost.Valid {
			trace.TimeCost = timeCost.String // 只有在有效时才赋值
		} else {
			trace.TimeCost = "" // 或者设置为默认值
		}

		// 处理 parentId 的值
		if parentId.Valid {
			trace.ParentId = parentId.Int64 // 只有在有效时才赋值
		} else {
			trace.ParentId = 0 // 或者设置为默认值
		}

		// 处理 createdAt 的值
		if createdAt.Valid {
			trace.CreatedAt = createdAt.String // 只有在有效时才赋值
		} else {
			trace.CreatedAt = "" // 或者设置为默认值
		}

		// 处理 seq 的值
		if seq.Valid {
			trace.Seq = seq.String // 只有在有效时才赋值
		} else {
			trace.Seq = "" // 或者设置为默认值
		}

		traces = append(traces, trace)
	}
	return traces, nil
}

func (d *TraceDB) GetAllGIDs(page int, limit int) ([]uint64, error) {
	var gids []uint64
	offset := (page - 1) * limit // 计算偏移量
	query := "SELECT DISTINCT gid FROM TraceData LIMIT ? OFFSET ?"
	rows, err := d.db.Query(query, limit, offset) // 使用 LIMIT 和 OFFSET 进行分页
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var gid uint64
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		gids = append(gids, gid)
	}
	return gids, nil
}

func (d *TraceDB) GetAllFunctionName() ([]string, error) {
	var functionNames []string
	rows, err := d.db.Query("SELECT DISTINCT name FROM TraceData") // 查询所有不同的函数名
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var functionName string
		if err := rows.Scan(&functionName); err != nil {
			return nil, err
		}
		functionNames = append(functionNames, functionName)
	}
	return functionNames, nil
}

func (d *TraceDB) GetParamsByID(id int32) ([]functrace.TraceParams, error) {
	var params []functrace.TraceParams
	var paramsJSON string
	rows, err := d.db.Query("SELECT params FROM TraceData WHERE id = ?", id) // 使用 sqlite3 查询
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&paramsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, err
		}
	}
	return params, nil
}

func (d *TraceDB) GetGidsByFunctionName(functionName string) ([]string, error) {
	var gids []string
	rows, err := d.db.Query("SELECT DISTINCT gid FROM TraceData WHERE name = ?", functionName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var gid string
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		gids = append(gids, gid)
	}
	return gids, nil
}

func (d *TraceDB) Close() error {
	return d.db.Close()
}

func (d *TraceDB) GetTotalGIDs() (int, error) {
	var total int
	query := "SELECT COUNT(DISTINCT gid) FROM TraceData"
	err := d.db.QueryRow(query).Scan(&total) // 查询总数
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (d *TraceDB) GetInitialFunc(gid uint64) (string, error) {
	var initialFunc string
	query := "SELECT name FROM TraceData WHERE gid = ? limit 1 offset 0 "
	err := d.db.QueryRow(query, gid).Scan(&initialFunc)
	if err != nil {
		return "", err
	}
	return initialFunc, nil
}

// GetTracesByParentId 根据父函数ID查询函数调用
func (d *TraceDB) GetTracesByParentId(parentId int64) ([]entity.TraceData, error) {
	var traces []entity.TraceData
	rows, err := d.db.Query("SELECT id, name, gid, indent, params, timeCost, parentId, createdAt, seq FROM TraceData WHERE parentId = ?", parentId) // 使用 sqlite3 查询
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var trace entity.TraceData
		var paramsJSON string
		var timeCost sql.NullString  // 使用 sql.NullString 处理可能的 NULL 值
		var parentId sql.NullInt64   // 使用 sql.NullInt64 处理可能的 NULL 值
		var createdAt sql.NullString // 使用 sql.NullString 处理可能的 NULL 值
		var seq sql.NullString       // 使用 sql.NullString 处理可能的 NULL 值
		if err := rows.Scan(&trace.ID, &trace.Name, &trace.GID, &trace.Indent, &paramsJSON, &timeCost, &parentId, &createdAt, &seq); err != nil {
			return nil, err
		}

		// 将 JSON 字符串解析为列表
		var params []functrace.TraceParams
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, err
		}
		trace.Params = params

		// 处理 timeCost 的值
		if timeCost.Valid {
			trace.TimeCost = timeCost.String
		} else {
			trace.TimeCost = ""
		}

		// 处理 parentId 的值
		if parentId.Valid {
			trace.ParentId = parentId.Int64
		} else {
			trace.ParentId = 0
		}

		// 处理 createdAt 的值
		if createdAt.Valid {
			trace.CreatedAt = createdAt.String
		} else {
			trace.CreatedAt = ""
		}

		// 处理 seq 的值
		if seq.Valid {
			trace.Seq = seq.String
		} else {
			trace.Seq = ""
		}

		traces = append(traces, trace)
	}
	return traces, nil
}

// GetAllParentIds 获取所有的父函数ID
func (d *TraceDB) GetAllParentIds() ([]int64, error) {
	var parentIds []int64
	rows, err := d.db.Query("SELECT DISTINCT parentId FROM TraceData WHERE parentId IS NOT NULL AND parentId != 0") // 查询所有不同的父函数ID
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var parentId int64
		if err := rows.Scan(&parentId); err != nil {
			return nil, err
		}
		parentIds = append(parentIds, parentId)
	}
	return parentIds, nil
}

// GetChildFunctions 获取函数的子函数
func (d *TraceDB) GetChildFunctions(parentId int64) ([]string, error) {
	var childFunctions []string
	rows, err := d.db.Query("SELECT DISTINCT name FROM TraceData WHERE parentId = ?", parentId) // 查询所有以指定函数ID为父函数的函数名
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var childFunction string
		if err := rows.Scan(&childFunction); err != nil {
			return nil, err
		}
		childFunctions = append(childFunctions, childFunction)
	}
	return childFunctions, nil
}

// GetHotFunctions 获取热点函数分析数据
func (d *TraceDB) GetHotFunctions(sortBy string) ([]entity.HotFunction, error) {
	var hotFunctions []entity.HotFunction

	// 构建查询SQL
	query := `
		SELECT 
			name,
			COUNT(*) as call_count,
			SUM(CASE WHEN timeCost != '' THEN CAST(REPLACE(REPLACE(timeCost, 'ms', ''), 's', '000') AS REAL) ELSE 0 END) as total_time
		FROM 
			TraceData
		GROUP BY 
			name
	`

	// 根据排序方式添加ORDER BY子句
	if sortBy == "time" {
		query += " ORDER BY total_time DESC"
	} else {
		// 默认按调用次数排序
		query += " ORDER BY call_count DESC"
	}

	// 限制返回数量
	query += " LIMIT 50"

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query hot functions failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var hotFunc entity.HotFunction
		var totalTime float64

		if err := rows.Scan(&hotFunc.Name, &hotFunc.CallCount, &totalTime); err != nil {
			return nil, fmt.Errorf("scan hot function data failed: %w", err)
		}

		// 提取包名
		parts := strings.Split(hotFunc.Name, ".")
		if len(parts) > 1 {
			hotFunc.Package = strings.Join(parts[:len(parts)-1], ".")
		} else {
			hotFunc.Package = "main"
		}

		// 格式化时间
		if totalTime > 1000 {
			hotFunc.TotalTime = fmt.Sprintf("%.2fs", totalTime/1000)
		} else {
			hotFunc.TotalTime = fmt.Sprintf("%.2fms", totalTime)
		}

		// 计算平均时间
		avgTime := 0.0
		if hotFunc.CallCount > 0 {
			avgTime = totalTime / float64(hotFunc.CallCount)
		}

		if avgTime > 1000 {
			hotFunc.AvgTime = fmt.Sprintf("%.2fs", avgTime/1000)
		} else {
			hotFunc.AvgTime = fmt.Sprintf("%.2fms", avgTime)
		}

		hotFunctions = append(hotFunctions, hotFunc)
	}

	return hotFunctions, nil
}

// GetGoroutineStats 获取Goroutine统计信息
func (d *TraceDB) GetGoroutineStats() (*entity.GoroutineStats, error) {
	stats := &entity.GoroutineStats{}

	// 获取活跃Goroutine数量
	err := d.db.QueryRow("SELECT COUNT(DISTINCT gid) FROM TraceData").Scan(&stats.Active)
	if err != nil {
		return nil, fmt.Errorf("query active goroutines failed: %w", err)
	}

	// 获取最大调用深度
	err = d.db.QueryRow("SELECT MAX(indent) FROM TraceData").Scan(&stats.MaxDepth)
	if err != nil {
		return nil, fmt.Errorf("query max depth failed: %w", err)
	}

	// 计算平均执行时间
	var avgTime float64
	err = d.db.QueryRow(`
		SELECT 
			AVG(CASE WHEN timeCost != '' THEN CAST(REPLACE(REPLACE(timeCost, 'ms', ''), 's', '000') AS REAL) ELSE 0 END) 
		FROM 
			TraceData
	`).Scan(&avgTime)
	if err != nil {
		return nil, fmt.Errorf("query avg time failed: %w", err)
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
func (d *TraceDB) GetFunctionAnalysis(functionName string, queryType string) ([]entity.FunctionNode, error) {
	var result []entity.FunctionNode

	// 生成唯一ID
	nextID := 1
	generateID := func() string {
		id := fmt.Sprintf("node_%d", nextID)
		nextID++
		return id
	}

	// 根据查询类型选择不同的查询方式
	if queryType == "caller" {
		// 查询调用指定函数的函数
		callers, err := d.getCallerFunctions(functionName)
		if err != nil {
			return nil, fmt.Errorf("get caller functions failed: %w", err)
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

		result = append(result, rootNode)
	} else {
		// 查询被指定函数调用的函数
		callees, err := d.getCalleeFunctions(functionName)
		if err != nil {
			return nil, fmt.Errorf("get callee functions failed: %w", err)
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

		result = append(result, rootNode)
	}

	return result, nil
}

// 获取调用指定函数的函数列表
func (d *TraceDB) getCallerFunctions(functionName string) ([]string, error) {
	var callers []string

	// 查询调用了指定函数的父函数
	query := `
		SELECT DISTINCT name 
		FROM TraceData 
		WHERE parentId IN (
			SELECT id FROM TraceData WHERE name = ?
		)
	`

	rows, err := d.db.Query(query, functionName)
	if err != nil {
		return nil, fmt.Errorf("query caller functions failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var caller string
		if err := rows.Scan(&caller); err != nil {
			return nil, fmt.Errorf("scan caller function failed: %w", err)
		}
		callers = append(callers, caller)
	}

	return callers, nil
}

// 获取被指定函数调用的函数列表
func (d *TraceDB) getCalleeFunctions(functionName string) ([]string, error) {
	var callees []string

	// 查询被指定函数调用的函数
	query := `
		SELECT DISTINCT name 
		FROM TraceData 
		WHERE parentId = (
			SELECT id FROM TraceData WHERE name = ? LIMIT 1
		)
	`

	rows, err := d.db.Query(query, functionName)
	if err != nil {
		return nil, fmt.Errorf("query callee functions failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var callee string
		if err := rows.Scan(&callee); err != nil {
			return nil, fmt.Errorf("scan callee function failed: %w", err)
		}
		callees = append(callees, callee)
	}

	return callees, nil
}

// 函数统计信息
type functionStats struct {
	CallCount int
	AvgTime   string
}

// 获取函数的调用次数和平均时间
func (d *TraceDB) getFunctionStats(functionName string) (*functionStats, error) {
	query := `
		SELECT 
			COUNT(*) as call_count,
			AVG(CASE WHEN timeCost != '' THEN CAST(REPLACE(REPLACE(timeCost, 'ms', ''), 's', '000') AS REAL) ELSE 0 END) as avg_time
		FROM 
			TraceData
		WHERE 
			name = ?
	`

	var stats functionStats
	var avgTime float64

	err := d.db.QueryRow(query, functionName).Scan(&stats.CallCount, &avgTime)
	if err != nil {
		return nil, fmt.Errorf("query function stats failed: %w", err)
	}

	// 格式化平均时间
	if avgTime > 1000 {
		stats.AvgTime = fmt.Sprintf("%.2fs", avgTime/1000)
	} else {
		stats.AvgTime = fmt.Sprintf("%.2fms", avgTime)
	}

	return &stats, nil
}

// GetFunctionCallGraph 获取函数调用关系图
func (d *TraceDB) GetFunctionCallGraph(functionName string, depth int, direction string) (*entity.FunctionCallGraph, error) {
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
		err := d.addCallerNodes(graph, nodeMap, functionName, rootID, depth, generateID)
		if err != nil {
			return nil, fmt.Errorf("add caller nodes failed: %w", err)
		}
	}

	// 处理被调用者（向下查询）
	if direction == "callee" || direction == "both" {
		err := d.addCalleeNodes(graph, nodeMap, functionName, rootID, depth, generateID)
		if err != nil {
			return nil, fmt.Errorf("add callee nodes failed: %w", err)
		}
	}

	return graph, nil
}

// 添加调用者节点（向上查询）
func (d *TraceDB) addCallerNodes(graph *entity.FunctionCallGraph, nodeMap map[string]string, funcName string, parentID string, depth int, generateID func() string) error {
	if depth <= 0 {
		return nil
	}

	// 获取调用者列表
	callers, err := d.getCallerFunctions(funcName)
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
			err := d.addCallerNodes(graph, nodeMap, caller, nodeID, depth-1, generateID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 添加被调用者节点（向下查询）
func (d *TraceDB) addCalleeNodes(graph *entity.FunctionCallGraph, nodeMap map[string]string, funcName string, parentID string, depth int, generateID func() string) error {
	if depth <= 0 {
		return nil
	}

	// 获取被调用者列表
	callees, err := d.getCalleeFunctions(funcName)
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
			err := d.addCalleeNodes(graph, nodeMap, callee, nodeID, depth-1, generateID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetGoroutineCallDepth 获取指定 Goroutine 的最大调用深度
func (d *TraceDB) GetGoroutineCallDepth(gid uint64) (int, error) {
	var maxIndent int
	query := "SELECT MAX(indent) FROM TraceData WHERE gid = ?"
	err := d.db.QueryRow(query, gid).Scan(&maxIndent)
	if err != nil {
		return 0, err
	}
	// 调用深度为最大缩进级别 + 1
	return maxIndent + 1, nil
}

// GetGoroutineExecutionTime 获取指定 Goroutine 的总执行时间
func parseTimeCost(timeCostStr string) (int64, error) {
	duration, err := time.ParseDuration(timeCostStr)
	if err != nil {
		return 0, err
	}
	return int64(duration.Milliseconds()), nil
}

func formatTime(totalTime int64) string {
	return fmt.Sprintf("%d ms", totalTime) // 返回总时间的字符串格式
}

func (d *TraceDB) GetGoroutineExecutionTime(gid uint64) (string, error) {
	var totalTime int64
	rows, err := d.db.Query(`
		SELECT timeCost FROM TraceData 
		WHERE gid = ? AND indent = 0
	`, gid)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var timeCostStr sql.NullString
		if err := rows.Scan(&timeCostStr); err != nil {
			return "", err
		}
		if timeCostStr.String == "" {
			continue
		}
		timeCost, err := parseTimeCost(timeCostStr.String)
		if err != nil {
			return "", err
		}
		totalTime += timeCost
	}

	return formatTime(totalTime), nil
}

// IsGoroutineFinished 检查指定的goroutine是否已完成
// 判断依据：如果一个goroutine中所有indent=0的函数都有timeCost值，则认为该goroutine已完成
func (d *TraceDB) IsGoroutineFinished(gid uint64) (bool, error) {
	// 查询indent=0的函数数量
	var totalCount int
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM TraceData 
		WHERE gid = ? AND indent = 0
	`, gid).Scan(&totalCount)
	if err != nil {
		return false, err
	}

	if totalCount == 0 {
		return false, nil // 没有找到indent=0的函数，认为未完成
	}

	// 查询有timeCost值的indent=0函数数量
	var finishedCount int
	err = d.db.QueryRow(`
		SELECT COUNT(*) FROM TraceData 
		WHERE gid = ? AND indent = 0 AND timeCost IS NOT NULL AND timeCost != ''
	`, gid).Scan(&finishedCount)
	if err != nil {
		return false, err
	}

	// 如果所有indent=0的函数都有timeCost值，则认为该goroutine已完成
	return finishedCount == totalCount, nil
}

// GetUnfinishedFunctions 获取未完成的函数列表
func (d *TraceDB) GetUnfinishedFunctions(threshold int64) ([]entity.UnfinishedFunction, error) {
	var functions []entity.UnfinishedFunction

	// 获取最后一条记录的时间
	var lastRecordTime string
	err := d.db.QueryRow(`
		SELECT createdAt FROM TraceData 
		ORDER BY createdAt DESC LIMIT 1
	`).Scan(&lastRecordTime)
	if err != nil {
		return nil, fmt.Errorf("get last record time failed: %w", err)
	}

	// 解析最后记录时间
	lastTime, err := time.Parse("2006-01-02 15:04:05", lastRecordTime)
	if err != nil {
		return nil, fmt.Errorf("parse last record time failed: %w", err)
	}

	// 查询所有未完成的函数（timeCost为空）
	rows, err := d.db.Query(`
		SELECT id, name, gid, indent, parentId, createdAt FROM TraceData 
		WHERE timeCost IS NULL OR timeCost = ''
	`)
	if err != nil {
		return nil, fmt.Errorf("query unfinished functions failed: %w", err)
	}
	defer rows.Close()

	// 处理每个未完成的函数
	for rows.Next() {
		var function entity.TraceData
		if err := rows.Scan(&function.ID, &function.Name, &function.GID, &function.Indent, &function.ParentId, &function.CreatedAt); err != nil {
			continue
		}

		// 计算运行时间（从创建时间到现在）
		var runningTime string
		var runningTimeMs int64
		createdTime, _ := time.Parse("2006-01-02 15:04:05", function.CreatedAt)
		duration := lastTime.Sub(createdTime)
		runningTimeMs = duration.Milliseconds()

		if runningTimeMs > 1000 {
			runningTime = fmt.Sprintf("%.2fs", float64(runningTimeMs)/1000)
		} else {
			runningTime = fmt.Sprintf("%dms", runningTimeMs)
		}

		// 只有运行时间超过阈值的函数才被视为未完成函数
		if runningTimeMs <= threshold {
			continue
		}

		// 检查是否为阻塞函数（运行时间超过阈值）
		isBlocking := runningTimeMs >= threshold

		// 添加到未完成函数列表
		functions = append(functions, entity.UnfinishedFunction{
			Name:        function.Name,
			GID:         function.GID,
			RunningTime: runningTime,
			IsBlocking:  isBlocking,
			FunctionID:  function.ID,
		})
	}

	return functions, nil
}
