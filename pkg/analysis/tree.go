package analysis

import (
	"fmt"
	"github.com/goccy/go-graphviz/cgraph"
)

/**
@file:
@author: levi.Tang
@time: 2024/4/2 16:05
@description:
**/

type FuncNode struct {
	Key  string `json:"key"` // 唯一表示
	Pkg  string `json:"pkg"`
	Name string `json:"name"`

	Parent   []string `json:"parent"`   // 通过key来索引
	Children []string `json:"children"` // 通过key来索引
}

// ProgramAst
// @Description: 用于存储当前程序的所有对象, 以map作为索引
type ProgramAst struct {
	Nodes   map[string]*FuncNode `json:"nodes"`
	MainKey string               `json:"main_key"` // main函数

	isVisited map[string]struct{}
}

func NewProgramAst() *ProgramAst {
	return &ProgramAst{
		Nodes:     make(map[string]*FuncNode),
		isVisited: make(map[string]struct{}),
	}
}

// ParentToPng
//
//	@Description: DFS, 遍历当前元素的parent元素
//	@receiver p
func (p *ProgramAst) ParentToPng(node *FuncNode, children *cgraph.Node, graph *cgraph.Graph) {
	mNode, err := graph.CreateNode(fmt.Sprintf("%s.%s", node.Pkg, node.Name))
	if err != nil {
		fmt.Printf("node: %s , create Node, err: %s", node.Name, err)
		return
	}
	mNode.SetLabel(node.Name)
	if children != nil {
		_, err = graph.CreateEdge(node.Pkg, mNode, children)
		if err != nil {
			fmt.Printf("node: %s , create edge, err: %s", node.Name, err)
			return
		}
	}
	// 判断item是否之前已经被遍历过了
	if _, ok := p.isVisited[node.Key]; ok {
		return
	}
	p.isVisited[node.Key] = struct{}{}
	// 判断是否存在子树
	for _, item := range node.Parent {
		prnt, ok := p.Nodes[item]
		if !ok {
			continue
		}
		p.ParentToPng(prnt, mNode, graph)
	}
}

// ChildrenToPng
//
//	@Description: DFS, 遍历当前元素的Children元素
//	@receiver p
func (p *ProgramAst) ChildrenToPng(node *FuncNode, parent *cgraph.Node, graph *cgraph.Graph) {
	mNode, err := graph.CreateNode(fmt.Sprintf("%s.%s", node.Pkg, node.Name))
	if err != nil {
		fmt.Printf("node: %s , create Node, err: %s", node.Name, err)
		return
	}
	mNode.SetLabel(node.Name)
	if parent != nil {
		_, err = graph.CreateEdge(node.Pkg, parent, mNode)
		if err != nil {
			fmt.Printf("node: %s , create edge, err: %s", node.Name, err)
			return
		}
	}
	// 判断item是否之前已经被遍历过了
	if _, ok := p.isVisited[node.Key]; ok {
		return
	}
	p.isVisited[node.Key] = struct{}{}
	// 判断是否存在子树
	for _, item := range node.Children {
		chd, ok := p.Nodes[item]
		if !ok {
			continue
		}
		p.ChildrenToPng(chd, mNode, graph)
	}
}
