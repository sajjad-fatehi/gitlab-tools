package bulkmr

import (
	"fmt"
	"log"

	"github.com/sajjad-fatehi/gitlab-tools/internal/gitlab"
)

type GitLabClient interface {
	GetProject(projectPath string) (*gitlab.Project, error)
	BranchExists(projectID int, branch string) (bool, error)
	CompareBranches(projectID int, sourceBranch, targetBranch string) (*gitlab.Compare, error)
	FindOpenMergeRequests(projectID int, sourceBranch, targetBranch string) ([]gitlab.MergeRequest, error)
	CreateMergeRequest(projectID int, sourceBranch, targetBranch, title, description string) (*gitlab.MergeRequest, error)
}

type Config struct {
	OriginBranch string
	TargetBranch string
	Projects     []string
	Verbose      bool
}

type ResultStatus string

const (
	StatusCreated         ResultStatus = "CREATED"
	StatusSkippedExists   ResultStatus = "SKIPPED_EXISTS"
	StatusSkippedDraft    ResultStatus = "SKIPPED_DRAFT"
	StatusSkippedBranch   ResultStatus = "SKIPPED_NO_BRANCH"
	StatusSkippedNoChange ResultStatus = "SKIPPED_NO_CHANGE"
	StatusError           ResultStatus = "ERROR"
)

type ProjectResult struct {
	Project         string
	Status          ResultStatus
	MergeRequestID  int
	MergeRequestIID int
	MergeRequestURL string
	ErrorMessage    string
	Details         string
}

type Summary struct {
	Total           int
	Created         int
	SkippedExists   int
	SkippedDraft    int
	SkippedBranch   int
	SkippedNoChange int
	Errors          int
}

type Service struct {
	client GitLabClient
	config Config
}

func NewService(client GitLabClient, config Config) *Service {
	return &Service{
		client: client,
		config: config,
	}
}

func (s *Service) ProcessProjects() ([]ProjectResult, Summary) {
	results := make([]ProjectResult, 0, len(s.config.Projects))
	summary := Summary{Total: len(s.config.Projects)}

	for _, projectPath := range s.config.Projects {
		result := s.processProject(projectPath)
		results = append(results, result)

		switch result.Status {
		case StatusCreated:
			summary.Created++
		case StatusSkippedExists:
			summary.SkippedExists++
		case StatusSkippedDraft:
			summary.SkippedDraft++
		case StatusSkippedBranch:
			summary.SkippedBranch++
		case StatusSkippedNoChange:
			summary.SkippedNoChange++
		case StatusError:
			summary.Errors++
		}
	}

	return results, summary
}

func (s *Service) processProject(projectPath string) ProjectResult {
	result := ProjectResult{
		Project: projectPath,
	}

	project, err := s.client.GetProject(projectPath)
	if err != nil {
		result.Status = StatusError
		result.ErrorMessage = err.Error()
		return result
	}

	if s.config.Verbose {
		log.Printf("[%s] Checking branches...", projectPath)
	}

	originExists, err := s.client.BranchExists(project.ID, s.config.OriginBranch)
	if err != nil {
		result.Status = StatusError
		result.ErrorMessage = fmt.Sprintf("failed to check origin branch: %v", err)
		return result
	}

	if !originExists {
		result.Status = StatusSkippedBranch
		result.Details = fmt.Sprintf("Origin branch '%s' does not exist", s.config.OriginBranch)
		return result
	}

	targetExists, err := s.client.BranchExists(project.ID, s.config.TargetBranch)
	if err != nil {
		result.Status = StatusError
		result.ErrorMessage = fmt.Sprintf("failed to check target branch: %v", err)
		return result
	}

	if !targetExists {
		result.Status = StatusSkippedBranch
		result.Details = fmt.Sprintf("Target branch '%s' does not exist", s.config.TargetBranch)
		return result
	}

	if s.config.Verbose {
		log.Printf("[%s] Checking existing merge requests...", projectPath)
	}

	existingMRs, err := s.client.FindOpenMergeRequests(project.ID, s.config.OriginBranch, s.config.TargetBranch)
	if err != nil {
		result.Status = StatusError
		result.ErrorMessage = fmt.Sprintf("failed to find existing merge requests: %v", err)
		return result
	}

	if len(existingMRs) > 0 {
		hasNonDraft := false
		var draftMR *gitlab.MergeRequest

		for i := range existingMRs {
			mr := &existingMRs[i]
			if !mr.IsDraft() {
				hasNonDraft = true
				result.Status = StatusSkippedExists
				result.MergeRequestID = mr.ID
				result.MergeRequestIID = mr.IID
				result.MergeRequestURL = mr.WebURL
				result.Details = fmt.Sprintf("Open MR already exists: !%d", mr.IID)
				break
			} else {
				draftMR = mr
			}
		}

		if !hasNonDraft && draftMR != nil {
			result.Status = StatusSkippedDraft
			result.MergeRequestID = draftMR.ID
			result.MergeRequestIID = draftMR.IID
			result.MergeRequestURL = draftMR.WebURL
			result.Details = fmt.Sprintf("Draft MR exists: !%d (%s)", draftMR.IID, draftMR.Title)
		}

		return result
	}

	if s.config.Verbose {
		log.Printf("[%s] Comparing branches...", projectPath)
	}

	compare, err := s.client.CompareBranches(project.ID, s.config.OriginBranch, s.config.TargetBranch)
	if err != nil {
		result.Status = StatusError
		result.ErrorMessage = fmt.Sprintf("failed to compare branches: %v", err)
		return result
	}

	if !compare.HasChanges() {
		result.Status = StatusSkippedNoChange
		result.Details = fmt.Sprintf("No changes between %s and %s", s.config.OriginBranch, s.config.TargetBranch)
		return result
	}

	if s.config.Verbose {
		log.Printf("[%s] Found %d commit(s) with changes, creating merge request...", projectPath, len(compare.Commits))
	}

	title := fmt.Sprintf("Merge %s into %s", s.config.OriginBranch, s.config.TargetBranch)
	description := fmt.Sprintf("This merge request was created automatically by gitlab-tools.\n\n**Source Branch**: `%s`\n**Target Branch**: `%s`",
		s.config.OriginBranch, s.config.TargetBranch)

	mr, err := s.client.CreateMergeRequest(project.ID, s.config.OriginBranch, s.config.TargetBranch, title, description)
	if err != nil {
		result.Status = StatusError
		result.ErrorMessage = fmt.Sprintf("failed to create merge request: %v", err)
		return result
	}

	result.Status = StatusCreated
	result.MergeRequestID = mr.ID
	result.MergeRequestIID = mr.IID
	result.MergeRequestURL = mr.WebURL
	result.Details = fmt.Sprintf("MR !%d: %s", mr.IID, mr.WebURL)

	return result
}
