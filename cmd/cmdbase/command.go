package cmdbase

import (
	"github.com/spf13/cobra"
)

// Command 命令接口，所有命令都需要实现此接口
type Command interface {
	// Init 初始化命令，设置标志等
	Init()
	// GetCobraCmd 获取cobra命令
	GetCobraCmd() *cobra.Command
}

// BaseCommand 基础命令结构体，提供通用实现
type BaseCommand struct {
	CobraCmd *cobra.Command
}

// GetCobraCmd 获取cobra命令
func (b *BaseCommand) GetCobraCmd() *cobra.Command {
	return b.CobraCmd
}

// CommandRegistry 命令注册表，保存所有注册的命令
type CommandRegistry struct {
	rootCmd  *cobra.Command
	commands []Command
}

// NewCommandRegistry 创建一个命令注册表
func NewCommandRegistry(rootCmd *cobra.Command) *CommandRegistry {
	return &CommandRegistry{
		rootCmd:  rootCmd,
		commands: make([]Command, 0),
	}
}

// Register 注册命令
func (r *CommandRegistry) Register(cmd Command) {
	cmd.Init()
	r.commands = append(r.commands, cmd)
	r.rootCmd.AddCommand(cmd.GetCobraCmd())
}

// RegisterCommands 注册所有命令
func (r *CommandRegistry) RegisterCommands() {
	// 已经在Register方法中注册，此方法仅为向后兼容
}

// Execute 执行根命令
func (r *CommandRegistry) Execute() error {
	return r.rootCmd.Execute()
}
