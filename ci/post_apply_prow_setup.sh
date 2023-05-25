#!/bin/bash

# This file should be sourced in a wrapper script (i.e apply_folder.sh***REMOVED***

echo
echo "[INFO] Running in OpenShift CI (Prow***REMOVED***."
echo "[INFO] Post apply setup"
echo

cluster_id_tf=$(terraform output -json | jq -r '.cluster_id'***REMOVED***
if [[ "${cluster_id_tf}" != "null" ]]; then
    echo "[INFO] Found cluster id... Setting up shared files."
    cluster_id=$(echo "${cluster_id_tf}" | jq -r ".value"***REMOVED***
    echo "[INFO] Cluster ID: ${cluster_id}"
    # Cluster ID
    echo "${cluster_id}" > "${SHARED_DIR}/cluster_id"
    # Kubeconfig
    ocm login --token="${OCM_TOKEN}" --url="${GATEWAY_URL}"

    creds=$(ocm get "/api/clusters_mgmt/v1/clusters/${cluster_id}/credentials"***REMOVED***

    echo "${creds}" | jq -r .kubeconfig  > "${SHARED_DIR}/kubeconfig"
    echo "${creds}" | jq -r .admin.password > "${SHARED_DIR}/kubeadmin-password"

    echo "[INFO] Kubeconfig and kubeadmin-password have been stored in shared directory"
fi

echo "[INFO] Done post PROW setup"
echo
