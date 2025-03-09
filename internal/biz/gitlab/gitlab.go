package gitlab

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/xanzy/go-gitlab"
)

// GitLabBiz GitLab业务逻辑
type GitLabBiz struct {
	conf      *conf.Biz
	log       *log.Helper
	client    *gitlab.Client
	cloneDir  string
	token     string
	gitlabURL string
}

// NewGitLabBiz 创建GitLab业务逻辑实例
func NewGitLabBiz(conf *conf.Biz, logger log.Logger) *GitLabBiz {
	log := log.NewHelper(logger)

	// 从配置中获取GitLab相关配置
	token := conf.GitlabToken
	gitlabURL := conf.GitlabUrl
	cloneDir := conf.GitlabCloneDir

	// 确保克隆目录存在
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		log.Errorf("Failed to create GitLab clone directory: %v", err)
	}

	// 创建GitLab客户端
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabURL))
	if err != nil {
		log.Errorf("Failed to create GitLab client: %v", err)
	}

	return &GitLabBiz{
		conf:      conf,
		log:       log,
		client:    client,
		cloneDir:  cloneDir,
		token:     token,
		gitlabURL: gitlabURL,
	}
}

// ListRepositories 获取有权限的GitLab仓库列表
func (g *GitLabBiz) ListRepositories() ([]entity.Repository, error) {
	g.log.Info("Getting GitLab repository list")

	if g.client == nil {
		return nil, fmt.Errorf("GitLab client not initialized")
	}

	// 获取当前用户信息
	user, _, err := g.client.Users.CurrentUser()
	if err != nil {
		g.log.Errorf("Failed to get current user info: %v", err)
		return nil, fmt.Errorf("Failed to get current user info: %v", err)
	}

	g.log.Infof("Current user: %s", user.Username)

	// 获取用户有权限的项目
	opt := &gitlab.ListProjectsOptions{
		OrderBy:        gitlab.String("last_activity_at"),
		Sort:           gitlab.String("desc"),
		MinAccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
		Statistics:     gitlab.Bool(true),
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var allProjects []*gitlab.Project
	page := 1

	for {
		opt.Page = page
		projects, resp, err := g.client.Projects.ListProjects(opt)
		if err != nil {
			g.log.Errorf("Failed to get project list: %v", err)
			return nil, fmt.Errorf("Failed to get project list: %v", err)
		}

		allProjects = append(allProjects, projects...)

		// 检查是否有下一页
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		page = resp.NextPage
	}

	// 转换为Repository结构
	var repositories []entity.Repository
	for _, project := range allProjects {
		repositories = append(repositories, entity.Repository{
			ID:            project.ID,
			Name:          project.Name,
			FullName:      project.NameWithNamespace,
			Description:   project.Description,
			DefaultBranch: project.DefaultBranch,
			WebURL:        project.WebURL,
			SSHURLToRepo:  project.SSHURLToRepo,
			HTTPURLToRepo: project.HTTPURLToRepo,
			Visibility:    string(project.Visibility),
			LastActivity:  project.LastActivityAt.String(),
		})
	}

	g.log.Infof("Found %d repositories", len(repositories))
	return repositories, nil
}

// CloneRepository 克隆指定的GitLab仓库
func (g *GitLabBiz) CloneRepository(repoURL, branch string) (string, error) {
	g.log.Infof("Cloning repository: %s, branch: %s", repoURL, branch)

	// 从URL中提取仓库名称
	repoName := extractRepoName(repoURL)
	if repoName == "" {
		return "", fmt.Errorf("Cannot extract repository name from URL: %s", repoURL)
	}

	// 创建目标目录
	targetDir := filepath.Join(g.cloneDir, repoName)

	// 如果目录已存在，先删除
	if _, err := os.Stat(targetDir); err == nil {
		g.log.Infof("Directory already exists, deleting: %s", targetDir)
		if err := os.RemoveAll(targetDir); err != nil {
			g.log.Errorf("Failed to delete directory: %v", err)
			return "", fmt.Errorf("Failed to delete directory: %v", err)
		}
	}

	g.log.Infof("Cloning to directory: %s", targetDir)

	// 使用go-git克隆仓库
	cloneOptions := &git.CloneOptions{
		URL:           repoURL,
		Progress:      os.Stdout,
		SingleBranch:  true,
		ReferenceName: getBranchReference(branch),
		Auth: &http.BasicAuth{
			Username: "oauth2",
			Password: g.token,
		},
	}

	_, err := git.PlainClone(targetDir, false, cloneOptions)
	if err != nil {
		// 如果go-git克隆失败，尝试使用git命令行
		g.log.Warnf("Failed to clone using go-git, trying git command line: %v", err)
		err = g.cloneWithGitCommand(repoURL, branch, targetDir)
		if err != nil {
			g.log.Errorf("Failed to clone repository: %v", err)
			return "", fmt.Errorf("Failed to clone repository: %v", err)
		}
	}

	g.log.Infof("Repository cloned successfully: %s", targetDir)
	return targetDir, nil
}

// 使用git命令行克隆仓库
func (g *GitLabBiz) cloneWithGitCommand(repoURL, branch, targetDir string) error {
	args := []string{"clone"}

	if branch != "" {
		args = append(args, "-b", branch)
	}

	// 添加认证信息到URL
	if g.token != "" {
		// 将https://gitlab.com/user/repo.git转换为https://oauth2:token@gitlab.com/user/repo.git
		if strings.HasPrefix(repoURL, "https://") {
			parts := strings.SplitN(repoURL, "//", 2)
			if len(parts) == 2 {
				repoURL = fmt.Sprintf("https://oauth2:%s@%s", g.token, parts[1])
			}
		}
	}

	args = append(args, repoURL, targetDir)

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		g.log.Errorf("Git command execution failed: %v, output: %s", err, string(output))
		return fmt.Errorf("Git command execution failed: %v", err)
	}

	return nil
}

// 从URL中提取仓库名称
func extractRepoName(repoURL string) string {
	// 移除.git后缀
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// 获取最后一个/后面的内容
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}

// 获取分支引用
func getBranchReference(branch string) plumbing.ReferenceName {
	if branch == "" {
		return plumbing.ReferenceName("refs/heads/master")
	}
	return plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
}
