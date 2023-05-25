FROM registry.access.redhat.com/ubi8/ubi:latest
WORKDIR /root

RUN dnf install -y \
    curl \
    tar \
    unzip \
    python3 \
    make \
    jq \
    httpd-tools \
    git

# oc
RUN curl -Ls https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable/openshift-client-linux.tar.gz | tar -C /usr/local/bin -xzf - oc

# aws
RUN curl -Ls "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o /tmp/awscliv2.zip && \
    unzip /tmp/awscliv2.zip && \
    ./aws/install && \
    rm -f /tmp/awscliv2.zip

# ocm
RUN curl -Ls https://github.com/openshift-online/ocm-cli/releases/download/v0.1.66/ocm-linux-amd64 -o /usr/local/bin/ocm && \
    chmod +x /usr/local/bin/ocm

# go
RUN curl -Ls https://go.dev/dl/go1.18.linux-amd64.tar.gz | tar -C /usr/local -xzf -
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/usr/local/go
RUN go env -w GO111MODULE=on
ENV TEST_OFFLINE_TOKEN=""

# terraform
RUN dnf install -y 'dnf-command(config-manager)' && \
    dnf config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo && \
    dnf install -y terraform

# terraform-provider-ocm repo
COPY . ./terraform-provider-ocm

RUN pip3 install PyYAML jinja2 && \
    go install github.com/onsi/ginkgo/v2/ginkgo@latest && \
    go install github.com/golang/mock/mockgen@v1.6.0 && \
    cd terraform-provider-ocm && make install && \
    chmod -R 777 $GOPATH
