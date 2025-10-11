package sqlite

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"github.com/klauspost/compress/zstd"
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
	"github.com/toheart/goanalysis/internal/biz/repo"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen/funcedge"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen/funcnode"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen/packagedep"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen/pkg"

	// 注册 SQLite 驱动
	_ "modernc.org/sqlite"
)

var _ repo.StaticDBStore = (*StaticEntDBImpl)(nil)

// compressSourceCode 使用 zstd 压缩源代码
func compressSourceCode(sourceCode string) (string, error) {
	if sourceCode == "" {
		return "", nil
	}

	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return "", fmt.Errorf("create zstd encoder failed: %w", err)
	}
	defer encoder.Close()

	compressed := encoder.EncodeAll([]byte(sourceCode), make([]byte, 0, len(sourceCode)))
	return string(compressed), nil
}

// decompressSourceCode 使用 zstd 解压缩源代码
func decompressSourceCode(compressedData string) (string, error) {
	if compressedData == "" {
		return "", nil
	}

	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return "", fmt.Errorf("create zstd decoder failed: %w", err)
	}
	defer decoder.Close()

	decompressed, err := decoder.DecodeAll([]byte(compressedData), nil)
	if err != nil {
		return "", fmt.Errorf("decompress source code failed: %w", err)
	}

	return string(decompressed), nil
}

// isInternalPackage 判断是否为内部包（需要保存源代码）
func (s *StaticEntDBImpl) isInternalPackage(pkgPath string) bool {
	if pkgPath == "" {
		return false
	}

	// 标准库包识别
	if !strings.Contains(pkgPath, ".") {
		return false // 标准库包不包含点号
	}

	// 获取当前项目的模块路径
	modulePath, err := s.getProjectModulePath()
	if err != nil {
		// 如果无法获取模块路径，使用启发式判断（推断目标项目）
		return s.inferTargetProject(pkgPath)
	}

	// 基于模块路径判断
	isInternal := strings.HasPrefix(pkgPath, modulePath)

	// 如果当前模块路径不匹配（跨项目分析），推断目标项目
	if !isInternal {
		return s.inferTargetProject(pkgPath)
	}

	return isInternal
}

// inferTargetProject 推断被分析的目标项目（用于跨项目分析）
func (s *StaticEntDBImpl) inferTargetProject(pkgPath string) bool {
	// 从已保存的函数节点中推断目标项目
	targetProjectPrefixes := s.getTargetProjectPrefixes()

	for _, prefix := range targetProjectPrefixes {
		if strings.HasPrefix(pkgPath, prefix) {
			return true
		}
	}

	return false
}

// getTargetProjectPrefixes 获取目标项目前缀（从数据库中推断）
func (s *StaticEntDBImpl) getTargetProjectPrefixes() []string {
	if s.client == nil {
		// 如果客户端未初始化，返回空数组
		return []string{}
	}

	ctx := context.Background()

	// 查询数据库中的包路径，找出最常见的项目前缀
	rows, err := s.client.FuncNode.
		Query().
		Select(funcnode.FieldPkg).
		All(ctx)

	if err != nil {
		// 如果查询失败，返回空数组，避免硬编码
		return []string{}
	}

	// 统计项目前缀
	projectCounts := make(map[string]int)
	for _, row := range rows {
		pkg := row.Pkg
		if strings.HasPrefix(pkg, "github.com/") || strings.HasPrefix(pkg, "gitlab.com/") || strings.HasPrefix(pkg, "bitbucket.org/") {
			parts := strings.Split(pkg, "/")
			if len(parts) >= 3 {
				// 提取项目前缀，如 github.com/user/project
				projectPrefix := strings.Join(parts[:3], "/")
				projectCounts[projectPrefix]++
			}
		}
	}

	if len(projectCounts) == 0 {
		return []string{}
	}

	// 找出出现频率最高的项目（可能是目标项目）
	var result []string
	maxCount := 0
	for prefix, count := range projectCounts {
		if count > maxCount {
			maxCount = count
			result = []string{prefix} // 重置为当前最高频的项目
		} else if count == maxCount {
			result = append(result, prefix) // 添加同样高频的项目
		}
	}

	return result
}

// getProjectModulePath 获取项目模块路径
func (s *StaticEntDBImpl) getProjectModulePath() (string, error) {
	return getModuleName()
}

// IsInternalPackage 公开方法用于测试
func (s *StaticEntDBImpl) IsInternalPackage(pkgPath string) bool {
	return s.isInternalPackage(pkgPath)
}

