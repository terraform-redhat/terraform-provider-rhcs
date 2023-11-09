#!/bin/bash
set -x

# This file should be sourced in a wrapper script (i.e apply_folder.sh)

echo
echo "[INFO] Running in OpenShift CI (Prow)."
echo "[INFO] Post apply setup"
echo

cluster_id_tf=$(terraform output -json | jq -r '.cluster_id')
if [[ "${cluster_id_tf}" != "null" ]]; then
    echo "[INFO] Found cluster id... Setting up shared files."
    cluster_id=$(echo "${cluster_id_tf}" | jq -r ".value")
    echo "[INFO] Cluster ID: ${cluster_id}"
    # Cluster ID
    echo "${cluster_id}" > "${SHARED_DIR}/cluster_id"
    # Kubeconfig
    ocm login --token="${OCM_TOKEN}" --url="${GATEWAY_URL}"

    #save the json in artifacts
    ocm get /api/clusters_mgmt/v1/clusters/${cluster_id} > $ARTIFACT_DIR/cluster.json

    creds=$(ocm get "/api/clusters_mgmt/v1/clusters/${cluster_id}/credentials")

    echo "${creds}" | jq -r .kubeconfig  > "${SHARED_DIR}/kubeconfig"
    echo "${creds}" | jq -r .admin.password > "${SHARED_DIR}/kubeadmin-password"

    echo "[INFO] Kubeconfig and kubeadmin-password have been stored in shared directory"

    echo "[INFO] wait for CVO availability"
    export KUBECONFIG="${SHARED_DIR}/kubeconfig"
    oc wait nodes --all --for=condition=Ready=true --timeout=30m &
    oc wait clusteroperators --all --for=condition=Progressing=false --timeout=30m &
    wait
fi

echo "[INFO] Done post PROW setup"
echo
