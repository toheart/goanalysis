package gitanalysis

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/callgraph"
	gdos "github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
	"github.com/toheart/goanalysis/internal/biz/repo"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// 用例4：静态分析端到端（默认跳过）
func TestE2E_StaticAnalysis_GenerateAndAnalyze(t *testing.T) {

	// 使用内置 example 目录作为分析对象（通过当前测试文件绝对路径定位）
	_, thisFile, _, _ := runtime.Caller(0)
	base := filepath.Dir(thisFile)
	dir := filepath.Join(base, "example")

	// 建立数据层
	d := data.NewData(log.DefaultLogger)
	dbPath := filepath.Join(t.TempDir(), "e2e.db")
	db := mustGetDB(t, d, dbPath)

	// 运行调用图生成
	pa := callgraph.NewProgramAnalysis(dir, log.NewHelper(log.DefaultLogger), db, callgraph.WithAlgo(callgraph.CallGraphTypeVta))
	status := make(chan []byte, 16)
	defer close(status)
	go func() {
		for range status { /* 丢弃状态 */
		}
	}()
	if err := pa.Execute(context.Background(), status); err != nil {
		t.Fatalf("execute callgraph: %v", err)
	}

	// 构造改动结果：声明 A 为有效改动函数
	res := &gdos.MrAnalysisResult{
		AffectedFunctions: []gdos.AffectedFunction{{
			Filename:   "a.go",
			ChangeType: gdos.Modified,
			Functions:  []*gdos.Function{{FunctionName: "A", IsValid: true}},
		}},
	}

	// 运行影响分析，期望能找到上游调用者（B）
	g := &GitAnalysis{logger: log.NewHelper(log.DefaultLogger), data: d}
	if err := g.AnalyzeFunctionCallRelations(dbPath, res); err != nil {
		t.Fatalf("analyze relations: %v", err)
	}
	if len(res.AffectedFunctions[0].Functions[0].UpstreamCallers) == 0 {
		t.Fatalf("expected upstream callers for A, got none")
	}
	// 输出上游调用链文字
	for _, af := range res.AffectedFunctions {
		for _, fn := range af.Functions {
			if len(fn.UpstreamCallers) == 0 {
				continue
			}
			for _, chain := range fn.UpstreamCallers {
				t.Logf("函数 %s 上游调用链: %s", fn.FunctionName, chain)
			}
		}
	}
}

// mustGetDB 获取一个临时DB（位于临时目录）
func mustGetDB(t *testing.T, d *data.Data, dbPath string) repo.StaticDBStore {
	t.Helper()
	db, err := d.GetFuncNodeDB(dbPath)
	if err != nil {
		t.Fatalf("get func node db: %v", err)
	}
	return db
}

// 递归复制目录
func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_ = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

// 在目录初始化 git 仓库并提交
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	r, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("git init: %v", err)
	}
	wt, err := r.Worktree()
	if err != nil {
		t.Fatalf("git worktree: %v", err)
	}
	if _, err = wt.Add("."); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err = wt.Commit("init", &git.CommitOptions{Author: &object.Signature{Name: "test", Email: "t@example.com", When: time.Now()}}); err != nil {
		t.Fatalf("git commit: %v", err)
	}
}

// 假 GitLab，仅提供 MR 元数据与 diff
type e2eFakeGit struct{}

func (e *e2eFakeGit) GetMergeRequest(projectID, mrID int) (*gitlab.MergeRequest, error) {
	return &gitlab.MergeRequest{}, nil
}
func (e *e2eFakeGit) ListMergeRequestDiffs(projectID, mrID, page int) ([]*gitlab.MergeRequestDiff, int, error) {
	if page != 1 {
		return nil, 0, nil
	}
	return []*gitlab.MergeRequestDiff{{NewPath: "a.go", OldPath: "a.go", Diff: "@@\n-func A() int { return 1 }\n+func A() int { return 2 }\n"}}, 0, nil
}
func (e *e2eFakeGit) GetRawFile(projectID int, path string, ref string) (string, error) {
	return "", nil
}
func (e *e2eFakeGit) CreateMergeRequestNote(projectID, mrID int, body string) error { return nil }

// E2E：使用真实 LLM + 本地 example + mock GitLab
func TestE2E_MR_WithExample_RealLLM_MockGitLab(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	apiBase := os.Getenv("OPENAI_API_BASE")
	model := os.Getenv("OPENAI_MODEL")
	if apiKey == "" || apiBase == "" || model == "" {
		t.Skip("missing OPENAI_* env; skip real LLM e2e test")
	}

	// cfg
	tmp := t.TempDir()
	cfg := &conf.Biz{Gitlab: &conf.GitLab{CloneDir: tmp}, Openai: &conf.OpenAI{ApiKey: apiKey, ApiBase: apiBase, Model: model}}

	// 准备克隆仓库内容
	_, thisFile, _, _ := runtime.Caller(0)
	base := filepath.Dir(thisFile)
	src := filepath.Join(base, "example")
	projectName := "example"
	clonePath := filepath.Join(tmp, projectName)
	copyDir(t, src, clonePath)
	initGitRepo(t, clonePath)

	// 运行静态分析，生成 DB
	d := data.NewData(log.DefaultLogger)
	g := &GitAnalysis{conf: cfg, logger: log.NewHelper(log.DefaultLogger), data: d}
	dbPath, err := g.RepoCallGraph(&gitlab.Project{Name: projectName}, "")
	if err != nil {
		t.Fatalf("RepoCallGraph: %v", err)
	}

	// 构造真实 LLM + 假 GitLab 的 MRAnalyzer
	llm, err := NewLLMAnalyzer(apiKey, apiBase, model)
	if err != nil {
		t.Fatalf("new llm: %v", err)
	}
	analyzer := NewMRAnalyzerWithAdapters(cfg, llm, &e2eFakeGit{})

	// 运行 MR 分析
	res, err := analyzer.AnalyzeMR(1, 1, false)
	if err != nil {
		t.Fatalf("AnalyzeMR: %v", err)
	}

	// 结合 DB 做影响面分析
	if err := g.AnalyzeFunctionCallRelations(dbPath, res); err != nil {
		t.Fatalf("AnalyzeFunctionCallRelations: %v", err)
	}

	// 宽松断言：存在至少一个有效函数有上游调用者
	ok := false
	for _, af := range res.AffectedFunctions {
		for _, fn := range af.Functions {
			if fn.IsValid && len(fn.UpstreamCallers) > 0 {
				ok = true
			}
		}
	}
	if !ok {
		t.Fatalf("expected at least one valid function with upstream callers")
	}
	// 输出上游调用链文字
	for _, af := range res.AffectedFunctions {
		for _, fn := range af.Functions {
			if !fn.IsValid || len(fn.UpstreamCallers) == 0 {
				continue
			}
			for _, chain := range fn.UpstreamCallers {
				t.Logf("函数 %s 上游调用链: %s", fn.FunctionName, chain)
			}
		}
	}
}
