package jira

// Issue represents a Jira issue
type Issue struct {
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

// Fields represents Jira issue fields
type Fields struct {
	IssueType     IssueType `json:"issuetype"`
	CVEId         string    `json:"customfield_10667"` // "CVE ID" field
	ComponentName string    `json:"customfield_10669"` // "Downstream Component Name" field
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
}

// ReleaseData represents parsed release data from Jira
type ReleaseData struct {
	Issues []string
	CVEs   []CVE
}
