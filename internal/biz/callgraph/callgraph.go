package callgraph

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

/**
@file:
@author: levi.Tang
@time: 2024/3/20 10:43
@description: 用于分析程序包, 整体逻辑
**/

// ProgramAnalysis 定义程序分析的主要结构
type ProgramAnalysis struct {
	algo        string   // 使用的分析算法
	Dir         string   // 项目目录
	ignorePaths []string // 需要忽略的路径
	onlyMethod  string   // 只分析特定方法
	log         *log.Helper

	cachePath  string // 缓存文件路径
	isCache    bool   // 是否使用缓存
	outputPath string // 输出文件路径

	tree map[string]*FuncNode // 语法树

	callGraph  *callgraph.Graph // 调用图
	data       DBStore          // 数据存储
	moduleName string           // 模块名

	edgeChan chan *FuncEdge // 边通道
	nodeChan chan *FuncNode // 节点通道
}

// NewProgramAnalysis 创建新的程序分析实例
func NewProgramAnalysis(dir string, logger log.Logger, data DBStore, opts ...ProgramOption) *ProgramAnalysis {
	p := &ProgramAnalysis{
		Dir:      dir,
		algo:     CallGraphTypeVta,
		data:     data,
		tree:     make(map[string]*FuncNode),
		nodeChan: make(chan *FuncNode, 100),
		edgeChan: make(chan *FuncEdge, 100),
		log:      log.NewHelper(log.With(logger, "module", "callgraph")),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *ProgramAnalysis) GetMainPackage(pkgs []*ssa.Package) ([]*ssa.Package, error) {
	p.log.Info("get main package")
	var mains []*ssa.Package
	for _, pkg := range pkgs {
		if pkg != nil && pkg.Pkg.Name() == "main" && pkg.Func("main") != nil {
			mains = append(mains, pkg)
		}
	}
	if len(mains) == 0 {
		return nil, fmt.Errorf("no main packages")
	}
	return mains, nil
}

func (p *ProgramAnalysis) NodeIsExist(key string) bool {
	if _, ok := p.tree[key]; ok {
		return true
	}
	return false
}

// GetNode 获取节点
func (p *ProgramAnalysis) GetNode(key string) *FuncNode {
	if !p.NodeIsExist(key) {
		return nil
	}
	return p.tree[key]
}

// GetModuleName 从go.mod文件中获取模块名
func (p *ProgramAnalysis) GetModuleName() (string, error) {
	// 从当前目录开始向上查找go.mod文件
	dir := p.Dir
	for {
		modPath := filepath.Join(dir, "go.mod")
		file, err := os.Open(modPath)
		if err == nil {
			defer file.Close()
			// 读取go.mod文件查找module名
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "module ") {
					return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
				}
			}
			if err := scanner.Err(); err != nil {
				return "", fmt.Errorf("error reading go.mod: %w", err)
			}
			return "", fmt.Errorf("module declaration not found in go.mod")
		}

		// 获取父目录
		parent := filepath.Dir(dir)
		if parent == dir {
			// 已经到达根目录
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// Analysis 执行程序分析
func (p *ProgramAnalysis) Analysis() error {
	p.log.Info("analysis")
	// 获取模块名
	moduleName, err := p.GetModuleName()
	if err != nil {
		p.log.Errorf("get module name failed: %v", err)
		return fmt.Errorf("get module name failed: %w", err)
	}
	p.moduleName = moduleName
	p.log.Infof("analyzing module: %s", p.moduleName)

	pkgs, err := p.loadPackages()
	if err != nil {
		p.log.Error("load packages failed: %w", err)
		return fmt.Errorf("load packages failed: %w", err)
	}

	prog, err := p.buildSSA(pkgs)
	if err != nil {
		p.log.Error("buildSSA failed: %w", err)
		return fmt.Errorf("buildSSA failed: %w", err)
	}

	if err := p.buildCallGraph(prog); err != nil {
		p.log.Error("build call graph failed: %w", err)
		return fmt.Errorf("build call graph failed: %w", err)
	}

	return nil
}

// loadPackages 加载项目包
func (p *ProgramAnalysis) loadPackages() ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: false,
		Dir:   p.Dir,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			if strings.HasSuffix(filename, "_test.go") {
				return nil, nil
			}
			return parser.ParseFile(fset, filename, src, parser.ParseComments)
		},
	}

	initial, err := packages.Load(cfg, p.Dir+"/...")
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(initial) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	return initial, nil
}

// buildSSA 构建SSA形式的程序表示
func (p *ProgramAnalysis) buildSSA(pkgs []*packages.Package) (*ssa.Program, error) {
	prog, _ := ssautil.AllPackages(pkgs, ssa.InstantiateGenerics)
	prog.Build()
	return prog, nil
}

// buildCallGraph 根据选择的算法构建调用图
func (p *ProgramAnalysis) buildCallGraph(prog *ssa.Program) error {
	p.log.Infof("build call graph, algo: %s", p.algo)
	switch p.algo {
	case CallGraphTypeStatic:
		p.callGraph = static.CallGraph(prog)
	case CallGraphTypeCha:
		p.callGraph = cha.CallGraph(prog)
	case CallGraphTypeRta:
		roots, err := p.getMainFunctions(prog)
		if err != nil {
			return err
		}
		p.callGraph = rta.Analyze(roots, true).CallGraph
	case CallGraphTypeVta:
		p.callGraph = vta.CallGraph(ssautil.AllFunctions(prog), cha.CallGraph(prog))
	default:
		return fmt.Errorf("invalid call graph type: %s", p.algo)
	}

	// 确保 callGraph 不为空
	if p.callGraph == nil {
		return fmt.Errorf("failed to build call graph")
	}

	return nil
}

