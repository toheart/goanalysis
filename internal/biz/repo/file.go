package repo

import "github.com/toheart/goanalysis/internal/biz/filemanager/dos"

type FileRepo interface {
	DeleteFileInfo(id int64) error
	GetFileInfoByID(id int64) (*dos.FileInfo, error)
	ListFileInfos(fileType dos.FileType, limit int, offset int) ([]*dos.FileInfo, error)
	SaveFileInfo(info *dos.FileInfo) error
}
