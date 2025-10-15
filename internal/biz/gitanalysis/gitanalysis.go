package gitanalysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/callgraph"
	"github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
	"github.com/toheart/goanalysis/internal/biz/repo"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
	gitapi "gitlab.com/gitlab-org/api/client-go"
)

type GitAnalysis struct {
	conf     *conf.Biz
	logger   *log.Helper
	gCli     *gitapi.Client
	data     *data.Data
	provider dos.Provider
}

func NewGitAnalysis(conf *conf.Biz, data *data.Data, provider dos.Provider) *GitAnalysis {
	gitClient, err := gitapi.NewClient(conf.Gitlab.Token, gitapi.WithBaseURL(conf.Gitlab.Url))
	if err != nil {
		panic(err)
	}
	return &GitAnalysis{
		conf:     conf,
		logger:   log.NewHelper(log.GetLogger()),
		gCli:     gitClient,
		data:     data,
		provider: provider,
	}
}

// NewGitAnalysisWithProvider 允许第三方注入自定义的 dos.Provider
// 中文注释：第三方在其程序中实现 Provider 接口，并通过此构造注入
// 兼容构造函数已移除

func (g *GitAnalysis) MRAnalyzer(projectID, mrID int, autoNotes bool) (*dos.MrAnalysisResult, error) {
	// 中文注释：切换为端到端 Eino 编排实现
	return g.analyzeMRFullWithEino(context.Background(), projectID, mrID, autoNotes)
}

// RepoCallGraph 对仓库进行静态分析
// project 项目信息P
// rootDir 程序启动目录, 为空时使用克隆路径
func (g *GitAnalysis) RepoCallGraph(project *gitapi.Project, rootDir string) (string, error) {
	g.logger.Infof("static analysis repo: %s", project.Name)

	// 创建一个状态通道用于接收分析进度信息
	statusChan := make(chan []byte, 100)
	defer close(statusChan)

	// 启动一个 goroutine 来处理和记录状态信息
	go func() {
		for status := range statusChan {
			g.logger.Info(string(status))
		}
	}()

	// 获取项目根目录
	if rootDir == "" {
		rootDir = g.ClonePath(project.Name)
	} else {
		rootDir = filepath.Join(g.ClonePath(project.Name), rootDir)
	}
	repo, err := git.PlainOpen(g.ClonePath(project.Name))
	if err != nil {
		return "", fmt.Errorf("open repo failed: %w", err)
	}

	// 获取HEAD引用
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("get HEAD failed: %w", err)
	}
	// 获取最新commit并计算 DB 路径（使用 projectID 维度存储）
	dbPath := g.CallDbPath(strconv.Itoa(project.ID), ref.Hash().String())

	// 若该 commit 的 DB 已存在，则跳过重新分析
	if _, err := os.Stat(dbPath); err == nil {
		g.logger.Infof("reuse existing callgraph db: %s", dbPath)
		return dbPath, nil
	}

	// 确保 DB 目录存在
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return "", fmt.Errorf("prepare db directory failed: %w", err)
	}
	// 初始化静态分析数据库
	dbStore, err := g.data.GetFuncNodeDB(dbPath)
	if err != nil {
		return "", fmt.Errorf("init db failed: %w", err)
	}

	// 创建程序分析实例，使用 VTA 算法以获得更高精度
	pa := callgraph.NewProgramAnalysis(
		rootDir,
		g.logger,
		dbStore,
		callgraph.WithAlgo(callgraph.CallGraphTypeVta),
		callgraph.WithIgnorePaths("vendor,third_party"),
	)
	err = pa.Execute(context.Background(), statusChan)
	if err != nil {
		return "", fmt.Errorf("call graph analysis failed: %w", err)
	}
	g.logger.Info("call graph analysis done")
	return dbPath, nil
}

