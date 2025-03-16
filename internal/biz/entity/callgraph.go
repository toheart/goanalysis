package entity

import "time"

type FuncEdge struct {
	CallerKey string `json:"caller_key"`
	CalleeKey string `json:"callee_key"`
}

// FuncNode 表示函数节点
type FuncNode struct {
	Key      string   `json:"key"`      // 唯一标识
	Pkg      string   `json:"pkg"`      // 包名
	Name     string   `json:"name"`     // 函数名
	Parent   []string `json:"parent"`   // 父节点索引
	Children []string `json:"children"` // 子节点索引
}

// 数据库文件结构
type StaticDbFile struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	CreateTime time.Time `json:"createTime"`
}

// 分析结果结构
type StaticAnalysisResult struct {
	TotalFunctions      int                 `json:"totalFunctions"`
	TotalCalls          int                 `json:"totalCalls"`
	TotalPackages       int                 `json:"totalPackages"`
	PackageDependencies []PackageDependency `json:"packageDependencies"`
	HotFunctions        []HotFunction       `json:"hotFunctions"`
}

// 包依赖结构
type PackageDependency struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Count  int    `json:"count"`
}
