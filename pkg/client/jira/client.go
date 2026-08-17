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

	"github.com/Masterminds/semver/v3"
	"github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/utils"
)

// Client represents a Jira client
type Client struct {
	httpClient     *http.Client
	email          string
	token          string
	tokenValidated bool
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

// NewTestClient creates a Jira client with a custom HTTP client for testing.
// The HTTP client can be intercepted by gock for mocking API responses.
func NewTestClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{
		httpClient: httpClient,
		email:      "test@example.com",
		token:      "test-token",
	}
}

var rhelPattern = regexp.MustCompile(`rhel\d+`)

// ValidSeverities are the allowed severity values for CVEs in Jira
var ValidSeverities = []string{"Low", "Moderate", "Important", "Critical"}

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

// isValidSeverity checks if severity matches expected values
func isValidSeverity(severity string) bool {
	for _, valid := range ValidSeverities {
		if severity == valid {
			return true
		}
	}
	return false
}

// ExtractCVEKeyAndYear parses a CVE ID and returns the formatted key and year
func ExtractCVEKeyAndYear(cve string) (string, string, error) {
	cvePattern := regexp.MustCompile(`(cve|CVE)-(\d{4})-(\d+)`)
	matches := cvePattern.FindStringSubmatch(cve)
	if len(matches) != 4 {
		return "", "", fmt.Errorf("invalid CVE ID: %s", cve)
	}
	return strings.ToUpper(matches[0]), matches[2], nil
}

// FetchReleaseData fetches issues and CVEs from Jira using JQL
func (c *Client) FetchReleaseData(snapshot konflux.Snapshot) (*ReleaseData, error) {
	if err := c.validateToken(); err != nil {
		return nil, fmt.Errorf("jira token validation failed: %w", err)
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
			severity := strings.TrimSpace(issue.Fields.Severity.Value)

			if cveID == "" {
				return nil, fmt.Errorf("CVE ID is empty for Vulnerability issue %s", issue.Key)
			}
			if jiraComponent == "" {
				return nil, fmt.Errorf("downstream Component Name is empty for Vulnerability issue %s", issue.Key)
			}
			if severity == "" {
				return nil, fmt.Errorf("severity is empty for Vulnerability issue %s (CVE: %s)", issue.Key, cveID)
			}

			// Validate severity value
			if !isValidSeverity(severity) {
				return nil, fmt.Errorf("unexpected Severity value %q for Vulnerability issue %s (CVE: %s). Expected one of: %s", severity, issue.Key, cveID, strings.Join(ValidSeverities, ", "))
			}

			// Parse CVE key and year
			cveKey, cveYear, err := ExtractCVEKeyAndYear(cveID)
			if err != nil {
				return nil, fmt.Errorf("failed to parse CVE ID for issue %s: %w", issue.Key, err)
			}

			// Convert Jira component name to base component name
			componentName, err := convertJiraComponentName(jiraComponent)
			if err != nil {
				return nil, fmt.Errorf("failed to convert component name for issue %s: %w", issue.Key, err)
			}

			releaseData.CVEs = append(releaseData.CVEs, CVE{
				Code:          cveID,
				Key:           cveKey,
				Year:          cveYear,
				ComponentName: componentName,
				Severity:      severity,
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
		searchURL := fmt.Sprintf("%s/rest/api/3/search/jql", config.JiraBaseURL)

		// Build query parameters
		params := url.Values{}
		params.Add("jql", jql)
		params.Add("fields", fmt.Sprintf("issuetype,%s,%s", config.JiraCVEIDField, config.JiraComponentNameField))
		params.Add("maxResults", fmt.Sprintf("%d", config.JiraPageSize))
		params.Add("startAt", fmt.Sprintf("%d", startAt))

		fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())

		// Create request
		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.addAuthHeader(req)

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("failed to close response body: %w", closeErr)
			}
		}()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("jira API returned status %d: %s", resp.StatusCode, string(body))
		}

		// Parse response
		var searchResp SearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

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
// The validation result is cached to avoid duplicate HTTP requests within the same client session.
func (c *Client) validateToken() (err error) {
	// Return immediately if token was already validated in this session
	if c.tokenValidated {
		return nil
	}

	validateURL := fmt.Sprintf("%s/rest/api/3/myself", config.JiraBaseURL)

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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("jira token is expired or invalid (401 Unauthorized)\nPlease create a new API token at: https://id.atlassian.com/manage-profile/security/api-tokens")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("jira token validation failed with status %d: %s\nPlease create a new API token at: https://id.atlassian.com/manage-profile/security/api-tokens", resp.StatusCode, string(body))
	}

	// Cache the validation result
	c.tokenValidated = true

	return nil
}

