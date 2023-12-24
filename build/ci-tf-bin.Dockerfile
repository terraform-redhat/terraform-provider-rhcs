FROM registry.access.redhat.com/ubi8/ubi-minimal:latest AS builder
WORKDIR /root

# Update the base image and install necessary packages
RUN microdnf update -y && \
    microdnf install -y git make go-toolset tar && \
    microdnf clean all

RUN curl -Ls https://go.dev/dl/go1.21.5.linux-amd64.tar.gz |tar -C /usr/local -xzf -
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/usr/local/go
ENV TEST_OFFLINE_TOKEN=""

# terraform-provider-rhcs repo
COPY . ./terraform-provider-rhcs


RUN cd terraform-provider-rhcs && go mod tidy && go mod vendor && make build &&\
    echo 'RUN done'

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
WORKDIR /root
COPY --from=builder /root/terraform-provider-rhcs* /root/
