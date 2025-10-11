package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FuncNode holds the schema definition for the FuncNode entity.
type FuncNode struct {
	ent.Schema
}

// Fields of the FuncNode.
func (FuncNode) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			Unique().
			NotEmpty().
			Comment("短格式唯一标识，如 n6796"),
		field.String("full_name").
			NotEmpty().
			Comment("完整的函数路径，如 crypto/hmac.New$1"),
		field.String("pkg").
			NotEmpty().
			Comment("包导入路径，如 net/http"),
		field.String("pkg_path").
			Optional().
			Comment("完整包路径，用于关联 PackageNode"),
		field.String("name").
			NotEmpty(),
		// 源码位置信息（可选，合成/内建函数可能不存在）
		field.String("filename").
			Optional().
			Comment("源文件路径"),
		field.Int("start_line").
			Optional().
			Comment("函数开始行"),
		field.Int("end_line").
			Optional().
			Comment("函数结束行"),
		// LLM 分析增强字段
		field.String("signature").
			Optional().
			Comment("函数签名，如 func(http.ResponseWriter, *http.Request)"),
		field.Text("source_code").
			Optional().
			Comment("函数的源代码（压缩存储）"),
		field.Time("CreatedAt").
			Default(time.Now),
		field.Time("UpdatedAt").
			Default(time.Now),
	}
}

// Edges of the FuncNode.
func (FuncNode) Edges() []ent.Edge {
	return nil
}

// Indexes of the FuncNode.
func (FuncNode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("pkg"),
		index.Fields("key").
			Unique(),
		index.Fields("full_name"),
		// 按文件与起始行的常用查询
		index.Fields("filename"),
		index.Fields("filename", "start_line"),
		// 按包路径查询，用于关联包级分析
		index.Fields("pkg_path"),
	}
}