// GetProjectModulePath 公开方法用于测试
func (s *StaticEntDBImpl) GetProjectModulePath() (string, error) {
	return s.getProjectModulePath()
}

// GetTargetProjectPrefixes 公开方法用于测试
func (s *StaticEntDBImpl) GetTargetProjectPrefixes() []string {
	return s.getTargetProjectPrefixes()
}

// getModuleName 从go.mod文件中获取模块名
func getModuleName() (string, error) {
	// 从当前目录开始向上查找go.mod文件
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

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
				return "", fmt.Errorf("failed to read go.mod: %w", err)
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

// StaticEntDBImpl 使用 Ent 框架的静态分析数据库实现
type StaticEntDBImpl struct {
	client *gen.Client
}

// NewStaticEntDBImpl 创建函数节点数据库（使用 Ent 框架）
func NewStaticEntDBImpl(dbPath string) (*StaticEntDBImpl, error) {
	// 使用 ent 支持的 SQLite 创建方式
	client, err := gen.Open(dialect.SQLite, ParseDBPath(dbPath))
	if err != nil {
		return nil, fmt.Errorf("create ent client failed: %w", err)
	}

	return &StaticEntDBImpl{client: client}, nil
}

// InitTable 初始化数据库表
func (s *StaticEntDBImpl) InitTable() error {
	ctx := context.Background()
	// 自动创建表结构
	if err := s.client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("创建表结构失败: %w", err)
	}
	return nil
}

// SaveFuncNode 保存函数节点
func (s *StaticEntDBImpl) SaveFuncNode(node *dos.FuncNode) error {
	ctx := context.Background()

	// 检查节点是否已存在
	exists, err := s.client.FuncNode.
		Query().
		Where(funcnode.Key(node.Key)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("检查函数节点是否存在失败: %w", err)
	}

	// 处理源代码压缩（仅针对内部包）
	var compressedSourceCode string
	if s.isInternalPackage(node.PkgPath) && node.SourceCode != "" {
		compressedSourceCode, err = compressSourceCode(node.SourceCode)
		if err != nil {
			return fmt.Errorf("压缩源代码失败: %w", err)
		}
	}

	if exists {
		// 更新节点
		upd := s.client.FuncNode.
			Update().
			Where(funcnode.Key(node.Key)).
			SetFullName(node.FullName).
			SetPkg(node.Pkg).
			SetName(node.Name).
			SetSignature(node.Signature).
			SetSourceCode(compressedSourceCode)
		if node.PkgPath != "" {
			upd = upd.SetPkgPath(node.PkgPath)
		}
		if node.Position != nil {
			upd = upd.
				SetFilename(node.Position.Filename).
				SetStartLine(node.Position.Line).
				SetEndLine(node.Position.EndLine)
		}
		_, err = upd.Save(ctx)
	} else {
		// 创建节点
		crt := s.client.FuncNode.
			Create().
			SetKey(node.Key).
			SetFullName(node.FullName).
			SetPkg(node.Pkg).
			SetName(node.Name).
			SetSignature(node.Signature).
			SetSourceCode(compressedSourceCode)
		if node.PkgPath != "" {
			crt = crt.SetPkgPath(node.PkgPath)
		}
		if node.Position != nil {
			crt = crt.
				SetFilename(node.Position.Filename).
				SetStartLine(node.Position.Line).
				SetEndLine(node.Position.EndLine)
		}
		_, err = crt.Save(ctx)
	}

	if err != nil {
		return fmt.Errorf("保存函数节点失败: %w", err)
	}

	return nil
}

// SaveFuncEdge 保存函数调用关系
func (s *StaticEntDBImpl) SaveFuncEdge(edge *dos.FuncEdge) error {
	ctx := context.Background()

	_, err := s.client.FuncEdge.Create().
		SetCallerKey(edge.CallerKey).
		SetCalleeKey(edge.CalleeKey).
		SetCallSiteFile(edge.CallSiteFile).
		SetCallSiteLine(edge.CallSiteLine).
		SetCallType(edge.CallType).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("save func edge failed: %w", err)
	}
	return nil
}

