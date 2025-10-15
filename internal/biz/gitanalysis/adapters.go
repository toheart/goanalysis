package gitanalysis

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GitLabAdapter 定义GitLab访问适配接口，便于测试注入
type GitLabAdapter interface {
	GetMergeRequest(projectID, mrID int) (*gitlab.MergeRequest, error)
	ListMergeRequestDiffs(projectID, mrID, page int) ([]*gitlab.MergeRequestDiff, int, error)
	GetRawFile(projectID int, path string, ref string) (string, error)
	CreateMergeRequestNote(projectID, mrID int, body string) error
}
