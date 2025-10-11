package repo

import (
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
)

// 添加用于存储的方法
type StaticDBStore interface {
	// SaveFuncNode 保存函数节点
	SaveFuncNode(node *dos.FuncNode) error

	// SaveFuncEdge 保存函数调用关系
	SaveFuncEdge(edge *dos.FuncEdge) error

	// BatchSaveFuncNodes 批量保存函数节点
	BatchSaveFuncNodes(nodes []*dos.FuncNode) error

	// BatchSaveFuncEdges 批量保存函数调用关系
	BatchSaveFuncEdges(edges []*dos.FuncEdge) error

	// GetFuncNodeByKey 根据Key获取函数节点
	GetFuncNodeByKey(key string) (*dos.FuncNode, error)

	// GetCallerEdges 获取调用该函数的所有边
	GetCallerEdges(calleeKey string) ([]*dos.FuncNode, error)

	// GetCalleeEdges 获取该函数调用的所有边
	GetCalleeEdges(callerKey string) ([]*dos.FuncNode, error)

	// GetAllFuncNodes 获取所有函数节点
	GetAllFuncNodes() ([]*dos.FuncNode, error)

	// GetAllFuncEdges 获取所有函数调用边
	GetAllFuncEdges() ([]*dos.FuncEdge, error)

	// SearchFuncNodes 模糊搜索函数节点
	SearchFuncNodes(query string, limit int) ([]*dos.FuncNode, error)

	// InitTable 初始化数据库表
	InitTable() error

	// === 包级查询方法 - 简化设计 ===

	// SavePackage 保存包信息
	SavePackage(pkg *dos.Package) error

	// SavePackageDep 保存包依赖关系
	SavePackageDep(dep *dos.PackageDep) error

	// BatchSavePackages 批量保存包信息
	BatchSavePackages(pkgs []*dos.Package) error

	// BatchSavePackageDeps 批量保存包依赖关系
	BatchSavePackageDeps(deps []*dos.PackageDep) error

	// GetPackageByPath 根据包路径获取包信息
	GetPackageByPath(pkgPath string) (*dos.Package, error)

	// ListImportedBy 获取指定包导入的所有包
	ListImportedBy(pkgPath string) ([]*dos.Package, error)

	// ListImportersOf 获取导入指定包的所有包
	ListImportersOf(pkgPath string) ([]*dos.Package, error)

	// GetPackageDeps 获取包依赖关系
	GetPackageDeps(importerPkg string) ([]*dos.PackageDep, error)

	// SearchPackages 模糊搜索包
	SearchPackages(query string, limit int) ([]*dos.Package, error)

	// GetAllPackages 获取所有包
	GetAllPackages() ([]*dos.Package, error)

	// GetAllPackageDeps 获取所有包依赖关系
	GetAllPackageDeps() ([]*dos.PackageDep, error)
}

// PackageQueryFilter 包查询过滤器
type PackageQueryFilter struct {
	InternalOnly bool // 仅查询内部包
	ExternalOnly bool // 仅查询外部包
}
