package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
	"github.com/toheart/goanalysis/internal/biz/callgraph"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
)

// CallGraphCommand 调用图命令
type CallGraphCommand struct {
	cmdbase.BaseCommand
	codeDir    string
	outputPath string
	cachePath  string
	isCache    bool
	onlyMethod string
	algo       string
	flagconf   string
}

// NewCallGraphCommand 创建调用图命令
func NewCallGraphCommand() *CallGraphCommand {
	cmd := &CallGraphCommand{}
	cmd.CobraCmd = &cobra.Command{
		Use:   "callgraph",
		Short: "generate call graph",
		Long:  `This command is used to generate a call graph to help users analyze function call relationships.`,
		Run:   cmd.Run,
	}
	return cmd
}

// Init 初始化调用图命令
func (c *CallGraphCommand) Init() {
	c.CobraCmd.Flags().StringVarP(&c.codeDir, "dir", "d", "", "code directory")
	c.CobraCmd.Flags().StringVarP(&c.outputPath, "output", "o", callgraph.DefaultOutput, "Image output path,default: ./default.png")
	c.CobraCmd.Flags().StringVarP(&c.cachePath, "cache", "c", callgraph.DefaultCache, "FuncNode cache output path,default: ./cache.json")
	c.CobraCmd.Flags().StringVarP(&c.onlyMethod, "method", "m", "", "Only output relevant package names and method names")
	c.CobraCmd.Flags().StringVarP(&c.algo, "algo", "a", callgraph.CallGraphTypeRta, fmt.Sprintf("The algorithm used to construct the call graph. Possible values inlcude: %q, %q, %q, %q, default: %q",
		callgraph.CallGraphTypeVta, callgraph.CallGraphTypeStatic, callgraph.CallGraphTypeCha, callgraph.CallGraphTypeRta, callgraph.CallGraphTypeVta))
	c.CobraCmd.Flags().BoolVarP(&c.isCache, "isCache", "i", true, "Whether to enable caching, default true")
	c.CobraCmd.Flags().StringVar(&c.flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

// Run 执行调用图命令
func (c *CallGraphCommand) Run(cmd *cobra.Command, args []string) {
	if c.codeDir == "" {
		fmt.Println("Please provide code directory")
		return
	}

	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)

	kconf := config.New(
		config.WithSource(
			file.NewSource(c.flagconf),
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
	fileName := filepath.Base(c.codeDir)
	if fileName == "." || fileName == "/" {
		// 处理路径末尾有斜杠的情况
		fileName = filepath.Base(filepath.Dir(c.codeDir))
	}

	dbPath := filepath.Join(entity.GetFileStoragePath(bc.Biz.FileStoragePath, false), fileName)
	funcNodeDB, err := db.GetFuncNodeDB(dbPath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting to generate call graph for %s...\n", c.codeDir)
	fmt.Println("Using legacy ProgramAnalysis (更可靠的业务分析器)")
	fmt.Printf("Algorithm: %s\n", c.algo)
	fmt.Println("Note: Large projects may take several minutes to analyze...")

	// 使用旧的业务分析器而不是标准分析器
	options := []callgraph.ProgramOption{
		callgraph.WithAlgo(c.algo),
		callgraph.WithOutputDir(c.outputPath),
		callgraph.WithCacheDir(c.cachePath),
		callgraph.WithCacheFlag(c.isCache),
	}
	if c.onlyMethod != "" {
		options = append(options, callgraph.WithOnlyPkg(c.onlyMethod))
	}

	// 创建程序分析实例
	pa := callgraph.NewProgramAnalysis(c.codeDir, log.NewHelper(logger), funcNodeDB, options...)

	// 执行完整的分析流程
	statusChan := make(chan []byte, 100)
	go func() {
		for status := range statusChan {
			fmt.Printf("Status: %s\n", string(status))
		}
	}()

	ctx := context.Background()
	err = pa.Execute(ctx, statusChan)
	if err != nil {
		fmt.Printf("Analysis failed: %v\n", err)
		os.Exit(1)
	}

	close(statusChan)
	fmt.Println("Call graph generation completed using legacy ProgramAnalysis")
}
