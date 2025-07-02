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
			NotEmpty(),
		field.String("name").
			NotEmpty(),
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
	}
}