// CloneRepository 使用go-git库克隆或更新仓库到本地目录并切换到指定分支
func (g *GitAnalysis) CloneRepository(project *gitapi.Project, branchName string) (*git.Repository, error) {
	var repo *git.Repository
	cloneDir := g.ClonePath(project.Name)
	// 检查目录是否存在
	if _, err := os.Stat(cloneDir); os.IsNotExist(err) {
		// 目录不存在,执行克隆
		g.logger.Infof("clone %s to %s", project.HTTPURLToRepo, cloneDir)

		// 确保目标目录的父目录存在
		if err := os.MkdirAll(filepath.Dir(cloneDir), 0755); err != nil {
			return nil, fmt.Errorf("create directory failed: %w", err)
		}

		// 克隆选项
		cloneOptions := &git.CloneOptions{
			URL:      project.HTTPURLToRepo,
			Progress: os.Stdout,
			Auth:     &http.BasicAuth{Username: "git", Password: g.conf.Gitlab.Token},
		}

		// 克隆仓库
		repo, err = git.PlainClone(cloneDir, false, cloneOptions)
		if err != nil {
			return nil, fmt.Errorf("clone repository failed: %w", err)
		}
	} else {
		// 目录存在,打开仓库
		g.logger.Infof("repository exists, pulling latest changes")
		repo, err = git.PlainOpen(cloneDir)
		if err != nil {
			return nil, fmt.Errorf("open repository failed: %w", err)
		}
	}

	// 如果指定了分支名,则切换到该分支
	if branchName != "" {
		g.logger.Infof("switch to branch: %s", branchName)

		// 先确保远端分支引用存在（fetch 指定分支）
		_ = repo.Fetch(&git.FetchOptions{
			RemoteName: "origin",
			Progress:   os.Stdout,
			Auth:       &http.BasicAuth{Username: "git", Password: g.conf.Gitlab.Token},
			RefSpecs:   []config.RefSpec{config.RefSpec("+refs/heads/" + branchName + ":refs/remotes/origin/" + branchName)},
		})

		// 获取工作区
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("get worktree failed: %w", err)
		}

		localRefName := plumbing.NewBranchReferenceName(branchName)
		remoteRefName := plumbing.NewRemoteReferenceName("origin", branchName)

		// 探测本地分支是否存在
		_, localErr := repo.Reference(localRefName, true)
		if localErr == nil {
			// 本地已存在：强制切换，避免工作区脏导致失败
			if err := worktree.Checkout(&git.CheckoutOptions{Branch: localRefName, Force: true}); err != nil {
				return nil, fmt.Errorf("switch to existing local branch failed: %w", err)
			}
		} else {
			// 本地不存在：尝试从远端创建
			rref, rerr := repo.Reference(remoteRefName, true)
			if rerr != nil {
				return nil, fmt.Errorf("switch to branch %s failed: reference not found", branchName)
			}
			if err := worktree.Checkout(&git.CheckoutOptions{Branch: localRefName, Create: true, Hash: rref.Hash(), Force: true}); err != nil {
				return nil, fmt.Errorf("create local branch from remote failed: %w", err)
			}
		}
	}

	g.logger.Infof("repository ready, current branch: %s", branchName)
	return repo, nil
}

func (g *GitAnalysis) ClonePath(name string) string {
	return filepath.Join(g.conf.Gitlab.CloneDir, name)
}

// ClonePathWithSuffix 返回带后缀的克隆目录路径（用于独立克隆目标分支）
// 中文注释：避免在同一工作副本切换分支，减少 Windows 文件锁冲突
func (g *GitAnalysis) ClonePathWithSuffix(name string, suffix string) string {
	if suffix == "" {
		return g.ClonePath(name)
	}
	return filepath.Join(g.conf.Gitlab.CloneDir, name+suffix)
}

func (g *GitAnalysis) CallDbPath(projectID string, commit string) string {
	return filepath.Join(g.conf.Gitlab.CloneDir, "db", projectID, commit+".db")
}

// 获取远端分支的 commit 哈希（尽量不切换本地工作区）
func (g *GitAnalysis) getRemoteCommitHash(project *gitapi.Project, branch string) (string, error) {
	repo, err := git.PlainOpen(g.ClonePath(project.Name))
	if err != nil {
		return "", fmt.Errorf("open repo failed: %w", err)
	}
	// 确保已 fetch 对应分支引用
	_ = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		Auth:       &http.BasicAuth{Username: "git", Password: g.conf.Gitlab.Token},
		RefSpecs:   []config.RefSpec{config.RefSpec("+refs/heads/" + branch + ":refs/remotes/origin/" + branch)},
	})
	ref, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", branch), true)
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}

