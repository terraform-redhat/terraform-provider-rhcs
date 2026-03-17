#
# Copyright (c) 2021 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Disable CGO so that we always generate static binaries:
export CGO_ENABLED=0
export version=""

ifeq ($(shell go env GOOS),windows)
	BINARY=terraform-provider-rhcs.exe
	DESTINATION_PREFIX=$(APPDATA)/terraform.d/plugins
else
	BINARY=terraform-provider-rhcs
	DESTINATION_PREFIX=$(HOME)/.terraform.d/plugins
endif
RHCS_LOCAL_DIR=$(DESTINATION_PREFIX)/terraform.local/local/rhcs

GO_ARCH=$(shell go env GOARCH)
TARGET_ARCH=$(shell go env GOOS)_${GO_ARCH}
GOBIN ?= $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
GCI_VERSION ?= v0.13.4
GCI ?= $(GOBIN)/gci
GOLANGCI_LINT_VERSION ?= v2.6.1
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
LINT_OUTPUT_FLAGS ?=
GO_SOURCE_TARGETS := main.go build internal logging provider subsystem tests tools
LINT_TARGETS := ./ ./build/... ./internal/... ./logging/... ./provider/...

# Import path of the project:
import_path:=github.com/terraform-redhat/terraform-provider-rhcs
GCI_FLAGS := -s standard -s default -s "prefix(k8s)" -s "prefix(sigs.k8s)" -s "prefix(github.com)" -s "prefix(gitlab)" -s "prefix($(import_path))" --custom-order --skip-generated --skip-vendor

# Version of the project:
version=$(shell git describe --abbrev=0 | sed 's/^v//' | sed 's/-prerelease\.[0-9]*//')
commit:=$(shell git rev-parse --short HEAD)
git_status:=$(shell git status --porcelain)
REL_VER=$(version)

# Set the linker flags so that the version will be included in the binaries:
ldflags:=\
	-X $(import_path)/build.Version=$(version) \
	-X $(import_path)/build.Commit=$(commit) \
	$(NULL)

.PHONY: build
build:
	go build -ldflags="$(ldflags)" -o ${BINARY}

.PHONY: install
install: clean build
	platform=$$(terraform version -json | jq -r .platform); \
	extension=""; \
	if [[ "$${platform}" =~ ^windows_.*$$ ]]; then \
		extension=".exe"; \
	fi; \
	if [ -z "${version}" ]; then \
    	version="0.0.2"; \
    fi; \
    dir="$(RHCS_LOCAL_DIR)/$${version}/$(TARGET_ARCH)"; \
	file="terraform-provider-rhcs$${extension}"; \
	mkdir -p "$${dir}"; \
	mv ${BINARY} "$${dir}/$${file}"

.PHONY: subsystem-test
subsystem-test: install
	ginkgo run \
		--succinct \
		-ldflags="$(ldflags)" \
		-r subsystem

.PHONY: unit-test
unit-test:
	ginkgo run \
		--succinct \
		-ldflags="$(ldflags)" \
		-r provider internal/...

.PHONY: unit-test-coverage
unit-test-coverage:
	ginkgo run \
		--succinct \
		--cover \
		--coverprofile coverage.out \
		-ldflags="$(ldflags)" \
		-r provider internal/...


.PHONY: test tests
test tests: unit-test subsystem-test

.PHONY: fmt_go
fmt_go: gci
	"$(GCI)" write $(GCI_FLAGS) $(GO_SOURCE_TARGETS)
	gofmt -s -w $(GO_SOURCE_TARGETS)

.PHONY: fmt_go_check
fmt_go_check: gci
	@files="$$(gofmt -s -l $(GO_SOURCE_TARGETS))"; \
	imports="$$("$(GCI)" list $(GCI_FLAGS) $(GO_SOURCE_TARGETS))"; \
	if [ -z "$$files" ] && [ -z "$$imports" ]; then \
		exit 0; \
	fi; \
	if [ -n "$$files" ]; then \
		echo "Files that need gofmt:"; \
		echo "$$files"; \
	fi; \
	if [ -n "$$imports" ]; then \
		echo "Files that need import formatting (gci):"; \
		echo "$$imports"; \
	fi; \
	exit 1

.PHONY: fmt_tf
fmt_tf:
	terraform fmt -recursive examples

.PHONY: fmt_tests
fmt_tests:
	terraform fmt -recursive tests

.PHONY: fmt
fmt: fmt_go fmt_tf fmt_tests

.PHONY: fmt-check
fmt-check: fmt_go_check

.PHONY: gci
gci:
	GOBIN="$(GOBIN)" go install github.com/daixiang0/gci@$(GCI_VERSION)

.PHONY: golangci-lint
golangci-lint:
	GOBIN="$(GOBIN)" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: golangci-lint
	"$(GOLANGCI_LINT)" run --timeout 5m0s $(LINT_OUTPUT_FLAGS) $(LINT_TARGETS)

.PHONY: clean
clean:
	rm -rf "$(RHCS_LOCAL_DIR)"

generate: tools
	go generate ./...

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate

.PHONY: tools
tools: gci golangci-lint
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.13.2
	go install go.uber.org/mock/mockgen@v0.3.0

.PHONY: e2e_sanity_test
e2e_sanity_test: tools install
	ginkgo run $(ginkgo_flags) \
		--timeout 5h \
		-r \
		--focus-file ci/e2e/.* \
		-- \
		--token-url=$(test_token_url) \
		--gateway-url=$(test_gateway_url) \
		--offline-token=$(test_token) \
		--openshift-version=$(openshift_version) \
		$(NULL)

.PHONY: e2e_clean_tf_states
e2e_clean_tf_states:
	find tests/tf-manifests -name 'terraform.tfstate*' -exec rm -rf {} \; || true

.PHONY: e2e_clean_tf_vars
e2e_clean_tf_vars:
	find tests/tf-manifests -name 'terraform.tfvars*' -exec rm -rf {} \; || true

.PHONY: e2e_clean_tf_init
e2e_clean_tf_init:
	find tests/tf-manifests -name '.terraform*' -exec rm -rf {} \; || true

.PHONY: e2e_clean_tf
e2e_clean_tf: e2e_clean_tf_init e2e_clean_tf_states e2e_clean_tf_vars

.PHONY: apply_folder
apply_folder: install
	bash ./ci/apply_folder.sh

.PHONY: destroy_folder
destroy_folder: install
	bash ./ci/destroy_folder.sh

.PHONY: binary
binary:
	podman run --pull=always --rm registry.ci.openshift.org/ci/rhcs-tf-bin:latest cat /root/terraform-provider-rhcs > ~/terraform-provider-rhcs && chmod +x ~/terraform-provider-rhcs

.PHONY: e2e_test
e2e_test: tools install
	ginkgo run \
        --label-filter $(LabelFilter)\
        --timeout 5h \
        -r \
        --focus-file tests/e2e/.* \
		$(NULL)

.PHONY: check-gen
check-gen: generate
	scripts/assert_no_diff.sh "generate"

.PHONY: prepare_release
prepare_release:
	import_path=${import_path} ldflags="${ldflags}" bash ./build/build_multiarch
