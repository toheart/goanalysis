package gitanalysis

import (
	"context"

	"github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// LLMClient 定义LLM适配接口，便于测试注入
type LLMClient interface {
	AnalyzeFile(ctx context.Context, fileContent string) (dos.LLMResponseList, error)
	AnalyzeMR(ctx context.Context, title, content string) (*dos.LLMReview, error)
	// GenerateTestCases 基于有效函数及其上游影响生成结构化测试用例
	GenerateTestCases(ctx context.Context, title, content string) ([]dos.TestCase, error)
}

// GitLabAdapter 定义GitLab访问适配接口，便于测试注入
type GitLabAdapter interface {
	GetMergeRequest(projectID, mrID int) (*gitlab.MergeRequest, error)
	ListMergeRequestDiffs(projectID, mrID, page int) ([]*gitlab.MergeRequestDiff, int, error)
	GetRawFile(projectID int, path string, ref string) (string, error)
	CreateMergeRequestNote(projectID, mrID int, body string) error
}
