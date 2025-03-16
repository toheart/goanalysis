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
