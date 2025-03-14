package entity

import "time"

// FileType 文件类型枚举
type FileType string

const (
	FileTypeRuntime FileType = "runtime" // 运行时文件
	FileTypeStatic  FileType = "static"  // 静态分析文件
)

// FileInfo 文件信息结构体
type FileInfo struct {
	ID          int64     `json:"id"`           // 文件ID
	FileName    string    `json:"file_name"`    // 文件名
	FilePath    string    `json:"file_path"`    // 文件存储路径
	FileType    FileType  `json:"file_type"`    // 文件类型
	FileSize    int64     `json:"file_size"`    // 文件大小（字节）
	ContentType string    `json:"content_type"` // 文件MIME类型
	UploadTime  time.Time `json:"upload_time"`  // 上传时间
	Description string    `json:"description"`  // 文件描述
}
