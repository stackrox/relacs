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

## MVP scope

* ✅  Membership and access: `relacs self-test` checks if you are in the right GitHub teams, connected to VPN, Kerberos ticket present, member of Comet (?), GitLab setup (can clone repo with SSH), Konflux maintainer, quay.io/rhacs-eng and registry.redhat.io configured and gives remediation steps if parts are missing (partially is okay for MVP)
* ✅  Listing open PRs targeting the release branch
* ✅  Finding custom snapshots based on tags
* ✅  Retrieve the operator-bundle image ref
* ✅  Retrieve Conforma Dev policy violations
* Listing and categorizing the MintMaker PRs for a given repository and release branch
* Ensure Konflux ReleasePlan and ReleasePlanAdmission resources exist on the cluster

## v1 scope

* Remaining work of Membership and access self-test.
* Ensure that custom snapshot exists on the cluster and is annotated to prevent deletion
* Check image vulnerabilities and give guidance
* Support: Test upgrade with upstream images
* Support: Test Operator install and upgrade with Konflux images
* ...
