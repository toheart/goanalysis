package dos

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
