#!/bin/bash

set -euo pipefail

BINARY_DIR="${HOME}/.local/bin"
BINARY_PATH="${BINARY_DIR}/relacs"

echo ">> This script installs the relacs binary."
echo ">> It expects you to have an authenticated gh CLI."

version="${1:-}"
if [ -z "${version}" ]; then
	echo ">> No version provided, installing the latest version."
else
    echo ">> Installing version: ${version}"
fi

# Ask user to confirm the installation.
read -p ">> Are you sure you want to install the relacs binary to '${BINARY_PATH}'? (y/n) " confirm
if [ "$confirm" != "y" ]; then
	echo ">> Installation cancelled."
	exit 1
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

echo ">> Downloading relacs binary for OS: '${GOOS}' and architecture: '${GOARCH}'"

tmp_dir=$(mktemp -d)
trap "rm -rf ${tmp_dir}" EXIT

gh release download \
	"${version}" \
	--repo stackrox/relacs \
	--pattern "relacs-${GOOS}-${GOARCH}" \
	--output "${tmp_dir}/relacs-${GOOS}-${GOARCH}" \
	2>/dev/null

chmod +x "${tmp_dir}/relacs-${GOOS}-${GOARCH}"
mkdir -p "$(dirname "${BINARY_PATH}")"
mv "${tmp_dir}/relacs-${GOOS}-${GOARCH}" "${BINARY_PATH}"

echo ">> relacs binary installed to '${BINARY_PATH}'"
echo ">> Confirming relacs binary is working:"
if ! "${BINARY_PATH}" version; then
	echo ">> Error: relacs binary is not working."
	exit 1
fi
echo ">> Done."

# Tell user to add ${BINARY_DIR} to their PATH if not already there.
if ! echo "${PATH}" | grep -q "${BINARY_DIR}"; then
	echo ">> Add '${BINARY_DIR}' to your PATH to use relacs:"
	echo "    export PATH=\"${BINARY_DIR}:\$PATH\""
fi
