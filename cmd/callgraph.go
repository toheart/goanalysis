package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/internal/biz/callgraph"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
)

var (
	codeDir    string
	outputPath string
	cachePath  string
	isCache    bool
	onlyMethod string
	algo       string
)

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
		c := callgraph.NewProgramAnalysis(codeDir, log.NewHelper(log.With(logger, "module", "callgraph")), funcNodeDB, callgraph.WithOutputDir(outputPath),
			callgraph.WithCacheDir(cachePath), callgraph.WithOnlyPkg(onlyMethod), callgraph.WithAlgo(algo), callgraph.WithCacheFlag(isCache))

		// 创建一个命令行状态通道，用于接收状态更新
		statusChan := make(chan []byte, 100)

		// 创建一个goroutine来处理状态更新
		go func() {
			for msg := range statusChan {
				fmt.Println(string(msg))
			}
		}()

		// 这里可以添加生成调用图的逻辑
		fmt.Printf("开始为 %s 生成调用图...\n", codeDir)

		// 使用 WaitGroup 等待所有任务完成
		var wg sync.WaitGroup
		wg.Add(1)

		// 设置完成标志
		var completed bool
		var mu sync.Mutex

		// 启动调用图生成
		go func() {
			defer wg.Done()
			if err := c.SetTree(statusChan); err != nil {
				errMsg := fmt.Sprintf("调用图生成失败: %v", err)
				fmt.Println(errMsg)
				return
			}

			// 保存数据
			if err := c.SaveData(context.Background(), statusChan); err != nil {
				errMsg := fmt.Sprintf("保存数据失败: %v", err)
				fmt.Println(errMsg)
				return
			}

			// 标记为完成
			mu.Lock()
			completed = true
			mu.Unlock()

			fmt.Println("分析任务完成")
		}()

		// 等待任务完成
		wg.Wait()

		// 检查是否成功完成
		mu.Lock()
		defer mu.Unlock()
		if !completed {
			fmt.Println("分析任务未成功完成")
			os.Exit(1)
		}

		fmt.Println("调用图生成完成")
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