// GetFuncNodeByKey 根据Key获取函数节点
func (s *StaticEntDBImpl) GetFuncNodeByKey(key string) (*dos.FuncNode, error) {
	ctx := context.Background()

	// 查询函数节点
	funcEnt, err := s.client.FuncNode.
		Query().
		Where(funcnode.Key(key)).
		Only(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return nil, nil // 节点不存在
		}
		return nil, fmt.Errorf("查询函数节点失败: %w", err)
	}

	// 解压缩源代码
	sourceCode, err := decompressSourceCode(funcEnt.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("解压缩源代码失败: %w", err)
	}

	// 转换为业务实体
	node := &dos.FuncNode{
		Key:        funcEnt.Key,
		FullName:   funcEnt.FullName,
		Pkg:        funcEnt.Pkg,
		Name:       funcEnt.Name,
		Signature:  funcEnt.Signature,
		SourceCode: sourceCode,
		Position: &dos.Position{
			Filename: funcEnt.Filename,
			Line:     funcEnt.StartLine,
			EndLine:  funcEnt.EndLine,
		},
	}

	// 获取父节点
	parents, err := s.GetCallerEdges(key)
	if err == nil {
		_ = parents // dos.FuncNode 无对应字段，若后续需要可在服务层使用
	}

	// 获取子节点
	children, err := s.GetCalleeEdges(key)
	if err == nil {
		_ = children // dos.FuncNode 无对应字段
	}

	return node, nil
}

