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
	fmt.Println("review:")
	fmt.Println(r.Review)
}

// GetAffectedFiles 获取受影响的文件
func (r *MrAnalysisResult) GetAffectedFilesMd() string {
	var sb strings.Builder
	sb.WriteString("## 改动影响面\n\n")
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
