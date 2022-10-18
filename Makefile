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

ifeq ($(shell go env GOOS),windows)
	BINARY=terraform-provider-ocm.exe
else
	BINARY=terraform-provider-ocm
endif

# Import path of the project:
import_path:=github.com/openshift-online/terraform-provider-ocm

# Version of the project:
version:=$(shell git describe --abbrev=0 | sed 's/^v//')
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
install: build
	platform=$$(terraform version -json | jq -r .platform); \
	extension=""; \
	if [[ "$${platform}" =~ ^windows_.*$$ ]]; then \
	  extension=".exe"; \
	fi; \
	dir=".terraform.d/plugins/localhost/openshift-online/ocm/$(version)/$${platform}"; \
	file="terraform-provider-ocm$${extension}"; \
	mkdir -p "$${dir}"; \
	mv ${BINARY} "$${dir}/$${file}"

.PHONY: test tests
test tests: install
	ginkgo run \
		--succinct \
		-ldflags="$(ldflags)" \
		-r

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
	rm -rf .terraform.d
