# Intended ci-operator build_root image (ubi9/go-toolset).
#
# Not active yet: openshift/release still uses build_root.from_repository: true,
# which reads .ci-operator.yaml (ocp/builder imagestream). After this lands on
# main, switch release configs to:
#
#   build_root:
#     project_image:
#       dockerfile_path: build/custom-ci-build-root.Dockerfile
#
# That pattern is widely used in openshift/release. Motivation: pin Go via
# go-toolset so we can bump the toolchain when we need to (including fixing the
# Snyk security presubmit, which runs go mod tidy/vendor on the build_root/src
# image).
FROM registry.access.redhat.com/ubi9/go-toolset:1.26.5-1784638038

USER root

RUN dnf install -y --setopt=install_weak_deps=False --nodocs git jq && \
    dnf clean all
