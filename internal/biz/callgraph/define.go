package callgraph

// 常量定义
const (
	DefaultOutput       = "./default.png"
	DefaultCache        = "./cachePath.json"
	CallGraphTypeStatic = "static"
	CallGraphTypeCha    = "cha"
	CallGraphTypeRta    = "rta" // 默认使用
	CallGraphTypeVta    = "vta"
)

// ProgramOption 定义程序分析的配置选项函数类型
type ProgramOption func(p *ProgramAnalysis)
