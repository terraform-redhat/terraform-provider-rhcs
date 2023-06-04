#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail
set -x
trap 'CHILDREN=$(jobs -p); if test -n "${CHILDREN}"; then kill ${CHILDREN} && wait; fi' TERM

function error_exit() {
    msg=${1}
    >&2 echo "[Error] ${msg}"
    exit 1
}

function prow_archive_state() {
    if [ -n "${OPENSHIFT_CI:-}" ]; then
        echo
        echo "[INFO] Running in OpenShift CI (Prow)."
        echo "[INFO] Archiving state for PROW."
        echo
        set -o xtrace
        tar cvfz "${STATE_ARCHIVE}" ./*.tf*  # For next steps to use
        set +o xtrace
        echo "[INFO] ${STATE_ARCHIVE} created."

        run_artifact_dir="${ARTIFACT_DIR}/destroy_folder_${CI_TIMESTAMP}"
        echo "[INFO] Uploading tf files to ARTIFACT_DIR @ ${run_artifact_dir}"
        mkdir "${run_artifact_dir}"
        cp ./*.tf "${run_artifact_dir}"/
    fi

    if [ -n "${1:-}" ]; then
        error_exit "TF command failed."
    fi
}

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
BASE_DIR=$(realpath "${THIS_DIR}/..")

echo
echo ">> Running destroy_folder script"
echo "   -----------------------------"
echo
if [ -n "${OPENSHIFT_CI:-}" ]; then
    source "${THIS_DIR}/setup_prow_env.sh"
fi

# Setup defaults
TERRAFORM_D_DIR=${TERRAFORM_D_DIR:-"${HOME}"}
WORK_DIR=${WORK_DIR:-"${BASE_DIR}/playground"}

echo "[INFO] Will apply folder."
echo "[INFO] - WORK_DIR: ${WORK_DIR}"
echo "[INFO] - TERRAFORM_D_DIR: ${TERRAFORM_D_DIR}"
echo "[INFO] - THIS_DIR: ${THIS_DIR}"
echo "[INFO] - BASE_DIR: ${BASE_DIR}"

if [[ ! -d ${WORK_DIR} ]]; then
    error_exit "can't load WORK_DIR"
fi

set -o xtrace
export TF_LOG=debug

cd "${WORK_DIR}"
HOME=${TERRAFORM_D_DIR} terraform init
HOME=${TERRAFORM_D_DIR} terraform destroy -auto-approve || prow_archive_state "true"

set +o xtrace

prow_archive_state

echo
echo "[INFO] Finished destroying the terraform folder ${WORK_DIR} successfully"
echo
