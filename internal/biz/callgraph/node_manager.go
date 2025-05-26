package callgraph

import (
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
)

// NodeManager 节点管理器，负责节点的创建和管理
type NodeManager struct {
	tree     map[string]*dos.FuncNode
	nodeChan chan *dos.FuncNode
}

// NewNodeManager 创建新的节点管理器
func NewNodeManager() *NodeManager {
	return &NodeManager{
		tree:     make(map[string]*dos.FuncNode),
		nodeChan: make(chan *dos.FuncNode, 100),
	}
}

// GetNodeChan 获取节点通道（用于外部消费）
func (nm *NodeManager) GetNodeChan() <-chan *dos.FuncNode {
	return nm.nodeChan
}

// NodeExists 检查节点是否存在
func (nm *NodeManager) NodeExists(key string) bool {
	_, exists := nm.tree[key]
	return exists
}

// GetNode 获取节点
func (nm *NodeManager) GetNode(key string) *dos.FuncNode {
	return nm.tree[key]
}

// CreateNode 创建节点
func (nm *NodeManager) CreateNode(key, pkg, name string) *dos.FuncNode {
	return &dos.FuncNode{
		Key:  key,
		Pkg:  pkg,
		Name: name,
	}
}

// AddNode 添加节点
func (nm *NodeManager) AddNode(node *dos.FuncNode) {	
	nm.tree[node.Key] = node
	nm.nodeChan <- node
}

// GetOrCreateNode 获取或创建节点
func (nm *NodeManager) GetOrCreateNode(key, pkg, name string) *dos.FuncNode {
	if nm.NodeExists(key) {
		return nm.GetNode(key)
	}

	node := nm.CreateNode(key, pkg, name)
	nm.AddNode(node)
	return node
}

// Close 关闭节点管理器
func (nm *NodeManager) Close() {
	close(nm.nodeChan)
}
