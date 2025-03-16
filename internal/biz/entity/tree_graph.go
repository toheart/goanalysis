package entity

// TreeNode 表示树状图节点
type TreeNode struct {
	Name      string      `json:"name"`                // 节点名称
	Value     int64       `json:"value,omitempty"`     // 值，只在tooltip中显示
	Collapsed bool        `json:"collapsed,omitempty"` // 是否默认折叠，true表示折叠
	Children  []*TreeNode `json:"children,omitempty"`  // 子节点
}

// TreeGraph 表示完整的树状图结构
type TreeGraph struct {
	Root *TreeNode `json:"root"` // 根节点
}
