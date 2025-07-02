package repo

import (
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
)

// 添加用于存储的方法
type StaticDBStore interface {
	// SaveFuncNode 保存函数节点
	SaveFuncNode(node *dos.FuncNode) error

	// SaveFuncEdge 保存函数调用关系
	SaveFuncEdge(edge *dos.FuncEdge) error

	// GetFuncNodeByKey 根据Key获取函数节点
	GetFuncNodeByKey(key string) (*dos.FuncNode, error)

	// GetCallerEdges 获取调用该函数的所有边
	GetCallerEdges(calleeKey string) ([]*dos.FuncNode, error)

	// GetCalleeEdges 获取该函数调用的所有边
	GetCalleeEdges(callerKey string) ([]*dos.FuncNode, error)

	// GetAllFuncNodes 获取所有函数节点
	GetAllFuncNodes() ([]*dos.FuncNode, error)

	// GetAllFuncEdges 获取所有函数调用边
	GetAllFuncEdges() ([]*dos.FuncEdge, error)

	// SearchFuncNodes 模糊搜索函数节点
	SearchFuncNodes(query string, limit int) ([]*dos.FuncNode, error)

	// InitTable 初始化数据库表
	InitTable() error
}
