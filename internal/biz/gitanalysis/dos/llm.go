package dos

type Function struct {
	FunctionName    string   `json:"function_name"`
	IsValid         bool     `json:"is_valid"`
	Reason          string   `json:"reason"`
	Suggestion      string   `json:"suggestion"`
	UpstreamCallers []string `json:"upstream_callers,omitempty"`
}

type LLMResponseList []*Function

// LLMReview MR级别审查与建议（由LLM以JSON返回）
type LLMReview struct {
	Review          string           `json:"review"`
	CommitMsgScore  int              `json:"commit_message_score"`
	GoIssuesScore   int              `json:"go_issues_score"`
	TestSuggestions []TestSuggestion `json:"test_suggestions"`
}
