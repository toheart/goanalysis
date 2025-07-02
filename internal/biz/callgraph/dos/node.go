package dos

type FuncEdge struct {
	CallerKey string `json:"caller_key"`
	CalleeKey string `json:"callee_key"`
}

// FuncNode 表示函数节点
type FuncNode struct {
	Key       string      `json:"key"`       // 短格式唯一标识，如 "n6796"
	FullName  string      `json:"full_name"` // 完整的函数路径，如 "crypto/hmac.New$1"
	Pkg       string      `json:"pkg"`       // 包名
	Name      string      `json:"name"`      // 函数名
	Parents   []*FuncNode `json:"parents"`   // 父节点
	Childrens []*FuncNode `json:"childrens"` // 子节点
}
