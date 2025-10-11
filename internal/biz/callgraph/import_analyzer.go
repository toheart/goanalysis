package callgraph

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
)

// ImportAnalyzer import语句分析器
type ImportAnalyzer struct {
	log        *log.Helper
	moduleName string                     // 当前项目模块名
	packages   map[string]*dos.Package    // 包信息缓存
	deps       map[string]*dos.PackageDep // 依赖关系缓存
}

// NewImportAnalyzer 创建import分析器
func NewImportAnalyzer(moduleName string, log *log.Helper) *ImportAnalyzer {
	return &ImportAnalyzer{
		log:        log,
		moduleName: moduleName,
		packages:   make(map[string]*dos.Package),
		deps:       make(map[string]*dos.PackageDep),
	}
}

// AnalyzeDirectory 分析目录中的所有Go文件的import语句
func (ia *ImportAnalyzer) AnalyzeDirectory(dir string) error {
	ia.log.Infof("analyze imports in directory: %s, module name: %s", dir, ia.moduleName)

	fileCount := 0
	// 遍历目录中的所有Go文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			ia.log.Errorf("walk directory error: %v, path: %s", err, path)
			return err
		}

		// 只处理.go文件，跳过测试文件
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileCount++
		ia.log.Infof("analyzing go file: %s", path)
		return ia.analyzeFile(path)
	})

	ia.log.Infof("analyzed %d go files, found %d packages and %d dependencies",
		fileCount, len(ia.packages), len(ia.deps))

	return err
}

// analyzeFile 分析单个Go文件的import语句
func (ia *ImportAnalyzer) analyzeFile(filePath string) error {
	// 解析Go文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		ia.log.Errorf("parse file failed: %s, error: %v", filePath, err)
		return err
	}

	// 获取当前文件所属的包路径
	currentPkg := ia.getPackagePathFromFile(filePath, node.Name.Name)
	if currentPkg == "" {
		ia.log.Warnf("cannot determine package path for file: %s", filePath)
		return nil
	}

	ia.log.Infof("file %s belongs to package: %s (declared as: %s)", filePath, currentPkg, node.Name.Name)

	// 记录当前包信息
	ia.recordPackage(currentPkg)

	importCount := 0
	// 分析import语句
	for _, imp := range node.Imports {
		if imp.Path == nil {
			continue
		}

		// 获取导入的包路径（去掉引号）
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			ia.log.Warnf("failed to unquote import path: %s", imp.Path.Value)
			continue
		}

		ia.log.Infof("processing import: %s", importPath)

		// 跳过标准库包（不包含点号的包）
		if !strings.Contains(importPath, ".") {
			ia.log.Infof("skipping standard library package: %s", importPath)
			continue
		}

		// 记录被导入的包
		ia.recordPackage(importPath)

		// 记录依赖关系
		ia.recordDependency(currentPkg, importPath)

		ia.log.Infof("found import: %s -> %s", currentPkg, importPath)
		importCount++
	}

	ia.log.Infof("file %s has %d imports processed", filePath, importCount)
	return nil
}

// getPackagePathFromFile 从文件路径推断包路径
// packageName 是从 AST 中解析出的包声明名称（如 "main", "callgraph" 等）
func (ia *ImportAnalyzer) getPackagePathFromFile(filePath string, packageName string) string {
	// 获取相对于项目根目录的路径
	relPath := ia.getRelativePathFromModule(filePath)
	if relPath == "" {
		ia.log.Warnf("failed to get relative path for file: %s", filePath)
		return ""
	}

	ia.log.Infof("relative path for %s: %s", filePath, relPath)

	// 去掉文件名，只保留目录路径
	dir := filepath.Dir(relPath)
	if dir == "." {
		// 文件在根目录
		if packageName == "main" {
			// main 包在根目录，使用模块名作为包路径
			ia.log.Infof("main package in root directory, using module name: %s", ia.moduleName)
			return ia.moduleName
		}
		// 其他包在根目录，也使用模块名
		ia.log.Infof("package %s in root directory, using module name: %s", packageName, ia.moduleName)
		return ia.moduleName
	}

	// 组合成完整包路径
	result := filepath.Join(ia.moduleName, dir)
	ia.log.Infof("computed package path: %s for package name: %s", result, packageName)
	return result
}

