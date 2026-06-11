package config

const (
	GitHubOrganization        = "stackrox"
	MainGitHubRepository      = "stackrox"
	KonfluxTenant             = "rh-acs-tenant"
	KonfluxApplicationPrefix  = "acs"
	KonfluxHost               = "konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com"
	ConformaDevScenarioPrefix = "acs-conforma-dev"
	ImageTagSuffix            = "fast"
	IssueSource               = "redhat.atlassian.net"

	// JQL query template for fetching release data from Jira
	// Use with fmt.Sprintf, passing the fixVersion (tag) as the argument
	// Query in the UI:
	// https://redhat.atlassian.net/issues?jql=project%20%3D%20%27Red%20Hat%20Advanced%20Cluster%20Security%27%20AND%20%20%20(%22Team%5BTeam%5D%22%20IS%20EMPTY%20OR%20%22Team%5BTeam%5D%22%20!%3D%20ec74d716-af36-4b3c-950f-f79213d08f71-1819)%20AND%20%20%20(component%20IS%20EMPTY%20OR%20component%20NOT%20IN%20(%22Documenation%22))%20AND%20(%27Release%20Note%20Type%5BDropdown%5D%27%20IS%20EMPTY%20OR%20%27Release%20Note%20Type%5BDropdown%5D%27%20NOT%20IN%20(%22Release%20Note%20Not%20Required%22%2C%20%22Do%20Not%20Include%20(note%3A%20this%20means%20to%20exclude%20from%20release%20notes%20and%20errata)%22))%20AND%20%20(%27Release%20Note%20Status%5BDropdown%5D%27%20IS%20EMPTY%20OR%20%27Release%20Note%20Status%5BDropdown%5D%27%20NOT%20IN%20(%22Not%20Required%22%2C%20Rejected))%20AND%20%20(%20%20%20%20issuetype%20IN%20(Feature%2C%20Vulnerability%2C%20Weakness)%20OR%20%20%20%20%27Release%20Note%20Type%5BDropdown%5D%27%20IS%20NOT%20EMPTY%20OR%20%20%20%20%27Release%20Note%20Status%5BDropdown%5D%27%20IS%20NOT%20EMPTY%20OR%20%20%20%20%27Release%20Note%20Text%27%20IS%20NOT%20EMPTY%20OR%20%20%20%20%22Product%20Documentation%20Required%5BDropdown%5D%22%20NOT%20IN%20(EMPTY%2C%20No)%20OR%20%20%20%20(SFDC_Cases_Counter%20IS%20NOT%20EMPTY%20AND%20SFDC_Cases_Counter%20!%3D%200)%20OR%20%20%20%20(issuetype%20%3D%20Bug%20AND%20priority%20IN%20(Blocker%2C%20Critical%2C%20Major))%20%20)%20AND%20(resolution%20IS%20EMPTY%20OR%20resolution%20IN%20(Done%2C%20Done-Errata%2C%20%27Test%20Pending%27))%20AND%20fixVersion%20%3D%20%224.11.0%22&src=confmacro
	JiraReleaseDataQuery = "project = 'Red Hat Advanced Cluster Security' AND " +
		"('Team[Team]' IS EMPTY OR 'Team[Team]' != 'ec74d716-af36-4b3c-950f-f79213d08f71-1819') AND " +
		"(component IS EMPTY OR component NOT IN ('Documenation', 'Documentation', 'documentation', 'DOCUMENTATION', 'Documentation ', 'Documentation_')) AND " +
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
		"(resolution IS EMPTY OR resolution IN (Done, Done-Errata, 'Test Pending')) AND " +
		"fixVersion = %s"
)
