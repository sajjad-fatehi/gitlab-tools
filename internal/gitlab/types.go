package gitlab

import "strings"

type Project struct {
	ID                int      `json:"id"`
	Name              string   `json:"name"`
	PathWithNamespace string   `json:"path_with_namespace"`
	WebURL            string   `json:"web_url"`
	Description       string   `json:"description"`
	Topics            []string `json:"topics"`
}

type Topic struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	TotalProjectsCount int    `json:"total_projects_count"`
}

type MergeRequest struct {
	ID           int    `json:"id"`
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	WebURL       string `json:"web_url"`
	State        string `json:"state"`
	Draft        bool   `json:"draft"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	ProjectID    int    `json:"project_id"`
}

type Commit struct {
	ID      string `json:"id"`
	ShortID string `json:"short_id"`
	Title   string `json:"title"`
}

type Compare struct {
	Commit  Commit   `json:"commit"`
	Commits []Commit `json:"commits"`
	Diffs   []Diff   `json:"diffs"`
}

type Diff struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	AMode       string `json:"a_mode"`
	BMode       string `json:"b_mode"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
}

func (c *Compare) HasChanges() bool {
	return len(c.Commits) > 0
}

func (mr *MergeRequest) IsDraft() bool {
	if mr.Draft {
		return true
	}

	titleLower := strings.ToLower(strings.TrimSpace(mr.Title))
	return strings.HasPrefix(titleLower, "draft:") || strings.HasPrefix(titleLower, "wip:")
}
