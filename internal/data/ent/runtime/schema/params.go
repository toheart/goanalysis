package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ParamStoreData 存储参数信息的实体
type ParamStoreData struct {
	ent.Schema
}

func (ParamStoreData) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable().
			Comment("唯一标识符"),

		field.Int64("traceId").
			StorageKey("traceId").
			Comment("关联的TraceData ID"),

		field.Int("position").
			StorageKey("position").
			Comment("参数位置"),

		field.String("data").
			Comment("参数JSON数据").
			Default(""),

		field.Bool("isReceiver").
			StorageKey("isReceiver").
			Comment("是否为接收者参数").
			Default(false),

		field.Int64("baseId").
			StorageKey("baseId").
			Optional().
			Nillable().
			Comment("基础参数ID（自关联，当参数为增量存储时使用）"),
	}
}

// Edges of the ParamStoreData.
func (ParamStoreData) Edges() []ent.Edge {
	// 使用逻辑关联, 不用外键
	return nil
}

func (ParamStoreData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "ParamStore",
		},
	}
}
