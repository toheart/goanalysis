package sqlite

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"github.com/toheart/goanalysis/internal/biz/callgraph/dos"
	"github.com/toheart/goanalysis/internal/biz/repo"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen/funcedge"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen/funcnode"
)

var _ repo.StaticDBStore = (*StaticEntDBImpl)(nil)

// StaticEntDBImpl 使用 Ent 框架的静态分析数据库实现
type StaticEntDBImpl struct {
	client *gen.Client
}

// NewStaticEntDBImpl 创建函数节点数据库（使用 Ent 框架）
func NewStaticEntDBImpl(dbPath string) (*StaticEntDBImpl, error) {
	// 创建 Ent 客户端
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

	if exists {
		// 更新节点
		_, err = s.client.FuncNode.
			Update().
			Where(funcnode.Key(node.Key)).
			SetFullName(node.FullName).
			SetPkg(node.Pkg).
			SetName(node.Name).
			Save(ctx)
	} else {
		// 创建节点
		_, err = s.client.FuncNode.
			Create().
			SetKey(node.Key).
			SetFullName(node.FullName).
			SetPkg(node.Pkg).
			SetName(node.Name).
			Save(ctx)
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

	// 转换为业务实体
	node := &dos.FuncNode{
		Key:      funcEnt.Key,
		FullName: funcEnt.FullName,
		Pkg:      funcEnt.Pkg,
		Name:     funcEnt.Name,
	}

	// 获取父节点
	parents, err := s.GetCallerEdges(key)
	if err == nil {
		node.Parents = parents
	}

	// 获取子节点
	children, err := s.GetCalleeEdges(key)
	if err == nil {
		node.Childrens = children
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
		node := &dos.FuncNode{
			Key:      funcNode.Key,
			FullName: funcNode.FullName,
			Pkg:      funcNode.Pkg,
			Name:     funcNode.Name,
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
		node := &dos.FuncNode{
			Key:      funcNode.Key,
			FullName: funcNode.FullName,
			Pkg:      funcNode.Pkg,
			Name:     funcNode.Name,
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
		node := &dos.FuncNode{
			Key:      funcEnt.Key,
			FullName: funcEnt.FullName,
			Pkg:      funcEnt.Pkg,
			Name:     funcEnt.Name,
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
			CallerKey: funcEdge.CallerKey,
			CalleeKey: funcEdge.CalleeKey,
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
		node := &dos.FuncNode{
			Key:      funcEnt.Key,
			FullName: funcEnt.FullName,
			Pkg:      funcEnt.Pkg,
			Name:     funcEnt.Name,
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// Close 关闭数据库连接
func (s *StaticEntDBImpl) Close() error {
	return s.client.Close()
}
