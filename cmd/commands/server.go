package commands

import (
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
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
		Short: "start the server",
		Run:   cmd.Run,
	}
	return cmd
}

// Init 初始化服务器命令
func (s *ServerCommand) Init() {
	s.CobraCmd.Flags().StringVar(&s.flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
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
	c := config.New(
		config.WithSource(
			file.NewSource(s.flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
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
