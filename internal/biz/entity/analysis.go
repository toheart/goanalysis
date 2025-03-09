package entity

import "time"

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
	DbPath      string
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
