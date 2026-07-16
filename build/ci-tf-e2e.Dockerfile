# Prow E2E runner image (rhcs-tf-e2e). Steps expect /root/terraform-provider-rhcs.
#
# Pinned versions below are Renovate-managed (see renovate.json). oc checksums are
# verified against the versioned mirror sha256sum.txt at build time.

# renovate: datasource=github-releases depName=hashicorp/terraform
ARG TERRAFORM_VERSION=1.15.4
# renovate: datasource=github-releases depName=openshift-online/ocm-cli
ARG OCM_VERSION=0.1.66
# Pin OpenShift client release (no first-class Renovate datasource for mirror.openshift.com).
# When bumping, pick a version under .../clients/ocp/<version>/; build verifies sha256sum.txt.
ARG OC_VERSION=4.22.3

FROM registry.access.redhat.com/ubi9/go-toolset:1.26.5-1784190466 AS builder
ARG TERRAFORM_VERSION
USER root
# Configure safe.directory before COPY: a worktree .git file is invalid inside the image build.
RUN git config --global --add safe.directory /root/terraform-provider-rhcs
WORKDIR /root/terraform-provider-rhcs
ENV GOBIN=/go/bin \
    HOME=/root \
    GOFLAGS=-buildvcs=false \
    PATH="/go/bin:${PATH}"
RUN mkdir -p /go/bin
RUN dnf install -y --setopt=install_weak_deps=False --nodocs \
        yum-utils jq make git gcc-c++ && \
    yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo && \
    dnf -y install "terraform-${TERRAFORM_VERSION}" && \
    dnf clean all && \
    terraform version
COPY . .
RUN go env -w GO111MODULE=on && \
    go mod download && go mod verify && \
    go install -mod=readonly github.com/onsi/ginkgo/v2/ginkgo && \
    go install -mod=readonly go.uber.org/mock/mockgen && \
    make install

FROM registry.access.redhat.com/ubi9/go-toolset:1.26.5-1784190466
ARG TERRAFORM_VERSION
ARG OCM_VERSION
ARG OC_VERSION
USER root
WORKDIR /root
# E2E steps copy /root/terraform-provider-rhcs to ~/terraform-provider-rhcs; HOME must
# not be /root or cp treats source and destination as the same path.
ENV HOME=/home/ci \
    GOBIN=/go/bin \
    PATH="/go/bin:/usr/local/bin:${PATH}" \
    TEST_OFFLINE_TOKEN=""
RUN mkdir -p /home/ci/.terraform.d/plugins/terraform.local/local/rhcs /tmp/cache && \
    chgrp -R 0 /home/ci /tmp/cache && \
    chmod -R g+rwX /home/ci /tmp/cache && \
    chmod g+s /home/ci /home/ci/.terraform.d
# oc — pin release; verify against published sha256sum.txt for that version
RUN OC_TAR=openshift-client-linux.tar.gz && \
    OC_BASE="https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/${OC_VERSION}" && \
    curl -fsSL "${OC_BASE}/sha256sum.txt" -o /tmp/sha256sum.txt && \
    curl -fsSL "${OC_BASE}/${OC_TAR}" -o "/tmp/${OC_TAR}" && \
    grep " ${OC_TAR}$" /tmp/sha256sum.txt | (cd /tmp && sha256sum -c -) && \
    tar -C /usr/local/bin -xzf "/tmp/${OC_TAR}" oc && \
    rm -f "/tmp/${OC_TAR}" /tmp/sha256sum.txt && \
    oc version --client
# ocm — pin release; verify published .sha256 asset
RUN curl -fsSL "https://github.com/openshift-online/ocm-cli/releases/download/v${OCM_VERSION}/ocm-linux-amd64" \
        -o /tmp/ocm-linux-amd64 && \
    curl -fsSL "https://github.com/openshift-online/ocm-cli/releases/download/v${OCM_VERSION}/ocm-linux-amd64.sha256" \
        -o /tmp/ocm-linux-amd64.sha256 && \
    (cd /tmp && sha256sum -c ocm-linux-amd64.sha256) && \
    install -m 0755 /tmp/ocm-linux-amd64 /usr/local/bin/ocm && \
    rm -f /tmp/ocm-linux-amd64 /tmp/ocm-linux-amd64.sha256 && \
    ocm version
COPY build/ci-tf-e2e-requirements.txt /tmp/ci-tf-e2e-requirements.txt
RUN dnf install -y --setopt=install_weak_deps=False --nodocs \
        yum-utils python3 python3-pip make jq httpd-tools git gcc-c++ && \
    yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo && \
    dnf -y install "terraform-${TERRAFORM_VERSION}" && \
    dnf clean all && \
    terraform version && \
    pip3 install --no-cache-dir -r /tmp/ci-tf-e2e-requirements.txt && \
    rm -f /tmp/ci-tf-e2e-requirements.txt
COPY --from=builder /go/bin/ginkgo /go/bin/mockgen /go/bin/
COPY --from=builder /root/terraform-provider-rhcs /root/terraform-provider-rhcs
COPY --from=builder /root/.terraform.d /root/.terraform.d
RUN git config --global --add safe.directory /root/terraform-provider-rhcs && \
    chgrp -R 0 /root /go /home/ci /tmp/cache && \
    chmod -R g+rwX /root /go /home/ci /tmp/cache
