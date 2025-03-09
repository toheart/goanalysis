package callgraph

type FuncEdge struct {
	CallerKey string `json:"caller_key"`
	CalleeKey string `json:"callee_key"`
}

// 添加用于存储的方法
type DBStore interface {
	// SaveFuncNode 保存函数节点
	SaveFuncNode(node *FuncNode) error

	// SaveFuncEdge 保存函数调用关系
	SaveFuncEdge(edge *FuncEdge) error

	// GetFuncNodeByKey 根据Key获取函数节点
	GetFuncNodeByKey(key string) (*FuncNode, error)

	// GetCallerEdges 获取调用该函数的所有边
	GetCallerEdges(calleeKey string) ([]*FuncNode, error)

	// GetCalleeEdges 获取该函数调用的所有边
	GetCalleeEdges(callerKey string) ([]*FuncNode, error)

	// GetAllFuncNodes 获取所有函数节点
	GetAllFuncNodes() ([]*FuncNode, error)

	// GetAllFuncEdges 获取所有函数调用边
	GetAllFuncEdges() ([]*FuncEdge, error)

	// InitTable 初始化数据库表
	InitTable() error
}