// GetCallerEdges 获取调用该函数的所有节点
func (s *StaticEntDBImpl) GetCallerEdges(calleeKey string) ([]*dos.FuncNode, error) {
	ctx := context.Background()

	// 查询调用该函数的节点
	callers, err := s.client.FuncEdge.
		Query().
		Where(funcedge.CalleeKey(calleeKey)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get caller edges failed: %w", err)
	}

	// 转换为业务实体
	var nodes []*dos.FuncNode
	for _, caller := range callers {
		funcNode, err := s.client.FuncNode.
			Query().
			Where(funcnode.Key(caller.CallerKey)).
			Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("get caller node failed: %w", err)
		}

		// 解压缩源代码
		sourceCode, _ := decompressSourceCode(funcNode.SourceCode)

		node := &dos.FuncNode{
			Key:        funcNode.Key,
			FullName:   funcNode.FullName,
			Pkg:        funcNode.Pkg,
			Name:       funcNode.Name,
			Signature:  funcNode.Signature,
			SourceCode: sourceCode,
			Position: &dos.Position{
				Filename: funcNode.Filename,
				Line:     funcNode.StartLine,
				EndLine:  funcNode.EndLine,
			},
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetCalleeEdges 获取该函数调用的所有节点
func (s *StaticEntDBImpl) GetCalleeEdges(callerKey string) ([]*dos.FuncNode, error) {
	ctx := context.Background()

	// 查询调用该函数的节点
	callers, err := s.client.FuncEdge.
		Query().
		Where(funcedge.CallerKey(callerKey)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get caller edges failed: %w", err)
	}

	// 转换为业务实体
	var nodes []*dos.FuncNode
	for _, caller := range callers {
		funcNode, err := s.client.FuncNode.
			Query().
			Where(funcnode.Key(caller.CalleeKey)).
			Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("get caller node failed: %w", err)
		}

		// 解压缩源代码
		sourceCode, _ := decompressSourceCode(funcNode.SourceCode)

		node := &dos.FuncNode{
			Key:        funcNode.Key,
			FullName:   funcNode.FullName,
			Pkg:        funcNode.Pkg,
			Name:       funcNode.Name,
			Signature:  funcNode.Signature,
			SourceCode: sourceCode,
			Position: &dos.Position{
				Filename: funcNode.Filename,
				Line:     funcNode.StartLine,
				EndLine:  funcNode.EndLine,
			},
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetAllFuncNodes 获取所有函数节点
func (s *StaticEntDBImpl) GetAllFuncNodes() ([]*dos.FuncNode, error) {
	ctx := context.Background()

	// 查询所有函数节点
	funcEnts, err := s.client.FuncNode.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询所有函数节点失败: %w", err)
	}

	// 转换为业务实体
	var nodes []*dos.FuncNode
	for _, funcEnt := range funcEnts {
		// 解压缩源代码
		sourceCode, _ := decompressSourceCode(funcEnt.SourceCode)

		node := &dos.FuncNode{
			Key:        funcEnt.Key,
			FullName:   funcEnt.FullName,
			Pkg:        funcEnt.Pkg,
			Name:       funcEnt.Name,
			Signature:  funcEnt.Signature,
			SourceCode: sourceCode,
			Position: &dos.Position{
				Filename: funcEnt.Filename,
				Line:     funcEnt.StartLine,
				EndLine:  funcEnt.EndLine,
			},
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetAllFuncEdges 获取所有函数调用边
func (s *StaticEntDBImpl) GetAllFuncEdges() ([]*dos.FuncEdge, error) {
	ctx := context.Background()

	// 查询所有函数调用边
	funcEdges, err := s.client.FuncEdge.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询所有函数调用边失败: %w", err)
	}

	// 转换为业务实体
	var edges []*dos.FuncEdge
	for _, funcEdge := range funcEdges {
		edge := &dos.FuncEdge{
			CallerKey:    funcEdge.CallerKey,
			CalleeKey:    funcEdge.CalleeKey,
			CallSiteFile: funcEdge.CallSiteFile,
			CallSiteLine: funcEdge.CallSiteLine,
			CallType:     funcEdge.CallType,
		}
		edges = append(edges, edge)
	}

	return edges, nil
}

// SearchFuncNodes 模糊搜索函数节点
func (s *StaticEntDBImpl) SearchFuncNodes(query string, limit int) ([]*dos.FuncNode, error) {
	ctx := context.Background()

	// 构建模糊查询条件
	funcEnts, err := s.client.FuncNode.
		Query().
		Where(
			funcnode.Or(
				funcnode.NameContainsFold(query), // 函数名模糊匹配（不区分大小写）
				funcnode.PkgContainsFold(query),  // 包名模糊匹配（不区分大小写）
			),
		).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("模糊搜索函数节点失败: %w", err)
	}

	// 转换为业务实体
	var nodes []*dos.FuncNode
	for _, funcEnt := range funcEnts {
		// 解压缩源代码
		sourceCode, _ := decompressSourceCode(funcEnt.SourceCode)

		node := &dos.FuncNode{
			Key:        funcEnt.Key,
			FullName:   funcEnt.FullName,
			Pkg:        funcEnt.Pkg,
			Name:       funcEnt.Name,
			Signature:  funcEnt.Signature,
			SourceCode: sourceCode,
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// BatchSaveFuncNodes 批量保存函数节点
func (s *StaticEntDBImpl) BatchSaveFuncNodes(nodes []*dos.FuncNode) error {
	if len(nodes) == 0 {
		return nil
	}

	ctx := context.Background()

	// 批量创建
	bulk := make([]*gen.FuncNodeCreate, len(nodes))
	for i, node := range nodes {
		// 处理源代码压缩（仅针对内部包）
		var compressedSourceCode string
		if s.isInternalPackage(node.PkgPath) && node.SourceCode != "" {
			var err error
			compressedSourceCode, err = compressSourceCode(node.SourceCode)
			if err != nil {
				return fmt.Errorf("批量保存时压缩源代码失败: %w", err)
			}
		}

		creator := s.client.FuncNode.Create().
			SetKey(node.Key).
			SetFullName(node.FullName).
			SetPkg(node.Pkg).
			SetName(node.Name).
			SetSignature(node.Signature).
			SetSourceCode(compressedSourceCode)

		if node.PkgPath != "" {
			creator = creator.SetPkgPath(node.PkgPath)
		}

		if node.Position != nil {
			creator = creator.
				SetFilename(node.Position.Filename).
				SetStartLine(node.Position.Line).
				SetEndLine(node.Position.EndLine)
		}

		bulk[i] = creator
	}

	// 执行批量插入
	_, err := s.client.FuncNode.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		return fmt.Errorf("批量保存函数节点失败: %w", err)
	}

	return nil
}

// BatchSaveFuncEdges 批量保存函数调用关系
func (s *StaticEntDBImpl) BatchSaveFuncEdges(edges []*dos.FuncEdge) error {
	if len(edges) == 0 {
		return nil
	}

	ctx := context.Background()

	// 批量创建
	bulk := make([]*gen.FuncEdgeCreate, len(edges))
	for i, edge := range edges {
		bulk[i] = s.client.FuncEdge.Create().
			SetCallerKey(edge.CallerKey).
			SetCalleeKey(edge.CalleeKey).
			SetCallSiteFile(edge.CallSiteFile).
			SetCallSiteLine(edge.CallSiteLine).
			SetCallType(edge.CallType).
			SetCreatedAt(time.Now()).
			SetUpdatedAt(time.Now())
	}

	// 执行批量插入
	_, err := s.client.FuncEdge.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		return fmt.Errorf("批量保存函数调用关系失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (s *StaticEntDBImpl) Close() error {
	return s.client.Close()
}

// === 包级查询方法实现 - 简化设计 ===

// SavePackage 保存包信息
func (s *StaticEntDBImpl) SavePackage(packageInfo *dos.Package) error {
	ctx := context.Background()

	// 检查包是否已存在
	exists, err := s.client.Pkg.
		Query().
		Where(pkg.PkgPath(packageInfo.PkgPath)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("check package exists failed: %w", err)
	}

	if exists {
		// 更新包信息
		_, err = s.client.Pkg.
			Update().
			Where(pkg.PkgPath(packageInfo.PkgPath)).
			SetIsInternal(packageInfo.IsInternal).
			Save(ctx)
	} else {
		// 创建包
		_, err = s.client.Pkg.
			Create().
			SetPkgPath(packageInfo.PkgPath).
			SetIsInternal(packageInfo.IsInternal).
			Save(ctx)
	}

	if err != nil {
		return fmt.Errorf("save package failed: %w", err)
	}

	return nil
}

// SavePackageDep 保存包依赖关系
func (s *StaticEntDBImpl) SavePackageDep(dep *dos.PackageDep) error {
	ctx := context.Background()

	// 检查依赖关系是否已存在
	exists, err := s.client.PackageDep.
		Query().
		Where(
			packagedep.ImporterPkg(dep.ImporterPkg),
			packagedep.ImportedPkg(dep.ImportedPkg),
		).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("check package dep exists failed: %w", err)
	}

	if exists {
		// 更新依赖关系（累加导入次数）
		existing, err := s.client.PackageDep.
			Query().
			Where(
				packagedep.ImporterPkg(dep.ImporterPkg),
				packagedep.ImportedPkg(dep.ImportedPkg),
			).
			Only(ctx)
		if err != nil {
			return fmt.Errorf("query existing package dep failed: %w", err)
		}

		_, err = s.client.PackageDep.
			Update().
			Where(
				packagedep.ImporterPkg(dep.ImporterPkg),
				packagedep.ImportedPkg(dep.ImportedPkg),
			).
			SetImportCount(existing.ImportCount + dep.ImportCount).
			Save(ctx)
	} else {
		// 创建依赖关系
		_, err = s.client.PackageDep.
			Create().
			SetImporterPkg(dep.ImporterPkg).
			SetImportedPkg(dep.ImportedPkg).
			SetImportCount(dep.ImportCount).
			Save(ctx)
	}

	if err != nil {
		return fmt.Errorf("save package dep failed: %w", err)
	}

	return nil
}

// BatchSavePackages 批量保存包信息
func (s *StaticEntDBImpl) BatchSavePackages(pkgs []*dos.Package) error {
	if len(pkgs) == 0 {
		return nil
	}

	// 检查客户端是否已初始化
	if s.client == nil {
		return fmt.Errorf("database client is not initialized")
	}

	ctx := context.Background()

	// 批量创建
	bulk := make([]*gen.PkgCreate, len(pkgs))
	for i, packageInfo := range pkgs {
		bulk[i] = s.client.Pkg.Create().
			SetPkgPath(packageInfo.PkgPath).
			SetIsInternal(packageInfo.IsInternal)
	}

	// 执行批量插入
	_, err := s.client.Pkg.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		return fmt.Errorf("batch save packages failed: %w", err)
	}

	return nil
}

// BatchSavePackageDeps 批量保存包依赖关系
func (s *StaticEntDBImpl) BatchSavePackageDeps(deps []*dos.PackageDep) error {
	if len(deps) == 0 {
		return nil
	}

	// 检查客户端是否已初始化
	if s.client == nil {
		return fmt.Errorf("database client is not initialized")
	}

	ctx := context.Background()

	// 批量创建
	bulk := make([]*gen.PackageDepCreate, len(deps))
	for i, dep := range deps {
		bulk[i] = s.client.PackageDep.Create().
			SetImporterPkg(dep.ImporterPkg).
			SetImportedPkg(dep.ImportedPkg).
			SetImportCount(dep.ImportCount)
	}

	// 执行批量插入
	_, err := s.client.PackageDep.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		return fmt.Errorf("batch save package deps failed: %w", err)
	}

	return nil
}

// GetPackageByPath 根据包路径获取包信息
func (s *StaticEntDBImpl) GetPackageByPath(pkgPath string) (*dos.Package, error) {
	ctx := context.Background()

	packageEnt, err := s.client.Pkg.
		Query().
		Where(pkg.PkgPath(pkgPath)).
		Only(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return nil, nil // 包不存在
		}
		return nil, fmt.Errorf("query package failed: %w", err)
	}

	return &dos.Package{
		PkgPath:    packageEnt.PkgPath,
		IsInternal: packageEnt.IsInternal,
	}, nil
}

// ListImportedBy 获取指定包导入的所有包
func (s *StaticEntDBImpl) ListImportedBy(pkgPath string) ([]*dos.Package, error) {
	ctx := context.Background()

	// 查询该包导入的依赖关系
	deps, err := s.client.PackageDep.
		Query().
		Where(packagedep.ImporterPkg(pkgPath)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query imported deps failed: %w", err)
	}

	// 获取被导入的包信息
	var packages []*dos.Package
	for _, dep := range deps {
		packageEnt, err := s.client.Pkg.
			Query().
			Where(pkg.PkgPath(dep.ImportedPkg)).
			Only(ctx)
		if err != nil {
			continue // 忽略找不到的包
		}

		packages = append(packages, &dos.Package{
			PkgPath:    packageEnt.PkgPath,
			IsInternal: packageEnt.IsInternal,
		})
	}

	return packages, nil
}

// ListImportersOf 获取导入指定包的所有包
func (s *StaticEntDBImpl) ListImportersOf(pkgPath string) ([]*dos.Package, error) {
	ctx := context.Background()

	// 查询导入该包的依赖关系
	deps, err := s.client.PackageDep.
		Query().
		Where(packagedep.ImportedPkg(pkgPath)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query importer deps failed: %w", err)
	}

	// 获取导入者包信息
	var packages []*dos.Package
	for _, dep := range deps {
		packageEnt, err := s.client.Pkg.
			Query().
			Where(pkg.PkgPath(dep.ImporterPkg)).
			Only(ctx)
		if err != nil {
			continue // 忽略找不到的包
		}

		packages = append(packages, &dos.Package{
			PkgPath:    packageEnt.PkgPath,
			IsInternal: packageEnt.IsInternal,
		})
	}

	return packages, nil
}

// GetPackageDeps 获取包依赖关系
func (s *StaticEntDBImpl) GetPackageDeps(importerPkg string) ([]*dos.PackageDep, error) {
	ctx := context.Background()

	// 查询指定包的依赖关系
	depEnts, err := s.client.PackageDep.
		Query().
		Where(packagedep.ImporterPkg(importerPkg)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query package deps failed: %w", err)
	}

	// 转换为业务实体
	var deps []*dos.PackageDep
	for _, depEnt := range depEnts {
		deps = append(deps, &dos.PackageDep{
			ImporterPkg: depEnt.ImporterPkg,
			ImportedPkg: depEnt.ImportedPkg,
			ImportCount: depEnt.ImportCount,
		})
	}

	return deps, nil
}

// SearchPackages 模糊搜索包
func (s *StaticEntDBImpl) SearchPackages(query string, limit int) ([]*dos.Package, error) {
	ctx := context.Background()

	// 构建模糊查询条件
	packageEnts, err := s.client.Pkg.
		Query().
		Where(pkg.PkgPathContainsFold(query)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("search packages failed: %w", err)
	}

	// 转换为业务实体
	var packages []*dos.Package
	for _, packageEnt := range packageEnts {
		packages = append(packages, &dos.Package{
			PkgPath:    packageEnt.PkgPath,
			IsInternal: packageEnt.IsInternal,
		})
	}

	return packages, nil
}

// GetAllPackages 获取所有包
func (s *StaticEntDBImpl) GetAllPackages() ([]*dos.Package, error) {
	ctx := context.Background()

	// 查询所有包
	packageEnts, err := s.client.Pkg.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query all packages failed: %w", err)
	}

	// 转换为业务实体
	var packages []*dos.Package
	for _, packageEnt := range packageEnts {
		packages = append(packages, &dos.Package{
			PkgPath:    packageEnt.PkgPath,
			IsInternal: packageEnt.IsInternal,
		})
	}

	return packages, nil
}

// GetAllPackageDeps 获取所有包依赖关系
func (s *StaticEntDBImpl) GetAllPackageDeps() ([]*dos.PackageDep, error) {
	ctx := context.Background()

	// 查询所有包依赖关系
	depEnts, err := s.client.PackageDep.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query all package deps failed: %w", err)
	}

	// 转换为业务实体
	var deps []*dos.PackageDep
	for _, depEnt := range depEnts {
		deps = append(deps, &dos.PackageDep{
			ImporterPkg: depEnt.ImporterPkg,
			ImportedPkg: depEnt.ImportedPkg,
			ImportCount: depEnt.ImportCount,
		})
	}

	return deps, nil
}