// CloneRepositoryTo 在独立目录克隆指定分支，避免在同一工作区切换分支导致的文件占用
// 返回：repo、cloneDir
func (g *GitAnalysis) CloneRepositoryTo(project *gitapi.Project, branchName string, dirSuffix string) (*git.Repository, string, error) {
	cloneDir := g.ClonePathWithSuffix(project.Name, dirSuffix)

	// 若目录已存在，直接删除以避免旧分支残留导致的切换失败
	if _, err := os.Stat(cloneDir); err == nil {
		g.logger.Infof("remove existing clone directory: %s", cloneDir)
		if rmErr := os.RemoveAll(cloneDir); rmErr != nil {
			return nil, "", fmt.Errorf("remove existing clone directory failed: %w", rmErr)
		}
	}

	if err := os.MkdirAll(filepath.Dir(cloneDir), 0755); err != nil {
		return nil, "", fmt.Errorf("create parent directory failed: %w", err)
	}

	g.logger.Infof("clone %s branch %s to %s", project.HTTPURLToRepo, branchName, cloneDir)
	repo, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:           project.HTTPURLToRepo,
		Progress:      os.Stdout,
		Auth:          &http.BasicAuth{Username: "git", Password: g.conf.Gitlab.Token},
		ReferenceName: plumbing.NewBranchReferenceName(branchName),
		SingleBranch:  true,
		Depth:         0,
	})
	if err != nil {
		return nil, "", fmt.Errorf("clone repository to target dir failed: %w", err)
	}
	return repo, cloneDir, nil
}

// RepoCallGraphInDir 在指定 repoDir 上执行静态分析，生成/复用 DB
func (g *GitAnalysis) RepoCallGraphInDir(project *gitapi.Project, repoDir string) (string, error) {
	g.logger.Infof("static analysis repo in dir: %s", repoDir)

	statusChan := make(chan []byte, 100)
	defer close(statusChan)
	go func() {
		for status := range statusChan {
			g.logger.Info(string(status))
		}
	}()

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", fmt.Errorf("open repo failed: %w", err)
	}
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("get HEAD failed: %w", err)
	}
	dbPath := g.CallDbPath(strconv.Itoa(project.ID), ref.Hash().String())
	if _, err := os.Stat(dbPath); err == nil {
		g.logger.Infof("reuse existing callgraph db: %s", dbPath)
		return dbPath, nil
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return "", fmt.Errorf("prepare db directory failed: %w", err)
	}

	dbStore, err := g.data.GetFuncNodeDB(dbPath)
	if err != nil {
		return "", fmt.Errorf("init db failed: %w", err)
	}

	pa := callgraph.NewProgramAnalysis(
		repoDir,
		g.logger,
		dbStore,
		callgraph.WithAlgo(callgraph.CallGraphTypeVta),
		callgraph.WithIgnorePaths("vendor,third_party"),
	)
	if err := pa.Execute(context.Background(), statusChan); err != nil {
		return "", fmt.Errorf("call graph analysis failed: %w", err)
	}
	g.logger.Info("call graph analysis done")
	return dbPath, nil
}

// AnalyzeFunctionCallRelations 分析函数调用关系，为MR分析结果中的有效改动函数查找上级调用者
func (g *GitAnalysis) AnalyzeFunctionCallRelations(dbPath string, result *dos.MrAnalysisResult) error {
	g.logger.Infof("analyzing function call relations using db: %s", dbPath)

	// 获取静态分析数据库连接
	dbStore, err := g.data.GetFuncNodeDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer dbStore.Close()

	// 获取所有函数调用边
	allEdges, err := dbStore.GetAllFuncEdges()
	if err != nil {
		return fmt.Errorf("failed to get function edges: %w", err)
	}

	// 创建被调用方到调用方的映射
	calleeToCallers := make(map[string][]string)
	for _, edge := range allEdges {
		calleeToCallers[edge.CalleeKey] = append(calleeToCallers[edge.CalleeKey], edge.CallerKey)
	}

	// 构建 key->函数名 映射，便于输出友好名称
	allNodes, err := dbStore.GetAllFuncNodes()
	if err != nil {
		return fmt.Errorf("failed to get all function nodes: %w", err)
	}
	keyToName := make(map[string]string, len(allNodes))
	for _, n := range allNodes {
		// 优先使用节点的 Name（已是短名，如 pkg.Recv.Method），若为空可回退 FullName
		name := n.Name
		if name == "" {
			name = n.FullName
		}
		if name == "" {
			name = n.Key
		}
		keyToName[n.Key] = name
	}

	// 遍历分析结果中受影响的函数
	for i, affectedFunction := range result.AffectedFunctions {
		g.logger.Infof("analyzing call relations for file: %s", affectedFunction.Filename)

		for j, function := range affectedFunction.Functions {
			// 只处理有效的函数
			if !function.IsValid {
				continue
			}

			g.logger.Infof("finding callers for function: %s", function.FunctionName)

			// 计算从各上游到当前函数的调用链（文本形式，如 C->B->A），不限制深度
			chains, err := g.findCallerChainsText(dbStore, function.FunctionName, calleeToCallers, keyToName)
			if err != nil {
				g.logger.Warnf("failed to find callers for function %s: %v", function.FunctionName, err)
				continue
			}

			if len(chains) > 0 {
				result.AffectedFunctions[i].Functions[j].UpstreamCallers = chains
				g.logger.Infof("function %s has %d caller chains", function.FunctionName, len(chains))
			} else {
				// 如果没有找到调用者，可能是入口函数
				entryInfo := "此函数可能是入口函数，没有发现其他函数调用它"
				if function.Suggestion != "" {
					result.AffectedFunctions[i].Functions[j].Suggestion += "; " + entryInfo
				} else {
					result.AffectedFunctions[i].Functions[j].Suggestion = entryInfo
				}
				g.logger.Infof("function %s appears to be an entry function", function.FunctionName)
			}
		}
	}

	g.logger.Info("function call relations analysis completed")
	return nil
}

