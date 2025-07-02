package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/toheart/goanalysis/api/filemanager/v1"
	"github.com/toheart/goanalysis/internal/biz/filemanager"
	"github.com/toheart/goanalysis/internal/biz/filemanager/dos"
	"github.com/toheart/goanalysis/internal/server/iface"
	"google.golang.org/grpc"
)

var _ iface.InitGrpcHttp = (*FileManagerService)(nil)

// FileManagerService 文件管理服务实现
type FileManagerService struct {
	v1.UnimplementedFileManagerServer

	fileBiz *filemanager.FileBiz
	log     *log.Helper
}

// NewFileManagerService 创建文件管理服务实例
func NewFileManagerService(fileBiz *filemanager.FileBiz, logger log.Logger) *FileManagerService {
	return &FileManagerService{
		fileBiz: fileBiz,
		log:     log.NewHelper(logger),
	}
}

// RegisterHttp 注册HTTP服务
func (s *FileManagerService) RegisterHttp(mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	// 注册自定义处理函数
	return v1.RegisterFileManagerHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
}

// RegisterGrpc 注册gRPC服务
func (s *FileManagerService) RegisterGrpc(svr *grpc.Server) {
	v1.RegisterFileManagerServer(svr, s)
}

// GetFileInfo 获取文件信息
func (s *FileManagerService) GetFileInfo(ctx context.Context, req *v1.GetFileInfoRequest) (*v1.GetFileInfoReply, error) {
	fileInfo, err := s.fileBiz.GetFileInfo(req.Id)
	if err != nil {
		return nil, fmt.Errorf("get file info failed: %w", err)
	}

	return &v1.GetFileInfoReply{
		FileInfo: convertToProtoFileInfo(fileInfo),
	}, nil
}

// ListFiles 获取文件列表
func (s *FileManagerService) ListFiles(ctx context.Context, req *v1.ListFilesRequest) (*v1.ListFilesReply, error) {
	// 转换文件类型
	fileType := dos.NewFileType(req.FileType)

	// 获取文件列表
	fileInfos, err := s.fileBiz.ListFiles(fileType, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, fmt.Errorf("list files failed: %w", err)
	}

	// 构建响应
	reply := &v1.ListFilesReply{
		Files: make([]*v1.FileInfo, 0, len(fileInfos)),
		Total: int64(len(fileInfos)), // 这里简化处理，实际应该从数据库获取总数
	}

	for _, fileInfo := range fileInfos {
		reply.Files = append(reply.Files, convertToProtoFileInfo(fileInfo))
	}

	return reply, nil
}

// DeleteFile 删除文件
func (s *FileManagerService) DeleteFile(ctx context.Context, req *v1.DeleteFileRequest) (*v1.DeleteFileReply, error) {
	err := s.fileBiz.DeleteFile(req.Id)
	if err != nil {
		return nil, fmt.Errorf("delete file failed: %w", err)
	}

	return &v1.DeleteFileReply{
		Success: true,
	}, nil
}

// DownloadFile 下载文件
func (s *FileManagerService) DownloadFile(ctx context.Context, req *v1.DownloadFileRequest) (*v1.DownloadFileReply, error) {
	// 获取文件信息
	fileInfo, err := s.fileBiz.GetFileInfo(req.Id)
	if err != nil {
		return nil, fmt.Errorf("get file info failed: %w", err)
	}

	// 读取文件内容
	fileContent, err := os.ReadFile(fileInfo.FilePath)
	if err != nil {
		return nil, fmt.Errorf("read file content failed: %w", err)
	}

	// 构建响应
	return &v1.DownloadFileReply{
		FileContent: base64.StdEncoding.EncodeToString(fileContent),
		FileName:    fileInfo.FileName,
		ContentType: fileInfo.ContentType,
	}, nil
}

// convertToProtoFileInfo 将实体文件信息转换为proto文件信息
func convertToProtoFileInfo(fileInfo *dos.FileInfo) *v1.FileInfo {
	var fileType v1.FileType
	switch fileInfo.FileType {
	case dos.FileTypeRuntime:
		fileType = v1.FileType_FILE_TYPE_RUNTIME
	case dos.FileTypeStatic:
		fileType = v1.FileType_FILE_TYPE_STATIC
	default:
		fileType = v1.FileType_FILE_TYPE_UNSPECIFIED
	}

	return &v1.FileInfo{
		Id:          fileInfo.ID,
		FileName:    fileInfo.FileName,
		FilePath:    fileInfo.FilePath,
		FileType:    fileType,
		FileSize:    fileInfo.FileSize,
		ContentType: fileInfo.ContentType,
		UploadTime:  fileInfo.UploadTime.Format(time.RFC3339),
		Description: fileInfo.Description,
	}
}
