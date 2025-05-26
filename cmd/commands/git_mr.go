package commands

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
	"github.com/toheart/goanalysis/internal/biz/gitanalysis"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
)

// MRCommand MR分析命令
type MRCommand struct {
	cmdbase.BaseCommand
	mrID       int
	projectID  int
	outputFile string
	flagconf   string
	autoNotes  bool
}

// NewMRCommand 创建MR分析命令
func NewMRCommand() *MRCommand {
	cmd := &MRCommand{}
	cmd.CobraCmd = &cobra.Command{
		Use:   "mr",
		Short: "Analyze GitLab MR changes",
		Long:  `This command analyzes functions affected by a GitLab Merge Request (MR).`,
		Run:   cmd.Run,
	}
	return cmd
}

// Init 初始化MR分析命令
func (m *MRCommand) Init() {
	m.CobraCmd.Flags().IntVarP(&m.mrID, "mr", "m", 0, "gitlab mr id")
	m.CobraCmd.Flags().IntVarP(&m.projectID, "project", "p", 0, "gitlab project id")
	m.CobraCmd.Flags().StringVarP(&m.outputFile, "output", "o", "", "output file path(JSON format)")
	m.CobraCmd.Flags().StringVarP(&m.flagconf, "conf", "c", "./configs/config.yaml", "config file path")
	m.CobraCmd.Flags().BoolVarP(&m.autoNotes, "auto-notes", "a", false, "auto create notes")
}

// Run 执行MR分析命令
func (m *MRCommand) Run(cmd *cobra.Command, args []string) {
	if m.mrID == 0 {
		fmt.Println("please provide gitlab mr id")
		return
	}

	if m.projectID == 0 {
		fmt.Println("please provide gitlab project id")
		return
	}

	_ = log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)

	// 加载配置
	kconf := config.New(
		config.WithSource(
			env.NewSource(""),
			file.NewSource(m.flagconf),
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

	data := data.NewData(log.DefaultLogger)
	gitAnalysis := gitanalysis.NewGitAnalysis(bc.Biz, data)
	// 执行 MR 分析
	fmt.Printf("start analyzing gitlab mr #%d...\n", m.mrID)
	result, err := gitAnalysis.MRAnalyzer(m.projectID, m.mrID, m.autoNotes)
	if err != nil {
		fmt.Printf("analyze failed: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	if m.outputFile != "" {
		if err := result.SaveToFile(m.outputFile); err != nil {
			fmt.Printf("save result to file failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("analyze result saved to: %s\n", m.outputFile)
	} else {
		// 直接输出到控制台
		result.PrintToConsole()
	}

	fmt.Println("analyze completed")
}
