FROM registry.access.redhat.com/ubi8/ubi:latest
WORKDIR /root

# oc
RUN curl -Ls https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable/openshift-client-linux.tar.gz |tar -C /usr/local/bin -xzf - oc

# ocm
RUN yum install -y wget &&\
    wget https://github.com/openshift-online/ocm-cli/releases/download/v0.1.66/ocm-linux-amd64 -O /usr/local/bin/ocm && \
    chmod +x /usr/local/bin/ocm

# go
RUN curl -Ls https://go.dev/dl/go1.18.linux-amd64.tar.gz |tar -C /usr/local -xzf -
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/usr/local/go
ENV TEST_OFFLINE_TOKEN=""

# terraform-provider-rhcs repo
COPY . ./terraform-provider-rhcs

RUN yum install -y yum-utils && \
    yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo &&\
    yum -y install terraform python3 make jq httpd-tools git &&\
    pip3 install PyYAML jinja2 &&\
    go env -w GO111MODULE=on &&\
    go install github.com/onsi/ginkgo/v2/ginkgo@latest &&\
    go install github.com/golang/mock/mockgen@v1.6.0 &&\
    cd terraform-provider-rhcs && go mod tidy && go mod vendor && make install &&\
    chmod -R 777 $GOPATH &&\
    echo 'RUN done'
