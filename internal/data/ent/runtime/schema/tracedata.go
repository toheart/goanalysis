package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TraceData holds the schema definition for the TraceData entity.
type TraceData struct {
	ent.Schema
}

// Fields of the TraceData.
func (TraceData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Unique(),
		field.String("name").
			NotEmpty(),
		field.Uint64("gid"),
		field.Int("indent").
			Default(0),
		field.Int("paramsCount").
			StorageKey("paramsCount").
			Default(0),
		field.String("timeCost").
			StorageKey("timeCost").
			Optional(),
		field.Int64("parentId").
			StorageKey("parentId").
			Optional(),
		field.String("createdAt").
			StorageKey("createdAt"),
		field.String("seq").
			StorageKey("seq").
			Optional(),
	}
}

func (TraceData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "traceData",
		},
	}
}

// Edges of the TraceData.
func (TraceData) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the TraceData.
func (TraceData) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("gid"),
		index.Fields("name"),
		index.Fields("parentId"),
	}
}
