package sqlite

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/repo"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data/ent/file/gen"
	"github.com/toheart/goanalysis/internal/data/ent/file/gen/fileinfo"
)

var _ repo.FileRepo = (*FileEntDB)(nil)

// FileEntDB 使用 Ent 框架的文件数据库管理结构
type FileEntDB struct {
	conf   *conf.Data
	client *gen.Client
}

// NewFileEntDB 创建文件数据库管理实例（使用 Ent 框架）
func NewFileEntDB(conf *conf.Data) (repo.FileRepo, error) {
	// 创建 Ent 客户端
	client, err := gen.Open(dialect.SQLite, ParseDBPath(conf.Dbpath))
	if err != nil {
		return nil, fmt.Errorf("create ent client failed: %w", err)
	}

	// 创建 FileEntDB 实例
	fileDB := &FileEntDB{
		client: client,
		conf:   conf,
	}

	// 初始化数据库表
	if err := fileDB.initTables(context.Background()); err != nil {
		client.Close()
		return nil, err
	}

	return fileDB, nil
}

// initTables 初始化数据库表
func (f *FileEntDB) initTables(ctx context.Context) error {
	// 自动创建表结构
	if err := f.client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}
	return nil
}

// SaveFileInfo 保存文件信息
func (f *FileEntDB) SaveFileInfo(info *entity.FileInfo) error {
	ctx := context.Background()

	// 使用 Ent 创建文件信息
	fileEnt, err := f.client.FileInfo.
		Create().
		SetFileName(info.FileName).
		SetFilePath(info.FilePath).
		SetFileType(string(info.FileType)).
		SetFileSize(info.FileSize).
		SetContentType(info.ContentType).
		SetUploadTime(info.UploadTime).
		SetDescription(info.Description).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("保存文件信息失败: %w", err)
	}

	// 更新 ID
	info.ID = fileEnt.ID
	return nil
}

// GetFileInfoByID 根据ID获取文件信息
func (f *FileEntDB) GetFileInfoByID(id int64) (*entity.FileInfo, error) {
	ctx := context.Background()

	// 查询文件信息
	fileEnt, err := f.client.FileInfo.Get(ctx, id)
	if err != nil {
		if gen.IsNotFound(err) {
			return nil, nil // 文件不存在
		}
		return nil, fmt.Errorf("查询文件信息失败: %w", err)
	}

	// 转换为业务实体
	info := &entity.FileInfo{
		ID:          fileEnt.ID,
		FileName:    fileEnt.FileName,
		FilePath:    fileEnt.FilePath,
		FileType:    entity.FileType(fileEnt.FileType),
		FileSize:    fileEnt.FileSize,
		ContentType: fileEnt.ContentType,
		UploadTime:  fileEnt.UploadTime,
		Description: fileEnt.Description,
	}

	return info, nil
}

// ListFileInfos 获取文件信息列表
func (f *FileEntDB) ListFileInfos(fileType entity.FileType, limit int, offset int) ([]*entity.FileInfo, error) {
	ctx := context.Background()

	// 如果 limit 为 0，设置一个默认值
	if limit <= 0 {
		limit = 10
	}

	// 确保 offset 不为负数
	if offset < 0 {
		offset = 0
	}

	// 查询文件列表
	fileEnts, err := f.client.FileInfo.
		Query().
		Where(fileinfo.FileType(string(fileType))).
		Order(gen.Desc(fileinfo.FieldUploadTime)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询文件列表失败: %w", err)
	}

	// 转换为业务实体
	var fileInfos []*entity.FileInfo
	for _, fileEnt := range fileEnts {
		info := &entity.FileInfo{
			ID:          int64(fileEnt.ID),
			FileName:    fileEnt.FileName,
			FilePath:    fileEnt.FilePath,
			FileType:    entity.FileType(fileEnt.FileType),
			FileSize:    fileEnt.FileSize,
			ContentType: fileEnt.ContentType,
			UploadTime:  fileEnt.UploadTime,
			Description: fileEnt.Description,
		}
		fileInfos = append(fileInfos, info)
	}

	// 如果没有找到记录，返回空切片而不是 nil
	if len(fileInfos) == 0 {
		return []*entity.FileInfo{}, nil
	}

	return fileInfos, nil
}

// DeleteFileInfo 删除文件信息
func (f *FileEntDB) DeleteFileInfo(id int64) error {
	ctx := context.Background()

	// 删除文件信息
	err := f.client.FileInfo.
		DeleteOneID(id).
		Exec(ctx)
	if err != nil {
		if gen.IsNotFound(err) {
			return nil // 文件不存在，视为删除成功
		}
		return fmt.Errorf("删除文件信息失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (f *FileEntDB) Close() error {
	return f.client.Close()
}

// ParseDBPath 解析数据库路径
func ParseDBPath(dbPath string) string {
	return fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&cache=shared&_journal_mode=WAL", dbPath)
}
