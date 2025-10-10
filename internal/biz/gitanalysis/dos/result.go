package dos

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ChangeType string

const (
	Added    ChangeType = "added"
	Deleted  ChangeType = "deleted"
	Modified ChangeType = "modified"
)

// MrAnalysisResult MR分析结果
type MrAnalysisResult struct {
	MergeRequestID    int                `json:"merge_request_id"`
	ProjectID         int                `json:"project_id"`
	AffectedFiles     []AffectedFile     `json:"affected_files"`
	AffectedFunctions []AffectedFunction `json:"affected_functions"`
	Review            string             `json:"review"`
	Impact            ImpactAnalysis     `json:"impact"`
	TestSuggestions   []TestSuggestion   `json:"test_suggestions"`
	TestCases         []TestCase         `json:"test_cases,omitempty"`
}

// AffectedFile 受影响的文件
type AffectedFile struct {
	Filename   string     `json:"file"`
	ChangeType ChangeType `json:"change_type"` // "added", "deleted", "modified"
}

// AffectedFunction 受影响的函数
type AffectedFunction struct {
	Filename   string      `json:"file"`
	ChangeType ChangeType  `json:"change_type"`
	Functions  []*Function `json:"functions"`
}

// ImpactAnalysis 影响指标
type ImpactAnalysis struct {
	TotalChangedFiles     int `json:"total_changed_files"`
	ValidChangedFunctions int `json:"valid_changed_functions"`
	UpstreamCallers       int `json:"upstream_callers"`
}

// 质量评分已移除

// TestSuggestion 测试建议
type TestSuggestion struct {
	Title            string   `json:"title"`
	Reason           string   `json:"reason"`
	Priority         string   `json:"priority"`
	RelatedFunctions []string `json:"related_functions"`
}

