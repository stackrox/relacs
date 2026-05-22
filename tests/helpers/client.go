package helpers

import (
	"net/http"

	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
)

// NewTestClient creates a client with a custom HTTP client for testing.
// The HTTP client can be intercepted by gock for mocking API responses.
func NewTestClient(httpClient *http.Client) (*client.Client, error) {
	// Create a default test config
	cfg := &config.Config{
		GitHubOrganization:        "stackrox",
		MainGitHubRepository:      "stackrox",
		KonfluxTenant:             "rh-acs-tenant",
		KonfluxApplicationPrefix:  "acs",
		KonfluxHost:               "konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com",
		ConformaDevScenarioPrefix: "acs-conforma-dev",
		ImageTagSuffix:            "fast",
		IssueSource:               "issues.redhat.com",
		JiraReleaseDataQuery:      "project = 'Red Hat Advanced Cluster Security' AND fixVersion = %s",
	}
	return client.NewTestClient(cfg, httpClient)
}
