package dos

type Function struct {
	FunctionName    string   `json:"function_name"`
	IsValid         bool     `json:"is_valid"`
	Reason          string   `json:"reason"`
	Suggestion      string   `json:"suggestion"`
	UpstreamCallers []string `json:"upstream_callers,omitempty"`
	// Issues 为函数级问题清单（由 LLM 产出）
	Issues []Issue `json:"issues,omitempty"`
}

type LLMResponseList []*Function

// LLMReview MR级别审查与建议（由LLM以JSON返回）
type LLMReview struct {
	Review          string           `json:"review"`
	CommitMsgScore  int              `json:"commit_message_score"`
	GoIssuesScore   int              `json:"go_issues_score"`
	TestSuggestions []TestSuggestion `json:"test_suggestions"`
}

// Issue 表示函数级问题项
type Issue struct {
	Title    string `json:"title"`    // 问题简要标题
	Severity string `json:"severity"` // 严重级别：P0/P1/P2 或 High/Medium/Low
	Detail   string `json:"detail"`   // 细节描述/证据
}

// TestCase 表示结构化测试用例
type TestCase struct {
	ID               string   `json:"id"`
	Module           string   `json:"module"`
	Title            string   `json:"title"`
	Prerequisites    []string `json:"prerequisites"`
	Steps            []string `json:"steps"`
	Expected         []string `json:"expected"`
	Priority         string   `json:"priority"`
	RelatedFunctions []string `json:"related_functions,omitempty"`
}
