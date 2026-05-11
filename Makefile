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
SHELL := bash

export CGO_ENABLED=0
export version=""

ifeq ($(shell go env GOOS),windows)
	BINARY=terraform-provider-rhcs.exe
	DESTINATION_PREFIX=$(APPDATA)/terraform.d/plugins
	BIN_EXT=.exe
else
	BINARY=terraform-provider-rhcs
	DESTINATION_PREFIX=$(HOME)/.terraform.d/plugins
	BIN_EXT=
endif
RHCS_LOCAL_DIR=$(DESTINATION_PREFIX)/terraform.local/local/rhcs

GO_ARCH=$(shell go env GOARCH)
TARGET_ARCH=$(shell go env GOOS)_${GO_ARCH}
LOCALBIN ?= $(CURDIR)/bin
LOCALBIN_ABS := $(abspath $(LOCALBIN))
RUN_CHECKS_SCRIPT := ./hack/run-checks.sh

GCI_VERSION ?= v0.13.4
GOLANGCI_LINT_VERSION ?= v2.6.1
ADDLICENSE_VERSION ?= v1.2.0

GCI := $(LOCALBIN)/gci$(BIN_EXT)
GINKGO := $(LOCALBIN)/ginkgo$(BIN_EXT)
MOCKGEN := $(LOCALBIN)/mockgen$(BIN_EXT)
GOLANGCI_LINT := $(LOCALBIN)/golangci-lint$(BIN_EXT)
ADDLICENSE := $(LOCALBIN)/addlicense$(BIN_EXT)

LINT_OUTPUT_FLAGS ?=
GO_SOURCE_TARGETS := main.go build internal logging provider subsystem tests tools
LINT_TARGETS := ./ ./build/... ./internal/... ./logging/... ./provider/... ./tests/...

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

$(LOCALBIN):
	mkdir -p "$(LOCALBIN)"

$(GCI): | $(LOCALBIN)
	GOBIN="$(LOCALBIN_ABS)" go install github.com/daixiang0/gci@$(GCI_VERSION)

$(GINKGO): | $(LOCALBIN)
	# resolve tool version from go.mod
	GOBIN="$(LOCALBIN_ABS)" go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo

$(MOCKGEN): | $(LOCALBIN)
	# resolve tool version from go.mod
	GOBIN="$(LOCALBIN_ABS)" go install -mod=mod go.uber.org/mock/mockgen

$(GOLANGCI_LINT): | $(LOCALBIN)
	GOBIN="$(LOCALBIN_ABS)" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(ADDLICENSE): | $(LOCALBIN)
	GOBIN="$(LOCALBIN_ABS)" go install github.com/google/addlicense@$(ADDLICENSE_VERSION)

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
subsystem-test: $(GINKGO) install
	$(GINKGO) run \
		--succinct \
		-ldflags="$(ldflags)" \
		-r subsystem

.PHONY: unit-test
unit-test: $(GINKGO)
	$(GINKGO) run \
		--succinct \
		-ldflags="$(ldflags)" \
		-r provider internal/...

.PHONY: unit-test-coverage
unit-test-coverage: $(GINKGO)
	$(GINKGO) run \
		--succinct \
		--cover \
		--coverprofile coverage.out \
		-ldflags="$(ldflags)" \
		-r provider internal/...


.PHONY: test tests
test tests: unit-test subsystem-test e2e-unit-test

.PHONY: fmt_go
fmt_go: $(GCI)
	"$(GCI)" write $(GCI_FLAGS) $(GO_SOURCE_TARGETS)
	gofmt -s -w $(GO_SOURCE_TARGETS)

.PHONY: fmt_go_check
fmt_go_check: $(GCI)
	@test -z "$$("$(GCI)" list $(GCI_FLAGS) $(GO_SOURCE_TARGETS))"
	@test -z "$$(gofmt -s -l $(GO_SOURCE_TARGETS))"

.PHONY: fmt_tf
fmt_tf:
	terraform fmt -recursive examples

.PHONY: fmt_tests
fmt_tests:
	terraform fmt -recursive tests

