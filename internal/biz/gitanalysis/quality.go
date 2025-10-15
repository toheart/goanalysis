package gitanalysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// 中文注释：统一的质量问题结构
type QualityFinding struct {
	File       string
	Line       int
	Rule       string
	Severity   string
	Message    string
	Suggestion string
	Source     string // tool name
}

// 中文注释：golangci-lint JSON 输出的部分字段
type golangciIssue struct {
	FromLinter string `json:"from_linter"`
	Text       string `json:"text"`
	Severity   string `json:"severity"`
	Pos        struct {
		Filename string `json:"filename"`
		Line     int    `json:"line"`
	} `json:"pos"`
}

type golangciOutput struct {
	Issues []golangciIssue `json:"Issues"`
}

// RunGolangCILint 在 repoRoot 内对 files 运行 golangci-lint 并解析 JSON 结果
// 中文注释：若未安装或执行失败，不中断流程，返回空结果与可观测日志
func RunGolangCILint(ctx context.Context, repoRoot string, files []string, timeout time.Duration, logger *log.Helper) ([]QualityFinding, error) {
	// 中文注释：仅处理 .go 文件
	var goFiles []string
	for _, f := range files {
		if strings.HasSuffix(f, ".go") {
			goFiles = append(goFiles, filepath.FromSlash(f))
		}
	}
	if len(goFiles) == 0 {
		return nil, nil
	}

	// 中文注释：构造命令
	args := append([]string{"run", "--out-format", "json"}, goFiles...)
	cmd := exec.CommandContext(ctx, "golangci-lint", args...)
	cmd.Dir = repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 中文注释：超时上下文
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, "golangci-lint", args...)
		cmd.Dir = repoRoot
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		// golangci-lint 返回非零也可能有 JSON 输出，先尝试解析；无输出再告警
		if stdout.Len() == 0 {
			logger.Warnf("golangci-lint failed: %v, stderr: %s", err, strings.TrimSpace(stderr.String()))
			return nil, nil
		}
	}

	var out golangciOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		logger.Warnf("parse golangci-lint json failed: %v", err)
		return nil, nil
	}
	findings := make([]QualityFinding, 0, len(out.Issues))
	for _, is := range out.Issues {
		sev := strings.ToUpper(strings.TrimSpace(is.Severity))
		if sev == "" {
			// 中文注释：MVP 将未知严重级别映射为 P1
			sev = "P1"
		}
		findings = append(findings, QualityFinding{
			File:       filepath.ToSlash(is.Pos.Filename),
			Line:       is.Pos.Line,
			Rule:       is.FromLinter,
			Severity:   sev,
			Message:    is.Text,
			Suggestion: "",
			Source:     "golangci-lint",
		})
	}
	return findings, nil
}

// BuildQualitySummaryMD 生成质量检查的汇总 Markdown（按严重级别排序，限制条数）
func BuildQualitySummaryMD(findings []QualityFinding, maxItems int) string {
	if len(findings) == 0 {
		return "## 质量报告\n\n无发现问题\n"
	}
	// 中文注释：按严重级别与规则排序，P0>P1>P2
	rank := func(sev string) int {
		switch strings.ToUpper(sev) {
		case "P0":
			return 0
		case "P1":
			return 1
		default:
			return 2
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		ri, rj := rank(findings[i].Severity), rank(findings[j].Severity)
		if ri != rj {
			return ri < rj
		}
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Rule < findings[j].Rule
	})

	// 统计
	total, p0, p1, p2 := len(findings), 0, 0, 0
	for _, f := range findings {
		switch strings.ToUpper(f.Severity) {
		case "P0":
			p0++
		case "P1":
			p1++
		default:
			p2++
		}
	}

	if maxItems <= 0 {
		maxItems = 50
	}
	var b strings.Builder
	b.WriteString("## 质量报告\n\n")
	b.WriteString(fmt.Sprintf("- 总问题数: %d\n", total))
	b.WriteString(fmt.Sprintf("- P0: %d, P1: %d, P2: %d\n\n", p0, p1, p2))
	b.WriteString("### Top 问题\n\n")
	limit := maxItems
	if len(findings) < limit {
		limit = len(findings)
	}
	for i := 0; i < limit; i++ {
		f := findings[i]
		// 中文注释：格式：- [P?][rule] file:line - message
		b.WriteString(fmt.Sprintf("- [%s][%s] %s:%d - %s\n", strings.ToUpper(f.Severity), f.Rule, f.File, f.Line, escapeMd(f.Message)))
	}
	if limit < len(findings) {
		b.WriteString(fmt.Sprintf("\n其余 %d 条问题已省略。\n", len(findings)-limit))
	}
	b.WriteString("\n> 注：本报告仅针对本次 MR 变更文件，建议优先处理 P0/P1。\n")
	return b.String()
}

// 中文注释：简单转义 Markdown 特殊字符，避免渲染异常
func escapeMd(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}
