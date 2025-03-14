package sqllite

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/repo"
)

var _ repo.StaticDBStore = (*StaticDBImpl)(nil)

// FuncNodeDB 函数节点的数据库模型
type FuncNodeDB struct {
	ID        uint      `gorm:"primarykey"` // 主键ID
	CreatedAt time.Time `gorm:"index"`      // 创建时间
	UpdatedAt time.Time // 更新时间
	Key       string    `gorm:"type:varchar(255);unique"` // 函数唯一标识
	Pkg       string    `gorm:"type:varchar(255);index"`  // 包名
	Name      string    `gorm:"type:varchar(255)"`        // 函数名
}

// 添加从数据库模型转换回内存模型的方法
func (f *FuncNodeDB) ToMemModel() *entity.FuncNode {
	return &entity.FuncNode{
		Key:      f.Key,
		Pkg:      f.Pkg,
		Name:     f.Name,
		Parent:   []string{}, // 需要通过边表查询填充
		Children: []string{}, // 需要通过边表查询填充
	}
}

func NodefromMemModel(node *entity.FuncNode) *FuncNodeDB {
	return &FuncNodeDB{
		Key:  node.Key,
		Pkg:  node.Pkg,
		Name: node.Name,
	}
}

// FuncEdgeDB 函数调用关系的数据库模型
type FuncEdgeDB struct {
	ID        uint      `gorm:"primarykey"` // 主键ID
	CreatedAt time.Time `gorm:"index"`      // 创建时间
	UpdatedAt time.Time // 更新时间
	CallerKey string    `gorm:"type:varchar(255);index"` // 调用方函数Key
	CalleeKey string    `gorm:"type:varchar(255);index"` // 被调用方函数Key
}

func EdgefromMemModel(edge *entity.FuncEdge) *FuncEdgeDB {
	return &FuncEdgeDB{
		CallerKey: edge.CallerKey,
		CalleeKey: edge.CalleeKey,
	}
}

type StaticDBImpl struct {
	db *sql.DB
}

func NewFuncNodeDB(dbPath string) (*StaticDBImpl, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db failed: %w", err)
	}

	return &StaticDBImpl{db: db}, nil
}

func (s *StaticDBImpl) InitTable() error {
	// 创建函数节点表
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS func_nodes (
			key TEXT PRIMARY KEY,
			pkg TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create func_nodes table failed: %w", err)
	}

	// 创建函数调用关系表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS func_edges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			caller_key TEXT NOT NULL,
			callee_key TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (caller_key) REFERENCES func_nodes(key),
			FOREIGN KEY (callee_key) REFERENCES func_nodes(key)
		)
	`)
	if err != nil {
		return fmt.Errorf("create func_edges table failed: %w", err)
	}

	// 创建索引
	_, err = s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_func_nodes_pkg ON func_nodes(pkg);
		CREATE INDEX IF NOT EXISTS idx_func_edges_caller ON func_edges(caller_key);
		CREATE INDEX IF NOT EXISTS idx_func_edges_callee ON func_edges(callee_key);
	`)
	if err != nil {
		return fmt.Errorf("create indexes failed: %w", err)
	}

	return nil
}