// findCallerChainsText 返回从上游到目标函数的调用链（文本），例如：C->B->A
func (g *GitAnalysis) findCallerChainsText(dbStore repo.StaticDBStore, functionName string, calleeToCallers map[string][]string, keyToName map[string]string) ([]string, error) {
	// 中文注释：优先进行严格匹配（完整等值、后缀包含接收者限定），再回退模糊匹配，降低误匹配
	allNodes, err := dbStore.GetAllFuncNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get all function nodes: %w", err)
	}

	var matchingKeys []string
	for _, node := range allNodes {
		// 严格等值匹配（包含接收者/包前缀的完整名）
		if node.Name == functionName {
			matchingKeys = append(matchingKeys, node.Key)
			continue
		}
		// 带接收者的后缀匹配：如 pkg.Type.Method 或 Type.Method
		if strings.HasSuffix(node.Name, "."+functionName) {
			matchingKeys = append(matchingKeys, node.Key)
			continue
		}
	}
	// 回退模糊匹配（仅当严格匹配为空时）
	if len(matchingKeys) == 0 {
		for _, node := range allNodes {
			if strings.Contains(node.Name, functionName) {
				matchingKeys = append(matchingKeys, node.Key)
			}
		}
	}

	// 深度优先遍历，收集所有链路（以 key 表示），然后转为名称文本
	type void struct{}
	seenChains := make(map[string]void)
	var chains []string
	for _, key := range matchingKeys {
		var dfs func(curr string, path []string, visited map[string]bool)
		dfs = func(curr string, path []string, visited map[string]bool) {
			if visited[curr] {
				return
			}
			visited[curr] = true

			// 追加当前节点（从目标向上）
			path = append(path, curr)

			callers := calleeToCallers[curr]
			if len(callers) == 0 {
				// 到达最上游，反转为上游->...->目标
				for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
					path[i], path[j] = path[j], path[i]
				}
				// 转名称
				var segs []string
				for _, k := range path {
					name := keyToName[k]
					if name == "" {
						name = k
					}
					segs = append(segs, name)
				}
				chain := strings.Join(segs, "->")
				if _, ok := seenChains[chain]; !ok {
					seenChains[chain] = void{}
					chains = append(chains, chain)
				}
				return
			}

			for _, caller := range callers {
				// 每条分支使用独立 visited 复制，避免错误剪枝
				nextVisited := make(map[string]bool, len(visited)+1)
				for k, v := range visited {
					nextVisited[k] = v
				}
				dfs(caller, append([]string{}, path...), nextVisited)
			}
		}
		dfs(key, []string{}, make(map[string]bool))
	}
	return chains, nil
}

