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

	// 排除标准库节点：如果调用者或被调用者是标准库，则排除
	if f.IsStandardLibrary(caller) || f.IsStandardLibrary(callee) {
		return false
	}

	// 排除忽略的包
	if f.ShouldIgnore(caller) || f.ShouldIgnore(callee) {
		return false
	}

	// 排除调用者、被调用者两者都不属于当前项目的方法
	callerIsInternal := f.IsInternal(caller)
	calleeIsInternal := f.IsInternal(callee)

	// 至少有一个必须是内部模块，否则排除
	if !callerIsInternal && !calleeIsInternal {
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
	// 标准库包路径特征：
	// 1. 不包含域名（没有点号）
	// 2. 常见的标准库包前缀

	// 如果包含点号，通常是第三方包
	if strings.Contains(path, ".") {
		return false
	}

	// 空路径或特殊路径
	if path == "" || path == "command-line-arguments" {
		return false
	}

	// 常见的标准库包
	stdPkgs := []string{
		"fmt", "os", "io", "net", "http", "time", "strings", "bytes",
		"context", "errors", "log", "math", "sort", "sync", "unsafe",
		"runtime", "reflect", "encoding", "crypto", "database", "go",
		"html", "image", "index", "mime", "path", "regexp", "strconv",
		"testing", "text", "unicode", "archive", "bufio", "builtin",
		"compress", "container", "debug", "expvar", "flag", "hash",
		"heap", "plugin", "syscall",
	}

	// 检查是否是标准库包或其子包
	for _, stdPkg := range stdPkgs {
		if path == stdPkg || strings.HasPrefix(path, stdPkg+"/") {
			return true
		}
	}

	// 其他判断：不包含斜杠的单个词通常是标准库
	if !strings.Contains(path, "/") && path != "" {
		return true
	}

	return false
}
