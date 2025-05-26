package commands

import (
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
)

// GitCommand Git集成命令
type GitCommand struct {
	cmdbase.BaseCommand
}

// NewGitCommand 创建Git命令
func NewGitCommand() *GitCommand {
	cmd := &GitCommand{}
	cmd.CobraCmd = &cobra.Command{
		Use:   "git",
		Short: "Git integration commands",
		Long:  `Commands that integrate with Git`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	return cmd
}

// Init 初始化Git命令
func (g *GitCommand) Init() {
	// 注册MR命令到Git命令
	mrCmd := NewMRCommand()
	mrCmd.Init()

	g.CobraCmd.AddCommand(mrCmd.GetCobraCmd())
}

// ExecGitCommand 执行一个git命令
func (g *GitCommand) ExecGitCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	return cmd.Output()
}
