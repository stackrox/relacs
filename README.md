# relacs

`relacs` automates the ACS release process.
It starts as a collection of utility commands, implemented in Go.
Use them to interact with GitHub, Konflux & co.

A minimum required version can be enforced in case of incompatible changes by bumping [MINIMUM_REQUIRED_VERSION](./MINIMUM_REQUIRED_VERSION).
relacs requires network access to GitHub and `GH_TOKEN` variable set on every command execution by design to verify the minimum required version.

## Prerequisites

* Completed [Membership and access](https://spaces.redhat.com/spaces/StackRox/pages/690979844/Upstream+Konflux+Release#Upstream%2BKonfluxRelease-membershipsMembershipsandaccess) section from the release instructions.
* Install [kubearchive kubectl plugin](https://kubearchive.github.io/kubearchive/main/cli/installation.html).
* Need to be logged in to our Konflux tenant with Maintainer role, for example with `oc login --web https://api.stone-prd-rh01.pg1f.p1.openshiftapps.com:6443/`.

## Installation

### Via script

You need an authenticated gh CLI for this.

Without cloning this repository:
```bash
bash <(gh api -H "Accept: application/vnd.github.v3.raw" "/repos/stackrox/relacs/contents/scripts/install-relacs.sh") [version]
```

or in a clone of this repository:

```bash
./scripts/install-relacs.sh [version]
```

`[version]` is an optional argument to select a specific [released version](https://github.com/stackrox/relacs/releases).
If not specified, the latest `relacs` version will be installed.

### Manual installation

1. Go to the [Releases page](https://github.com/stackrox/relacs/releases).
2. Select and download the appropriate binary for your OS and architecture.
3. Move it to your PATH.
4. Confirm it's working with `relacs version`.

### Build and install locally
```bash
git clone git@github.com:stackrox/relacs.git
cd relacs
go install -ldflags "-X 'github.com/stackrox/relacs/cmd/relacs/version.version=$(make version)'" ./cmd/relacs
```
This places the `relacs` binary into your `$GOPATH/bin`. Make sure that directory is in your `PATH`.

## Configuration

* `GH_TOKEN`: A personal GitHub token. Run:
    ```bash
    export GH_TOKEN=$(gh auth token)
    ```
* `KUBECTL_PLUGIN_KA_HOST`: KubeArchive host for the tenant. Run:
    ```bash
    export KUBECTL_PLUGIN_KA_HOST="https://kubearchive-api-server-product-kubearchive.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com"
    ```
* `JIRA_EMAIL`. An email from Jira account. Run:
    ```bash
    export JIRA_EMAIL="<YOUR_EMAIL>"
    ```
* `JIRA_TOKEN`. A personal API Jira token. Can be created on the [API tokens page](https://id.atlassian.com/manage-profile/security/api-tokens). Run:
    ```bash
    export JIRA_TOKEN="<YOUR_JIRA_API_TOKEN>"
    ```
* Other configuration values are hard-coded in [`pkg/config/config.go`](./pkg/config/config.go).

## Available Commands

See [COMMANDS.md](./COMMANDS.md).

### Command Types

Commands in `relacs` interact with resources in the Konflux cluster through **KubeArchive commands**:
The read-only commands query historical data via KubeArchive (using `kubectl ka`). These commands require the `KUBECTL_PLUGIN_KA_HOST` environment variable.
Commands that modify live cluster resources through Kubernetes API may be added in the future.

## Development

When introducing breaking configuration changes or fixing critical bugs that require users to update to the fixed version, release and update the [`MINIMUM_REQUIRED_VERSION`](./MINIMUM_REQUIRED_VERSION) to the next expected version.

1. Merge your PR.
2. Switch to `main` branch. `git checkout main; git pull --ff-only origin main; git fetch --tags --force`
3. Find the most recent `vN.M.K` tag on it. Come up with the following tag number `vX.Y.Z`.
3. Tag the HEAD commit with `git tag -m "vX.Y.Z" vX.Y.Z` and push the tag.
4. CI will build and publish a release with downloadable binaries. Check that it's there on the Releases page.
5. Open a PR bumping [`MINIMUM_REQUIRED_VERSION`](./MINIMUM_REQUIRED_VERSION) to `vX.Y.Z` and get it merged.
