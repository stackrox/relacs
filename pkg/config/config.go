package config

const (
	GitHubOrganization        = "stackrox"
	MainGitHubRepository      = "stackrox"
	KonfluxTenant             = "rh-acs-tenant"
	KonfluxApplicationPrefix  = "acs"
	KonfluxHost               = "konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com"
	ConformaDevScenarioPrefix = "acs-conforma-dev"
	ImageTagSuffix            = "fast"
	IssueSource               = "issues.redhat.com"

	// JQL query template for fetching release data from Jira
	// Use with fmt.Sprintf, passing the fixVersion (tag) as the argument
	JiraReleaseDataQuery = "project = 'Red Hat Advanced Cluster Security' AND " +
		"((labels IN (SecurityTracking) OR type = Vulnerability) OR " +
		"labels IN (docs-RN-needed, doc_RN_needed) OR " +
		"'Release Note Text' IS NOT EMPTY OR " +
		"SFDC_Cases_Counter > 0) AND " +
		"(status = Resolved OR status = Closed OR status = \"Release Pending\") AND " +
		"resolution = Done AND " +
		"fixVersion = %s"
)
