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

ifeq ($(OS),Windows_NT)
	BINARY=terraform-provider-ocm.exe
else
	BINARY=terraform-provider-ocm
endif

BINARY=terraform-provider-ocm
PLATFORM=$(shell go env GOOS)_$(shell go env GOARCH)
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
	mkdir -p ~/.terraform.d/plugins/localhost/openshift-online/ocm/${version}/${PLATFORM}
	mv ${BINARY} ~/.terraform.d/plugins/localhost/openshift-online/ocm/${version}/${PLATFORM}

.PHONY: clean
clean:
	$(shell go clean)
	@echo "==> Removing ~/.terraform.d/plugins/localhost/openshift-online/ocm/${version}/${PLATFORM} directory"
	@rm -rf ~/.terraform.d/plugins/localhost/openshift-online/ocm/${version}/${PLATFORM}


.PHONY: test tests
test tests: build
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