// getRelativePathFromModule 获取相对于模块根目录的路径
func (ia *ImportAnalyzer) getRelativePathFromModule(filePath string) string {
	// 这里需要找到项目根目录（包含go.mod的目录）
	// 简化实现：假设当前工作目录就是项目根目录
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		ia.log.Errorf("failed to get absolute path for %s: %v", filePath, err)
		return ""
	}

	ia.log.Infof("absolute path: %s", absPath)

	// 查找go.mod文件所在目录
	moduleRoot := ia.findModuleRoot(absPath)
	if moduleRoot == "" {
		ia.log.Warnf("no module root found for path: %s", absPath)
		return ""
	}

	ia.log.Infof("module root: %s", moduleRoot)

	// 获取相对路径
	relPath, err := filepath.Rel(moduleRoot, absPath)
	if err != nil {
		ia.log.Errorf("failed to get relative path from %s to %s: %v", moduleRoot, absPath, err)
		return ""
	}

	result := filepath.ToSlash(relPath) // 转换为Unix风格路径
	ia.log.Infof("final relative path: %s", result)
	return result
}

// findModuleRoot 查找模块根目录（包含go.mod的目录）
func (ia *ImportAnalyzer) findModuleRoot(startPath string) string {
	dir := filepath.Dir(startPath)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// 已到达根目录
			break
		}
		dir = parent
	}
	return ""
}

// recordPackage 记录包信息
func (ia *ImportAnalyzer) recordPackage(pkgPath string) {
	if _, exists := ia.packages[pkgPath]; exists {
		ia.log.Infof("package already recorded: %s", pkgPath)
		return
	}

	isInternal := ia.isInternalPackage(pkgPath)
	pkg := &dos.Package{
		PkgPath:    pkgPath,
		IsInternal: isInternal,
	}

	ia.packages[pkgPath] = pkg
	ia.log.Infof("recorded package: %s (internal: %v)", pkgPath, isInternal)
}

// recordDependency 记录依赖关系
func (ia *ImportAnalyzer) recordDependency(importerPkg, importedPkg string) {
	key := importerPkg + "->" + importedPkg

	if existing, exists := ia.deps[key]; exists {
		// 增加导入次数
		existing.ImportCount++
		ia.log.Infof("updated dependency count: %s (count: %d)", key, existing.ImportCount)
	} else {
		// 创建新的依赖关系
		dep := &dos.PackageDep{
			ImporterPkg: importerPkg,
			ImportedPkg: importedPkg,
			ImportCount: 1,
		}
		ia.deps[key] = dep
		ia.log.Infof("recorded new dependency: %s", key)
	}
}

// isInternalPackage 判断是否为内部包
func (ia *ImportAnalyzer) isInternalPackage(pkgPath string) bool {
	if pkgPath == "" {
		return false
	}

	// 标准库包识别（不包含点号）
	if !strings.Contains(pkgPath, ".") {
		return false
	}

	// 基于模块名判断
	return strings.HasPrefix(pkgPath, ia.moduleName)
}

// GetPackages 获取收集到的包信息
func (ia *ImportAnalyzer) GetPackages() []*dos.Package {
	packages := make([]*dos.Package, 0, len(ia.packages))
	for _, pkg := range ia.packages {
		packages = append(packages, pkg)
	}
	return packages
}

// GetDependencies 获取收集到的依赖关系
func (ia *ImportAnalyzer) GetDependencies() []*dos.PackageDep {
	deps := make([]*dos.PackageDep, 0, len(ia.deps))
	for _, dep := range ia.deps {
		deps = append(deps, dep)
	}
	return deps
}

// Clear 清空缓存
func (ia *ImportAnalyzer) Clear() {
	ia.packages = make(map[string]*dos.Package)
	ia.deps = make(map[string]*dos.PackageDep)
}
