package callgraph

import (
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
)

// EdgeManager 边管理器，负责边的创建和关系管理
type EdgeManager struct {
	edgeChan chan *dos.FuncEdge
}

// NewEdgeManager 创建新的边管理器
func NewEdgeManager() *EdgeManager {
	return &EdgeManager{
		edgeChan: make(chan *dos.FuncEdge, 100),
	}
}

// GetEdgeChan 获取边通道（用于外部消费）
func (em *EdgeManager) GetEdgeChan() <-chan *dos.FuncEdge {
	return em.edgeChan
}

// AddEdge 添加边
func (em *EdgeManager) AddEdge(callerKey, calleeKey string) {
	em.edgeChan <- &dos.FuncEdge{
		CallerKey: callerKey,
		CalleeKey: calleeKey,
	}
}

// BuildRelationship 建立节点间的父子关系
func (em *EdgeManager) BuildRelationship(caller, callee *dos.FuncNode) {
	if caller != nil && callee != nil {
		// 建立父子关系
		caller.Children = append(caller.Children, callee.Key)
		callee.Parent = append(callee.Parent, caller.Key)

		// 添加边到通道
		em.AddEdge(caller.Key, callee.Key)
	}
}

// Close 关闭边管理器
func (em *EdgeManager) Close() {
	close(em.edgeChan)
}
