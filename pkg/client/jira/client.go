package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/utils"
)

const (
	JiraBaseURL = "https://redhat.atlassian.net"

	// Custom field IDs in Red Hat Jira
	// you can fetch all fields IDs from "https://redhat.atlassian.net/rest/api/3/field" API endpoint
	CVEIDField         = "customfield_10667"
	ComponentNameField = "customfield_10669"

	PageSize = 1000
)

// Client represents a Jira client
type Client struct {
	httpClient *http.Client
	email      string
	token      string
}

// NewClient creates a new Jira client
func NewClient() (*Client, error) {
	email := os.Getenv("JIRA_EMAIL")
	if email == "" {
		return nil, fmt.Errorf("JIRA_EMAIL environment variable is required")
	}

	token := os.Getenv("JIRA_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("JIRA_TOKEN environment variable is required")
	}

	return &Client{
		httpClient: &http.Client{},
		email:      email,
		token:      token,
	}, nil
}

var rhelPattern = regexp.MustCompile(`rhel\d+`)

// convertJiraComponentName converts Jira component name to Konflux component name
// Jira formats:
//   - "advanced-cluster-security/rhacs-<ComponentName>-rhel8" -> "main"
//   - "advanced-cluster-security/rhacs-rhel8-operator" -> "operator"
//   - "rhacs-eng/release-<ComponentName>" -> "<ComponentName>" (e.g. "release-scanner-v4" -> "scanner-v4")
//
// Returns just the component base name (e.g., "main", "scanner-db", "operator")
func convertJiraComponentName(jiraComponentName string) (string, error) {
	// Extract component name from Jira format
	// "advanced-cluster-security/rhacs-main-rhel8" -> "rhacs-main-rhel8"
	parts := strings.Split(jiraComponentName, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Jira component name format: %s", jiraComponentName)
	}

	projectKey := parts[0]
	componentPart := parts[1]

	// "rhacs-eng/release-scanner-v4" -> Konflux component "scanner-v4" (not "release-scanner-v4")
	const releasePrefix = "release-"
	if strings.EqualFold(projectKey, "rhacs-eng") && len(componentPart) >= len(releasePrefix) &&
		strings.EqualFold(componentPart[:len(releasePrefix)], releasePrefix) {
		componentPart = componentPart[len(releasePrefix):]
	}

	// Remove all "rhacs" occurrences
	componentPart = strings.ReplaceAll(componentPart, "rhacs", "")

	// Remove all "rhel<number>" patterns (e.g., "rhel8", "rhel9")
	componentPart = rhelPattern.ReplaceAllString(componentPart, "")

	// Trim leading and trailing dashes
	componentPart = strings.Trim(componentPart, "-")

	if componentPart == "" {
		return "", fmt.Errorf("component name is empty after conversion from: %s", jiraComponentName)
	}

	return componentPart, nil
}

// FetchReleaseData fetches issues and CVEs from Jira using JQL
func (c *Client) FetchReleaseData(snapshot konflux.Snapshot) (*ReleaseData, error) {
	if err := c.validateToken(); err != nil {
		return nil, fmt.Errorf("jira token validation failed: %w\nPlease create a new API token at: https://id.atlassian.com/manage-profile/security/api-tokens", err)
	}

	// Build the default JQL query using the product version from the snapshot
	version, err := utils.NewVersion(snapshot.Metadata.Labels["acs-product-version"])
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	jql := fmt.Sprintf(config.JiraReleaseDataQuery, version)
	if jql == "" {
		return nil, fmt.Errorf("JQL query cannot be empty")
	}

	issues, err := c.searchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	releaseData := &ReleaseData{
		Issues: []string{},
		CVEs:   []CVE{},
	}

	for _, issue := range issues {
		// All issues go to the issues list
		releaseData.Issues = append(releaseData.Issues, issue.Key)

		// Additionally, Vulnerability issues also go to CVEs
		if issue.Fields.IssueType.Name == "Vulnerability" {
			cveID := strings.TrimSpace(issue.Fields.CVEId)
			jiraComponent := strings.TrimSpace(issue.Fields.ComponentName)

			if cveID == "" {
				return nil, fmt.Errorf("CVE ID is empty for Vulnerability issue %s", issue.Key)
			}
			if jiraComponent == "" {
				return nil, fmt.Errorf("downstream Component Name is empty for Vulnerability issue %s", issue.Key)
			}

			// Convert Jira component name to base component name
			componentName, err := convertJiraComponentName(jiraComponent)
			if err != nil {
				return nil, fmt.Errorf("failed to convert component name for issue %s: %w", issue.Key, err)
			}

			releaseData.CVEs = append(releaseData.CVEs, CVE{
				Code:          cveID,
				ComponentName: componentName,
			})
		}
	}

	return releaseData, nil
}

// searchIssues performs a JQL search and returns all issues using pagination
func (c *Client) searchIssues(jql string) ([]Issue, error) {
	var allIssues []Issue
	startAt := 0

	for {
		// Build the search URL using v3 API
		searchURL := fmt.Sprintf("%s/rest/api/3/search/jql", JiraBaseURL)

		// Build query parameters
		params := url.Values{}
		params.Add("jql", jql)
		params.Add("fields", fmt.Sprintf("issuetype,%s,%s", CVEIDField, ComponentNameField))
		params.Add("maxResults", fmt.Sprintf("%d", PageSize))
		params.Add("startAt", fmt.Sprintf("%d", startAt))

		fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())

		// Create request
		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add Basic Auth header
		credentials := make([]byte, 0, len(c.email)+len(c.token)+1)
		credentials = fmt.Appendf(credentials, "%s:%s", c.email, c.token)
		auth := base64.StdEncoding.EncodeToString(credentials)
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
		req.Header.Add("Content-Type", "application/json")

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		// Check response status
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			//nolint:errcheck
			resp.Body.Close()
			return nil, fmt.Errorf("jira API returned status %d: %s", resp.StatusCode, string(body))
		}

		// Parse response
		var searchResp SearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			//nolint:errcheck
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		//nolint:errcheck
		resp.Body.Close()

		// Add issues from this page
		allIssues = append(allIssues, searchResp.Issues...)

		// Check if we need to fetch more pages
		if len(searchResp.Issues) == 0 || startAt+len(searchResp.Issues) >= searchResp.Total {
			break
		}

		// Move to next page
		startAt += len(searchResp.Issues)
	}

	return allIssues, nil
}

// validateToken checks if the Jira token is valid by calling an authenticated endpoint
// Returns an error if the token is expired or invalid
// Note: Calling API with expired token still works but only limited to public tickets.
// This method ensures we can access private tickets as well.
func (c *Client) validateToken() error {
	validateURL := fmt.Sprintf("%s/rest/api/3/myself", JiraBaseURL)

	req, err := http.NewRequest("GET", validateURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	credentials := make([]byte, 0, len(c.email)+len(c.token)+1)
	credentials = fmt.Appendf(credentials, "%s:%s", c.email, c.token)
	auth := base64.StdEncoding.EncodeToString(credentials)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute validation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("jira token is expired or invalid (401 Unauthorized)")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("jira token validation failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
