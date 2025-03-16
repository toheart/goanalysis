package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// GoroutineTrace holds the schema definition for the GoroutineTrace entity.
type GoroutineTrace struct {
	ent.Schema
}

// Fields of the GoroutineTrace.
func (GoroutineTrace) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Positive().
			Unique(),
		field.Uint64("originGid").
			StorageKey("originGid").
			Positive(),
		field.String("timeCost").
			StorageKey("timeCost").
			Optional(),
		field.String("createTime").
			StorageKey("createTime").
			Optional(),
		field.Int("isFinished").
			StorageKey("isFinished").
			Optional(),
		field.String("initFuncName").
			StorageKey("initFuncName").
			Optional(),
	}
}

func (GoroutineTrace) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "goroutineTrace",
		},
	}
}

// Edges of the GoroutineTrace.
func (GoroutineTrace) Edges() []ent.Edge {
	return nil
}
