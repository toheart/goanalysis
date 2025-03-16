package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FileInfo holds the schema definition for the FileInfo entity.
type FileInfo struct {
	ent.Schema
}

// Fields of the FileInfo.
func (FileInfo) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable(),
		field.String("file_name").
			NotEmpty(),
		field.String("file_path").
			NotEmpty(),
		field.String("file_type").
			NotEmpty(),
		field.Int64("file_size").
			Positive(),
		field.String("content_type").
			NotEmpty(),
		field.Time("upload_time").
			Default(time.Now),
		field.String("description").
			Optional(),
	}
}

// Edges of the FileInfo.
func (FileInfo) Edges() []ent.Edge {
	return nil
}

// Indexes of the FileInfo.
func (FileInfo) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_type"),
		index.Fields("upload_time"),
	}
}
