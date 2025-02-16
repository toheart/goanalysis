package main

import (
	"fmt"

	"github.com/toheart/goanalysis/internal/biz/rewrite"

	"github.com/spf13/cobra"
)

var rewriteCmd = &cobra.Command{
	Use:   "rewrite --dir [directory]",
	Short: "重写命令",
	Run: func(cmd *cobra.Command, args []string) {
		dir := cmd.Flag("dir").Value.String()
		if dir == "" {
			fmt.Println("请指定目录")
			return
		}
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
