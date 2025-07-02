package entity

import (
	"strings"
	"time"
)

const (
	AnalysisEventStart = iota
	AnalysisEventProcessing
	AnalysisEventCompleted
	AnalysisEventFailed
)

// AnalysisTask 分析任务
type AnalysisTask struct {
	ID          string
	ProjectPath string
	Filename    string
	Options     *AnalysisOptions // 分析选项
}

// AnalysisOptions 分析选项
type AnalysisOptions struct {
	Algo         string // 分析算法
	IgnoreMethod string // 忽略分析特定方法
}

// TaskStatus 任务状态
const (
	TaskStatusStarting   = 0  // 启动中
	TaskStatusProcessing = 1  // 处理中
	TaskStatusCompleted  = 2  // 已完成
	TaskStatusFailed     = -1 // 失败
	TaskStatusNotFound   = -2 // 未找到
)

// AnalysisTaskStatus 分析任务状态
type AnalysisTaskStatus struct {
	Status   int     // 状态
	Progress float64 // 进度
	Message  string  // 消息
}

// DbFileInfo 数据库文件信息
type DbFileInfo struct {
	Path       string    // 路径
	Name       string    // 名称
	Size       int64     // 大小
	CreateTime time.Time // 创建时间
}

type AnalysisEvent struct {
	Type    int    // 类型
	Message string // 消息
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

// HotPathInfo 热点路径信息
type HotPathInfo struct {
	Path      []string // 路径上的函数名称序列
	CallCount int      // 调用次数
	TotalTime string   // 总耗时
	AvgTime   string   // 平均耗时
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

// PerformanceAnomaly 性能异常信息
type PerformanceAnomaly struct {
	Name        string            // 函数名称
	Package     string            // 包名
	AnomalyType string            // 异常类型: "time_variance", "depth_anomaly", "frequency_anomaly"
	Description string            // 异常描述
	Severity    float64           // 严重程度 (0.0-1.0)
	Details     map[string]string // 详细信息
}

type TraceParams struct {
	ID         int64  `json:"id"`         // 唯一标识符
	TraceID    int64  `json:"traceId"`    // 关联的TraceData ID
	Position   int    `json:"position"`   // 参数位置
	Data       string `json:"data"`       // 参数JSON数据
	IsReceiver bool   `json:"isReceiver"` // 是否为接收者参数
	BaseID     int64  `json:"baseId"`     // 基础参数ID（自关联，当参数为增量存储时使用）
}
