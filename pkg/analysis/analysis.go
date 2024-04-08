package analysis

import (
	"encoding/json"
	"fmt"
	"github.com/goccy/go-graphviz"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"io"
	"log"
	"os"
	"strings"
)

/**
@file:
@author: levi.Tang
@time: 2024/3/20 10:43
@description: 用于分析程序包, 整体逻辑
**/

type ProgramOption func(p *ProgramAnalysis)

type ProgramAnalysis struct {
	algo        string
	Dir         string
	ignorePaths []string
	onlyMethod  string

	cachePath  string
	isCache    bool
	outputPath string

	tree      *ProgramAst
	callGraph *callgraph.Graph
}

func NewProgramAnalysis(dir string, opts ...ProgramOption) *ProgramAnalysis {
	p := &ProgramAnalysis{
		Dir:  dir,
		algo: CallGraphTypeVta,
		tree: NewProgramAst(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *ProgramAnalysis) GetMainPackage(pkgs []*ssa.Package) ([]*ssa.Package, error) {
	var mains []*ssa.Package
	for _, p := range pkgs {
		if p != nil && p.Pkg.Name() == "main" && p.Func("main") != nil {
			mains = append(mains, p)
		}
	}
	if len(mains) == 0 {
		return nil, fmt.Errorf("no main packages")
	}
	return mains, nil
}

func (p *ProgramAnalysis) LoadCache() error {
	var err error
	defer func() {
		//错误时，表明文件出现问题, 删除文件
		if err != nil {
			if err := os.RemoveAll(p.cachePath); err != nil {
				fmt.Println("delete cachePath failed", err)
			}
		}
	}()
	// 判断是否使用缓存
	// 打开文件
	file, err := os.Open(p.cachePath)
	if err != nil {
		fmt.Println("open file err", err)
		return err
	}
	defer file.Close()

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("can't read file content", err)
		return err
	}

	// 将内容转换为 JSON 格式
	if err := json.Unmarshal(data, p.tree); err != nil {
		fmt.Println("json unmarshal found err", err)
		return err
	}
	return err
}

func (p *ProgramAnalysis) Analysis() error {
	// 1. 加载项目代码所有的package名称
	initial, _ := packages.Load(&packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedTypesSizes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Tests: false,
		Dir:   p.Dir,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			// 忽略测试文件
			if strings.HasSuffix(filename, "_test.go") {
				return nil, nil
			}
			return parser.ParseFile(fset, filename, src, parser.ParseComments)
		},
	}, p.Dir+"/...")
	if packages.PrintErrors(initial) > 0 {
		return fmt.Errorf("packages contain errors")
	}
	// 2. 基于指定的package名称，创建SSA项目（包含所有引用的包）
	prog, _ := ssautil.AllPackages(initial, 0)
	prog.Build()

	switch p.algo {
	case CallGraphTypeStatic:
		p.callGraph = static.CallGraph(prog)
	case CallGraphTypeCha:
		p.callGraph = cha.CallGraph(prog)
	case CallGraphTypeRta:
		mains, err := p.GetMainPackage(prog.AllPackages())
		if err != nil {
			return err
		}
		var roots []*ssa.Function
		for _, main := range mains {
			roots = append(roots, main.Func("main"))
		}

		p.callGraph = rta.Analyze(roots, true).CallGraph
	case CallGraphTypeVta:
		p.callGraph = vta.CallGraph(ssautil.AllFunctions(prog), cha.CallGraph(prog))
	default:
		return fmt.Errorf("invalid call graph type: %s", p.algo)
	}

	return nil
}

func (p *ProgramAnalysis) SaveCache() error {
	// 不管使不使用都输出文件
	// 打开（创建）一个文件用于写入
	file, err := os.Create(p.cachePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	// 使用json编码器将Person对象编码为JSON，并写入文件
	encoder := json.NewEncoder(file)
	err = encoder.Encode(p.tree)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return err
	}

	fmt.Printf("JSON data has been written to <%s> \n", p.cachePath)
	return nil
}

