package jira

// Issue represents a Jira issue
type Issue struct {
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

// Fields represents Jira issue fields
type Fields struct {
	IssueType     IssueType       `json:"issuetype"`
	CVEId         string          `json:"customfield_10667"` // "CVE ID" field (config.JiraCVEIDField)
	ComponentName string          `json:"customfield_10669"` // "Downstream Component Name" field (config.JiraComponentNameField)
	Severity      SeverityOption  `json:"customfield_10840"` // "Severity" field (config.JiraSeverityField)
	Status        Status          `json:"status"`
	FixVersions   []FixVersion    `json:"fixVersions"`
	Resolution    Resolution      `json:"resolution"`
}

// SeverityOption represents the Jira Severity dropdown field value
type SeverityOption struct {
	Value string `json:"value"`
}

// Status represents Jira issue status
type Status struct {
	Name string `json:"name"`
}

// FixVersion represents Jira fix version
type FixVersion struct {
	Name     string `json:"name"`
	Released bool   `json:"released"`
	Archived bool   `json:"archived"`
}

// Resolution represents Jira issue resolution
type Resolution struct {
	Name string `json:"name"`
}

// ProjectVersionsResponse represents a paginated Jira project versions response
type ProjectVersionsResponse struct {
	Values     []FixVersion `json:"values"`
	StartAt    int          `json:"startAt"`
	MaxResults int          `json:"maxResults"`
	Total      int          `json:"total"`
	IsLast     bool         `json:"isLast"`
}

// IssueType represents Jira issue type
type IssueType struct {
	Name string `json:"name"`
}

// SearchResponse represents Jira search API response
type SearchResponse struct {
	Issues     []Issue `json:"issues"`
	Total      int     `json:"total"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
}

// CVE represents a CVE with component information
type CVE struct {
	Code          string // CVE identifier (e.g., "CVE-2025-12345")
	ComponentName string // Konflux component name (e.g., "main")
	Severity      string // Severity from Jira ("Low", "Moderate", "Important", "Critical")
}

// ReleaseData represents parsed release data from Jira
type ReleaseData struct {
	Issues []string
	CVEs   []CVE
}
