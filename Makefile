#
# Copyright (c***REMOVED*** 2021 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

ifeq ($(shell go env GOOS***REMOVED***,windows***REMOVED***
	BINARY=terraform-provider-rhcs.exe
	DESTINATION_PREFIX=$(APPDATA***REMOVED***/terraform.d/plugins
else
	BINARY=terraform-provider-rhcs
	DESTINATION_PREFIX=$(HOME***REMOVED***/.terraform.d/plugins
endif
RHCS_LOCAL_DIR=$(DESTINATION_PREFIX***REMOVED***/terraform.local/local/rhcs

GO_ARCH=$(shell go env GOARCH***REMOVED***
TARGET_ARCH=$(shell go env GOOS***REMOVED***_${GO_ARCH}

# Import path of the project:
import_path:=github.com/terraform-redhat/terraform-provider-rhcs

# Version of the project:
version=$(shell git describe --abbrev=0 | sed 's/^v//' | sed 's/-prerelease\.[0-9]*//'***REMOVED***
commit:=$(shell git rev-parse --short HEAD***REMOVED***

# Set the linker flags so that the version will be included in the binaries:
ldflags:=\
	-X $(import_path***REMOVED***/build.Version=$(version***REMOVED*** \
	-X $(import_path***REMOVED***/build.Commit=$(commit***REMOVED*** \
	$(NULL***REMOVED***

.PHONY: build
build:
	go build -ldflags="$(ldflags***REMOVED***" -o ${BINARY}

.PHONY: install
install: clean build
	platform=$$(terraform version -json | jq -r .platform***REMOVED***; \
	extension=""; \
	if [[ "$${platform}" =~ ^windows_.*$$ ]]; then \
		extension=".exe"; \
	fi; \
	if [ -z "${version}" ]; then \
    	version="0.0.2"; \
    fi; \
    dir="$(RHCS_LOCAL_DIR***REMOVED***/$${version}/$(TARGET_ARCH***REMOVED***"; \
	file="terraform-provider-rhcs$${extension}"; \
	mkdir -p "$${dir}"; \
	mv ${BINARY} "$${dir}/$${file}"

.PHONY: subsystem-test
subsystem-test: install
	ginkgo run \
		--succinct \
		-ldflags="$(ldflags***REMOVED***" \
		-r subsystem

.PHONY: unit-test
unit-test:
	ginkgo run \
		--succinct \
		-ldflags="$(ldflags***REMOVED***" \
		-r provider internal/...

.PHONY: test tests
test tests: unit-test subsystem-test

.PHONY: fmt_go
fmt_go:
	gofmt -s -l -w $$(find . -name '*.go'***REMOVED***

.PHONY: fmt_tf
fmt_tf:
	terraform fmt -recursive examples

.PHONY: fmt
fmt: fmt_go fmt_tf

.PHONY: clean
clean:
	rm -rf "$(RHCS_LOCAL_DIR***REMOVED***"

generate:
	go generate ./...

.PHONY: tools
tools:
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.1.1
	go install github.com/golang/mock/mockgen@v1.6.0

.PHONY: e2e_sanity_test
e2e_sanity_test: tools install
	ginkgo run $(ginkgo_flags***REMOVED*** \
		--timeout 5h \
		-r \
		--focus-file ci/e2e/.* \
		-- \
		--token-url=$(test_token_url***REMOVED*** \
		--gateway-url=$(test_gateway_url***REMOVED*** \
		--offline-token=$(test_token***REMOVED*** \
		--openshift-version=$(openshift_version***REMOVED*** \
		$(NULL***REMOVED***

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
        --label-filter $(LabelFilter***REMOVED***\
        --timeout 5h \
        -r \
        --focus-file tests/e2e/.* \
		$(NULL***REMOVED***