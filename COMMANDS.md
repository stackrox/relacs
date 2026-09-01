# Commands

## GitHub group: `relacs github [subcommand]`

Commands for interacting with GitHub repositories, teams, and organizations.

| Command | Description |
|---------|-------------|
| add-team-member | Add a user to a GitHub team in the organization |
| check-release-pr | Check if a release branch PR satisfies the criteria |
| get-commit-for-tag | Get the commit SHA for a given Git tag |
| get-commit-from-acs-version | Get the commit SHA for a given ACS version (from make tag) |
| list-prs | List the pull requests for a given repository and base branch |
| list-vulns-by-cve | List CVEs from image vulnerability scans for a given Git tag |
| list-vulns-by-image | List image vulnerability scan results for a given Git tag |

## Konflux group: `relacs konflux [subcommand]`

Commands for interacting with Konflux build system

| Command | Description |
|---------|-------------|
| get-custom-snapshot | Get the custom snapshot for a given Git tag |
| get-operator-bundle-image | Get the operator bundle image for a given Git tag |
| get-release | Get information about a given Konflux release |
| get-released-operator-bundle-image | Get the operator bundle image for a given release |
| list-conforma-runs | List conforma pipeline runs for a given Git tag |
| list-conforma-violations | Get the Conforma violations for a given Git tag |
| list-releases | List Konflux releases by commit SHA or tag |
| list-pipelineruns-for-tag | Get the pipeline runs for a given Git tag |
| prepare-stage-release | Prepare a custom resource YAML for a given Konflux stage release from a snapshot |
| promote-release-to-prod | Promote a custom resource YAML for a given Konflux stage release to production |

## Miscellaneous commands: `relacs [command]`

Miscellaneous commands directly under the command root.

| Command | Description |
|---------|-------------|
| version | Print the version number of relacs |
