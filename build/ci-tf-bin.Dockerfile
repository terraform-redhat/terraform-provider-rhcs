FROM registry.access.redhat.com/ubi8/ubi-minimal:latest AS builder
WORKDIR /root

# Update the base image and install necessary packages
RUN ARCH=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/***REMOVED*** && \
    microdnf update -y && \
    microdnf install -y git yum-utils python3-pyyaml python3-jinja2 make jq tar gzip go-toolset && \
    yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo && \
    microdnf install -y terraform && \
    microdnf clean all && \
    #Install oc CLI
    curl -sS https://mirror.openshift.com/pub/openshift-v4/$ARCH/clients/ocp/stable/openshift-client-linux.tar.gz | tar -C /usr/local/bin -xzf - oc && \
    curl -Ls https://github.com/openshift-online/ocm-cli/releases/download/v0.1.66/ocm-linux-$ARCH --output  /usr/local/bin/ocm && \
    chmod +x /usr/local/bin/ocm 

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/usr/local/go
ENV TEST_OFFLINE_TOKEN=""

# terraform-provider-ocm repo
COPY . ./terraform-provider-ocm

RUN go env -w GO111MODULE=on &&\
    go install github.com/onsi/ginkgo/v2/ginkgo@latest &&\
    go install github.com/golang/mock/mockgen@v1.6.0 &&\
    cd terraform-provider-ocm && go mod tidy && go mod vendor && make install &&\
    chmod -R 777 $GOPATH &&\
    echo 'RUN done'

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
WORKDIR /root
COPY --from=builder /root/.terraform.d/plugins/terraform.local/local/ocm/1.0.1 /root/.terraform.d/plugins/terraform.local/local/ocm/1.0.1
