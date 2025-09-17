package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Pkg 包信息表 - 简化设计
type Pkg struct {
	ent.Schema
}

// Fields of the Pkg.
func (Pkg) Fields() []ent.Field {
	return []ent.Field{
		field.String("pkg_path").
			NotEmpty().
			Unique().
			Comment("包路径，如 github.com/example/pkg"),
		field.Bool("is_internal").
			Default(false).
			Comment("是否为内部包"),
	}
}

// Edges of the Pkg.
func (Pkg) Edges() []ent.Edge {
	return nil
}

// Indexes of the Pkg.
func (Pkg) Indexes() []ent.Index {
	return []ent.Index{
		// 按内部/外部分类查询
		index.Fields("is_internal"),
	}
}
