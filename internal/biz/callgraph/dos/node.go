package dos

// Position 表示函数的位置信息
type Position struct {
	Filename string `json:"filename"` // 文件名
	Line     int    `json:"line"`     // 起始行号
	EndLine  int    `json:"end_line"` // 结束行号
}

// Package 表示包信息 - 简化设计
type Package struct {
	PkgPath    string `json:"pkg_path"`    // 包路径
	IsInternal bool   `json:"is_internal"` // 是否为内部包
}

// PackageDep 表示包依赖关系 - 简化设计
type PackageDep struct {
	ImporterPkg string `json:"importer_pkg"` // 导入者包路径
	ImportedPkg string `json:"imported_pkg"` // 被导入包路径
	ImportCount int    `json:"import_count"` // 导入次数
}

// FuncEdge 表示函数调用边
type FuncEdge struct {
	CallerKey    string `json:"caller_key"`     // 调用者键
	CalleeKey    string `json:"callee_key"`     // 被调用者键
	CallSiteFile string `json:"call_site_file"` // 调用点文件
	CallSiteLine int    `json:"call_site_line"` // 调用点行号
	CallType     string `json:"call_type"`      // 调用类型
}

// FuncNode 表示函数节点
type FuncNode struct {
	Key        string      `json:"key"`         // 短格式唯一标识，如 "n6796"
	FullName   string      `json:"full_name"`   // 完整的函数路径，如 "crypto/hmac.New$1"
	Pkg        string      `json:"pkg"`         // 包名
	PkgPath    string      `json:"pkg_path"`    // 包路径
	Name       string      `json:"name"`        // 函数名
	Signature  string      `json:"signature"`   // 函数签名
	SourceCode string      `json:"source_code"` // 源代码（仅内部包）
	Position   *Position   `json:"position"`    // 位置信息
	Parents    []*FuncNode `json:"parents"`     // 父节点
	Childrens  []*FuncNode `json:"childrens"`   // 子节点
}