// SetTree
//
//	@Description: 生成项目整体的语法树
//	@receiver p
//	@return error
func (p *ProgramAnalysis) SetTree() error {
	// 加载内存
	if p.isCache {
		if err := p.LoadCache(); err == nil {
			fmt.Printf("use cache file: <%s> \n", p.cachePath)
			return nil
		}
	}
	if err := p.Analysis(); err != nil {
		return err
	}

	var isInter = func(edge *callgraph.Edge) bool {
		//caller := edge.Caller
		callee := edge.Callee
		if callee.Func.Object() != nil && !callee.Func.Object().Exported() {
			return true
		}
		return false
	}

	var inIgnores = func(node *callgraph.Node) bool {
		pkgPath := node.Func.String()
		for _, p := range p.ignorePaths {
			if strings.HasPrefix(pkgPath, p) {
				return true
			}
		}
		return false
	}

	err := callgraph.GraphVisitEdges(p.callGraph, func(edge *callgraph.Edge) error {
		caller := edge.Caller
		callee := edge.Callee

		if isSynthetic(edge) {
			return nil
		}
		// 排除标准库
		if inStd(caller) || inStd(callee) {
			return nil
		}

		if isInter(edge) {
			return nil
		}
		if inIgnores(caller) || inIgnores(callee) {
			return nil
		}
		// caller是否存在
		var pNode, qNode *FuncNode
		var ok bool
		// 如果不存在, 则创建
		if pNode, ok = p.tree.Nodes[caller.String()]; !ok {
			pNode = &FuncNode{
				Key:  caller.String(),
				Pkg:  caller.Func.Pkg.Pkg.Path(),
				Name: caller.Func.RelString(caller.Func.Pkg.Pkg),
			}
			p.tree.Nodes[pNode.Key] = pNode
		}
		if qNode, ok = p.tree.Nodes[callee.String()]; !ok {
			qNode = &FuncNode{
				Key:  callee.String(),
				Pkg:  callee.Func.Pkg.Pkg.Path(),
				Name: callee.Func.RelString(callee.Func.Pkg.Pkg),
			}
			p.tree.Nodes[qNode.Key] = qNode
		}

		if strings.HasSuffix(caller.String(), "main") && p.tree.MainKey == "" {
			p.tree.MainKey = caller.String()
		}
		pNode.Children = append(pNode.Children, qNode.Key)
		qNode.Parent = append(qNode.Parent, pNode.Key)
		fmt.Printf("%s to %s \n", caller, callee)
		return nil
	})
	err = p.SaveCache()
	return err
}

func (p *ProgramAnalysis) Print() {
	if err := p.SetTree(); err != nil {
		fmt.Println(err)
		return
	}
	if p.onlyMethod == "" {
		p.printPng()
		return
	}
	// 如果只想找某个方法的上下游关系
	p.printOnlyMethod()
}

func (p *ProgramAnalysis) printOnlyMethod() {
	if p.onlyMethod == "" {
		fmt.Println("no input onlyMethod")
		return
	}
	method, ok := p.tree.Nodes[p.onlyMethod]
	if !ok {
		fmt.Printf("method: <%s> not found, please open cache<%s> verify \n", p.onlyMethod, p.cachePath)
		return
	}
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()
	p.tree.ParentToPng(method, nil, graph)
	// 重置访问节点
	p.tree.isVisited = make(map[string]struct{})
	p.tree.ChildrenToPng(method, nil, graph)
	// 1. write to file directly
	if err := g.RenderFilename(graph, graphviz.PNG, p.outputPath); err != nil {
		log.Fatal(err)
	}
}

func (p *ProgramAnalysis) printPng() {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()
	main := p.tree.Nodes[p.tree.MainKey]
	p.tree.ChildrenToPng(main, nil, graph)
	// init函数的遍历
	for key, value := range p.tree.Nodes {
		// 排除已遍历程序
		if _, ok := p.tree.isVisited[key]; ok {
			continue
		}
		// 最后以为以init开始时
		keySpl := strings.Split(key, ".")
		initKey := keySpl[len(keySpl)-1]
		if !strings.HasPrefix(initKey, "init") {
			continue
		}
		p.tree.ChildrenToPng(value, nil, graph)
	}
	// 1. write to file directly
	if err := g.RenderFilename(graph, graphviz.PNG, p.outputPath); err != nil {
		log.Fatal(err)
	}
}

func isSynthetic(edge *callgraph.Edge) bool {
	return edge.Caller.Func.Pkg == nil || edge.Callee.Func.Synthetic != ""
}

func inStd(node *callgraph.Node) bool {
	pkg, _ := build.Import(node.Func.Pkg.Pkg.Path(), "", 0)
	return pkg.Goroot
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
