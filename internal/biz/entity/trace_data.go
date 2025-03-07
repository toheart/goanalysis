package entity

import (
	"github.com/toheart/goanalysis/functrace"
)

type TraceData struct {
	ID             int
	Name           string
	GID            int
	Indent         int
	Params         []functrace.TraceParams
	TimeCost       string
	ParentFuncname string
}

// HotFunction 热点函数信息
type HotFunction struct {
	Name      string // 函数名称
	Package   string // 包名
	CallCount int    // 调用次数
	TotalTime string // 总耗时
	AvgTime   string // 平均耗时
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

// FunctionCallGraph 函数调用关系图
type FunctionCallGraph struct {
	Nodes []FunctionGraphNode // 图节点
	Edges []FunctionGraphEdge // 图边
}