// SaveToFile 将分析结果保存到文件
func (r *MrAnalysisResult) SaveToFile(filepath string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// PrintToConsole 将分析结果输出到控制台
func (r *MrAnalysisResult) PrintToConsole() {
	fmt.Printf("MR #%d \n", r.MergeRequestID)
	fmt.Printf("project id: %d\n", r.ProjectID)
	fmt.Println("\n-------------------------------")
	fmt.Println("add new functions:")
	for _, function := range r.AffectedFunctions {
		if function.ChangeType == Added {
			fmt.Printf("- file: %s (%s)\n", function.Filename, function.ChangeType)
			for _, f := range function.Functions {
				if !f.IsValid {
					fmt.Printf("    * %s\t %s\t %s\n", f.FunctionName, f.Reason, f.Suggestion)
				}
			}
		}
	}
	fmt.Println("\n-------------------------------")
	fmt.Println("affected functions:")
	for _, function := range r.AffectedFunctions {
		if function.ChangeType == Modified {
			fmt.Printf("- file: %s (%s)\n", function.Filename, function.ChangeType)
			fmt.Println("   functions:")
			for _, f := range function.Functions {
				if f.IsValid {
					fmt.Printf("    * %s\t %s\t %s\n", f.FunctionName, f.Reason, f.Suggestion)
					if len(f.UpstreamCallers) > 0 {
						fmt.Println("      upstream callers:")
						for _, c := range f.UpstreamCallers {
							fmt.Printf("        - %s\n", c)
						}
					}
				}
			}
		}
	}
	fmt.Println("\n-------------------------------")
	fmt.Println("no valid functions:")
	for _, function := range r.AffectedFunctions {
		fmt.Printf("- file: %s (%s)\n", function.Filename, function.ChangeType)
		for _, f := range function.Functions {
			if !f.IsValid {
				fmt.Printf("    * %s\t %s\t %s\n", f.FunctionName, f.Reason, f.Suggestion)
			}
		}
	}

	fmt.Println("\n-------------------------------")
	fmt.Println("impact:")
	fmt.Printf("- total changed files: %d\n", r.Impact.TotalChangedFiles)
	fmt.Printf("- valid changed functions: %d\n", r.Impact.ValidChangedFunctions)
	fmt.Printf("- upstream callers: %d\n", r.Impact.UpstreamCallers)

	fmt.Println("\n-------------------------------")
	// 质量评分已移除

	fmt.Println("\n-------------------------------")
	fmt.Println("review:")
	fmt.Println(r.Review)

	fmt.Println("\n-------------------------------")
	fmt.Println("test suggestions:")
	if len(r.TestSuggestions) == 0 {
		fmt.Println("(none)")
	} else {
		for _, ts := range r.TestSuggestions {
			fmt.Printf("- %s\n", ts.Title)
			if ts.Reason != "" {
				fmt.Printf("  reason: %s\n", ts.Reason)
			}
			if ts.Priority != "" {
				fmt.Printf("  priority: %s\n", ts.Priority)
			}
			if len(ts.RelatedFunctions) > 0 {
				fmt.Println("  related functions:")
				for _, rf := range ts.RelatedFunctions {
					fmt.Printf("    - %s\n", rf)
				}
			}
		}
	}

	fmt.Println("\n-------------------------------")
	fmt.Println("test cases:")
	if len(r.TestCases) == 0 {
		fmt.Println("(none)")
	} else {
		for _, tc := range r.TestCases {
			fmt.Printf("- [%s] %s | module=%s | priority=%s\n", tc.ID, tc.Title, tc.Module, tc.Priority)
		}
	}
}

// GetAffectedFiles 获取受影响的文件
func (r *MrAnalysisResult) GetAffectedFilesMd() string {
	var sb strings.Builder
	sb.WriteString("## 改动影响面\n\n")
	// 影响指标
	sb.WriteString("### 影响指标\n\n")
	sb.WriteString(fmt.Sprintf("- 变更文件数: %d\n", r.Impact.TotalChangedFiles))
	sb.WriteString(fmt.Sprintf("- 有效变更函数数: %d\n", r.Impact.ValidChangedFunctions))
	sb.WriteString(fmt.Sprintf("- 上游调用者数: %d\n\n", r.Impact.UpstreamCallers))

	// 质量评分已移除
	sb.WriteString("\n")

	sb.WriteString("### 受影响的文件\n\n")

	for _, function := range r.AffectedFunctions {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", function.Filename, function.ChangeType))
		for _, f := range function.Functions {
			if f.IsValid {
				sb.WriteString(fmt.Sprintf("  - %s\n", f.FunctionName))
				sb.WriteString(fmt.Sprintf("    - 原因: %s\n", f.Reason))
				if f.Suggestion != "" {
					sb.WriteString(fmt.Sprintf("    - 建议: %s\n", f.Suggestion))
				}
				// 函数级问题列表
				if len(f.Issues) > 0 {
					sb.WriteString("    - 问题:\n")
					for _, is := range f.Issues {
						sb.WriteString(fmt.Sprintf("      - [%s] %s: %s\n", is.Severity, is.Title, is.Detail))
					}
				}
				// 新增逻辑：展示上游调用者
				if len(f.UpstreamCallers) > 0 {
					sb.WriteString("    - **上游调用者**:\n")
					for _, caller := range f.UpstreamCallers {
						sb.WriteString(fmt.Sprintf("      - `%s`\n", caller))
					}
				}
			}
		}
	}

	sb.WriteString("\n### 无效的函数\n\n")
	for _, function := range r.AffectedFunctions {
		hasInvalidFuncs := false
		for _, f := range function.Functions {
			if !f.IsValid {
				if !hasInvalidFuncs {
					sb.WriteString(fmt.Sprintf("- %s (%s)\n", function.Filename, function.ChangeType))
					hasInvalidFuncs = true
				}
				sb.WriteString(fmt.Sprintf("  - %s\n", f.FunctionName))
				sb.WriteString(fmt.Sprintf("    - 原因: %s\n", f.Reason))
				if f.Suggestion != "" {
					sb.WriteString(fmt.Sprintf("    - 建议: %s\n", f.Suggestion))
				}
			}
		}
	}

	return sb.String()
}

// GetTestSuggestionsMd 获取测试建议（Markdown）
func (r *MrAnalysisResult) GetTestSuggestionsMd() string {
	var sb strings.Builder
	sb.WriteString("## 改动测试建议\n\n")
	if len(r.TestSuggestions) == 0 {
		sb.WriteString("暂无测试建议\n")
		return sb.String()
	}
	sb.WriteString("| 标题 | 原因 | 优先级 | 相关函数 |\n")
	sb.WriteString("| :--- | :--- | :---: | :--- |\n")
	for _, ts := range r.TestSuggestions {
		rf := ""
		if len(ts.RelatedFunctions) > 0 {
			for i, f := range ts.RelatedFunctions {
				if i > 0 {
					rf += "<br>"
				}
				rf += f
			}
		} else {
			rf = "-"
		}
		reason := ts.Reason
		if reason == "" {
			reason = "-"
		}
		prio := ts.Priority
		if prio == "" {
			prio = "-"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", ts.Title, reason, prio, rf))
	}
	return sb.String()
}

// GetTestCasesMd 以 Markdown 表格输出结构化测试用例
func (r *MrAnalysisResult) GetTestCasesMd() string {
	var sb strings.Builder
	sb.WriteString("## 结构化测试用例\n\n")
	if len(r.TestCases) == 0 {
		sb.WriteString("暂无\n")
		return sb.String()
	}
	sb.WriteString("| 用例ID | 模块 | 用例标题 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n")
	for _, tc := range r.TestCases {
		pre := strings.Join(tc.Prerequisites, "<br>")
		steps := strings.Join(tc.Steps, "<br>")
		expected := strings.Join(tc.Expected, "<br>")
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |\n", tc.ID, tc.Module, tc.Title, pre, steps, expected, tc.Priority))
	}
	return sb.String()
}
