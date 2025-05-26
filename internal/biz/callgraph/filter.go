package callgraph

import (
	"strings"

	"golang.org/x/tools/go/callgraph"
)

// Filter 过滤器结构
type Filter struct {
	config *FilterConfig
}

// NewFilter 创建新的过滤器
func NewFilter(config *FilterConfig) *Filter {
	return &Filter{config: config}
}

// ShouldIgnore 检查节点是否应该被忽略
func (f *Filter) ShouldIgnore(node *callgraph.Node) bool {
	pkgPath := node.Func.String()
	for _, ignorePath := range f.config.IgnorePaths {
		if strings.HasPrefix(pkgPath, ignorePath) {
			return true
		}
	}
	return false
}

// IsStandardLibrary 检查节点是否为标准库
func (f *Filter) IsStandardLibrary(node *callgraph.Node) bool {
	return isStdPkgPath(node.Func.Pkg.Pkg.Path())
}

// IsInternal 检查节点是否为内部模块
func (f *Filter) IsInternal(node *callgraph.Node) bool {
	return strings.Contains(node.Func.String(), f.config.ModuleName)
}

// ShouldProcessEdge 检查边是否应该被处理
func (f *Filter) ShouldProcessEdge(edge *callgraph.Edge) bool {
	// 跳过合成边
	if isSynthetic(edge) {
		return false
	}

	caller := edge.Caller
	callee := edge.Callee

	// 排除标准库
	if f.IsStandardLibrary(caller) && f.IsStandardLibrary(callee) {
		return false
	}

	// 排除忽略的包
	if f.ShouldIgnore(caller) || f.ShouldIgnore(callee) {
		return false
	}

	// 至少有一个是内部模块
	if !f.IsInternal(caller) && !f.IsInternal(callee) {
		return false
	}

	return true
}

// isSynthetic 检查边是否为合成边
func isSynthetic(edge *callgraph.Edge) bool {
	return edge.Caller.Func.Pkg == nil ||
		edge.Callee.Func.Pkg == nil ||
		edge.Callee.Func.Synthetic != ""
}

// isStdPkgPath 检查包路径是否为标准库
func isStdPkgPath(path string) bool {
	// 标准库包路径通常不包含域名（没有点号）
	if strings.Contains(path, ".") {
		return false
	}
	return true
}
