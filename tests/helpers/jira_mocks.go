package helpers

import (
	"strconv"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/config"
)

// MockJiraTokenValid mocks a successful Jira token validation (/myself)
func MockJiraTokenValid() {
	gock.New(config.JiraBaseURL).
		Get("/rest/api/3/myself").
		Reply(200).
		JSON(map[string]interface{}{
			"accountId": "test-account",
			"emailAddress": "test@example.com",
		})
}

// MockJiraTokenInvalid mocks an expired/invalid Jira token
func MockJiraTokenInvalid() {
	gock.New(config.JiraBaseURL).
		Get("/rest/api/3/myself").
		Reply(401).
		JSON(map[string]interface{}{
			"message": "Unauthorized",
		})
}

// MockUnreleasedProjectVersions mocks the paginated unreleased project versions endpoint
func MockUnreleasedProjectVersions(versionNames ...string) {
	values := make([]map[string]interface{}, len(versionNames))
	for i, name := range versionNames {
		values[i] = map[string]interface{}{
			"name":     name,
			"released": false,
			"archived": false,
		}
	}

	gock.New(config.JiraBaseURL).
		Get("/rest/api/3/project/"+config.JiraProjectKey+"/version").
		MatchParam("status", "unreleased").
		MatchParam("maxResults", strconv.Itoa(config.JiraPageSize)).
		MatchParam("startAt", "0").
		Reply(200).
		JSON(map[string]interface{}{
			"startAt":    0,
			"maxResults": config.JiraPageSize,
			"total":      len(values),
			"isLast":     true,
			"values":     values,
		})
}

// MockJiraIssue mocks a successful GetIssueFromKey response
func MockJiraIssue(key, status, resolution string, fixVersions ...string) {
	versions := make([]map[string]interface{}, len(fixVersions))
	for i, name := range fixVersions {
		versions[i] = map[string]interface{}{
			"name": name,
		}
	}

	gock.New(config.JiraBaseURL).
		Get("/rest/api/3/issue/" + key).
		Reply(200).
		JSON(map[string]interface{}{
			"key": key,
			"fields": map[string]interface{}{
				"status": map[string]interface{}{
					"name": status,
				},
				"resolution": map[string]interface{}{
					"name": resolution,
				},
				"fixVersions": versions,
			},
		})
}

// MockJiraIssueNotFound mocks a 404 for an unknown Jira issue
func MockJiraIssueNotFound(key string) {
	gock.New(config.JiraBaseURL).
		Get("/rest/api/3/issue/" + key).
		Reply(404).
		JSON(map[string]interface{}{
			"errorMessages": []string{"Issue does not exist or you do not have permission to see it."},
		})
}

// MockJiraIssueServerError mocks a 503 Jira server error
func MockJiraIssueServerError(key string) {
	gock.New(config.JiraBaseURL).
		Get("/rest/api/3/issue/" + key).
		Reply(503).
		JSON(map[string]interface{}{
			"message": "Service Unavailable",
		})
}
