package gitanalysis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/compose"
	gitapi "gitlab.com/gitlab-org/api/client-go"

	"github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
)

// 中文注释：全链路编排状态
type gaState struct {
	ProjectID int
	MRID      int
	AutoNotes bool

	MergeRequest *gitapi.MergeRequest
	Project      *gitapi.Project

	Result       *dos.MrAnalysisResult
	HasDeleted   bool
	DbPathSource string
	DbPathTarget string
}

// 中文注释：使用 Eino 对 GitAnalysis.MRAnalyzer 进行端到端编排
func (g *GitAnalysis) analyzeMRFullWithEino(ctx context.Context, projectID, mrID int, autoNotes bool) (*dos.MrAnalysisResult, error) {
	initState := &gaState{
		ProjectID: projectID,
		MRID:      mrID,
		AutoNotes: autoNotes,
		Result: &dos.MrAnalysisResult{
			MergeRequestID: mrID,
			ProjectID:      projectID,
		},
	}

	chain := compose.NewChain[*gaState, *dos.MrAnalysisResult]()

	// 节点1：创建 MRAnalyzer，拉取 MR 与 Project 元信息
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*gaState, error) {
		analyzer, err := NewMRAnalyzer(g.conf, g.gCli, g.provider)
		if err != nil {
			return nil, err
		}
		mr, err := analyzer.GetMergeRequest(s.ProjectID, s.MRID)
		if err != nil {
			return nil, err
		}
		s.MergeRequest = mr
		prj, _, err := g.gCli.Projects.GetProject(s.ProjectID, &gitapi.GetProjectOptions{})
		if err != nil {
			return nil, err
		}
		s.Project = prj
		return s, nil
	}))

	// 节点2：执行 MR 主链路分析（已在 MRAnalyzer.AnalyzeMR 内部使用 Eino 编排）
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*gaState, error) {
		analyzer, err := NewMRAnalyzer(g.conf, g.gCli, g.provider)
		if err != nil {
			return nil, err
		}
		res, err := analyzer.AnalyzeMR(s.ProjectID, s.MRID, s.AutoNotes)
		if err != nil {
			return nil, err
		}
		s.Result = res
		// 是否包含删除变更
		for _, f := range s.Result.AffectedFiles {
			if f.ChangeType == dos.Deleted {
				s.HasDeleted = true
				break
			}
		}
		return s, nil
	}))

	// 节点3：克隆并分析 Source 分支（生成/复用调用图 DB），并回填上游调用链
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*gaState, error) {
		if _, err := g.CloneRepository(s.Project, s.MergeRequest.SourceBranch); err != nil {
			return nil, err
		}
		dbPath, err := g.RepoCallGraph(s.Project, "")
		if err != nil {
			return nil, err
		}
		s.DbPathSource = dbPath
		if err := g.AnalyzeFunctionCallRelations(dbPath, s.Result); err != nil {
			g.logger.Warnf("failed to analyze function call relations: %v", err)
		}
		return s, nil
	}))

	// 节点4：如存在删除，分析 Target 分支调用者（独立克隆，避免原地切分支导致的文件占用）
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*gaState, error) {
		if !s.HasDeleted {
			return s, nil
		}
		var dbPathTarget string
		// 优先使用远端引用复用 DB
		if commit, err := g.getRemoteCommitHash(s.Project, s.MergeRequest.TargetBranch); err == nil {
			candidate := g.CallDbPath(fmt.Sprintf("%d", s.ProjectID), commit)
			if _, stErr := osStat(candidate); stErr == nil {
				g.logger.Infof("reuse existing target callgraph db: %s", candidate)
				dbPathTarget = candidate
			}
		}
		if dbPathTarget == "" {
			// 中文注释：使用独立目录克隆目标分支（如 @target）
			_, targetDir, err := g.CloneRepositoryTo(s.Project, s.MergeRequest.TargetBranch, "@target")
			if err != nil {
				return nil, err
			}
			p, perr := g.RepoCallGraphInDir(s.Project, targetDir)
			if perr != nil {
				return nil, perr
			}
			dbPathTarget = p
		}
		s.DbPathTarget = dbPathTarget
		if err := g.analyzeDeletedFunctionCallRelations(dbPathTarget, s.Result); err != nil {
			g.logger.Warnf("failed to analyze deleted function callers: %v", err)
		}
		return s, nil
	}))

	// 节点5：计算影响指标
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*gaState, error) {
		g.computeImpact(s.Result)
		return s, nil
	}))

	// 节点6：使用 Source 分支 DB 的精确函数源码增强 LLM 问题识别（只补充无问题的 Modified/Added 函数）
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*gaState, error) {
		if s.DbPathSource == "" || s.Result == nil {
			return s, nil
		}
		dbStore, err := g.data.GetFuncNodeDB(s.DbPathSource)
		if err != nil {
			return nil, err
		}
		defer dbStore.Close()

		nodes, err := dbStore.GetAllFuncNodes()
		if err != nil {
			return nil, err
		}
		name2src := make(map[string]string, len(nodes))
		for _, n := range nodes {
			if strings.TrimSpace(n.SourceCode) != "" {
				// 以节点 Name 作为匹配键（与 MRAnalyzer 的函数名一致策略）
				name2src[n.Name] = n.SourceCode
			}
		}

		// 遍历受影响函数，补充 issues
		for i, af := range s.Result.AffectedFunctions {
			if af.ChangeType != dos.Modified && af.ChangeType != dos.Added {
				continue
			}
			for j, fn := range af.Functions {
				if len(fn.Issues) > 0 {
					continue
				}
				src := name2src[fn.FunctionName]
				if strings.TrimSpace(src) == "" {
					continue
				}
				// 构造仅针对该函数的输入（函数源码优先）
				input := fmt.Sprintf("# Function Source\nFile: %s\nFunction: %s\n\n%s", af.Filename, fn.FunctionName, src)
				lres, lerr := g.provider.AnalyzeFile(ctx, input)
				if lerr != nil || len(lres) == 0 {
					continue
				}
				// 匹配同名函数，回填 issues
				for _, r := range lres {
					if r != nil && r.FunctionName == fn.FunctionName && len(r.Issues) > 0 {
						s.Result.AffectedFunctions[i].Functions[j].Issues = r.Issues
						break
					}
				}
			}
		}
		return s, nil
	}))

	// 节点7：生成结构化测试用例（如存在受影响函数）
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*gaState, error) {
		if len(s.Result.AffectedFunctions) == 0 {
			return s, nil
		}
		b, _ := json.Marshal(struct {
			AffectedFunctions []dos.AffectedFunction `json:"affected_functions"`
			Impact            dos.ImpactAnalysis     `json:"impact"`
		}{AffectedFunctions: s.Result.AffectedFunctions, Impact: s.Result.Impact})
		cases, err := g.provider.GenerateTestCases(ctx, s.MergeRequest.Title, string(b))
		if err != nil {
			g.logger.Warnf("failed to generate test cases: %v", err)
			return s, nil
		}
		s.Result.TestCases = cases
		return s, nil
	}))

	// 节点8：自动发布 MR 笔记
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*gaState, error) {
		if !s.AutoNotes {
			return s, nil
		}
		analyzer, err := NewMRAnalyzer(g.conf, g.gCli, g.provider)
		if err != nil {
			return nil, err
		}
		if err := analyzer.CreateMRNotes(s.ProjectID, s.MRID, s.Result); err != nil {
			g.logger.Warnf("failed to create MR notes: %v", err)
		}
		return s, nil
	}))

	// 终结节点：输出结果
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, s *gaState) (*dos.MrAnalysisResult, error) {
		return s.Result, nil
	}))

	r, err := chain.Compile(ctx)
	if err != nil {
		return nil, err
	}
	return r.Invoke(ctx, initState)
}

// 中文注释：便于替身注入；默认指向 os.Stat
var osStat = os.Stat
