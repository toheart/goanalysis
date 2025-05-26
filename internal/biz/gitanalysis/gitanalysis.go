package gitanalysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/callgraph"
	"github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
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
	err = g.RepoCallGraph(project, "")
	if err != nil {
		return nil, err
	}

	return analyzer.AnalyzeMR(projectID, mrID, autoNotes)
}

// RepoCallGraph 对仓库进行静态分析
// project 项目信息
// rootDir 程序启动目录, 为空时使用克隆路径
func (g *GitAnalysis) RepoCallGraph(project *gitapi.Project, rootDir string) error {
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
		return fmt.Errorf("open repo failed: %w", err)
	}

	// 获取HEAD引用
	ref, err := repo.Head()
	if err != nil {
		return fmt.Errorf("get HEAD failed: %w", err)
	}
	// 获取最新commit
	dbPath := g.CallDbPath(project.Name, ref.Hash().String())
	// 初始化静态分析数据库
	dbStore, err := g.data.GetFuncNodeDB(dbPath)
	if err != nil {
		return fmt.Errorf("init db failed: %w", err)
	}
	defer dbStore.Close()

	// 创建程序分析实例，使用默认的 RTA 算法
	programAnalysis := callgraph.NewProgramAnalysis(
		rootDir,
		g.logger,
		dbStore,
		callgraph.WithAlgo(callgraph.CallGraphTypeRta),
		callgraph.WithIgnorePaths("vendor,third_party"),
	)
	stopChan := make(chan struct{})
	go func() {
		defer func() {
			stopChan <- struct{}{}
		}()
		if err := programAnalysis.SetTree(statusChan); err != nil {
			statusChan <- []byte(err.Error())
			g.logger.Error(err.Error())
			return
		}
	}()
	go func() {
		if err := programAnalysis.SaveData(context.Background(), statusChan); err != nil {
			statusChan <- []byte(err.Error())
			g.logger.Error(err.Error())
			return
		}
	}()
	stopChan <- struct{}{}
	g.logger.Info("call graph analysis done")
	return nil
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
