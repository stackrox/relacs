# relacs

`relacs` automates the ACS release process.
It starts as a collection of utility commands, implemented in Go.
Use them to interact with GitHub, Konflux & co.

## Prerequisites

* Completed [Membership and access](https://spaces.redhat.com/spaces/StackRox/pages/690979844/Upstream+Konflux+Release#Upstream%2BKonfluxRelease-membershipsMembershipsandaccess) section from the release instructions.
* Install [kubearchive kubectl plugin](https://kubearchive.github.io/kubearchive/main/cli/installation.html)
* Need to be logged in to our Konflux tenant with Maintainer role, for example with `oc login --web https://api.stone-prd-rh01.pg1f.p1.openshiftapps.com:6443/`.

## Installation

```bash
GOOS=$(go env GOOS) # linux, darwin
GOARCH=$(go env GOARCH) # amd64, arm64

curl -LO https://github.com/stackrox/relacs/releases/latest/download/relacs-$GOOS-$GOARCH
chmod +x relacs-*

# If you are on a Mac
xattr -d com.apple.quarantine relacs-*
```

## Configuration

* `GH_TOKEN`: A personal GitHub token.
* `KUBECTL_PLUGIN_KA_HOST`: KubeArchive host for the tenant, for example: `https://kubearchive-api-server-product-kubearchive.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com`

## Available Commands

See [COMMANDS.md](./COMMANDS.md).

## v1 scope

* Where applicable, also make commands available for input with just a commit.
* Membership and access: `relacs self-test` checks if you are in the right GitHub teams, connected to VPN, Kerberos ticket present, member of Comet (?), GitLab setup (can clone repo with SSH), Konflux maintainer, quay.io/rhacs-eng and registry.redhat.io configured and gives remediation steps if parts are missing
* Listing and categorizing the MintMaker PRs for a given repository and release branch
* Ensure Konflux ReleasePlan and ReleasePlanAdmission resources exist on the cluster
* Ensure that custom snapshot exists on the cluster and is annotated to prevent deletion
* Support: Test upgrade with upstream images
* Support: Test Operator install and upgrade with Konflux images
* ...
