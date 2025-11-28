# relacs

`relacs` automates the ACS release process.
It starts as a collection of utility commands, implemented in Go.
Use them to interact with GitHub, Konflux & co.

## Prerequisites

* Install [kubearchive kubectl plugin](https://kubearchive.github.io/kubearchive/main/cli/installation.html)

## Configuration

* `GH_TOKEN`: A personal GitHub token.
* `KUBECTL_PLUGIN_KA_HOST`: KubeArchive host for the tenant, for example: `https://kubearchive-api-server-product-kubearchive.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com`
* Need to be logged in to our Konflux tenant, for example with `oc login --web https://api.stone-prd-rh01.pg1f.p1.openshiftapps.com:6443/`.

## Commands

### self-test

Check if the current user has the necessary setup to prepare a release:

```bash
# Basic setup check
relacs self-test check-setup

# Output in JSON format
relacs self-test check-setup -o json
```

The command verifies:

* GitHub token (GH_TOKEN) is configured
* GitHub teams membership
* kubectl/oc is installed and available
* Connected to Konflux cluster
* KubeArchive plugin is installed
* KubeArchive host (KUBECTL_PLUGIN_KA_HOST) is configured

### github list-prs

List open pull requests targeting a release branch:

```bash
# List PRs targeting a specific release branch
relacs github list-prs --org stackrox --repo stackrox --branch release-4.6

# Include draft PRs
relacs github list-prs --org stackrox --repo stackrox --branch release-4.6 --include-drafts

# Output in JSON format
relacs github list-prs --org stackrox --repo stackrox --branch release-4.6 -o json

# Output in YAML format
relacs github list-prs --org stackrox --repo stackrox --branch release-4.6 -o yaml
```

The command displays:

* Title of the pull request
* Author of the pull request
* Review Status (Approved, Changes Requested, Commented, No Reviews)
* Checks Status (✓ Passed, ✗ Failed, ⋯ Pending, - No Checks)
* Draft status (yes/no, only shown in text output when --include-drafts is used)
* URL to the pull request

Results are sorted by PR number in ascending order. Supports text, JSON, and YAML output formats.

## MVP scope

* ✅  Membership and access: `relacs self-test` checks if you are in the right GitHub teams, connected to VPN, Kerberos ticket present, member of Comet (?), GitLab setup (can clone repo with SSH), Konflux maintainer, quay.io/rhacs-eng and registry.redhat.io configured and gives remediation steps if parts are missing (partially is okay for MVP)
* ✅  Listing open PRs targeting the release branch
* ✅  Finding custom snapshots based on tags
* Retrieve the operator-bundle image ref
* Listing and categorizing the MintMaker PRs for a given repository and release branch
* Ensure Konflux ReleasePlan and ReleasePlanAdmission resources exist on the cluster

## v1 scope

* Remaining work of Membership and access self-test.
* Ensure that custom snapshot exists on the cluster and is annotated to prevent deletion
* Check image vulnerabilities and give guidance
* Support: Test upgrade with upstream images
* Support: Test Operator install and upgrade with Konflux images
* ...
