package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// 生成ent对象，不使用外键关联
type FuncEdge struct {
	ent.Schema
}

func (FuncEdge) Fields() []ent.Field {
	return []ent.Field{
		field.Time("CreatedAt"),
		field.Time("UpdatedAt"),
		field.String("CallerKey"),
		field.String("CalleeKey"),
		// LLM 分析增强字段 - 调用上下文信息
		field.String("call_site_file").
			Optional().
			Comment("调用发生的文件路径"),
		field.Int("call_site_line").
			Optional().
			Comment("调用发生的行号"),
		field.String("call_type").
			Default("static").
			Comment("调用类型: static, dynamic, go, defer"),
	}
}

func (FuncEdge) Edges() []ent.Edge {
	return nil
}

func (FuncEdge) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("CallerKey"),
		index.Fields("CalleeKey"),
	}
}
