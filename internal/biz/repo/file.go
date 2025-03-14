package repo

import "github.com/toheart/goanalysis/internal/biz/entity"

type FileRepo interface {
	DeleteFileInfo(id int64) error
	GetFileInfoByID(id int64) (*entity.FileInfo, error)
	ListFileInfos(fileType entity.FileType, limit int, offset int) ([]*entity.FileInfo, error)
	SaveFileInfo(info *entity.FileInfo) error
}
