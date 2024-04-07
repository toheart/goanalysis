/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/pkg/analysis"
	"log"
)

var (
	ignoreFlag string
	onlyMethod string
	algoFlag   string
	outputPath string
	cachePath  string

	cacheFlag bool
)

// analysisCmd represents the analysis command
var analysisCmd = &cobra.Command{
	Use:   "analysis [options] <programDir>",
	Short: "A brief description of your command",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("please input program dir")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("analysis dir: %s", args[0])
		a := analysis.NewProgramAnalysis(args[0],
			analysis.WithAlgo(algoFlag),
			analysis.WithIgnorePaths(ignoreFlag),
			analysis.WithOnlyPkg(onlyMethod),
			analysis.WithOutputDir(outputPath),
			analysis.WithCacheDir(cachePath),
			analysis.WithCacheFlag(cacheFlag),
		)
		a.Print()
	},
}

func init() {
	analysisCmd.Flags().StringVarP(&algoFlag, "algo", "a", "vta", fmt.Sprintf("The algorithm used to construct the call graph. Possible values inlcude: %q, %q, %q, %q, default: %q",
		analysis.CallGraphTypeVta, analysis.CallGraphTypeStatic, analysis.CallGraphTypeCha, analysis.CallGraphTypeRta, analysis.CallGraphTypeVta))
	analysisCmd.Flags().StringVarP(&ignoreFlag, "ignore", "i", "", "Ignore methods paths containing given suffix")
	analysisCmd.Flags().StringVarP(&onlyMethod, "onlyMethod", "p", "", "Only output relevant package names and method names")
	analysisCmd.Flags().StringVarP(&outputPath, "outputPath", "o", analysis.DefaultOutput, "Image output path,default: ./default.png")
	analysisCmd.Flags().StringVarP(&cachePath, "cachePath", "c", analysis.DefaultCache, "FuncNode cache output path,default: ./cache.json")
	analysisCmd.Flags().BoolVar(&cacheFlag, "cacheFlag", true, "Whether to enable caching, default true")
	rootCmd.AddCommand(analysisCmd)
}
