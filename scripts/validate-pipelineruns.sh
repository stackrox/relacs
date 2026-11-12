#!/bin/bash

set -euxo pipefail

MAJOR_VERSION=${1}
MINOR_VERSION=${2}

if [ -z "${MAJOR_VERSION}" ] || [ -z "${MINOR_VERSION}" ]; then
    echo "Usage: $0 <major_version> <minor_version>"
    exit 1
fi

ERRORS=0

for file in .tekton/*.yaml; do
    if [ ! -f "${file}" ]; then
        continue
    fi
    if [ "$(yq '.kind' "${file}")" != "PipelineRun" ]; then
        continue
    fi

    application_name="$(yq '.metadata.labels["appstudio.openshift.io/application"]' "${file}")"
    if [[ ! "${application_name}" =~ .*-${MAJOR_VERSION}-${MINOR_VERSION}$ ]]; then
        echo "${file}: .metadata.labels["appstudio.openshift.io/application"] does not end with '-${MAJOR_VERSION}-${MINOR_VERSION}' (got '${application_name}')"
        ERRORS=$((ERRORS + 1))
    fi

    component_name="$(yq '.metadata.labels["appstudio.openshift.io/component"]' "${file}")"
    if [[ "${component_name}" != "null" && ! "${component_name}" =~ .*-${MAJOR_VERSION}-${MINOR_VERSION}$ ]]; then
        echo "${file}: .metadata.labels["appstudio.openshift.io/component"] does not end with '-${MAJOR_VERSION}-${MINOR_VERSION}' (got '${component_name}')"
        ERRORS=$((ERRORS + 1))
    fi

    service_account_name="$(yq '.spec.taskRunTemplate.serviceAccountName' "${file}")"
    if [[ ! "${service_account_name}" =~ .*-${MAJOR_VERSION}-${MINOR_VERSION}$ ]]; then
        echo "${file}: .spec.taskRunTemplate.serviceAccountName does not end with '-${MAJOR_VERSION}-${MINOR_VERSION}' (got '${service_account_name}')"
        ERRORS=$((ERRORS + 1))
    fi

    # Validate that application, component, and service account exist on Konflux cluster
    # Validate that the RPA contains an entry for the application
done

if [ "${ERRORS}" -gt 0 ]; then
    echo "Found ${ERRORS} errors"
    exit 1
fi

echo "No errors found"
exit 0