func (c *Client) GetIssueFromKey(key string) (*Issue, error) {

	if err := c.validateToken(); err != nil {
		return nil, fmt.Errorf("jira token validation failed: %w", err)
	}

	issueURL := fmt.Sprintf("%s/rest/api/3/issue/%s", config.JiraBaseURL, key)

	req, err := http.NewRequest("GET", issueURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue request: %w", err)
	}

	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute issue request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira API returned status %d: %s", resp.StatusCode, string(body))
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

func (c *Client) addAuthHeader(req *http.Request) {
	credentials := c.email + ":" + c.token
	auth := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
}

// GetLowestUnreleasedFixVersion returns the lowest unreleased fixVersion matching major.minor.
func (c *Client) GetLowestUnreleasedFixVersion(major, minor int) (string, error) {
	if err := c.validateToken(); err != nil {
		return "", fmt.Errorf("jira token validation failed: %w", err)
	}

	versions, err := c.listUnreleasedProjectVersions()
	if err != nil {
		return "", fmt.Errorf("failed to list unreleased project versions: %w", err)
	}

	lowest, err := findLowestMatchingVersion(versions, major, minor)
	if err != nil {
		return "", err
	}
	return lowest, nil
}

func (c *Client) listUnreleasedProjectVersions() ([]FixVersion, error) {
	var allVersions []FixVersion
	startAt := 0

	for {
		versionsURL := fmt.Sprintf("%s/rest/api/3/project/%s/version", config.JiraBaseURL, config.JiraProjectKey)

		params := url.Values{}
		params.Add("status", "unreleased")
		params.Add("maxResults", fmt.Sprintf("%d", config.JiraPageSize))
		params.Add("startAt", fmt.Sprintf("%d", startAt))

		fullURL := fmt.Sprintf("%s?%s", versionsURL, params.Encode())

		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.addAuthHeader(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("failed to close response body: %w", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("jira API returned status %d: %s", resp.StatusCode, string(body))
		}

		var versionsResp ProjectVersionsResponse
		if err := json.NewDecoder(resp.Body).Decode(&versionsResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allVersions = append(allVersions, versionsResp.Values...)

		if versionsResp.IsLast || len(versionsResp.Values) == 0 {
			break
		}

		startAt += len(versionsResp.Values)
	}

	return allVersions, nil
}

// findLowestMatchingVersion returns the name of the lowest unreleased semver version
// among versions that match the given major.minor. Released, archived, and non-semver
// names are ignored.
func findLowestMatchingVersion(versions []FixVersion, major, minor int) (string, error) {
	var lowest *semver.Version
	var lowestName string
	found := false

	for _, version := range versions {
		if version.Released || version.Archived {
			continue
		}
		parsed, err := semver.NewVersion(version.Name)
		if err != nil {
			continue
		}
		if int(parsed.Major()) != major || int(parsed.Minor()) != minor {
			continue
		}
		if !found || parsed.LessThan(lowest) {
			lowest = parsed
			lowestName = version.Name
			found = true
		}
	}

	if !found {
		return "", fmt.Errorf("no unreleased fixVersion found for %d.%d. Confirm that the base version is correct or request a new fixVersion", major, minor)
	}
	return lowestName, nil
}
