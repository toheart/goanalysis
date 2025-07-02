package filemanager

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/biz/filemanager/dos"
	"github.com/toheart/goanalysis/internal/biz/repo"
	"github.com/toheart/goanalysis/internal/conf"
)

// FileBiz 文件管理业务逻辑
type FileBiz struct {
	conf *conf.Biz
	log  *log.Helper
	repo repo.FileRepo
}

// NewFileBiz 创建文件管理业务逻辑实例
func NewFileBiz(bizConf *conf.Biz, logger log.Logger, fileRepo repo.FileRepo) *FileBiz {
	return &FileBiz{
		repo: fileRepo,
		conf: bizConf,
		log:  log.NewHelper(logger),
	}
}

func (f *FileBiz) GetUploadDir(runtime bool) string {
	return entity.GetFileStoragePath(f.conf.FileStoragePath, runtime)
}

// UploadFile 上传文件
func (f *FileBiz) SaveFileInfo(fileInfo *dos.FileInfo) (*dos.FileInfo, error) {
	if err := f.repo.SaveFileInfo(fileInfo); err != nil {
		// 删除已上传的文件
		os.Remove(fileInfo.FilePath)
		return nil, fmt.Errorf("save file info failed: %w", err)
	}

	return fileInfo, nil
}

// GetFileInfo 获取文件信息
func (f *FileBiz) GetFileInfo(id int64) (*dos.FileInfo, error) {
	fileInfo, err := f.repo.GetFileInfoByID(id)
	if err != nil {
		return nil, fmt.Errorf("get file info failed: %w", err)
	}

	if fileInfo == nil {
		return nil, fmt.Errorf("file not found")
	}

	return fileInfo, nil
}

// ListFiles 获取文件列表
func (f *FileBiz) ListFiles(fileType dos.FileType, limit int, offset int) ([]*dos.FileInfo, error) {
	fileInfos, err := f.repo.ListFileInfos(fileType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list file infos failed: %w", err)
	}

	return fileInfos, nil
}

// DeleteFile 删除文件
func (f *FileBiz) DeleteFile(id int64) error {
	// 获取文件信息
	fileInfo, err := f.GetFileInfo(id)
	if err != nil {
		return err
	}

	// 删除物理文件
	if err := os.Remove(fileInfo.FilePath); err != nil {
		f.log.Warnf("delete physical file failed: %s", err)
	}

	// 删除数据库记录
	if err := f.repo.DeleteFileInfo(id); err != nil {
		return fmt.Errorf("delete file info failed: %w", err)
	}

	return nil
}
