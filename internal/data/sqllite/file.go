package sqllite

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/glebarez/go-sqlite" // 引入 sqlite3 驱动
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/repo"
)

var _ repo.FileRepo = (*FileDB)(nil)

// FileDB 文件数据库管理结构
type FileDB struct {
	db *sql.DB
}

// NewFileDB 创建文件数据库管理实例
func NewFileDB(dbPath string) (*FileDB, error) {
	// 打开数据库连接
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite3 database failed: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	fileDB := &FileDB{
		db: db,
	}

	// 初始化数据库表
	if err := fileDB.initTables(); err != nil {
		db.Close()
		return nil, err
	}

	return fileDB, nil
}

// initTables 初始化数据库表
func (f *FileDB) initTables() error {
	// 创建文件信息表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS file_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_type TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		content_type TEXT NOT NULL,
		upload_time TIMESTAMP NOT NULL,
		description TEXT
	);
	`

	_, err := f.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("create file_info table failed: %w", err)
	}

	return nil
}

// SaveFileInfo 保存文件信息
func (f *FileDB) SaveFileInfo(info *entity.FileInfo) error {
	// 插入文件信息
	insertSQL := `
	INSERT INTO file_info (
		file_name, file_path, file_type, file_size, 
		content_type, upload_time, description
	) VALUES (?, ?, ?, ?, ?, ?, ?);
	`

	result, err := f.db.Exec(
		insertSQL,
		info.FileName,
		info.FilePath,
		info.FileType,
		info.FileSize,
		info.ContentType,
		info.UploadTime,
		info.Description,
	)
	if err != nil {
		return fmt.Errorf("insert file info failed: %w", err)
	}

	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id failed: %w", err)
	}

	info.ID = id
	return nil
}

// GetFileInfoByID 根据ID获取文件信息
func (f *FileDB) GetFileInfoByID(id int64) (*entity.FileInfo, error) {
	querySQL := `
	SELECT id, file_name, file_path, file_type, file_size, 
	       content_type, upload_time, description
	FROM file_info
	WHERE id = ?;
	`

	var info entity.FileInfo
	var uploadTimeStr string

	err := f.db.QueryRow(querySQL, id).Scan(
		&info.ID,
		&info.FileName,
		&info.FilePath,
		&info.FileType,
		&info.FileSize,
		&info.ContentType,
		&uploadTimeStr,
		&info.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 文件不存在
		}
		return nil, fmt.Errorf("query file info failed: %w", err)
	}

	// 解析时间字符串
	uploadTime, err := time.Parse(time.RFC3339, uploadTimeStr)
	if err != nil {
		return nil, fmt.Errorf("parse upload time failed: %w", err)
	}
	info.UploadTime = uploadTime

	return &info, nil
}

// ListFileInfos 获取文件信息列表
func (f *FileDB) ListFileInfos(fileType entity.FileType, limit int, offset int) ([]*entity.FileInfo, error) {
	var querySQL string
	var args []interface{}

	// 按文件类型筛选
	querySQL = `
	SELECT id, file_name, file_path, file_type, file_size, 
			content_type, upload_time, description
	FROM file_info
	WHERE file_type = ?
	ORDER BY upload_time DESC
	LIMIT ? OFFSET ?;
	`
	args = append(args, fileType, limit, offset)
	rows, err := f.db.Query(querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query file infos failed: %w", err)
	}
	defer rows.Close()

	var fileInfos []*entity.FileInfo
	for rows.Next() {
		var info entity.FileInfo
		var uploadTimeStr string

		err := rows.Scan(
			&info.ID,
			&info.FileName,
			&info.FilePath,
			&info.FileType,
			&info.FileSize,
			&info.ContentType,
			&uploadTimeStr,
			&info.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("scan file info failed: %w", err)
		}

		// 解析时间字符串
		uploadTime, err := time.Parse(time.RFC3339, uploadTimeStr)
		if err != nil {
			return nil, fmt.Errorf("parse upload time failed: %w", err)
		}
		info.UploadTime = uploadTime

		fileInfos = append(fileInfos, &info)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows failed: %w", err)
	}

	return fileInfos, nil
}

// DeleteFileInfo 删除文件信息
func (f *FileDB) DeleteFileInfo(id int64) error {
	deleteSQL := `DELETE FROM file_info WHERE id = ?;`

	_, err := f.db.Exec(deleteSQL, id)
	if err != nil {
		return fmt.Errorf("delete file info failed: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (f *FileDB) Close() error {
	return f.db.Close()
}
