# Build from repository root (so go.mod is in context):
#   docker build -f build/ci-tf-e2e.Dockerfile .
FROM registry.access.redhat.com/ubi9/ubi:latest
WORKDIR /root

# oc
RUN curl -Ls https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable/openshift-client-linux.tar.gz |tar -C /usr/local/bin -xzf - oc

# ocm
RUN yum install -y wget &&\
    wget https://github.com/openshift-online/ocm-cli/releases/download/v0.1.66/ocm-linux-amd64 -O /usr/local/bin/ocm && \
    chmod +x /usr/local/bin/ocm

# go — toolchain version follows the `go` line in go.mod (no separate pin to drift)
COPY go.mod /tmp/go.mod
RUN GO_VER=$(awk '/^go / { print $2; exit }' /tmp/go.mod) && \
    curl -fsSL "https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz" | tar -C /usr/local -xzf - && \
    rm -f /tmp/go.mod
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/usr/local/go
ENV TEST_OFFLINE_TOKEN=""

# terraform-provider-rhcs repo
COPY . ./terraform-provider-rhcs

RUN yum install -y yum-utils && \
    yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo &&\
    yum -y install terraform python3 python3-pip make jq httpd-tools git &&\
    yum clean all && \
    pip3 install PyYAML jinja2 &&\
    cd terraform-provider-rhcs && \
    go env -w GO111MODULE=on && \
    # Keep tool installs aligned with go.mod resolution.
    go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo && \
    go install -mod=mod go.uber.org/mock/mockgen && \
    # Do not mutate module state during image build.
    go mod download && go mod verify && make install &&\
    chmod -R 777 $GOPATH &&\
    echo 'RUN done'
