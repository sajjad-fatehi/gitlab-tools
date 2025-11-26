package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	verbose    bool
}

func NewClient(baseURL, token string, verbose bool) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		verbose: verbose,
	}
}

func (c *Client) GetProject(projectPath string) (*Project, error) {
	encodedPath := url.PathEscape(projectPath)
	endpoint := fmt.Sprintf("%s/api/v4/projects/%s", c.baseURL, encodedPath)

	var project Project
	if err := c.doRequest("GET", endpoint, nil, &project); err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}

	return &project, nil
}

func (c *Client) BranchExists(projectID int, branchName string) (bool, error) {
	encodedBranch := url.PathEscape(branchName)
	endpoint := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches/%s", c.baseURL, projectID, encodedBranch)

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check branch %s: %w", branchName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return true, nil
}

func (c *Client) FindOpenMergeRequests(projectID int, sourceBranch, targetBranch string) ([]MergeRequest, error) {
	endpoint := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests?state=opened&source_branch=%s&target_branch=%s",
		c.baseURL,
		projectID,
		url.QueryEscape(sourceBranch),
		url.QueryEscape(targetBranch),
	)

	var mergeRequests []MergeRequest
	if err := c.doRequest("GET", endpoint, nil, &mergeRequests); err != nil {
		return nil, fmt.Errorf("failed to find merge requests: %w", err)
	}

	return mergeRequests, nil
}

func (c *Client) CompareBranches(projectID int, sourceBranch, targetBranch string) (*Compare, error) {
	endpoint := fmt.Sprintf("%s/api/v4/projects/%d/repository/compare?from=%s&to=%s",
		c.baseURL,
		projectID,
		url.QueryEscape(targetBranch),
		url.QueryEscape(sourceBranch),
	)

	var compare Compare
	if err := c.doRequest("GET", endpoint, nil, &compare); err != nil {
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}

	return &compare, nil
}

func (c *Client) ListOpenMergeRequestsByTarget(projectID int, targetBranch string) ([]MergeRequest, error) {
	endpoint := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests?state=opened&target_branch=%s",
		c.baseURL,
		projectID,
		url.QueryEscape(targetBranch),
	)

	var mergeRequests []MergeRequest
	if err := c.doRequest("GET", endpoint, nil, &mergeRequests); err != nil {
		return nil, fmt.Errorf("failed to list merge requests: %w", err)
	}

	return mergeRequests, nil
}

func (c *Client) AcceptMergeRequest(projectID, mrIID int) (*MergeRequest, error) {
	endpoint := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/merge", c.baseURL, projectID, mrIID)

	var mergeRequest MergeRequest
	if err := c.doRequest("PUT", endpoint, nil, &mergeRequest); err != nil {
		return nil, fmt.Errorf("failed to accept merge request: %w", err)
	}

	return &mergeRequest, nil
}

func (c *Client) CreateMergeRequest(projectID int, sourceBranch, targetBranch, title, description string) (*MergeRequest, error) {
	endpoint := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests", c.baseURL, projectID)

	payload := map[string]interface{}{
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"title":         title,
		"description":   description,
	}

	var mergeRequest MergeRequest
	if err := c.doRequest("POST", endpoint, payload, &mergeRequest); err != nil {
		return nil, fmt.Errorf("failed to create merge request: %w", err)
	}

	return &mergeRequest, nil
}

func (c *Client) doRequest(method, endpoint string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}

	resp, err := c.makeRequest(method, endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		fmt.Printf("[DEBUG] %s %s -> %d\n", method, endpoint, resp.StatusCode)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) ListTopics(page, perPage int) ([]Topic, error) {
	endpoint := fmt.Sprintf("%s/api/v4/topics?page=%d&per_page=%d", c.baseURL, page, perPage)

	var topics []Topic
	if err := c.doRequest("GET", endpoint, nil, &topics); err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	return topics, nil
}

func (c *Client) ListProjectsByTopic(topicName string, page, perPage int) ([]Project, error) {
	endpoint := fmt.Sprintf("%s/api/v4/projects?topic=%s&page=%d&per_page=%d",
		c.baseURL,
		url.QueryEscape(topicName),
		page,
		perPage,
	)

	var projects []Project
	if err := c.doRequest("GET", endpoint, nil, &projects); err != nil {
		return nil, fmt.Errorf("failed to list projects for topic %s: %w", topicName, err)
	}

	return projects, nil
}

func (c *Client) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return resp, nil
}