// getMainFunctions 获取主函数
func (p *ProgramAnalysis) getMainFunctions(prog *ssa.Program) ([]*ssa.Function, error) {
	p.log.Info("get main functions")
	mains, err := p.GetMainPackage(prog.AllPackages())
	if err != nil {
		p.log.Error("get main package failed: %w", err)
		return nil, err
	}

	var roots []*ssa.Function
	for _, main := range mains {
		roots = append(roots, main.Func("main"))
	}
	p.log.Infof("get main functions: %v", roots)
	return roots, nil
}

// SetTree
//
//	@Description: 生成项目整体的语法树
//	@receiver p
//	@return error
func (p *ProgramAnalysis) SetTree() error {
	p.log.Info("set tree")
	if err := p.Analysis(); err != nil {
		return err
	}

	var inIgnores = func(node *callgraph.Node) bool {
		pkgPath := node.Func.String()

		for _, ignorePath := range p.ignorePaths {
			if strings.HasPrefix(pkgPath, ignorePath) {
				return true
			}
		}
		return false
	}

	err := callgraph.GraphVisitEdges(p.callGraph, func(edge *callgraph.Edge) error {
		if isSynthetic(edge) {
			return nil
		}
		caller := edge.Caller
		callee := edge.Callee
		// 排除标准库
		if inStd(caller) || inStd(callee) {
			return nil
		}
		p.log.Infof("caller: %s, callee: %s", caller.String(), callee.String())
		// 排除忽略的包
		if inIgnores(caller) || inIgnores(callee) {
			return nil
		}
		// caller是否存在
		var pNode, qNode *FuncNode
		// 如果不存在, 则创建
		if !p.NodeIsExist(caller.String()) {
			pNode = &FuncNode{
				Key:  caller.String(),
				Pkg:  caller.Func.Pkg.Pkg.Path(),
				Name: caller.Func.RelString(caller.Func.Pkg.Pkg),
			}
			p.tree[pNode.Key] = pNode
			p.nodeChan <- pNode
		} else {
			pNode = p.GetNode(caller.String())
		}
		// 如果不存在, 则创建
		if !p.NodeIsExist(callee.String()) {
			qNode = &FuncNode{
				Key:  callee.String(),
				Pkg:  callee.Func.Pkg.Pkg.Path(),
				Name: callee.Func.RelString(callee.Func.Pkg.Pkg),
			}
			p.tree[qNode.Key] = qNode
			p.nodeChan <- qNode
		} else {
			qNode = p.GetNode(callee.String())
		}
		pNode.Children = append(pNode.Children, qNode.Key)
		qNode.Parent = append(qNode.Parent, pNode.Key)
		p.edgeChan <- &FuncEdge{
			CallerKey: pNode.Key,
			CalleeKey: qNode.Key,
		}
		p.log.Infof("set edge caller: %s, --> callee: %s", pNode.Key, qNode.Key)

		return nil
	})
	if err != nil {
		return err
	}
	// 关闭通道
	close(p.nodeChan)
	close(p.edgeChan)
	p.log.Infof("set tree success")

	return err
}

func (p *ProgramAnalysis) isInternal(node *callgraph.Node) bool {
	// 如果pkgPath不包含moduleName, 则忽略
	return !strings.Contains(node.Func.String(), p.moduleName)
}

// SaveData 异步保存数据到数据库
func (p *ProgramAnalysis) SaveData(ctx context.Context) error {
	wg := sync.WaitGroup{}
	// 启动goroutine处理节点保存
	wg.Add(1)
	go func() {
		defer wg.Done()
		for node := range p.nodeChan {
			p.log.Infof("save node: %s", node.Key)
			if err := p.data.SaveFuncNode(node); err != nil {
				p.log.Error("save node failed: %v", err, "node", node)
			}
		}
	}()

	// 启动goroutine处理边保存
	wg.Add(1)
	go func() {
		defer wg.Done()
		for edge := range p.edgeChan {
			p.log.Infof("save edge: %s --> %s", edge.CallerKey, edge.CalleeKey)
			if err := p.data.SaveFuncEdge(edge); err != nil {
				p.log.Error("save edge failed: %v", err, "edge", edge)
			}
		}
	}()

	wg.Wait()
	p.log.Infof("save data exit")
	return nil
}

func isSynthetic(edge *callgraph.Edge) bool {
	return edge.Caller.Func.Pkg == nil || edge.Callee.Func.Pkg == nil || edge.Callee.Func.Synthetic != ""
}

func inStd(node *callgraph.Node) bool {
	return isStdPkgPath(node.Func.Pkg.Pkg.Path())
}

func isStdPkgPath(path string) bool {
	if strings.Contains(path, ".") {
		return false
	}
	return true
}

func WithAlgo(algo string) ProgramOption {
	return func(p *ProgramAnalysis) {
		p.algo = algo
	}
}

func WithIgnorePaths(ignorePath string) ProgramOption {
	var ignorePaths []string
	for _, item := range strings.Split(ignorePath, ",") {
		p := strings.TrimSpace(item)
		if p != "" {
			ignorePaths = append(ignorePaths, p)
		}
	}
	return func(p *ProgramAnalysis) {
		p.ignorePaths = ignorePaths
	}
}

func WithOnlyPkg(onlyPkg string) ProgramOption {
	return func(p *ProgramAnalysis) {
		p.onlyMethod = onlyPkg
	}
}

func WithOutputDir(output string) ProgramOption {
	return func(p *ProgramAnalysis) {
		p.outputPath = output
	}
}

func WithCacheDir(cache string) ProgramOption {
	return func(p *ProgramAnalysis) {
		p.cachePath = cache
	}
}

func WithCacheFlag(flag bool) ProgramOption {
	return func(p *ProgramAnalysis) {
		p.isCache = flag
	}
}
