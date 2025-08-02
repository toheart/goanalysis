package commands

import (
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/pkg/logger"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func newApp(logger log.Logger, servers []transport.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(servers...),
	)
}

// ServerCommand 服务器命令
type ServerCommand struct {
	cmdbase.BaseCommand
	flagconf string
	id       string
	Name     string
	Version  string
}

// NewServerCommand 创建服务器命令
func NewServerCommand() *ServerCommand {
	hostname, _ := os.Hostname()
	cmd := &ServerCommand{
		id: hostname,
	}
	cmd.CobraCmd = &cobra.Command{
		Use:   "server",
		Short: "启动 GoAnalysis 服务器",
		Long: `启动 GoAnalysis 服务器，支持通过命令行参数、环境变量或配置文件进行配置。

配置优先级：命令行参数 > 环境变量 > 配置文件 > 默认值

示例：
  # 使用默认配置启动
  goanalysis server
  
  # 自定义端口和日志级别
  goanalysis server --http-addr=0.0.0.0:8080 --log-level=info
  
  # 使用配置文件
  goanalysis server --conf=configs/config.yaml`,
		Run: cmd.Run,
	}
	return cmd
}

// Init 初始化服务器命令
func (s *ServerCommand) Init() {
	// 配置文件参数
	s.CobraCmd.Flags().StringVar(&s.flagconf, "conf", "", "配置文件路径，例如: -conf config.yaml")

	// 服务器配置参数
	s.CobraCmd.Flags().String("http-addr", "0.0.0.0:8001", "HTTP服务地址")
	s.CobraCmd.Flags().String("grpc-addr", "0.0.0.0:9000", "gRPC服务地址")
	s.CobraCmd.Flags().String("http-timeout", "1s", "HTTP超时时间")
	s.CobraCmd.Flags().String("grpc-timeout", "1s", "gRPC超时时间")

	// 日志配置参数
	s.CobraCmd.Flags().String("log-level", "debug", "日志级别 (debug, info, warn, error)")
	s.CobraCmd.Flags().String("log-file", "./logs/app.log", "日志文件路径")
	s.CobraCmd.Flags().Bool("log-console", true, "是否输出到控制台")
	s.CobraCmd.Flags().Int32("log-max-size", 100, "日志文件最大大小(MB)")
	s.CobraCmd.Flags().Int32("log-max-age", 7, "日志保留天数")
	s.CobraCmd.Flags().Int32("log-max-backups", 10, "保留的日志文件数量")
	s.CobraCmd.Flags().Bool("log-compress", true, "是否压缩日志文件")

	// GitLab配置参数
	s.CobraCmd.Flags().String("gitlab-token", "", "GitLab访问令牌 (也可通过 GITLAB_TOKEN 环境变量设置)")
	s.CobraCmd.Flags().String("gitlab-url", "", "GitLab API地址 (也可通过 GITLAB_API_URL 环境变量设置)")
	s.CobraCmd.Flags().String("gitlab-clone-dir", "./data", "GitLab克隆目录")

	// OpenAI配置参数
	s.CobraCmd.Flags().String("openai-api-key", "", "OpenAI API密钥 (也可通过 OPENAI_API_KEY 环境变量设置)")
	s.CobraCmd.Flags().String("openai-api-base", "", "OpenAI API地址 (也可通过 OPENAI_API_BASE 环境变量设置)")
	s.CobraCmd.Flags().String("openai-model", "", "OpenAI模型名称 (也可通过 OPENAI_MODEL 环境变量设置)")

	// 存储路径配置参数
	s.CobraCmd.Flags().String("static-store-path", "./data/static", "静态分析存储路径")
	s.CobraCmd.Flags().String("runtime-store-path", "./data/runtime", "运行时分析存储路径")
	s.CobraCmd.Flags().String("file-storage-path", "./data/files", "文件存储路径")

	// 数据配置参数
	s.CobraCmd.Flags().String("db-path", "./goanalysis.db", "数据库文件路径")
}

// newApp 创建Kratos应用
func (s *ServerCommand) newApp(logger log.Logger, servers []transport.Server) *kratos.App {
	return kratos.New(
		kratos.ID(s.id),
		kratos.Name(s.Name),
		kratos.Version(s.Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(servers...),
	)
}

// Run 执行服务器命令
func (s *ServerCommand) Run(cmd *cobra.Command, args []string) {
	// 使用新的配置加载器
	bc, err := conf.LoadConfig(cmd)
	if err != nil {
		panic(err)
	}

	// 验证配置
	if err := conf.ValidateConfig(bc); err != nil {
		panic(err)
	}

	// 启动服务器的逻辑
	log := logger.NewLogger(bc.Logger)
	// 使用原始cmd包中的wireApp函数
	app, cleanup, err := wireApp(bc.Server, bc.Biz, bc.Data, log)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
