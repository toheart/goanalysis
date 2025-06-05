package gitanalysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
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
	conf   *conf.Biz
	logger *log.Helper
	gCli   *gitapi.Client
	data   *data.Data
}

func NewGitAnalysis(conf *conf.Biz, data *data.Data) *GitAnalysis {
	gitClient, err := gitapi.NewClient(conf.Gitlab.Token, gitapi.WithBaseURL(conf.Gitlab.Url))
	if err != nil {
		panic(err)
	}
	return &GitAnalysis{
		conf:   conf,
		logger: log.NewHelper(log.GetLogger()),
		gCli:   gitClient,
		data:   data,
	}
}

func (g *GitAnalysis) MRAnalyzer(projectID, mrID int, autoNotes bool) (*dos.MrAnalysisResult, error) {

	analyzer, err := NewMRAnalyzer(g.conf, g.gCli)
	if err != nil {
		return nil, err
	}

	// 获取合并请求信息
	mergeRequest, err := analyzer.GetMergeRequest(projectID, mrID)
	if err != nil {
		return nil, err
	}

	// 获取项目信息
	project, _, err := g.gCli.Projects.GetProject(projectID, &gitapi.GetProjectOptions{})
	if err != nil {
		return nil, err
	}

	// 克隆仓库
	_, err = g.CloneRepository(project, mergeRequest.TargetBranch)
	if err != nil {
		return nil, err
	}

	// 对克隆仓库进行静态分析
	dbPath, err := g.RepoCallGraph(project, "")
	if err != nil {
		return nil, err
	}

	// 进行MR分析，获取LLM分析结果
	result, err := analyzer.AnalyzeMR(projectID, mrID, autoNotes)
	if err != nil {
		return nil, err
	}

	// 使用静态分析数据库查找受影响函数的调用关系
	err = g.AnalyzeFunctionCallRelations(dbPath, result)
	if err != nil {
		g.logger.Warnf("failed to analyze function call relations: %v", err)
	}

	return result, nil
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
	// 获取最新commit
	dbPath := g.CallDbPath(project.Name, ref.Hash().String())
	// 初始化静态分析数据库
	dbStore, err := g.data.GetFuncNodeDB(dbPath)
	if err != nil {
		return "", fmt.Errorf("init db failed: %w", err)
	}
	defer dbStore.Close()

	// 创建程序分析实例，使用默认的 RTA 算法
	pa := callgraph.NewProgramAnalysis(
		rootDir,
		g.logger,
		dbStore,
		callgraph.WithAlgo(callgraph.CallGraphTypeRta),
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
			Auth:     &http.BasicAuth{Password: g.conf.Gitlab.Token},
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

		// 获取工作区
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("get worktree failed: %w", err)
		}

		// 拉取最新代码
		err = worktree.Pull(&git.PullOptions{
			RemoteName: "origin",
			Progress:   os.Stdout,
			Auth:       &http.BasicAuth{Password: g.conf.Gitlab.Token},
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return nil, fmt.Errorf("pull latest changes failed: %w", err)
		}
	}

	// 如果指定了分支名,则切换到该分支
	if branchName != "" {
		g.logger.Infof("switch to branch: %s", branchName)

		// 获取工作区
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("get worktree failed: %w", err)
		}

		// 切换到指定分支
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branchName),
		})
		if err != nil {
			return nil, fmt.Errorf("switch to branch %s failed: %w", branchName, err)
		}
	}

	g.logger.Infof("repository ready, current branch: %s", branchName)
	return repo, nil
}

func (g *GitAnalysis) ClonePath(name string) string {
	return filepath.Join(g.conf.Gitlab.CloneDir, name)
}

func (g *GitAnalysis) CallDbPath(name string, commit string) string {
	return filepath.Join(g.conf.Gitlab.CloneDir, name, commit+".db")
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

	// 遍历分析结果中受影响的函数
	for i, affectedFunction := range result.AffectedFunctions {
		g.logger.Infof("analyzing call relations for file: %s", affectedFunction.Filename)

		for j, function := range affectedFunction.Functions {
			// 只处理有效的函数
			if !function.IsValid {
				continue
			}

			g.logger.Infof("finding callers for function: %s", function.FunctionName)

			// 查找调用当前函数的上级函数
			callers, err := g.findAllCallers(dbStore, function.FunctionName, calleeToCallers, 3) // 最大深度3
			if err != nil {
				g.logger.Warnf("failed to find callers for function %s: %v", function.FunctionName, err)
				continue
			}

			if len(callers) > 0 {
				// 将调用者信息添加到建议中
				callersInfo := fmt.Sprintf("此函数被以下%d个函数调用: %v", len(callers), callers)
				if function.Suggestion != "" {
					result.AffectedFunctions[i].Functions[j].Suggestion += "; " + callersInfo
				} else {
					result.AffectedFunctions[i].Functions[j].Suggestion = callersInfo
				}
				g.logger.Infof("function %s has %d callers: %v", function.FunctionName, len(callers), callers)
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

// findAllCallers 递归查找所有调用指定函数的上级函数
func (g *GitAnalysis) findAllCallers(dbStore repo.StaticDBStore, functionName string, calleeToCallers map[string][]string, maxDepth int) ([]string, error) {
	if maxDepth <= 0 {
		return nil, nil
	}

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

	// 对每个匹配的函数键，递归查找调用者
	for _, key := range matchingKeys {
		g.findCallersRecursive(key, calleeToCallers, visited, &callers, maxDepth)
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
func (g *GitAnalysis) findCallersRecursive(functionKey string, calleeToCallers map[string][]string, visited map[string]bool, result *[]string, depth int) {
	if depth <= 0 || visited[functionKey] {
		return
	}

	visited[functionKey] = true

	if callerKeys, exists := calleeToCallers[functionKey]; exists {
		for _, callerKey := range callerKeys {
			*result = append(*result, callerKey)
			// 递归查找上级调用者
			g.findCallersRecursive(callerKey, calleeToCallers, visited, result, depth-1)
		}
	}
}
