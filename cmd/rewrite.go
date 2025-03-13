package cmd

import (
	"fmt"

	"github.com/toheart/goanalysis/internal/biz/rewrite"

	"github.com/spf13/cobra"
)

var rewriteCmd = &cobra.Command{
	Use:   "rewrite --dir [directory]",
	Short: "rewrite code, you can use --dir to specify the directory",
	Run: func(cmd *cobra.Command, args []string) {
		dir := cmd.Flag("dir").Value.String()
		if dir == "" {
			fmt.Println("please specify the directory")
			return
		}
		rewrite.RewriteDir(dir)
	},
}

func init() {
	rewriteCmd.Flags().StringP("dir", "d", "", "specify the directory")
	rootCmd.AddCommand(rewriteCmd)
}
