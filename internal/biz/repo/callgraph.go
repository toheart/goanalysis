package repo

import "github.com/toheart/goanalysis/internal/biz/entity"

// 添加用于存储的方法
type StaticDBStore interface {
	// SaveFuncNode 保存函数节点
	SaveFuncNode(node *entity.FuncNode) error

	// SaveFuncEdge 保存函数调用关系
	SaveFuncEdge(edge *entity.FuncEdge) error

	// GetFuncNodeByKey 根据Key获取函数节点
	GetFuncNodeByKey(key string) (*entity.FuncNode, error)

	// GetCallerEdges 获取调用该函数的所有边
	GetCallerEdges(calleeKey string) ([]*entity.FuncNode, error)

	// GetCalleeEdges 获取该函数调用的所有边
	GetCalleeEdges(callerKey string) ([]*entity.FuncNode, error)

	// GetAllFuncNodes 获取所有函数节点
	GetAllFuncNodes() ([]*entity.FuncNode, error)

	// GetAllFuncEdges 获取所有函数调用边
	GetAllFuncEdges() ([]*entity.FuncEdge, error)

	// InitTable 初始化数据库表
	InitTable() error
}
