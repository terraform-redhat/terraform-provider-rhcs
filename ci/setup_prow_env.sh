#!/bin/bash
set -x

# This file should be sourced in a wrapper script (i.e apply/destroy_folder.sh)

CI_TIMESTAMP=$(date +%s)
echo
echo "[INFO] Running in OpenShift CI (Prow)."
echo "[INFO] Setting up for PROW. timestamp ${CI_TIMESTAMP}"
echo

# Validate some mandatory variables
if [ -z "${ARCHIVE_NAME:-}" ]; then
    error_exit "missing mandatory variable \$ARCHIVE_NAME"
fi

if [ -z "${TF_FOLDER:-}" ]; then
    error_exit "missing mandatory variable \$TF_FOLDER"
fi

if [ -z "${GATEWAY_URL:-}" ]; then
    error_exit "missing mandatory variable \$GATEWAY_URL"
fi
echo "[INFO] OCM gateway url: ${GATEWAY_URL}"

CLUSTER_NAME_FILE="${SHARED_DIR}/cluster-name"
if [[ ! -f "${CLUSTER_NAME_FILE}" ]]; then
    echo "rhcsci-$(mktemp -u XXXXX | tr '[:upper:]' '[:lower:]')" > "${CLUSTER_NAME_FILE}"
fi
CLUSTER_NAME=$(cat "${CLUSTER_NAME_FILE}")

set +x

OCM_TOKEN=$(cat "${CLUSTER_PROFILE_DIR}/ocm-token")
if [ -z "${OCM_TOKEN:-}" ]; then
    error_exit "missing mandatory variable \$OCM_TOKEN"
fi
export TF_VAR_token=${OCM_TOKEN}

#Expose aws credentials as explicit TF_VAR format, and do not use .awscred file
TF_VAR_aws_access_key=$(cat ${CLUSTER_PROFILE_DIR}/.awscred | awk '/\[default\]/{line=1; next} line && /^\[/{exit} line' | grep aws_access_key_id     | awk -F '=' '{print $2}'| sed 's/ //g')
export TF_VAR_aws_access_key
TF_VAR_aws_secret_key=$(cat ${CLUSTER_PROFILE_DIR}/.awscred | awk '/\[default\]/{line=1; next} line && /^\[/{exit} line' | grep aws_secret_access_key | awk -F '=' '{print $2}'| sed 's/ //g')
export TF_VAR_aws_secret_key

set -x

CLOUD_PROVIDER_REGION=${LEASED_RESOURCE}
export AWS_DEFAULT_REGION="${CLOUD_PROVIDER_REGION}"
export TF_VAR_aws_region="${CLOUD_PROVIDER_REGION}"

export TERRAFORM_D_DIR=/root # location of .terraform.d folder
WORK_DIR=${SHARED_DIR}/work
STATE_ARCHIVE="${SHARED_DIR}/${ARCHIVE_NAME}.tar.gz" 
TFVARS_FILE="${WORK_DIR}/terraform.tfvars"

rm -rf "${WORK_DIR}"
mkdir "${WORK_DIR}"

if [[ -f "${STATE_ARCHIVE}" ]]; then
    echo "[INFO] Found TF state archive for '${ARCHIVE_NAME}' @ ${STATE_ARCHIVE}, extracting and ignoring \$TF_VARS"
    tar xvfz "${SHARED_DIR}/${ARCHIVE_NAME}.tar.gz" -C "${WORK_DIR}"
else
    echo "[INFO] Did not find TF state, generating tfvars file."
    { echo "account_role_prefix = \"${CLUSTER_NAME}\"";echo "operator_role_prefix = \"${CLUSTER_NAME}\"";echo "cluster_name = \"${CLUSTER_NAME}\"";} >> "${TFVARS_FILE}"
    echo "${TF_VARS:-}" >> "${TFVARS_FILE}"
fi

echo "[INFO] TF vars:"
cat "${TFVARS_FILE}"

cp -r "${BASE_DIR}/${TF_FOLDER}"/* "${WORK_DIR}/"

echo "[INFO] Done setting up for PROW run"
echo
