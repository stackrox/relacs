# Commands

## GitHub group: `relacs github [subcommand]`

Commands for interacting with GitHub repositories, teams, and organizations.

| Command | Description |
|---------|-------------|
| get-commit-for-tag | Get the commit SHA for a given Git tag |
| list-prs | List the pull requests for a given repository and base branch |

## Konflux group: `relacs konflux [subcommand]`

Commands for interacting with Konflux build system

| Command | Description |
|---------|-------------|
| get-conforma-violations | Get the Conforma violations for a given Git tag |
| get-custom-snapshot | Get the custom snapshot for a given Git tag |
| get-operator-bundle-image | Get the operator bundle image for a given Git tag |
| get-release | Get information about a given Konflux release |
| get-released-operator-bundle-image | Get the operator bundle image for a given release and Git tag |

## Miscellaneous commands: `relacs [command]`

Miscellaneous commands directly under the command root.

| Command | Description |
|---------|-------------|
| version | Print the version number of relacs |
