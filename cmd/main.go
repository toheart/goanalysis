package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goanalysis",
	Short: "goanalysis is a tool for analyzing Go applications",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Usage: goanalysis [command] [options]")
		fmt.Println("Available commands:")
		fmt.Println("  callgraph - Generate call graph")
		fmt.Println("  rewrite - Rewrite code")
		fmt.Println("  server - Start the server")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
