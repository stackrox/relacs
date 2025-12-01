# relacs Commands Reference

This document provides a complete list of all available `relacs` commands, organized by command group.

## Command Groups

- [GitHub Commands](#github-commands)
- [Konflux Commands](#konflux-commands)
- [Self-test Commands](#self-test-commands)

---

## GitHub Commands

Commands for interacting with GitHub repositories, teams, and organizations.

### `github get-commit-for-tag`

Get the commit SHA for a given Git tag.

Retrieve the commit SHA associated with a specific Git tag in a GitHub repository.

**Usage:**
```bash
relacs github get-commit-for-tag --org <org> --repo <repo> --tag <tag>
```

**Required Flags:**
- `--org` - GitHub organization
- `--repo` - GitHub repository
- `--tag` - Git tag

**Output formats:** text, JSON, YAML

---

### `github is-team-member`

Check if a user is a member of a GitHub team.

Check whether a specified user (or the authenticated user) is a member of a team in a GitHub organization.

**Usage:**
```bash
relacs github is-team-member --org <org> --team <team> [--user <user>]
```

**Required Flags:**
- `--org` - GitHub organization
- `--team` - Team slug/name

**Optional Flags:**
- `--user` - Username to check (defaults to authenticated user)

**Output formats:** text, JSON, YAML

---

### `github list-prs`

List open pull requests targeting a release branch.

Retrieve and display all open pull requests targeting a specific release branch with their review status.

**Usage:**
```bash
relacs github list-prs --org <org> --repo <repo> --branch <branch> [--include-drafts]
```

**Required Flags:**
- `--org` - GitHub organization
- `--repo` - GitHub repository
- `--branch` - Target branch name

**Optional Flags:**
- `--include-drafts` - Include draft pull requests in the output

**Output formats:** text, JSON, YAML

The command displays:
- Title of the pull request
- Author of the pull request
- Review Status (Approved, Changes Requested, Commented, No Reviews)
- Checks Status (✓ Passed, ✗ Failed, ⋯ Pending, - No Checks)
- Draft status (yes/no, only shown in text output when --include-drafts is used)
- URL to the pull request

Results are sorted by PR number in ascending order.

---

## Konflux Commands

Commands for interacting with Konflux build system and snapshots.

### `konflux find-custom-snapshot-for-tag`

Find Konflux custom snapshot for a Git tag.

Find the Konflux custom snapshot associated with a specific Git tag.

**Usage:**
```bash
relacs konflux find-custom-snapshot-for-tag --base-application-name <name> --org <org> --repo <repo> --tag <tag>
```

**Required Flags:**
- `--base-application-name` - Base application name
- `--org` - GitHub organization
- `--repo` - GitHub repository
- `--tag` - Git tag

**Output formats:** text, JSON, YAML

---

### `konflux get-operator-bundle-image-from-release`

Get the operator bundle image ref for a Konflux release.

Get the operator bundle image ref for a Konflux release by name and namespace.

**Usage:**
```bash
relacs konflux get-operator-bundle-image-from-release --name <name> --namespace <namespace>
```

**Required Flags:**
- `--name` - Release name
- `--namespace` - Release namespace

**Output formats:** text, JSON, YAML

---

## Self-test Commands

Commands to check if the current user has the necessary setup to prepare a release.

### `self-test check-setup`

Check if the current user has the necessary setup to prepare a release.

Verifies the following prerequisites:
- GitHub token (GH_TOKEN) is configured
- GitHub teams membership
- kubectl/oc is installed and available
- Connected to Konflux cluster
- Konflux maintainer role
- KubeArchive plugin is installed
- KubeArchive host (KUBECTL_PLUGIN_KA_HOST) is configured
- VPN is active
- Kerberos ticket present

**Usage:**
```bash
relacs self-test check-setup
```

**Output formats:** text, JSON, YAML

The command performs multiple checks and provides remediation steps for any failed checks.

---

## Global Flags

All commands support the following global flag:

- `-o, --output` - Output format (text|json|yaml) [default: text]

**Example:**
```bash
relacs github list-prs --org stackrox --repo stackrox --branch release-4.6 -o json
```

---

## Configuration

`relacs` requires the following environment variables and configurations:

- `GH_TOKEN` - A personal GitHub token
- `KUBECTL_PLUGIN_KA_HOST` - KubeArchive host for the tenant
- Logged in to Konflux tenant (e.g., via `oc login`)

See [README.md](README.md) for more details on configuration and prerequisites.
