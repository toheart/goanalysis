package main

import (
	"fmt"
	"os"

	"github.com/toheart/goanalysis/cmd/server"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(server.ServerCmd)
}

var rootCmd = &cobra.Command{
	Use:   "goanalysis",
	Short: "goanalysis is a tool for analyzing Go applications",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to goanalysis!")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
