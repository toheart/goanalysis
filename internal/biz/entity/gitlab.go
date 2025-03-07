package entity

// Repository GitLab仓库信息
type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	WebURL        string `json:"web_url"`
	SSHURLToRepo  string `json:"ssh_url_to_repo"`
	HTTPURLToRepo string `json:"http_url_to_repo"`
	Visibility    string `json:"visibility"`
	LastActivity  string `json:"last_activity_at"`
}

