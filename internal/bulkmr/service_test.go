package bulkmr

import (
	"testing"

	"github.com/sajjad-fatehi/gitlab-tools/internal/gitlab"
)

type mockGitLabClient struct {
	projects      map[string]*gitlab.Project
	branches      map[int]map[string]bool
	mergeRequests map[int][]gitlab.MergeRequest
	createError   error
}

func newMockClient() *mockGitLabClient {
	return &mockGitLabClient{
		projects:      make(map[string]*gitlab.Project),
		branches:      make(map[int]map[string]bool),
		mergeRequests: make(map[int][]gitlab.MergeRequest),
	}
}

func (m *mockGitLabClient) addProject(path string, id int) {
	m.projects[path] = &gitlab.Project{
		ID:                id,
		Name:              path,
		PathWithNamespace: path,
		WebURL:            "https://gitlab.example.com/" + path,
	}
	m.branches[id] = make(map[string]bool)
}

func (m *mockGitLabClient) addBranch(projectID int, branchName string) {
	if m.branches[projectID] == nil {
		m.branches[projectID] = make(map[string]bool)
	}
	m.branches[projectID][branchName] = true
}

func (m *mockGitLabClient) addMergeRequest(projectID int, mr gitlab.MergeRequest) {
	m.mergeRequests[projectID] = append(m.mergeRequests[projectID], mr)
}

func (m *mockGitLabClient) GetProject(projectPath string) (*gitlab.Project, error) {
	project, ok := m.projects[projectPath]
	if !ok {
		return nil, nil
	}
	return project, nil
}

func (m *mockGitLabClient) BranchExists(projectID int, branch string) (bool, error) {
	branches, ok := m.branches[projectID]
	if !ok {
		return false, nil
	}
	return branches[branch], nil
}

func (m *mockGitLabClient) CompareBranches(projectID int, sourceBranch, targetBranch string) (*gitlab.Compare, error) {
	return &gitlab.Compare{
		Commits: []gitlab.Commit{
			{ID: "abc123", ShortID: "abc123", Title: "Test commit"},
		},
	}, nil
}

func (m *mockGitLabClient) FindOpenMergeRequests(projectID int, sourceBranch, targetBranch string) ([]gitlab.MergeRequest, error) {
	mrs, ok := m.mergeRequests[projectID]
	if !ok {
		return []gitlab.MergeRequest{}, nil
	}
	return mrs, nil
}

func (m *mockGitLabClient) CreateMergeRequest(projectID int, sourceBranch, targetBranch, title, description string) (*gitlab.MergeRequest, error) {
	if m.createError != nil {
		return nil, m.createError
	}
	mr := &gitlab.MergeRequest{
		ID:     999,
		IID:    99,
		Title:  title,
		WebURL: "https://gitlab.example.com/merge_requests/99",
		State:  "opened",
		Draft:  false,
	}
	return mr, nil
}

func TestProcessProject_BothBranchesExist_NoExistingMR_CreatesMR(t *testing.T) {
	client := newMockClient()
	client.addProject("group/repo-a", 1)
	client.addBranch(1, "op-stage")
	client.addBranch(1, "op-rc")

	config := Config{
		OriginBranch: "op-stage",
		TargetBranch: "op-rc",
		Projects:     []string{"group/repo-a"},
		Verbose:      false,
	}

	service := &Service{
		client: client,
		config: config,
	}

	result := service.processProject("group/repo-a")

	if result.Status != StatusCreated {
		t.Errorf("Expected status CREATED, got %s", result.Status)
	}
}

func TestProcessProject_ExistingNonDraftMR_Skips(t *testing.T) {
	client := newMockClient()
	client.addProject("group/repo-b", 2)
	client.addBranch(2, "op-stage")
	client.addBranch(2, "op-rc")
	client.addMergeRequest(2, gitlab.MergeRequest{
		ID:    100,
		IID:   10,
		Title: "Merge op-stage into op-rc",
		Draft: false,
	})

	config := Config{
		OriginBranch: "op-stage",
		TargetBranch: "op-rc",
		Projects:     []string{"group/repo-b"},
		Verbose:      false,
	}

	service := &Service{
		client: client,
		config: config,
	}

	result := service.processProject("group/repo-b")

	if result.Status != StatusSkippedExists {
		t.Errorf("Expected status SKIPPED_EXISTS, got %s", result.Status)
	}
}

func TestProcessProject_ExistingDraftMR_Skips(t *testing.T) {
	client := newMockClient()
	client.addProject("group/repo-c", 3)
	client.addBranch(3, "op-stage")
	client.addBranch(3, "op-rc")
	client.addMergeRequest(3, gitlab.MergeRequest{
		ID:    200,
		IID:   20,
		Title: "Draft: Merge op-stage into op-rc",
		Draft: false,
	})

	config := Config{
		OriginBranch: "op-stage",
		TargetBranch: "op-rc",
		Projects:     []string{"group/repo-c"},
		Verbose:      false,
	}

	service := &Service{
		client: client,
		config: config,
	}

	result := service.processProject("group/repo-c")

	if result.Status != StatusSkippedDraft {
		t.Errorf("Expected status SKIPPED_DRAFT, got %s", result.Status)
	}
}

func TestProcessProject_MissingOriginBranch_Skips(t *testing.T) {
	client := newMockClient()
	client.addProject("group/repo-d", 4)
	client.addBranch(4, "op-rc")

	config := Config{
		OriginBranch: "op-stage",
		TargetBranch: "op-rc",
		Projects:     []string{"group/repo-d"},
		Verbose:      false,
	}

	service := &Service{
		client: client,
		config: config,
	}

	result := service.processProject("group/repo-d")

	if result.Status != StatusSkippedBranch {
		t.Errorf("Expected status SKIPPED_NO_BRANCH, got %s", result.Status)
	}
}
