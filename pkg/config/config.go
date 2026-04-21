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
		"('Team[Team]' IS EMPTY OR 'Team[Team]' != 'ec74d716-af36-4b3c-950f-f79213d08f71-1819') AND " +
		"(component IS EMPTY OR component NOT IN ('Documentation')) AND " +
		"('Release Note Type[Dropdown]' IS EMPTY OR 'Release Note Type[Dropdown]' NOT IN ('Release Note Not Required', 'Do Not Include (note: this means to exclude from release notes and errata)')) AND " +
		"('Release Note Status[Dropdown]' IS EMPTY OR 'Release Note Status[Dropdown]' NOT IN ('Not Required', 'Rejected')) AND " +
		"(" +
		"issuetype IN ('Feature', 'Vulnerability', 'Weakness') OR " +
		"'Release Note Type[Dropdown]' IS NOT EMPTY OR " +
		"'Release Note Status[Dropdown]' IS NOT EMPTY OR " +
		"'Release Note Text' IS NOT EMPTY OR " +
		"'Product Documentation Required[Dropdown]' NOT IN ('EMPTY', 'No') OR " +
		"(SFDC_Cases_Counter IS NOT EMPTY AND SFDC_Cases_Counter != 0) OR " +
		"(issuetype = 'Bug' AND priority IN ('Blocker', 'Critical', 'Major'))" +
		") AND " +
		"fixVersion = %s"
)
