#!/bin/bash

set -euo pipefail

BINARY_DIR="${HOME}/.local/bin"
BINARY_PATH="${BINARY_DIR}/relacs"

echo ">> This script installs the relacs binary."
echo ">> It expects you to have an authenticated gh CLI."

version="${1:-}"
if [[ -z "${version}" ]]; then
	echo ">> No version provided, installing the latest version."
else
    echo ">> Installing version: ${version}"
fi

if [[ -z "${GH_TOKEN:-}" ]]; then
	echo ">> No GH_TOKEN provided, please set it in your environment."
	exit 1
fi

# Ask user to confirm the installation.
# Only ask if the script is running in an interactive terminal.
if [[ -t 0 ]]; then
	read -p ">> Are you sure you want to install the relacs binary to '${BINARY_PATH}'? (y/n) " confirm
	if [[ "$confirm" != "y" ]]; then
		echo ">> Installation cancelled."
		exit 1
	fi
fi

# Determine the OS and architecture.
case "$(uname -s)" in
	Darwin) GOOS=darwin ;;
	Linux) GOOS=linux ;;
	*) echo "error: unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
	x86_64) GOARCH=amd64 ;;
	arm64|aarch64) GOARCH=arm64 ;;
	*) echo "error: unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

# Download the relacs binary from GitHub Releases.
echo ">> Downloading relacs binary for OS: '${GOOS}' and architecture: '${GOARCH}'"

tmp_dir="$(mktemp -d)"
trap "rm -rf ${tmp_dir}" EXIT

gh release download \
	"${version}" \
	--repo stackrox/relacs \
	--pattern "relacs-${GOOS}-${GOARCH}" \
	--output "${tmp_dir}/relacs-${GOOS}-${GOARCH}"

# Download the sha256sums.txt file from GitHub Releases.
gh release download \
	"${version}" \
	--repo stackrox/relacs \
	--pattern "sha256sums.txt" \
	--output "${tmp_dir}/sha256sums.txt"

# Verify the checksum of the downloaded binary.
echo ">> Verifying checksum of the downloaded binary..."
expected=$(awk "/  bin\/relacs-${GOOS}-${GOARCH}\$/ {print \$1}" "${tmp_dir}/sha256sums.txt")
actual=$(sha256sum "${tmp_dir}/relacs-${GOOS}-${GOARCH}" | awk '{print $1}')
if [[ "$expected" != "$actual" ]]; then
	echo ">> Checksum mismatch: expected ${expected}, got ${actual}"
	exit 1
fi

chmod +x "${tmp_dir}/relacs-${GOOS}-${GOARCH}"
echo ">> Confirming relacs binary is working:"
if ! "${tmp_dir}/relacs-${GOOS}-${GOARCH}" version; then
	echo ">> Error: relacs binary is not working."
	exit 1
fi

# Move the downloaded binary to the binary directory on the user's PATH.
mkdir -p "${BINARY_DIR}"
mv "${tmp_dir}/relacs-${GOOS}-${GOARCH}" "${BINARY_PATH}"
echo ">> relacs binary installed to '${BINARY_PATH}'"
echo ">> Done."

# Tell user to add ${BINARY_DIR} to their PATH if not already there.
if ! echo ":${PATH}:" | grep -qF ":${BINARY_DIR}:"; then
	echo ">> Add '${BINARY_DIR}' to your PATH to use relacs:"
	echo "    export PATH=\"${BINARY_DIR}:\$PATH\""
fi
