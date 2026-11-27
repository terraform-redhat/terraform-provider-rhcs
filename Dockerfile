FROM registry.access.redhat.com/ubi10/go-toolset:1.27.4-1795507689 AS builder
ENV GOTOOLCHAIN=auto
COPY . .

ENV GOFLAGS=-buildvcs=false
RUN git config --global --add safe.directory /opt/app-root/src && \
    make prepare_release

# Runtime base is deliberately decoupled from the builder above: CGO_ENABLED=0 is
# exported in the Makefile and build/build_multiarch cross-compiles, so this stage
# only receives release zips of static binaries that it never executes.
FROM registry.access.redhat.com/ubi9/ubi-micro:latest
LABEL description="Terraform Provider RHCS"
LABEL io.k8s.description="Terraform Provider RHCS"
LABEL com.redhat.component="terraform-provider-rhcs"
LABEL distribution-scope="release"
LABEL name="terraform-provider-rhcs" release="X.Y" url="https://github.com/terraform-redhat/terraform-provider-rhcs"
LABEL vendor="Red Hat, Inc."
LABEL version="X.Y"

COPY --from=builder /opt/app-root/src/releases /releases
