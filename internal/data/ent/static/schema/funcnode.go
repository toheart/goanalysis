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
			NotEmpty(),
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
	}
}
