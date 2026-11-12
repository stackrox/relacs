#!/usr/bin/env bash
# Extract binary from the ACS release image and print its Go toolchain version.
set -euo pipefail

readonly IMAGE="${1:-}"
readonly BINARY_PATH="${2:-}"

if [[ -z "${IMAGE}" ]]; then
	echo "error: IMAGE is required" >&2
	exit 1
fi

if [[ -z "${BINARY_PATH}" ]]; then
	echo "error: BINARY_PATH is required" >&2
	exit 1
fi

echo ">> Detecting go version of the binary '${BINARY_PATH}' in image '${IMAGE}'..."

if command -v podman &>/dev/null; then
	container_cmd=podman
elif command -v docker &>/dev/null; then
	container_cmd=docker
else
	echo "error: podman or docker is required" >&2
	exit 1
fi

if ! command -v go &>/dev/null; then
	echo "error: go is required" >&2
	exit 1
fi

binary_copy=$(mktemp)
cid=""

cleanup() {
	if [[ -n "${cid}" ]]; then
		"${container_cmd}" rm -f "${cid}" >/dev/null 2>&1 || true
	fi
	rm -f "${binary_copy}"
}
trap cleanup EXIT

cid=$("${container_cmd}" create "${IMAGE}")
"${container_cmd}" cp "${cid}:${BINARY_PATH}" "${binary_copy}"
"${container_cmd}" rm -f "${cid}"
cid=""

chmod +x "${binary_copy}"
go version "${binary_copy}"
echo ">> Done."
