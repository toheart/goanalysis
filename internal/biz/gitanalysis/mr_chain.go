package gitanalysis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/compose"
	"github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
)

// 中文注释：用于 Eino 编排的 MR 处理状态
type mrProcState struct {
	ProjectID int
	MRID      int
	AutoNotes bool

	// 中文注释：链路中间数据
	MergeTitle string
	Tasks      []fileProcessTask

	Result *dos.MrAnalysisResult
}

// 中文注释：使用 Eino 链编排 AnalyzeMR 主链路（任务搜集→并发处理→审查生成）
func (a *MRAnalyzer) analyzeMRWithEino(ctx context.Context, projectID, mrID int, autoNotes bool) (*dos.MrAnalysisResult, error) {
	// 中文注释：初始化状态
	initState := &mrProcState{
		ProjectID: projectID,
		MRID:      mrID,
		AutoNotes: autoNotes,
		Result: &dos.MrAnalysisResult{
			MergeRequestID: mrID,
			ProjectID:      projectID,
		},
	}

	chain := compose.NewChain[*mrProcState, *dos.MrAnalysisResult]()

	// 节点1：获取 MR 元数据并收集任务
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *mrProcState) (*mrProcState, error) {
		a.logger.Infof("fetching MR #%d for project %d", s.MRID, s.ProjectID)
		mr, err := a.git.GetMergeRequest(s.ProjectID, s.MRID)
		if err != nil {
			return nil, fmt.Errorf("failed to get MR: %w", err)
		}
		s.MergeTitle = mr.Title

		// 分页拉取变更列表并构建任务
		var tasks []fileProcessTask
		page := 1
		for page > 0 {
			changes, nextPage, err := a.git.ListMergeRequestDiffs(s.ProjectID, s.MRID, page)
			if err != nil {
				return nil, fmt.Errorf("failed to get MR changes: %w", err)
			}
			page = nextPage
			for _, change := range changes {
				// 仅处理Go文件
				if !isGoFile(change.NewPath) && !isGoFile(change.OldPath) {
					continue
				}
				if a.filterFileByName(change.NewPath) || a.filterFileByName(change.OldPath) {
					continue
				}
				var (
					changeType dos.ChangeType
					filename   string
					ref        string
					path       string
					content    string
				)
				switch {
				case change.NewFile:
					changeType = dos.Added
					filename = change.NewPath
					ref = mr.SourceBranch
					path = change.NewPath
				case change.DeletedFile:
					changeType = dos.Deleted
					filename = change.OldPath
					ref = mr.TargetBranch
					path = change.OldPath
				default:
					changeType = dos.Modified
					filename = change.NewPath
					content = change.Diff
					// 中文注释：为 Modified 任务补充 Source 分支 ref 与路径，便于后续获取上下文源码
					ref = mr.SourceBranch
					path = change.NewPath
					a.logger.Infof("Modified file %s detected, need AI analysis", filename)
				}
				s.Result.AffectedFiles = append(s.Result.AffectedFiles, dos.AffectedFile{
					Filename:   filename,
					ChangeType: changeType,
				})
				tasks = append(tasks, fileProcessTask{
					ProjectID:  s.ProjectID,
					Path:       path,
					Ref:        ref,
					ChangeType: changeType,
					Filename:   filename,
					Content:    content,
				})
			}
		}
		s.Tasks = tasks
		return s, nil
	}))

	// 节点2：并发处理任务（Modified 走 LLM，其他走 AST 提取）
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *mrProcState) (*mrProcState, error) {
		if len(s.Tasks) == 0 {
			return s, nil
		}
		a.logger.Infof("processing %d changed files", len(s.Tasks))
		resultCh := make(chan fileProcessResult, 5)
		go func() {
			p := newWorkerPool(10)
			for _, t := range s.Tasks {
				tt := t
				p.Go(func() { resultCh <- a.fetchAndProcessFile(tt) })
			}
			p.Wait()
			close(resultCh)
		}()
		for res := range resultCh {
			if res.Error != nil {
				a.logger.Errorf("Failed to process file %s: %v", res.Filename, res.Error)
				continue
			}
			if len(res.Functions) > 0 {
				s.Result.AffectedFunctions = append(s.Result.AffectedFunctions, dos.AffectedFunction{
					Filename:   res.Filename,
					ChangeType: res.ChangeType,
					Functions:  res.Functions,
				})
			}
		}
		return s, nil
	}))

	// 节点3：调用 LLM 生成审查与测试建议
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *mrProcState) (*mrProcState, error) {
		a.logger.Infof("prepare to analyze MR: %s", s.MergeTitle)
		functions, _ := json.Marshal(s.Result.AffectedFunctions)
		llmReview, err := a.provider.AnalyzeMR(ctx, s.MergeTitle, string(functions))
		if err != nil {
			return nil, fmt.Errorf("failed to analyze MR: %w", err)
		}
		s.Result.Review = llmReview.Review
		s.Result.TestSuggestions = llmReview.TestSuggestions
		return s, nil
	}))

	// 终结节点：返回结果
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *mrProcState) (*dos.MrAnalysisResult, error) {
		return s.Result, nil
	}))

	r, err := chain.Compile(ctx)
	if err != nil {
		return nil, err
	}
	out, err := r.Invoke(ctx, initState)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// 中文注释：判断是否是 Go 文件
func isGoFile(path string) bool {
	if path == "" {
		return false
	}
	n := len(path)
	if n < 3 {
		return false
	}
	return path[n-3:] == ".go"
}

// 中文注释：简单的 goroutine worker 池（避免引入额外依赖）
type workerPool struct{ ch chan struct{} }

func newWorkerPool(n int) *workerPool {
	if n <= 0 {
		n = 1
	}
	return &workerPool{ch: make(chan struct{}, n)}
}

func (p *workerPool) Go(fn func()) {
	p.ch <- struct{}{}
	go func() {
		defer func() { <-p.ch }()
		fn()
	}()
}

func (p *workerPool) Wait() {
	// 中文注释：等待池清空
	for i := 0; i < cap(p.ch); i++ {
		p.ch <- struct{}{}
	}
	for i := 0; i < cap(p.ch); i++ {
		<-p.ch
	}
}
