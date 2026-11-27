# ci-operator build_root image (ubi10/go-toolset).
#
# Active for the main config: openshift/release
# ci-operator/config/terraform-redhat/terraform-provider-rhcs/
# terraform-redhat-terraform-provider-rhcs-main.yaml already sets
#
#   build_root:
#     project_image:
#       dockerfile_path: build/custom-ci-build-root.Dockerfile
#
# so this file — not .ci-operator.yaml — controls Go for the pre-push-checks,
# govulncheck and Snyk `security` jobs. The two e2e variants
# (...-main__e2e.yaml and ...-main__e2e-presubmits.yaml) still use
# build_root.from_repository: true and therefore still read .ci-operator.yaml;
# their runner image compiles the provider in its own ubi10/go-toolset builder
# stage, so the ocp/builder Go version does not gate them.
# Motivation: pin Go via go-toolset so we can bump the toolchain when we need
# to (including the Snyk security presubmit, which runs go mod tidy/vendor on
# the build_root/src image).
FROM registry.access.redhat.com/ubi10/go-toolset:1.27.4-1795507689
ENV GOTOOLCHAIN=auto

USER root

RUN dnf install -y --setopt=install_weak_deps=False --nodocs git jq && \
    dnf clean all
