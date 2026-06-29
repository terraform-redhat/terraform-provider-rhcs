FROM registry.access.redhat.com/ubi9/ubi:9.8-1782365825
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
    GO_TAR="go${GO_VER}.linux-amd64.tar.gz" && \
    curl -fsSL "https://dl.google.com/go/${GO_TAR}" -o "/tmp/${GO_TAR}" && \
    curl -fsSL "https://dl.google.com/go/${GO_TAR}.sha256" -o /tmp/go.sha256 && \
    printf '%s  %s\n' "$(tr -d '\n' < /tmp/go.sha256)" "${GO_TAR}" | (cd /tmp && sha256sum -c -) && \
    tar -C /usr/local -xzf "/tmp/${GO_TAR}" && \
    rm -f "/tmp/${GO_TAR}" /tmp/go.sha256 /tmp/go.mod && \
    /usr/local/go/bin/go version
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/usr/local/go
ENV TEST_OFFLINE_TOKEN=""

# terraform-provider-rhcs repo
COPY . ./terraform-provider-rhcs

RUN yum install -y yum-utils && \
    yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo &&\
    yum -y install terraform python3 python3-pip make jq httpd-tools git gcc-c++ &&\
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
