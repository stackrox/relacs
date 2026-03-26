package research

import (
	"context"
	"fmt"

	"github.com/stackrox/relacs/pkg/client"
)

type SecurityTracker struct {
	Key string `json:"key" yaml:"key"`
	Component string `json:"component" yaml:"component"`
	CVE string `json:"cve" yaml:"cve"`
	Severity string `json:"severity" yaml:"severity"`
	DueDate string `json:"dueDate" yaml:"dueDate"`
}

const securityTrackersQuery = `project = 'Red Hat Advanced Cluster Security' AND type = Vulnerability AND fixVersion = %s ORDER BY key ASC`
const newSecurityTrackersQuery = `project = 'Red Hat Advanced Cluster Security' AND type = Vulnerability AND fixVersion IS EMPTY AND resolution = Unresolved ORDER BY key ASC`

func FindSecurityTrackers(ctx context.Context, c *client.Client, version string) ([]SecurityTracker, error) {
	var trackers []SecurityTracker
	jiraClient := c.GetJira()
	if jiraClient == nil {
		return nil, fmt.Errorf("jira client not configured. Please set JIRA_EMAIL and JIRA_TOKEN environment variables")
	}

	// jql := fmt.Sprintf(securityTrackersQuery, version)
	jql := newSecurityTrackersQuery
	issues, err := jiraClient.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	for _, issue := range issues {
		trackers = append(trackers, SecurityTracker{
			Key: issue.Key,
			Component: issue.Fields.ComponentName,
			CVE: issue.Fields.CVEId,
			Severity: issue.Fields.Severity.Value,
			DueDate: issue.Fields.DueDate,
		})
	}

	return trackers, nil
}
