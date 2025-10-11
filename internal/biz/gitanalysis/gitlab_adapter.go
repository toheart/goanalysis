package gitanalysis

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type realGitLabAdapter struct {
	cli *gitlab.Client
}

func newRealGitLabAdapter(cli *gitlab.Client) *realGitLabAdapter {
	return &realGitLabAdapter{cli: cli}
}

func (r *realGitLabAdapter) GetMergeRequest(projectID, mrID int) (*gitlab.MergeRequest, error) {
	mr, _, err := r.cli.MergeRequests.GetMergeRequest(projectID, mrID, &gitlab.GetMergeRequestsOptions{})
	if err != nil {
		return nil, err
	}
	return mr, nil
}

func (r *realGitLabAdapter) ListMergeRequestDiffs(projectID, mrID, page int) ([]*gitlab.MergeRequestDiff, int, error) {
	diffs, resp, err := r.cli.MergeRequests.ListMergeRequestDiffs(projectID, mrID, &gitlab.ListMergeRequestDiffsOptions{
		ListOptions: gitlab.ListOptions{Page: page},
	})
	if err != nil {
		return nil, 0, err
	}
	return diffs, resp.NextPage, nil
}

func (r *realGitLabAdapter) GetRawFile(projectID int, path string, ref string) (string, error) {
	content, _, err := r.cli.RepositoryFiles.GetRawFile(projectID, path, &gitlab.GetRawFileOptions{Ref: gitlab.Ptr(ref)})
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (r *realGitLabAdapter) CreateMergeRequestNote(projectID, mrID int, body string) error {
	_, _, err := r.cli.Notes.CreateMergeRequestNote(projectID, mrID, &gitlab.CreateMergeRequestNoteOptions{Body: gitlab.Ptr(body)})
	return err
}
