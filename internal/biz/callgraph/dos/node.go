package dos

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
