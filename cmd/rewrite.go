package main

import (
	"github.com/toheart/goanalysis/internal/biz/rewrite"

	"github.com/spf13/cobra"
)

var rewriteCmd = &cobra.Command{
	Use:   "rewrite --dir [directory]",
	Short: "重写命令",
	Args:  cobra.ExactArgs(1), // 确保只接受一个参数
	Run: func(cmd *cobra.Command, args []string) {
		dir := cmd.Flag("dir").Value.String()
		rewrite.RewriteDir(dir)
	},
}

func init() {
	rewriteCmd.Flags().StringP("dir", "d", "", "指定目录")
	rootCmd.AddCommand(rewriteCmd)
}

func init() {
	rootCmd.AddCommand(rewriteCmd)
}
