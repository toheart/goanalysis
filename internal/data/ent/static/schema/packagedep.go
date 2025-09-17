package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PackageDep 包依赖关系表 - 简化设计
type PackageDep struct {
	ent.Schema
}

// Fields of the PackageDep.
func (PackageDep) Fields() []ent.Field {
	return []ent.Field{
		field.String("importer_pkg").
			NotEmpty().
			Comment("导入者包路径"),
		field.String("imported_pkg").
			NotEmpty().
			Comment("被导入包路径"),
		field.Int("import_count").
			Default(1).
			Comment("导入次数（在不同文件中的导入次数）"),
	}
}

// Edges of the PackageDep.
func (PackageDep) Edges() []ent.Edge {
	return nil
}

// Indexes of the PackageDep.
func (PackageDep) Indexes() []ent.Index {
	return []ent.Index{
		// 复合唯一约束：(importer_pkg, imported_pkg)
		index.Fields("importer_pkg", "imported_pkg").
			Unique(),
		// 按导入者查询 - 查看某个包导入了哪些包
		index.Fields("importer_pkg"),
		// 按被导入者查询 - 查看某个包被哪些包导入
		index.Fields("imported_pkg"),
	}
}
