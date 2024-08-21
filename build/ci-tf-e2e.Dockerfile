FROM registry.access.redhat.com/ubi9/ubi:latest
WORKDIR /root

# oc
RUN curl -Ls https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable/openshift-client-linux.tar.gz |tar -C /usr/local/bin -xzf - oc

# ocm
RUN yum install -y wget &&\
    wget https://github.com/openshift-online/ocm-cli/releases/download/v0.1.66/ocm-linux-amd64 -O /usr/local/bin/ocm && \
    chmod +x /usr/local/bin/ocm

# go
RUN curl -Ls https://go.dev/dl/go1.21.5.linux-amd64.tar.gz |tar -C /usr/local -xzf -
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/usr/local/go
ENV TEST_OFFLINE_TOKEN=""

# terraform-provider-rhcs repo
COPY . ./terraform-provider-rhcs

RUN yum install -y yum-utils && \
    yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo &&\
    #yum -y install terraform python3 python3-pip make jq httpd-tools git &&\
    yum -y install terraform python3 python3-pip make jq httpd-tools git unzip &&\
    pip3 install PyYAML jinja2 &&\
    go env -w GO111MODULE=on &&\
    go install github.com/onsi/ginkgo/v2/ginkgo@v2.13.2 &&\
    go install go.uber.org/mock/mockgen@v0.3.0 &&\
    cd terraform-provider-rhcs && go mod tidy && go mod vendor && make install &&\
    chmod -R 777 $GOPATH &&\
    echo 'RUN done'

# [WORKAROUND]install terraform version v1.8.5 due to latest version v1.9.5 cannot fetch registry with multiple provider versions
RUN wget https://releases.hashicorp.com/terraform/1.8.5/terraform_1.8.5_linux_amd64.zip &&\
    unzip terraform_1.8.5_linux_amd64.zip &&\
    ls &&\
    mv ./terraform /usr/local/bin &&\
    echo 'Install terraform finished' &&\
    terraform version