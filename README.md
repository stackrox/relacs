# relacs

`relacs` automates the ACS release process.
It starts as a collection of utility commands, implemented in Go.
Use them to interact with GitHub, Konflux & co.

## Prerequisites

* Completed [Membership and access](https://spaces.redhat.com/spaces/StackRox/pages/690979844/Upstream+Konflux+Release#Upstream%2BKonfluxRelease-membershipsMembershipsandaccess) section from the release instructions.
* Install [kubearchive kubectl plugin](https://kubearchive.github.io/kubearchive/main/cli/installation.html).
* Need to be logged in to our Konflux tenant with Maintainer role, for example with `oc login --web https://api.stone-prd-rh01.pg1f.p1.openshiftapps.com:6443/`.
* The `relacs-config` ConfigMap must be deployed to the `rh-acs-tenant` namespace in the Konflux cluster (see [Configuration](#configuration) below).

## Installation

### From GitHub Releases
```bash
GOOS=$(go env GOOS) # linux, darwin
GOARCH=$(go env GOARCH) # amd64, arm64

# For private repositories, authentication is required.
# Option 1: Use gh CLI
#   Optionally set --clobber flag to override existing downloads.
gh release download --repo stackrox/relacs --pattern "relacs-$GOOS-$GOARCH" [--clobber]

# Option 2: Use browser
#   1. Go to https://github.com/stackrox/relacs/releases/latest,
#   2. Download binary for your OS and arch
#   3. Point your shell to your Downloads directory

chmod +x relacs-$GOOS-$GOARCH

# Then move the binary to your PATH, e.g.
mkdir -p ~/.local/bin
mv relacs-$GOOS-$GOARCH ~/.local/bin/relacs
```
### Build and install locally
```bash
git clone git@github.com:stackrox/relacs.git
cd relacs
go install -ldflags "-X 'github.com/stackrox/relacs/cmd/relacs/version.version=$(git describe --tags)'" ./cmd/relacs
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

### ConfigMap Configuration

The `relacs` tool fetches most of its configuration from a ConfigMap named `relacs-config` in the `rh-acs-tenant` namespace of the Konflux cluster. This allows configuration updates without requiring all users to update their `relacs` binary.

**Required**: You must have `kubectl` configured with access to the `rh-acs-tenant` namespace in the Konflux cluster.

The ConfigMap is defined in [`konflux-release-data`](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/tree/main/tenants-config/cluster/stone-prd-rh01/tenants/rh-acs-tenant?ref_type=heads).
The schema is defined in [`pkg.Config`](./pkg/config/config.go) and contains the following configuration values:
- `github-organization`: GitHub organization
- `main-github-repository`: Main GitHub repository
- `konflux-application-prefix`: Konflux application prefix
- `conforma-dev-scenario-prefix`: Conforma dev scenario prefix
- `image-tag-suffix`: Image tag suffix
- `issue-source`: RH Jira base URL
- `jira-release-data-query`: JQL query template for fetching release data from Jira

**Note**: The `KonfluxTenant` (rh-acs-tenant) and `KonfluxHost` remain hardcoded in the CLI as they are needed to know where to fetch the configuration from.

If the ConfigMap is not available or cannot be fetched, `relacs` commands will fail with an error message.

## Available Commands

See [COMMANDS.md](./COMMANDS.md).

### Command Types

Commands in `relacs` interact with resources in two different ways:

1. **KubeArchive commands** - Most read-only commands query historical data via KubeArchive (using `kubectl ka`). These commands require the `KUBECTL_PLUGIN_KA_HOST` environment variable.

2. **Direct cluster commands** - Some commands need to modify resources on the live cluster and connect directly (using `kubectl` or `oc` without the `ka` plugin). These commands require an active cluster connection via `oc login`.

Commands that require direct cluster access:
- `konflux prune-snapshots` - Modifies snapshot annotations on the live cluster

## v1 scope

* Where applicable, also make commands available for input with just a commit.
* Membership and access: `relacs self-test` checks if you are in the right GitHub teams, connected to VPN, Kerberos ticket present, member of Comet (?), GitLab setup (can clone repo with SSH), Konflux maintainer, quay.io/rhacs-eng and registry.redhat.io configured and gives remediation steps if parts are missing
* Listing and categorizing the MintMaker PRs for a given repository and release branch
* Ensure Konflux ReleasePlan and ReleasePlanAdmission resources exist on the cluster
* Ensure that custom snapshot exists on the cluster and is annotated to prevent deletion
* Support: Test upgrade with upstream images
* Support: Test Operator install and upgrade with Konflux images
* ...