// findAllCallers 递归查找所有调用指定函数的上级函数
func (g *GitAnalysis) findAllCallers(dbStore repo.StaticDBStore, functionName string, calleeToCallers map[string][]string, maxDepth int) ([]string, error) {

	// 首先尝试通过函数名直接查找
	var callers []string
	var visited = make(map[string]bool)

	// 获取所有函数节点，找到匹配的函数键
	allNodes, err := dbStore.GetAllFuncNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get all function nodes: %w", err)
	}

	// 查找匹配的函数键
	var matchingKeys []string
	for _, node := range allNodes {
		// 支持多种匹配方式：完全匹配函数名，或者函数名包含在节点名称中
		if node.Name == functionName ||
			strings.Contains(node.Name, functionName) ||
			strings.HasSuffix(node.Name, "."+functionName) {
			matchingKeys = append(matchingKeys, node.Key)
		}
	}

	// 对每个匹配的函数键，递归查找调用者（内联递归，移除未使用的辅助方法）
	var dfs func(curr string, depth int)
	dfs = func(curr string, depth int) {
		if (depth == 0) || visited[curr] {
			return
		}
		visited[curr] = true
		if callerKeys, exists := calleeToCallers[curr]; exists {
			for _, ck := range callerKeys {
				callers = append(callers, ck)
				nextDepth := depth
				if depth > 0 {
					nextDepth = depth - 1
				}
				dfs(ck, nextDepth)
			}
		}
	}
	for _, key := range matchingKeys {
		dfs(key, maxDepth)
	}

	// 去重
	uniqueCallers := make(map[string]bool)
	var result []string
	for _, caller := range callers {
		if !uniqueCallers[caller] {
			uniqueCallers[caller] = true
			result = append(result, caller)
		}
	}

	return result, nil
}

// findCallersRecursive 递归查找调用者
// 删除：findCallersRecursive（逻辑已内联到 findAllCallers）

// analyzeDeletedFunctionCallRelations 使用 Target 分支的 DB，为删除的函数查找上游调用者
func (g *GitAnalysis) analyzeDeletedFunctionCallRelations(dbPath string, result *dos.MrAnalysisResult) error {
	g.logger.Infof("analyzing deleted function callers using db: %s", dbPath)

	dbStore, err := g.data.GetFuncNodeDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer dbStore.Close()

	allEdges, err := dbStore.GetAllFuncEdges()
	if err != nil {
		return fmt.Errorf("failed to get function edges: %w", err)
	}
	calleeToCallers := make(map[string][]string)
	for _, edge := range allEdges {
		calleeToCallers[edge.CalleeKey] = append(calleeToCallers[edge.CalleeKey], edge.CallerKey)
	}

	allNodes, err := dbStore.GetAllFuncNodes()
	if err != nil {
		return fmt.Errorf("failed to get all function nodes: %w", err)
	}
	keyToName := make(map[string]string, len(allNodes))
	for _, n := range allNodes {
		name := n.Name
		if name == "" {
			name = n.FullName
		}
		if name == "" {
			name = n.Key
		}
		keyToName[n.Key] = name
	}

	// 仅处理 Deleted 的函数
	for i, af := range result.AffectedFunctions {
		if af.ChangeType != dos.Deleted {
			continue
		}
		for j, fn := range af.Functions {
			chains, err := g.findCallerChainsText(dbStore, fn.FunctionName, calleeToCallers, keyToName)
			if err != nil {
				g.logger.Warnf("failed to find callers for deleted function %s: %v", fn.FunctionName, err)
				continue
			}
			if len(chains) > 0 {
				result.AffectedFunctions[i].Functions[j].UpstreamCallers = chains
				if fn.Suggestion == "" {
					result.AffectedFunctions[i].Functions[j].Suggestion = "可能存在破坏性调用，需要更新调用点"
				}
			}
		}
	}
	g.logger.Info("deleted function callers analysis completed")
	return nil
}

// computeImpact 根据受影响函数与上游调用者，计算影响指标
func (g *GitAnalysis) computeImpact(result *dos.MrAnalysisResult) {
	// 统计有效函数与上游调用者
	totalChangedFiles := len(result.AffectedFiles)
	validChangedFunctions := 0
	uniqueUpstream := make(map[string]struct{})

	for _, af := range result.AffectedFunctions {
		for _, fn := range af.Functions {
			if fn.IsValid {
				validChangedFunctions++
				for _, caller := range fn.UpstreamCallers {
					if caller == "" {
						continue
					}
					uniqueUpstream[caller] = struct{}{}
				}
			}
		}
	}

	upstreamCount := len(uniqueUpstream)
	result.Impact = dos.ImpactAnalysis{
		TotalChangedFiles:     totalChangedFiles,
		ValidChangedFunctions: validChangedFunctions,
		UpstreamCallers:       upstreamCount,
	}

	g.logger.Infof("computed impact: files=%d, funcs=%d, callers=%d", totalChangedFiles, validChangedFunctions, upstreamCount)
}
