# relacs

`relacs` automates the ACS release process.
It starts as a collection of utility commands, implemented in Go.
Use them to interact with GitHub, Konflux & co.

## Prerequisites

* Completed [Membership and access](https://spaces.redhat.com/spaces/StackRox/pages/690979844/Upstream+Konflux+Release#Upstream%2BKonfluxRelease-membershipsMembershipsandaccess) section from the release instructions.
* Install [kubearchive kubectl plugin](https://kubearchive.github.io/kubearchive/main/cli/installation.html).
* Need to be logged in to our Konflux tenant with Maintainer role, for example with `oc login --web https://api.stone-prd-rh01.pg1f.p1.openshiftapps.com:6443/`.

## Installation

You need an authenticated gh CLI installed.

Without cloning this repository:
```bash
bash <(gh api -H "Accept: application/vnd.github.v3.raw" "/repos/stackrox/relacs/contents/scripts/install-relacs.sh?ref=tm/install-relacs-script") [version]
```

or in a clone of this repository:

```bash
./scripts/install-relacs.sh [version]
```

`[version]` is an optional argument to select a specific [released version](https://github.com/stackrox/relacs/releases).
If not specified, the latest `relacs` version will be installed.

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
* Other configuration values are hard-coded in [`pkg/config/config.go`](./pkg/config/config.go).

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
