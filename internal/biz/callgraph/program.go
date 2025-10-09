package callgraph

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
	"github.com/toheart/goanalysis/internal/biz/repo"
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
	// 配置相关
	algo        string   // 使用的分析算法
	Dir         string   // 项目目录
	ignorePaths []string // 需要忽略的路径
	onlyMethod  string   // 只分析特定方法
	cachePath   string   // 缓存文件路径
	isCache     bool     // 是否使用缓存
	outputPath  string   // 输出文件路径

	// 依赖注入
	log  *log.Helper
	data repo.StaticDBStore // 数据存储

	// 分析结果
	callGraph  *callgraph.Graph // 调用图
	moduleName string           // 模块名

	// 组件
	nodeManager    *NodeManager
	edgeManager    *EdgeManager
	filter         *Filter
	reporter       *StatusReporter
	tracker        *ProgressTracker
	importAnalyzer *ImportAnalyzer

	// 状态跟踪
	isVisited map[string]bool // 是否访问过
}

// NewProgramAnalysis 创建新的程序分析实例
func NewProgramAnalysis(dir string, log *log.Helper, data repo.StaticDBStore, opts ...ProgramOption) *ProgramAnalysis {
	log.Infof("NewProgramAnalysis, dir: %s", dir)

	p := &ProgramAnalysis{
		Dir:         dir,
		algo:        CallGraphTypeVta,
		data:        data,
		isVisited:   make(map[string]bool),
		log:         log,
		nodeManager: NewNodeManager(),
		edgeManager: NewEdgeManager(),
		tracker:     &ProgressTracker{},
	}

	// 应用选项
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

// NodeExists 检查节点是否存在
func (p *ProgramAnalysis) NodeExists(key string) bool {
	return p.nodeManager.NodeExists(key)
}

// GetNode 获取节点
func (p *ProgramAnalysis) GetNode(key string) *dos.FuncNode {
	return p.nodeManager.GetNode(key)
}

// Execute 执行完整的调用图分析流程
// 这是对外提供的主要接口，内聚了所有内部操作
func (p *ProgramAnalysis) Execute(ctx context.Context, statusChan chan []byte) error {
	p.log.Info("execute call graph analysis")

	// 首先获取模块名（必须在 analyzeImports 之前）
	moduleName, err := p.GetModuleName()
	if err != nil {
		p.log.Errorf("get module name failed: %v", err)
		return fmt.Errorf("get module name failed: %w", err)
	}
	p.moduleName = moduleName
	p.log.Infof("analyzing module: %s", p.moduleName)

	// 初始化数据库表
	if err := p.data.InitTable(); err != nil {
		return fmt.Errorf("failed to init database table: %w", err)
	}

	// 启动数据消费者（并发执行）
	errChan := make(chan error)
	go func() {
		errChan <- p.consumeData(ctx, statusChan)
	}()

	// 分析import依赖关系（必须在 moduleName 初始化之后）
	if err := p.analyzeImports(); err != nil {
		p.log.Errorf("failed to analyze imports: %v", err)
		return fmt.Errorf("failed to analyze imports: %w", err)
	}

	// 生产数据到channels
	if err := p.produceData(statusChan); err != nil {
		p.log.Errorf("failed to produce data: %v", err)
		return fmt.Errorf("failed to produce data: %w", err)
	}

	// 关闭channels，通知消费者数据生产完毕
	p.nodeManager.Close()
	p.edgeManager.Close()

	// 等待数据消费完成
	if err := <-errChan; err != nil {
		p.log.Errorf("failed to consume data: %v", err)
		return fmt.Errorf("failed to consume data: %w", err)
	}

	p.log.Info("call graph analysis completed successfully")
	return nil
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

	// 初始化import分析器
	p.importAnalyzer = NewImportAnalyzer(p.moduleName, p.log)

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

	initial, err := packages.Load(cfg, "./...")
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

// setTree 构建调用图树结构（内部方法）
func (p *ProgramAnalysis) setTree(statusChan chan []byte) error {
	p.log.Info("set tree")
	if err := p.Analysis(); err != nil {
		return err
	}

	// 初始化组件
	p.reporter = NewStatusReporter(statusChan)
	p.filter = NewFilter(&FilterConfig{
		IgnorePaths: p.ignorePaths,
		ModuleName:  p.moduleName,
	})

	// 发送分析开始状态
	p.reporter.ReportStatus(fmt.Sprintf("Starting to build call graph, using algorithm: %s", p.algo))

	// 统计信息
	nodeCount := 0
	edgeCount := 0
	p.tracker.TotalNodes = len(p.callGraph.Nodes)

	err := callgraph.GraphVisitEdges(p.callGraph, func(edge *callgraph.Edge) error {
		caller := edge.Caller
		callee := edge.Callee

		callerKey := fmt.Sprintf("n%d", caller.ID)
		if !p.isVisited[callerKey] {
			p.isVisited[callerKey] = true
			p.tracker.ProcessedNodes++
		}

		// 使用过滤器检查是否应该处理这条边
		if !p.filter.ShouldProcessEdge(edge) {
			return nil
		}

		p.log.Infof("caller: %s, callee: %s", caller.String(), callee.String())

		// 处理caller节点
		callerFullName := caller.String()
		callerPkg := caller.Func.Pkg.Pkg.Path()
		callerName := caller.Func.RelString(caller.Func.Pkg.Pkg)
		if !p.nodeManager.NodeExists(callerKey) {
			nodeCount++
			// 每处理10个节点发送一次状态更新
			if nodeCount%NodeStatusUpdateInterval == 0 {
				p.reporter.ReportStatus(fmt.Sprintf("Processed %d nodes, %d edges", nodeCount, edgeCount))
			}
		}
		callerSignature := p.extractSignature(caller.Func)
		callerPosition := p.extractPosition(caller.Func)
		callerSourceCode := p.extractSourceCode(caller.Func)
		callerNode := p.nodeManager.GetOrCreateNode(caller.ID, callerFullName, callerPkg, callerPkg, callerName, callerSignature, callerPosition, callerSourceCode)

		// 处理callee节点
		calleeFullName := callee.String()
		calleePkg := callee.Func.Pkg.Pkg.Path()
		calleeName := callee.Func.RelString(callee.Func.Pkg.Pkg)

		calleeKey := fmt.Sprintf("n%d", callee.ID)
		if !p.nodeManager.NodeExists(calleeKey) {
			nodeCount++
			// 每处理10个节点发送一次状态更新
			if nodeCount%NodeStatusUpdateInterval == 0 {
				p.reporter.ReportStatus(fmt.Sprintf("Processed %d nodes, %d edges", nodeCount, edgeCount))
			}
		}
		calleeSignature := p.extractSignature(callee.Func)
		calleePosition := p.extractPosition(callee.Func)
		calleeSourceCode := p.extractSourceCode(callee.Func)
		calleeNode := p.nodeManager.GetOrCreateNode(callee.ID, calleeFullName, calleePkg, calleePkg, calleeName, calleeSignature, calleePosition, calleeSourceCode)

		// 提取调用点信息
		callSiteFile, callSiteLine, callType := p.extractCallSiteInfo(edge)

		// 建立边关系 - 使用EdgeManager封装逻辑
		p.edgeManager.BuildRelationship(callerNode, calleeNode, callSiteFile, callSiteLine, callType)
		edgeCount++

		// 每处理20条边发送一次状态更新
		if edgeCount%EdgeStatusUpdateInterval == 0 {
			p.reporter.ReportStatus(fmt.Sprintf("Processed %d nodes, %d edges", nodeCount, edgeCount))
		}

		p.log.Infof("set edge caller: %s, --> callee: %s", callerNode.Key, calleeNode.Key)

		return nil
	})

	if err != nil {
		p.reporter.ReportStatus(fmt.Sprintf("Call graph build error: %v", err))
		return err
	}

	// 关闭管理器
	p.nodeManager.Close()
	p.edgeManager.Close()

	// 发送完成状态
	p.reporter.ReportStatus(fmt.Sprintf("Call graph build completed, processed %d nodes, %d edges", nodeCount, edgeCount))

	p.log.Infof("set tree success")
	return nil
}

func (p *ProgramAnalysis) GetProgress() float64 {
	return float64(p.tracker.ProcessedNodes) / float64(p.tracker.TotalNodes)
}

// saveData 异步保存数据到数据库（内部方法）
func (p *ProgramAnalysis) saveData(ctx context.Context, statusChan chan []byte) error {
	wg := sync.WaitGroup{}
	if err := p.data.InitTable(); err != nil {
		return err
	}
	if statusChan != nil {
		statusChan <- []byte("Starting to save data to database...")
	}
	// 启动goroutine处理节点保存
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodeCount := 0
		for node := range p.nodeManager.GetNodeChan() {
			p.log.Infof("save node: %s", node.Key)
			if err := p.data.SaveFuncNode(node); err != nil {
				p.log.Error("save node failed: %v", err, "node", node)
			}
			nodeCount++

			// 每保存10个节点发送一次状态更新
			if statusChan != nil && nodeCount%10 == 0 {
				statusChan <- []byte(fmt.Sprintf("Saved %d nodes", nodeCount))
			}
		}
	}()

	// 启动goroutine处理边保存
	wg.Add(1)
	go func() {
		defer wg.Done()
		edgeCount := 0
		for edge := range p.edgeManager.GetEdgeChan() {
			p.log.Infof("save edge: %s --> %s", edge.CallerKey, edge.CalleeKey)
			if err := p.data.SaveFuncEdge(edge); err != nil {
				p.log.Error("save edge failed: %v", err, "edge", edge)
			}
			edgeCount++

			// 每保存20条边发送一次状态更新
			if statusChan != nil && edgeCount%20 == 0 {
				statusChan <- []byte(fmt.Sprintf("Saved %d edges", edgeCount))
			}
		}
	}()

	wg.Wait()

	if statusChan != nil {
		statusChan <- []byte("Data saving completed")
	}

	p.log.Infof("save data exit")
	return nil
}

// SaveData 异步保存数据到数据库
func (p *ProgramAnalysis) SaveData(ctx context.Context, statusChan chan []byte) error {
	return p.saveData(ctx, statusChan)
}

// SetTree 构建调用图树结构（向后兼容，建议使用Execute方法）
func (p *ProgramAnalysis) SetTree(statusChan chan []byte) error {
	return p.setTree(statusChan)
}

// Configuration option functions

func WithAlgo(algo string) ProgramOption {
	return func(p *ProgramAnalysis) {
		p.algo = algo
	}
}

func WithIgnorePaths(ignorePath string) ProgramOption {
	var ignorePaths []string
	for _, item := range strings.Split(ignorePath, ",") {
		path := strings.TrimSpace(item)
		if path != "" {
			ignorePaths = append(ignorePaths, path)
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

// produceData 生产调用图数据到channels（内部方法）
func (p *ProgramAnalysis) produceData(statusChan chan []byte) error {
	p.log.Info("produce call graph data")

	// 执行分析
	if err := p.Analysis(); err != nil {
		return err
	}

	// 初始化组件
	p.reporter = NewStatusReporter(statusChan)
	p.filter = NewFilter(&FilterConfig{
		IgnorePaths: p.ignorePaths,
		ModuleName:  p.moduleName,
	})

	// 发送分析开始状态
	p.reporter.ReportStatus(fmt.Sprintf("Starting to build call graph, using algorithm: %s", p.algo))

	// 统计信息
	nodeCount := 0
	edgeCount := 0
	p.tracker.TotalNodes = len(p.callGraph.Nodes)

	err := callgraph.GraphVisitEdges(p.callGraph, func(edge *callgraph.Edge) error {
		caller := edge.Caller
		callee := edge.Callee

		callerKey := fmt.Sprintf("n%d", caller.ID)
		if !p.isVisited[callerKey] {
			p.isVisited[callerKey] = true
			p.tracker.ProcessedNodes++
		}

		// 使用过滤器检查是否应该处理这条边
		if !p.filter.ShouldProcessEdge(edge) {
			return nil
		}
		p.log.Infof("caller: %s, callee: %s", caller.String(), callee.String())

		// 处理caller节点
		callerFullName := caller.String()
		callerPkg := caller.Func.Pkg.Pkg.Path()
		callerName := caller.Func.RelString(caller.Func.Pkg.Pkg)
		if !p.nodeManager.NodeExists(callerKey) {
			nodeCount++
			// 每处理10个节点发送一次状态更新
			if nodeCount%NodeStatusUpdateInterval == 0 {
				p.reporter.ReportStatus(fmt.Sprintf("Processed %d nodes, %d edges", nodeCount, edgeCount))
			}
		}
		callerSignature := p.extractSignature(caller.Func)
		callerPosition := p.extractPosition(caller.Func)
		callerSourceCode := p.extractSourceCode(caller.Func)
		callerNode := p.nodeManager.GetOrCreateNode(caller.ID, callerFullName, callerPkg, callerPkg, callerName, callerSignature, callerPosition, callerSourceCode)

		// 处理callee节点
		calleeFullName := callee.String()
		calleePkg := callee.Func.Pkg.Pkg.Path()
		calleeName := callee.Func.RelString(callee.Func.Pkg.Pkg)

		calleeKey := fmt.Sprintf("n%d", callee.ID)
		if !p.nodeManager.NodeExists(calleeKey) {
			nodeCount++
			// 每处理10个节点发送一次状态更新
			if nodeCount%NodeStatusUpdateInterval == 0 {
				p.reporter.ReportStatus(fmt.Sprintf("Processed %d nodes, %d edges", nodeCount, edgeCount))
			}
		}
		calleeSignature := p.extractSignature(callee.Func)
		calleePosition := p.extractPosition(callee.Func)
		calleeSourceCode := p.extractSourceCode(callee.Func)
		calleeNode := p.nodeManager.GetOrCreateNode(callee.ID, calleeFullName, calleePkg, calleePkg, calleeName, calleeSignature, calleePosition, calleeSourceCode)

		// 提取调用点信息
		callSiteFile, callSiteLine, callType := p.extractCallSiteInfo(edge)

		// 建立边关系 - 使用EdgeManager封装逻辑
		p.edgeManager.BuildRelationship(callerNode, calleeNode, callSiteFile, callSiteLine, callType)
		edgeCount++

		// 每处理20条边发送一次状态更新
		if edgeCount%EdgeStatusUpdateInterval == 0 {
			p.reporter.ReportStatus(fmt.Sprintf("Processed %d nodes, %d edges", nodeCount, edgeCount))
		}

		p.log.Infof("set edge caller: %s, --> callee: %s", callerNode.Key, calleeNode.Key)

		return nil
	})

	if err != nil {
		p.reporter.ReportStatus(fmt.Sprintf("Call graph build error: %v", err))
		return err
	}

	// 发送完成状态
	p.reporter.ReportStatus(fmt.Sprintf("Call graph data production completed, processed %d nodes, %d edges", nodeCount, edgeCount))

	p.log.Infof("produce data success")
	return nil
}

// consumeData 消费channels中的数据并保存到数据库（内部方法）
func (p *ProgramAnalysis) consumeData(ctx context.Context, statusChan chan []byte) error {
	p.log.Info("consume call graph data")

	wg := sync.WaitGroup{}
	if statusChan != nil {
		statusChan <- []byte("Starting to save data to database...")
	}

	// 启动goroutine处理节点保存
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodeCount := 0
		for node := range p.nodeManager.GetNodeChan() {
			p.log.Infof("save node: %s", node.Key)
			if err := p.data.SaveFuncNode(node); err != nil {
				p.log.Error("save node failed: %v", err, "node", node)
			}
			nodeCount++

			// 每保存10个节点发送一次状态更新
			if statusChan != nil && nodeCount%10 == 0 {
				statusChan <- []byte(fmt.Sprintf("Saved %d nodes", nodeCount))
			}
		}
	}()

	// 启动goroutine处理边保存
	wg.Add(1)
	go func() {
		defer wg.Done()
		edgeCount := 0
		for edge := range p.edgeManager.GetEdgeChan() {
			p.log.Infof("save edge: %s --> %s", edge.CallerKey, edge.CalleeKey)
			if err := p.data.SaveFuncEdge(edge); err != nil {
				p.log.Error("save edge failed: %v", err, "edge", edge)
			}
			edgeCount++

			// 每保存20条边发送一次状态更新
			if statusChan != nil && edgeCount%20 == 0 {
				statusChan <- []byte(fmt.Sprintf("Saved %d edges", edgeCount))
			}
		}
	}()

	wg.Wait()

	if statusChan != nil {
		statusChan <- []byte("Data saving completed")
	}

	p.log.Infof("consume data success")
	return nil
}

// extractSourceCode 提取函数源代码
func (p *ProgramAnalysis) extractSourceCode(fn *ssa.Function) string {
	if fn == nil || fn.Pos() == 0 {
		return ""
	}

	// 获取函数位置信息
	pos := fn.Prog.Fset.Position(fn.Pos())

	// 只为内部包提取源代码
	if !p.isInternalPackage(fn.Pkg.Pkg.Path()) {
		return ""
	}

	// 读取源文件
	content, err := ioutil.ReadFile(pos.Filename)
	if err != nil {
		p.log.Errorf("read source file failed: %v", err)
		return ""
	}

	// 获取函数结束位置 - 通过AST解析获取精确位置
	endLine := p.getFunctionEndLine(pos.Filename, pos.Line)
	if endLine == 0 {
		// 如果无法通过AST获取，使用启发式方法
		endLine = p.estimateFunctionEndLine(string(content), pos.Line)
	}

	// 按行分割内容
	lines := strings.Split(string(content), "\n")

	// 确保行号在有效范围内
	startLine := pos.Line - 1 // 转换为0基索引
	if startLine < 0 || startLine >= len(lines) {
		return ""
	}

	if endLine > len(lines) {
		endLine = len(lines)
	}

	// 提取函数源代码
	funcLines := lines[startLine:endLine]
	return strings.Join(funcLines, "\n")
}

// extractPosition 提取函数位置信息
func (p *ProgramAnalysis) extractPosition(fn *ssa.Function) *dos.Position {
	if fn == nil || fn.Pos() == 0 {
		return nil
	}

	pos := fn.Prog.Fset.Position(fn.Pos())

	// 通过AST解析获取精确的函数结束位置
	endLine := p.getFunctionEndLine(pos.Filename, pos.Line)
	if endLine == 0 {
		// 如果无法通过AST获取，使用启发式方法
		content, err := ioutil.ReadFile(pos.Filename)
		if err == nil {
			endLine = p.estimateFunctionEndLine(string(content), pos.Line)
		} else {
			endLine = pos.Line // 回退到起始行
		}
	}

	return &dos.Position{
		Filename: pos.Filename,
		Line:     pos.Line,
		EndLine:  endLine,
	}
}

// extractSignature 提取函数签名
func (p *ProgramAnalysis) extractSignature(fn *ssa.Function) string {
	if fn == nil {
		return ""
	}

	// 使用 SSA 函数的字符串表示作为签名
	return fn.String()
}

// extractCallSiteInfo 提取调用点信息
func (p *ProgramAnalysis) extractCallSiteInfo(edge *callgraph.Edge) (string, int, string) {
	if edge == nil || edge.Pos() == 0 {
		return "", 0, "unknown"
	}

	pos := edge.Caller.Func.Prog.Fset.Position(edge.Pos())
	callType := "static" // 默认调用类型

	// 根据调用指令类型确定调用类型
	if edge.Site != nil {
		switch edge.Site.(type) {
		case *ssa.Call:
			callType = "direct"
		case *ssa.Go:
			callType = "goroutine"
		case *ssa.Defer:
			callType = "defer"
		default:
			callType = "indirect"
		}
	}

	return pos.Filename, pos.Line, callType
}

// isInternalPackage 判断是否为内部包
func (p *ProgramAnalysis) isInternalPackage(pkgPath string) bool {
	if pkgPath == "" {
		return false
	}

	// 标准库包识别
	if !strings.Contains(pkgPath, ".") {
		return false // 标准库包不包含点号
	}

	// 基于模块名判断
	return strings.HasPrefix(pkgPath, p.moduleName)
}

// analyzeImports 分析import依赖关系
func (p *ProgramAnalysis) analyzeImports() error {
	p.log.Info("analyzing import dependencies")

	// 确保import分析器已初始化
	if p.importAnalyzer == nil {
		p.importAnalyzer = NewImportAnalyzer(p.moduleName, p.log)
	}

	// 分析项目目录中的import语句
	if err := p.importAnalyzer.AnalyzeDirectory(p.Dir); err != nil {
		return fmt.Errorf("analyze directory imports failed: %w", err)
	}

	p.log.Infof("import analysis completed, found %d packages and %d dependencies",
		len(p.importAnalyzer.GetPackages()),
		len(p.importAnalyzer.GetDependencies()))

	// 分析完成后立即保存packages数据
	if err := p.savePackageData(); err != nil {
		p.log.Errorf("save package data failed: %v", err)
		return fmt.Errorf("save package data failed: %w", err)
	}

	return nil
}

// savePackageData 保存包依赖数据
func (p *ProgramAnalysis) savePackageData() error {
	if p.importAnalyzer == nil {
		p.log.Warn("importAnalyzer is nil, skipping package data save")
		return nil
	}

	p.log.Info("saving package data")

	// 批量保存包信息
	packages := p.importAnalyzer.GetPackages()
	p.log.Infof("got %d packages from analyzer", len(packages))
	if len(packages) > 0 {
		if err := p.data.BatchSavePackages(packages); err != nil {
			p.log.Errorf("batch save packages failed: %v", err)
			return err
		}
		p.log.Infof("successfully saved %d packages", len(packages))
	} else {
		p.log.Warn("no packages to save")
	}

	// 批量保存包依赖关系
	deps := p.importAnalyzer.GetDependencies()
	p.log.Infof("got %d dependencies from analyzer", len(deps))
	if len(deps) > 0 {
		if err := p.data.BatchSavePackageDeps(deps); err != nil {
			p.log.Errorf("batch save package deps failed: %v", err)
			return err
		}
		p.log.Infof("successfully saved %d package dependencies", len(deps))
	} else {
		p.log.Warn("no package dependencies to save")
	}

	return nil
}

// getFunctionEndLine 通过AST解析获取函数的精确结束行号
func (p *ProgramAnalysis) getFunctionEndLine(filename string, startLine int) int {
	// 解析源文件的AST
	fset := token.NewFileSet()
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0
	}

	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return 0
	}

	var endLine int

	// 遍历AST查找匹配起始行的函数声明
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Pos() != 0 {
				funcStartPos := fset.Position(x.Pos())
				if funcStartPos.Line == startLine {
					// 找到匹配的函数，获取结束位置
					funcEndPos := fset.Position(x.End())
					endLine = funcEndPos.Line
					return false // 找到后停止遍历
				}
			}
		}
		return true
	})

	return endLine
}

// estimateFunctionEndLine 使用启发式方法估算函数结束行号
func (p *ProgramAnalysis) estimateFunctionEndLine(content string, startLine int) int {
	lines := strings.Split(content, "\n")
	if startLine <= 0 || startLine > len(lines) {
		return startLine
	}

	// 从函数开始行向下查找
	braceDepth := 0
	foundOpenBrace := false

	for i := startLine - 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}

		// 计算大括号的深度
		for _, char := range line {
			switch char {
			case '{':
				braceDepth++
				foundOpenBrace = true
			case '}':
				braceDepth--
				// 如果找到了开括号且括号深度回到0，说明函数结束
				if foundOpenBrace && braceDepth == 0 {
					return i + 1 // 返回1基索引
				}
			}
		}

		// 如果没有找到开括号且遇到分号，可能是函数声明
		if !foundOpenBrace && strings.HasSuffix(line, ";") {
			return i + 1
		}

		// 限制搜索范围，避免无限循环
		if i-startLine > 500 { // 最多向下搜索500行
			break
		}
	}

	// 如果无法确定结束位置，返回起始行后几行作为估计
	return startLine + 10
}
