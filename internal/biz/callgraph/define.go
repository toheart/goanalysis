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

// FuncNode 表示函数节点
type FuncNode struct {
	Key      string   `json:"key"`      // 唯一标识
	Pkg      string   `json:"pkg"`      // 包名
	Name     string   `json:"name"`     // 函数名
	Parent   []string `json:"parent"`   // 父节点索引
	Children []string `json:"children"` // 子节点索引
}
