package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
)

// ConfigCommand 配置文件命令
type ConfigCommand struct {
	cmdbase.BaseCommand
	output string
}

// NewConfigCommand 创建配置文件命令
func NewConfigCommand() *ConfigCommand {
	cmd := &ConfigCommand{}
	cmd.CobraCmd = &cobra.Command{
		Use:   "config",
		Short: "generate default configuration file",
		Long:  "Generate a default configuration file with all available options",
		Run:   cmd.Run,
	}
	return cmd
}

// Init 初始化配置文件命令
func (c *ConfigCommand) Init() {
	c.CobraCmd.Flags().StringVarP(&c.output, "output", "o", "./configs/config.yaml", "output path for the configuration file")
}

// defaultConfigContent 默认配置文件内容
const defaultConfigContent = `server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s

logger:
  level: debug                     # 日志级别: debug, info, warn, error
  file_path: ./logs/app.log       # 日志文件路径
  console: true                    # 同时输出到控制台
  max_size: 100                    # 单个日志文件最大大小(MB)
  max_age: 7                      # 日志文件保留天数
  max_backups: 10                 # 保留的旧日志文件最大数量
  compress: true           
         # 压缩旧日志文件


biz:
  gitlab:
    token: "${GITLAB_TOKEN}"                # GitLab访问令牌，用于API认证
    url: "${GITLAB_API_URL}"  # GitLab API URL
    clone_dir: ./data
  openai:
    api_key: "${OPENAI_API_KEY}"
    api_base: "${OPENAI_API_BASE}"
    model: "${OPENAI_MODEL}"
  staticStorePath: ./data/static
  runtimeStorePath: ./data/runtime
  file_storage_path: ./data/files

data:
  dbpath: ./goanalysis.db
`

// Run 执行配置文件生成命令
func (c *ConfigCommand) Run(cmd *cobra.Command, args []string) {
	// 检查输出路径是否存在，如果不存在则创建目录
	outputDir := filepath.Dir(c.output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	// 检查文件是否已存在
	if _, err := os.Stat(c.output); err == nil {
		fmt.Printf("Configuration file already exists at %s\n", c.output)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	// 写入配置文件
	if err := os.WriteFile(c.output, []byte(defaultConfigContent), 0644); err != nil {
		fmt.Printf("Error writing configuration file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration file generated successfully at: %s\n", c.output)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the configuration file to match your environment")
	fmt.Println("2. Set the required environment variables:")
	fmt.Println("   - GITLAB_TOKEN: Your GitLab access token")
	fmt.Println("   - GITLAB_API_URL: Your GitLab API URL")
	fmt.Println("   - OPENAI_API_KEY: Your OpenAI API key")
	fmt.Println("   - OPENAI_API_BASE: OpenAI API base URL")
	fmt.Println("   - OPENAI_MODEL: OpenAI model name")
	fmt.Printf("3. Start the server: goanalysis server --conf=%s\n", c.output)
}
