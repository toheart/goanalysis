package dos

import "strings"

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

// FunctionNode 函数调用节点
type FunctionNode struct {
	ID        string         // 节点ID
	Name      string         // 函数名称
	Package   string         // 包名
	CallCount int            // 调用次数
	AvgTime   string         // 平均耗时
	Children  []FunctionNode // 子节点
}

// FunctionInfo 函数在Goroutine中的信息
type FunctionInfo struct {
	ID        int64        `json:"id"`        // 函数ID
	Name      string       `json:"name"`      // 函数名称
	Indent    int          `json:"indent"`    // 函数在调用链中的实际深度
	Found     bool         `json:"found"`     // 是否在当前深度范围内找到
	ParentIds []ParentInfo `json:"parentIds"` // 父函数ID+深度列表（去重）
}

// ParentInfo 父函数信息
type ParentInfo struct {
	ParentId int64  `json:"parentId"` // 父函数ID
	Depth    int    `json:"depth"`    // 深度
	Name     string `json:"name"`     // 父函数名称
}

// 按频率排序，取前5个
type ModuleStat struct {
	Name  string
	Count int
}

type TraceParams struct {
	ID         int64  `json:"id"`         // 唯一标识符
	TraceID    int64  `json:"traceId"`    // 关联的TraceData ID
	Position   int    `json:"position"`   // 参数位置
	Data       string `json:"data"`       // 参数JSON数据
	IsReceiver bool   `json:"isReceiver"` // 是否为接收者参数
	BaseID     int64  `json:"baseId"`     // 基础参数ID（自关联，当参数为增量存储时使用）
}

type Function struct {
	Id         int64  // 函数ID
	Name       string // 函数名称
	Package    string // 包名
	ParentId   int64  // 父函数ID
	CallCount  int    // 调用次数
	TotalTime  string // 总耗时
	AvgTime    string // 平均耗时
	ParamCount int    // 参数数量
	Depth      int    // 深度
	Seq        string // 序列号
}

func NewFunction(id int64, name string, callCount int, totalTime string, avgTime string) *Function {
	f := &Function{
		Id:        id,
		Name:      name,
		CallCount: callCount,
		TotalTime: totalTime,
		AvgTime:   avgTime,
	}
	f.SetPackage()
	return f
}

func (f *Function) SetPackage() {
	parts := strings.Split(f.Name, "/")
	packageName := "main"
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1] // 取最后一个部分作为函数名
		packageNames := strings.Split(lastPart, ".")
		packageName = packageNames[0]
	}
	f.Package = packageName
}

// FunctionCallStats 函数调用统计信息
type FunctionCallStats struct {
	Name        string  // 函数名称
	Package     string  // 包名
	CallCount   int     // 调用次数
	CallerCount int     // 调用方数量
	CalleeCount int     // 被调用方数量
	AvgTime     string  // 平均执行时间
	MaxTime     string  // 最大执行时间
	MinTime     string  // 最小执行时间
	TimeStdDev  float64 // 执行时间标准差
}
