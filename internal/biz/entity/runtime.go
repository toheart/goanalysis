package entity

// TraceData 存储跟踪数据的结构体
type TraceData struct {
	ID         int64  `json:"id"`         // 唯一标识符
	Name       string `json:"name"`       // 函数名称
	GID        uint64 `json:"gid"`        // Goroutine ID
	Indent     int    `json:"indent"`     // 缩进级别
	ParamCount int    `json:"paramCount"` // 参数数量
	TimeCost   string `json:"timeCost"`   // 执行时间
	ParentId   uint64 `json:"parentId"`   // 父函数ID
	CreatedAt  string `json:"createdAt"`  // 创建时间
	Seq        string `json:"seq"`        // 序列号
}

// GoroutineTrace 存储goroutine信息的结构体
type GoroutineTrace struct {
	ID           int64  `json:"id"`           // 自增ID
	GID          uint64 `json:"gid"`          // Goroutine ID
	TimeCost     string `json:"timeCost"`     // 执行时间
	CreateTime   string `json:"createTime"`   // 创建时间
	IsFinished   int    `json:"isFinished"`   // 是否完成
	InitFuncName string `json:"initFuncName"` // 初始函数名
	Depth        int    `json:"depth"`        // 调用深度
}

// GoroutineStats Goroutine统计信息
type GoroutineStats struct {
	Active   int    // 活跃Goroutine数量
	AvgTime  string // 平均执行时间
	MaxDepth int    // 最大调用深度
}

// FunctionNode 函数调用节点
type FunctionNode struct {
	ID        string         // 节点ID
	Name      string         // 函数名称
	Package   string         // 包名
	CallCount int            // 调用次数
	AvgTime   string         // 平均耗时
	Children  []FunctionNode // 子节点
}

// FunctionGraphNode 函数调用关系图节点
type FunctionGraphNode struct {
	ID        string // 节点ID
	Name      string // 函数名称
	Package   string // 包名
	CallCount int    // 调用次数
	AvgTime   string // 平均耗时
	NodeType  string // 节点类型: "root", "caller", "callee"
}

// FunctionGraphEdge 函数调用关系图边
type FunctionGraphEdge struct {
	Source   string // 源节点ID
	Target   string // 目标节点ID
	Label    string // 边标签
	EdgeType string // 边类型: "caller_to_root", "root_to_callee"
}
