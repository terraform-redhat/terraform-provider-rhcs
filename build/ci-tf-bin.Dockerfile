FROM registry.access.redhat.com/ubi8/ubi-minimal:latest AS builder
WORKDIR /root

# Update the base image and install necessary packages
RUN ARCH=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    microdnf update -y && \
    microdnf install -y git yum-utils python3-pyyaml python3-jinja2 make jq tar gzip go-toolset && \
    microdnf clean all

# terraform-provider-ocm repo
COPY . ./terraform-provider-ocm


RUN cd terraform-provider-ocm && go mod tidy && go mod vendor && make build &&\
    echo 'RUN done'

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
WORKDIR /root
COPY --from=builder /root/terraform-provider-ocm* /root/
