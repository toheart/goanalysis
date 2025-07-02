package cmd

import (
	"github.com/toheart/goanalysis/cmd/cmdbase"
	"github.com/toheart/goanalysis/cmd/commands"
)

var (
	// Registry 命令注册表
	Registry *cmdbase.CommandRegistry
)

func init() {
	// 创建根命令
	rootCmd := commands.NewRootCommand()
	// 创建命令注册表
	Registry = cmdbase.NewCommandRegistry(rootCmd.GetCobraCmd())
	// 注册所有命令
	RegisterAllCommands()
}

// RegisterAllCommands 注册所有命令
func RegisterAllCommands() {
	// 注册服务器命令
	Registry.Register(commands.NewServerCommand())

	// 注册配置文件生成命令
	Registry.Register(commands.NewConfigCommand())

	// 注册调用图命令
	Registry.Register(commands.NewCallGraphCommand())

	// 注册重写命令
	Registry.Register(commands.NewRewriteCommand())

	// 注册Git命令
	Registry.Register(commands.NewGitCommand())

	// 注册同步命令
	Registry.Register(commands.NewSyncCommand())

}

// Execute 执行根命令
func Execute() error {
	return Registry.Execute()
}