.PHONY: fmt
fmt:
	./hack/fmt.sh

.PHONY: fmt-staged
fmt-staged:
	./hack/fmt-staged.sh

.PHONY: fmt-check
fmt-check: fmt_go_check
	@for dir in examples tests; do \
		if [ -d "$$dir" ]; then \
			terraform fmt -check -recursive "$$dir" >/dev/null; \
		fi; \
	done

.PHONY: gci
gci: $(GCI)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)

.PHONY: lint
lint: $(GOLANGCI_LINT)
	"$(GOLANGCI_LINT)" run $(LINT_OUTPUT_FLAGS) $(LINT_TARGETS)

.PHONY: clean
clean:
	rm -rf "$(RHCS_LOCAL_DIR)"

generate: $(MOCKGEN)
	PATH="$(LOCALBIN_ABS):$$PATH" go generate ./...

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate

.PHONY: tools
tools:
	@$(MAKE) --no-print-directory $(GCI) $(GINKGO) $(MOCKGEN) $(GOLANGCI_LINT) $(ADDLICENSE)

.PHONY: e2e-unit-test
e2e-unit-test: $(GINKGO)
	$(GINKGO) run \
		--succinct \
		-ldflags="$(ldflags)" \
		-r tests/utils/...

.PHONY: e2e_sanity_test
e2e_sanity_test: $(GINKGO) install
	$(GINKGO) run $(ginkgo_flags) \
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
e2e_test: $(GINKGO) install
	$(GINKGO) run \
        --label-filter $(LabelFilter)\
        --timeout 5h \
        -r \
        --focus-file tests/e2e/.* \
		$(NULL)

.PHONY: coverage-changed-files
coverage-changed-files:
	./hack/coverage-changed-files.sh

.PHONY: check-gen
check-gen:
	@before_unstaged=$$(mktemp); \
	before_staged=$$(mktemp); \
	before_untracked=$$(mktemp); \
	trap 'rm -f "$$before_unstaged" "$$before_staged" "$$before_untracked"' EXIT; \
	git diff --binary --no-ext-diff > "$$before_unstaged"; \
	git diff --cached --binary --no-ext-diff > "$$before_staged"; \
	git ls-files --others --exclude-standard > "$$before_untracked"; \
	$(MAKE) --no-print-directory generate; \
	scripts/assert_no_diff.sh "generate" "$$before_unstaged" "$$before_staged" "$$before_untracked"

.PHONY: install-hooks
install-hooks:
	@./hack/install-git-hooks.sh

.PHONY: license-check
license-check: $(ADDLICENSE)
	@echo "Checking for missing license headers..."
	@ADDLICENSE="$(ADDLICENSE)" bash scripts/add-license-header.sh -check

.PHONY: license-add
license-add: $(ADDLICENSE)
	@echo "Adding license headers to files..."
	@ADDLICENSE="$(ADDLICENSE)" bash scripts/add-license-header.sh

.PHONY: license-add-staged
license-add-staged: $(ADDLICENSE)
	@ADDLICENSE="$(ADDLICENSE)" ./hack/license-add-staged.sh

.PHONY: basic-checks
basic-checks:
	@$(RUN_CHECKS_SCRIPT) basic

.PHONY: pre-commit-checks
pre-commit-checks:
	@$(MAKE) --no-print-directory fmt-staged
	@$(MAKE) --no-print-directory license-add-staged

.PHONY: pre-push-checks
pre-push-checks:
	@$(RUN_CHECKS_SCRIPT) pre-push

.PHONY: run-checks
run-checks:
	@$(RUN_CHECKS_SCRIPT) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: commits/check
commits/check:
	@./hack/commit-msg-verify.sh

RUN_CHECKS_PASSTHROUGH_ARGS := basic pre-push --dry-run --list-steps -h --help
.PHONY: $(RUN_CHECKS_PASSTHROUGH_ARGS)
$(RUN_CHECKS_PASSTHROUGH_ARGS):
	@:

.PHONY: prepare_release
prepare_release:
	import_path=${import_path} ldflags="${ldflags}" bash ./build/build_multiarch
