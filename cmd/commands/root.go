package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
)

// RootCommand 根命令实现
type RootCommand struct {
	cmdbase.BaseCommand
}

// NewRootCommand 创建根命令
func NewRootCommand() *RootCommand {
	cmd := &RootCommand{}
	cmd.CobraCmd = &cobra.Command{
		Use:   "goanalysis",
		Short: "goanalysis is a tool for analyzing Go applications",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Help())
		},
	}
	return cmd
}

// Init 初始化根命令
func (r *RootCommand) Init() {
	// 根命令初始化不需要额外操作
}

// Execute 执行根命令
func (r *RootCommand) Execute() {
	if err := r.CobraCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
