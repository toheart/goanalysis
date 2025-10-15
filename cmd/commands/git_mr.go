package commands

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"github.com/toheart/goanalysis/cmd/cmdbase"
	"github.com/toheart/goanalysis/internal/biz/gitanalysis"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
	einoprovider "github.com/toheart/goanalysis/internal/data/llm/eino"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// MRCommand MR分析命令
type MRCommand struct {
	cmdbase.BaseCommand
	mrID       int
	projectID  int
	outputFile string
	flagconf   string
	autoNotes  bool
	mrURL      string
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
	m.CobraCmd.Flags().StringVarP(&m.mrURL, "url", "u", "", "gitlab MR url, e.g. https://gitlab.example.com/group/proj/-/merge_requests/123")
	m.CobraCmd.Flags().StringVarP(&m.outputFile, "output", "o", "", "output file path(JSON format)")
	m.CobraCmd.Flags().StringVarP(&m.flagconf, "conf", "c", "./configs/config.yaml", "config file path")
	m.CobraCmd.Flags().BoolVarP(&m.autoNotes, "auto-notes", "a", false, "auto create notes")
}

// Run 执行MR分析命令
func (m *MRCommand) Run(cmd *cobra.Command, args []string) {
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

	// 如果提供了 MR URL，尝试解析 projectId 与 mrId（优先级高于命令行数值）
	// 中文注释：从 Web URL 提取项目路径与 MR IID，再通过 API 将路径解析为数字项目ID
	if m.mrURL != "" {
		if projPath, iid, err := parseMRURL(m.mrURL); err == nil {
			// 仅在未显式提供时填充
			if m.mrID == 0 {
				m.mrID = iid
			}
			if m.projectID == 0 && projPath != "" {
				cli, err := gitlab.NewClient(bc.Biz.Gitlab.Token, gitlab.WithBaseURL(bc.Biz.Gitlab.Url))
				if err == nil {
					if p, _, e := cli.Projects.GetProject(projPath, &gitlab.GetProjectOptions{}); e == nil {
						m.projectID = p.ID
					} else {
						fmt.Printf("resolve project id from path failed: %v\n", e)
					}
				} else {
					fmt.Printf("create gitlab client failed: %v\n", err)
				}
			}
		} else {
			fmt.Printf("parse gitlab MR url failed: %v\n", err)
		}
	}

	// 参数校验：允许使用 URL 自动推断
	if m.mrID == 0 {
		fmt.Println("please provide gitlab mr id or a valid MR url")
		return
	}
	if m.projectID == 0 {
		fmt.Println("please provide gitlab project id or a valid MR url")
		return
	}

	data := data.NewData(log.DefaultLogger)
	// 中文注释：使用 EinoProvider，允许第三方通过 WithModel 形式注入自定义 LLM
	provider, err := einoprovider.NewEinoProvider(
		bc.Biz.Openai.ApiKey,
		bc.Biz.Openai.ApiBase,
		bc.Biz.Openai.Model,
	)
	if err != nil {
		fmt.Printf("init eino provider failed: %v\n", err)
		os.Exit(1)
	}
	gitAnalysis := gitanalysis.NewGitAnalysis(bc.Biz, data, provider)
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

// parseMRURL 解析 GitLab MR Web URL，提取项目路径与 MR IID
// 例：
//
//	https://gitlab.example.com/group/sub/project/-/merge_requests/123
//	https://gitlab.example.com/group/sub/project/merge_requests/123
func parseMRURL(raw string) (string, int, error) {
	// 中文注释：容错处理 query/fragment，主要依赖 path 结构
	u, err := url.Parse(raw)
	if err != nil {
		return "", 0, err
	}
	segs := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(segs) < 2 {
		return "", 0, fmt.Errorf("invalid mr url path")
	}
	// 找到 merge_requests 段
	mrIdx := -1
	for i, s := range segs {
		if s == "merge_requests" || s == "merge-request" {
			mrIdx = i
			break
		}
	}
	if mrIdx == -1 || mrIdx+1 >= len(segs) {
		return "", 0, fmt.Errorf("merge_requests segment not found")
	}
	// IID 在 merge_requests 后一段
	iid, err := strconv.Atoi(segs[mrIdx+1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid mr iid")
	}
	// 项目路径：从起始到 merge_requests 之前；若前一段为 "-"，则排除之
	end := mrIdx
	if end > 0 && segs[end-1] == "-" {
		end = end - 1
	}
	if end <= 0 {
		return "", 0, fmt.Errorf("project path not found")
	}
	projPath := strings.Join(segs[:end], "/")
	return projPath, iid, nil
}