// SaveFuncNode 保存函数节点
func (s *StaticDBImpl) SaveFuncNode(node *entity.FuncNode) error {
	query := `INSERT OR REPLACE INTO func_nodes (key, pkg, name) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, node.Key, node.Pkg, node.Name)
	if err != nil {
		return fmt.Errorf("save func node failed: %w", err)
	}
	return nil
}

// SaveFuncEdge 保存函数调用关系
func (s *StaticDBImpl) SaveFuncEdge(edge *entity.FuncEdge) error {
	query := `INSERT INTO func_edges (caller_key, callee_key) VALUES (?, ?)`
	_, err := s.db.Exec(query, edge.CallerKey, edge.CalleeKey)
	if err != nil {
		return fmt.Errorf("save func edge failed: %w", err)
	}
	return nil
}

// GetFuncNodeByKey 根据Key获取函数节点
func (s *StaticDBImpl) GetFuncNodeByKey(key string) (*entity.FuncNode, error) {
	query := `SELECT key, pkg, name FROM func_nodes WHERE key = ?`
	row := s.db.QueryRow(query, key)

	node := &entity.FuncNode{}
	err := row.Scan(&node.Key, &node.Pkg, &node.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get func node failed: %w", err)
	}

	// 获取父节点
	parents, err := s.GetCallerEdges(key)
	if err != nil {
		log.Printf("get parent nodes failed: %v", err)
	} else {
		for _, p := range parents {
			node.Parent = append(node.Parent, p.Key)
		}
	}

	// 获取子节点
	children, err := s.GetCalleeEdges(key)
	if err != nil {
		log.Printf("get children nodes failed: %v", err)
	} else {
		for _, c := range children {
			node.Children = append(node.Children, c.Key)
		}
	}

	return node, nil
}

// GetCallerEdges 获取调用该函数的所有节点
func (s *StaticDBImpl) GetCallerEdges(calleeKey string) ([]*entity.FuncNode, error) {
	query := `
		SELECT n.key, n.pkg, n.name 
		FROM func_nodes n
		JOIN func_edges e ON n.key = e.caller_key
		WHERE e.callee_key = ?
	`
	rows, err := s.db.Query(query, calleeKey)
	if err != nil {
		return nil, fmt.Errorf("get caller edges failed: %w", err)
	}
	defer rows.Close()

	var nodes []*entity.FuncNode
	for rows.Next() {
		node := &entity.FuncNode{}
		err := rows.Scan(&node.Key, &node.Pkg, &node.Name)
		if err != nil {
			return nil, fmt.Errorf("scan caller node failed: %w", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetCalleeEdges 获取该函数调用的所有节点
func (s *StaticDBImpl) GetCalleeEdges(callerKey string) ([]*entity.FuncNode, error) {
	query := `
		SELECT n.key, n.pkg, n.name 
		FROM func_nodes n
		JOIN func_edges e ON n.key = e.callee_key
		WHERE e.caller_key = ?
	`
	rows, err := s.db.Query(query, callerKey)
	if err != nil {
		return nil, fmt.Errorf("get callee edges failed: %w", err)
	}
	defer rows.Close()

	var nodes []*entity.FuncNode
	for rows.Next() {
		node := &entity.FuncNode{}
		err := rows.Scan(&node.Key, &node.Pkg, &node.Name)
		if err != nil {
			return nil, fmt.Errorf("scan callee node failed: %w", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// Close 关闭数据库连接
func (s *StaticDBImpl) Close() error {
	return s.db.Close()
}

// GetAllFuncNodes 获取所有函数节点
func (s *StaticDBImpl) GetAllFuncNodes() ([]*entity.FuncNode, error) {
	query := `SELECT key, pkg, name FROM func_nodes`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all func nodes failed: %w", err)
	}
	defer rows.Close()

	var nodes []*entity.FuncNode
	for rows.Next() {
		node := &entity.FuncNode{}
		err := rows.Scan(&node.Key, &node.Pkg, &node.Name)
		if err != nil {
			return nil, fmt.Errorf("scan func node failed: %w", err)
		}
		nodes = append(nodes, node)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate func nodes failed: %w", err)
	}

	return nodes, nil
}

// GetAllFuncEdges 获取所有函数调用边
func (s *StaticDBImpl) GetAllFuncEdges() ([]*entity.FuncEdge, error) {
	query := `SELECT caller_key, callee_key FROM func_edges`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all func edges failed: %w", err)
	}
	defer rows.Close()

	var edges []*entity.FuncEdge
	for rows.Next() {
		edge := &entity.FuncEdge{}
		err := rows.Scan(&edge.CallerKey, &edge.CalleeKey)
		if err != nil {
			return nil, fmt.Errorf("scan func edge failed: %w", err)
		}
		edges = append(edges, edge)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate func edges failed: %w", err)
	}

	return edges, nil
}
