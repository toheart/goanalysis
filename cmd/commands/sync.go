package commands

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
)

// SyncCommand 同步前端代码命令
type SyncCommand struct {
	cmdbase.BaseCommand
	outputDir string
}

// GitHubRelease GitHub发布信息结构
type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset 发布资源结构
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// NewSyncCommand 创建同步命令
func NewSyncCommand() *SyncCommand {
	cmd := &SyncCommand{}
	cmd.CobraCmd = &cobra.Command{
		Use:   "sync-web",
		Short: "sync web",
		Long:  `sync web from github`,
		Run:   cmd.Run,
	}
	return cmd
}

// Init 初始化同步命令
func (s *SyncCommand) Init() {
	s.CobraCmd.Flags().StringVarP(&s.outputDir, "output", "o", "web", "output directory (default: web)")
}

// Run 执行同步命令
func (s *SyncCommand) Run(cmd *cobra.Command, args []string) {
	fmt.Println("start sync web...")

	// 创建输出目录
	if err := s.createOutputDir(); err != nil {
		fmt.Printf("create output directory failed: %v\n", err)
		os.Exit(1)
	}

	// 创建resty客户端
	client := resty.New()

	// 获取最新发布版本信息
	fmt.Println("get latest release...")
	release, err := s.getLatestRelease(client)
	if err != nil {
		fmt.Printf("get latest release failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("get latest release: %s\n", release.TagName)

	// 查找zip文件下载链接
	downloadURL, err := s.findZipDownloadURL(release)
	if err != nil {
		fmt.Printf("find download url failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("start download: %s\n", downloadURL)

	// 下载文件
	zipFile := "dist.zip"
	if err := s.downloadFile(client, downloadURL, zipFile); err != nil {
		fmt.Printf("download file failed: %v\n", err)
		os.Exit(1)
	}

	// 解压文件
	fmt.Println("extract file to temp directory...")
	tempDir := "dist_temp"
	if err := s.extractZip(zipFile, tempDir); err != nil {
		fmt.Printf("extract file failed: %v\n", err)
		s.cleanup(zipFile, tempDir)
		os.Exit(1)
	}

	// 复制文件到目标目录
	fmt.Printf("copy file to %s directory...\n", s.outputDir)
	if err := s.copyFiles(tempDir, s.outputDir); err != nil {
		fmt.Printf("copy file failed: %v\n", err)
		s.cleanup(zipFile, tempDir)
		os.Exit(1)
	}

	// 清理临时文件
	s.cleanup(zipFile, tempDir)

	fmt.Println("sync web completed.")
}

// createOutputDir 创建输出目录
func (s *SyncCommand) createOutputDir() error {
	if _, err := os.Stat(s.outputDir); os.IsNotExist(err) {
		return os.MkdirAll(s.outputDir, 0755)
	}
	return nil
}

// getLatestRelease 获取最新发布版本
func (s *SyncCommand) getLatestRelease(client *resty.Client) (*GitHubRelease, error) {
	resp, err := client.R().
		SetHeader("Accept", "application/vnd.github.v3+json").
		Get("https://api.github.com/repos/toheart/goanalysis-web/releases/latest")

	if err != nil {
		return nil, fmt.Errorf("request github api failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("github api return error status code: %d", resp.StatusCode())
	}

	var release GitHubRelease
	if err := json.Unmarshal(resp.Body(), &release); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if release.TagName == "" {
		return nil, fmt.Errorf("no valid release found")
	}

	return &release, nil
}

// findZipDownloadURL 查找zip文件下载链接
func (s *SyncCommand) findZipDownloadURL(release *GitHubRelease) (string, error) {
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no zip file download url found")
}

// downloadFile 下载文件
func (s *SyncCommand) downloadFile(client *resty.Client, url, filename string) error {
	resp, err := client.R().
		SetOutput(filename).
		Get(url)

	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("download return error status code: %d", resp.StatusCode())
	}

	return nil
}

// extractZip 解压zip文件
func (s *SyncCommand) extractZip(src, dest string) error {
	// 删除可能存在的临时目录
	os.RemoveAll(dest)

	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// 解压文件
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, file.Name)

		// 确保路径安全
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			rc.Close()
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			rc.Close()
			continue
		}

		// 创建文件目录
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			rc.Close()
			return err
		}

		// 创建文件
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// copyFiles 复制文件
func (s *SyncCommand) copyFiles(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// 目标路径
		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// 复制文件
		return s.copyFile(path, destPath)
	})
}

// copyFile 复制单个文件
func (s *SyncCommand) copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// cleanup 清理临时文件
func (s *SyncCommand) cleanup(files ...string) {
	for _, file := range files {
		os.RemoveAll(file)
	}
}
