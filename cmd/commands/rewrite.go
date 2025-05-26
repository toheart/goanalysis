package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
	"github.com/toheart/goanalysis/internal/biz/rewrite"
)

// RewriteCommand 代码重写命令
type RewriteCommand struct {
	cmdbase.BaseCommand
	dir string
}

// NewRewriteCommand 创建代码重写命令
func NewRewriteCommand() *RewriteCommand {
	cmd := &RewriteCommand{}
	cmd.CobraCmd = &cobra.Command{
		Use:   "rewrite --dir [directory]",
		Short: "rewrite code, you can use --dir to specify the directory",
		Run:   cmd.Run,
	}
	return cmd
}

// Init 初始化代码重写命令
func (r *RewriteCommand) Init() {
	r.CobraCmd.Flags().StringVarP(&r.dir, "dir", "d", "", "specify the directory")
}

// Run 执行代码重写命令
func (r *RewriteCommand) Run(cmd *cobra.Command, args []string) {
	if r.dir == "" {
		fmt.Println("请指定目录")
		return
	}
	rewrite.RewriteDir(r.dir)
}
