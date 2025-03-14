package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/toheart/goanalysis/internal/biz/entity"
)

func (h *HttpServer) BaseHandler(mux *runtime.ServeMux, frontendDir string, fileServer http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查是否是API路径
		if isAPIPath(r.URL.Path) {
			// 如果是API路径，交给gRPC-Gateway处理
			mux.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/runtime/file/upload" {
			h.HandleChunkUpload(w, r)
			return
		}

		// 构建静态资源的完整路径
		path := filepath.Join(frontendDir, r.URL.Path)

		// 检查请求的文件是否存在
		if fileExists(path) && !strings.HasSuffix(path, "/") {
			// 如果文件存在，直接提供该文件
			fileServer.ServeHTTP(w, r)
			return
		}

		// 如果是目录或文件不存在，返回index.html（SPA应用通常需要这样处理）
		indexPath := filepath.Join(frontendDir, "index.html")
		if fileExists(indexPath) {
			http.ServeFile(w, r, indexPath)
			return
		}

		// 如果index.html也不存在，返回404
		http.NotFound(w, r)
	}
}

// isAPIPath 判断是否为API路径
func isAPIPath(path string) bool {
	// 假设所有API路径都以/api/开头
	return len(path) >= 5 && path[:5] == "/api/"
}

func (h *HttpServer) HandleChunkUpload(w http.ResponseWriter, r *http.Request) {
	// 解析分块参数
	fileID := r.FormValue("file_id")
	chunkIndex := r.FormValue("chunk_index")
	totalChunks := r.FormValue("total_chunks")
	fileName := r.FormValue("file_name")
	description := r.FormValue("description")
	contentType := r.FormValue("content_type")

	// 验证必要参数
	if fileID == "" || chunkIndex == "" || totalChunks == "" {
		http.Error(w, "missing required parameters", http.StatusBadRequest)
		return
	}

	// 创建临时目录存储分块
	tmpDir := fmt.Sprintf("./tmp/%s", fileID)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		http.Error(w, "create temp directory failed", http.StatusInternalServerError)
		return
	}

	// 保存分块文件
	file, _, err := r.FormFile("chunk")
	if err != nil {
		http.Error(w, "get chunk file failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	chunkPath := fmt.Sprintf("%s/%s.part", tmpDir, chunkIndex)
	dst, err := os.Create(chunkPath)
	if err != nil {
		http.Error(w, "create chunk file failed", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "save chunk file failed", http.StatusInternalServerError)
		return
	}

	// 检查是否为最后一个分块
	chunkIndexInt, _ := strconv.Atoi(chunkIndex)
	totalChunksInt, _ := strconv.Atoi(totalChunks)
	isLast := chunkIndexInt == totalChunksInt-1 // 分块索引从0开始，所以最后一个是totalChunks-1

	// 如果是最后一个分块，合并所有分块
	if isLast {
		// 创建上传目录
		uploadDir := h.fileBiz.GetUploadDir(true)
		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			http.Error(w, "create upload directory failed", http.StatusInternalServerError)
			return
		}

		// 如果没有提供文件名，生成一个
		if fileName == "" {
			fileName = fmt.Sprintf("file_%d", time.Now().Unix())
		}

		// 合并后的文件路径
		filePath := fmt.Sprintf("%s/%s", uploadDir, fileName)

		// 合并分块
		fileSize, err := h.mergeChunks(tmpDir, filePath, totalChunksInt)
		if err != nil {
			http.Error(w, fmt.Sprintf("merge chunks failed: %v", err), http.StatusInternalServerError)
			return
		}

		// 读取文件头以检测是否为SQLite数据库
		isSQLite := false
		f, err := os.Open(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("open file failed: %v", err), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		header := make([]byte, 16)
		_, err = f.Read(header)
		if err != nil {
			http.Error(w, fmt.Sprintf("read file header failed: %v", err), http.StatusInternalServerError)
			return
		}
		isSQLite = isSQLiteFile(header)
		if !isSQLite {
			http.Error(w, "file is not a sqlite database", http.StatusBadRequest)
			return
		}

		// 确定文件类型
		fileTypeEnum := entity.FileTypeRuntime

		// 创建文件信息
		fileInfo := &entity.FileInfo{
			FileName:    fileName,
			FilePath:    filePath,
			FileType:    fileTypeEnum,
			FileSize:    fileSize,
			ContentType: contentType,
			UploadTime:  time.Now(),
			Description: description,
		}

		// 保存文件信息到数据库
		savedFileInfo, err := h.fileBiz.SaveFileInfo(fileInfo)
		if err != nil {
			// 删除已上传的文件
			os.Remove(filePath)
			http.Error(w, fmt.Sprintf("save file info failed: %v", err), http.StatusInternalServerError)
			return
		}

		// 返回文件信息
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "completed",
			"file_info": savedFileInfo,
		})
	} else {
		// 如果不是最后一个分块，返回成功状态
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Chunk %s uploaded successfully", chunkIndex),
		})
	}
}

// 合并分块文件
func (h *HttpServer) mergeChunks(tmpDir, destPath string, totalChunks int) (int64, error) {
	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return 0, fmt.Errorf("create destination file failed: %w", err)
	}
	defer destFile.Close()

	var totalSize int64 = 0

	// 按顺序合并所有分块
	for i := 0; i < totalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/%d.part", tmpDir, i)

		// 打开分块文件
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return 0, fmt.Errorf("open chunk file %s failed: %w", chunkPath, err)
		}

		// 获取分块大小
		chunkInfo, err := chunkFile.Stat()
		if err != nil {
			chunkFile.Close()
			return 0, fmt.Errorf("get chunk file info failed: %w", err)
		}
		totalSize += chunkInfo.Size()

		// 复制分块内容到目标文件
		if _, err := io.Copy(destFile, chunkFile); err != nil {
			chunkFile.Close()
			return 0, fmt.Errorf("copy chunk content failed: %w", err)
		}

		// 关闭分块文件
		chunkFile.Close()

		// 删除分块文件
		os.Remove(chunkPath)
	}

	// 删除临时目录
	os.RemoveAll(tmpDir)

	return totalSize, nil
}

// 检测是否为SQLite数据库文件
func isSQLiteFile(fileContent []byte) bool {
	// SQLite文件的魔数是 "SQLite format 3\000"
	if len(fileContent) < 16 {
		return false
	}

	sqliteHeader := "SQLite format 3"
	header := string(fileContent[:15])
	return header == sqliteHeader
}
