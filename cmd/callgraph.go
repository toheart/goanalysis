package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/internal/biz/callgraph"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
)

var codeDir string
var outputPath string
var cachePath string
var isCache bool
var onlyMethod string
var algo string

var callGraphCmd = &cobra.Command{
	Use:   "callgraph",
	Short: "generate call graph",
	Long:  `This command is used to generate a call graph to help users analyze function call relationships.`,
	Run: func(cmd *cobra.Command, args []string) {
		if codeDir == "" {
			fmt.Println("please provide the code directory.")
			return
		}
		logger := log.With(log.NewStdLogger(os.Stdout),
			"ts", log.DefaultTimestamp,
			"caller", log.DefaultCaller,
		)
		kconf := config.New(
			config.WithSource(
				file.NewSource(flagconf),
			),
		)
		defer kconf.Close()

		if err := kconf.Load(); err != nil {
			panic(err)
		}

		var bc conf.Bootstrap
		if err := kconf.Scan(&bc); err != nil {
			panic(err)
		}
		db := data.NewData(logger)
		// 获取codeDir最后一个目录
		dir := filepath.Base(codeDir)
		if dir == "." || dir == "/" {
			// 处理路径末尾有斜杠的情况
			dir = filepath.Base(filepath.Dir(codeDir))
		}
		dbPath := filepath.Join(bc.Biz.StaticDBpath, dir)
		funcNodeDB, err := db.GetFuncNodeDB(dbPath)
		if err != nil {
			panic(err)
		}
		c := callgraph.NewProgramAnalysis(codeDir, logger, funcNodeDB, callgraph.WithOutputDir(outputPath),
			callgraph.WithCacheDir(cachePath), callgraph.WithOnlyPkg(onlyMethod), callgraph.WithAlgo(algo), callgraph.WithCacheFlag(isCache))
		// 这里可以添加生成调用图的逻辑
		fmt.Printf("start to generate call graph for %s...\n", codeDir)

		go func() {
			if err := c.SetTree(); err != nil {
				panic(err)
			}
		}()
		c.SaveData(context.Background())
	},
}

func init() {
	callGraphCmd.Flags().StringVarP(&codeDir, "dir", "d", "", "code directory")
	callGraphCmd.Flags().StringVarP(&outputPath, "output", "o", callgraph.DefaultOutput, "Image output path,default: ./default.png")
	callGraphCmd.Flags().StringVarP(&cachePath, "cache", "c", callgraph.DefaultCache, "FuncNode cache output path,default: ./cache.json")
	callGraphCmd.Flags().StringVarP(&onlyMethod, "method", "m", "", "Only output relevant package names and method names")
	callGraphCmd.Flags().StringVarP(&algo, "algo", "a", callgraph.CallGraphTypeRta, fmt.Sprintf("The algorithm used to construct the call graph. Possible values inlcude: %q, %q, %q, %q, default: %q",
		callgraph.CallGraphTypeVta, callgraph.CallGraphTypeStatic, callgraph.CallGraphTypeCha, callgraph.CallGraphTypeRta, callgraph.CallGraphTypeVta))
	callGraphCmd.Flags().BoolVarP(&isCache, "isCache", "i", true, "Whether to enable caching, default true")
	rootCmd.AddCommand(callGraphCmd)
}
