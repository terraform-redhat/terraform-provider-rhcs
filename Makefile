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

# Import path of the project:
import_path:=github.com/terraform-redhat/terraform-provider-rhcs

# Version of the project:
version=$(shell git describe --abbrev=0 | sed 's/^v//' | sed 's/-prerelease\.[0-9]*//')
commit:=$(shell git rev-parse --short HEAD)

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

.PHONY: test tests
test tests: unit-test subsystem-test

.PHONY: fmt_go
fmt_go:
	gofmt -s -l -w $$(find . -name '*.go')

.PHONY: fmt_tf
fmt_tf:
	terraform fmt -recursive examples

.PHONY: fmt
fmt: fmt_go fmt_tf

.PHONY: clean
clean:
	rm -rf "$(RHCS_LOCAL_DIR)"

generate:
	go generate ./...

.PHONY: tools
tools:
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.1.1
	go install github.com/golang/mock/mockgen@v1.6.0

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