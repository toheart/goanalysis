package callgraph

// 调用图分析算法常量
const (
	DefaultOutput       = "./default.png"
	DefaultCache        = "./cachePath.json"
	CallGraphTypeStatic = "static"
	CallGraphTypeCha    = "cha"
	CallGraphTypeRta    = "rta" // 默认使用
	CallGraphTypeVta    = "vta"
)

// 状态更新频率常量
const (
	NodeStatusUpdateInterval = 10
	EdgeStatusUpdateInterval = 20
)

// ProgramOption 定义程序分析的配置选项函数类型
type ProgramOption func(p *ProgramAnalysis)

// FilterConfig 过滤配置
type FilterConfig struct {
	IgnorePaths []string
	ModuleName  string
}

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	TotalNodes     int
	ProcessedNodes int
}

// StatusReporter 状态报告器
type StatusReporter struct {
	statusChan chan []byte
}

// NewStatusReporter 创建状态报告器
func NewStatusReporter(statusChan chan []byte) *StatusReporter {
	return &StatusReporter{statusChan: statusChan}
}

// ReportStatus 报告状态
func (sr *StatusReporter) ReportStatus(message string) {
	if sr.statusChan != nil {
		sr.statusChan <- []byte(message)
	}
}
