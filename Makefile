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

# Version of the provider:
version:=0.1.0

.PHONY: build
build:
	platform=$$(terraform version -json | jq -r .platform***REMOVED***; \
	ext=""; \
	if [[ "$${platform}" =~ ^windows_.*$$ ]]; then \
	  ext=".exe"; \
	fi; \
	dir=".terraform.d/plugins/localhost/redhat/ocm/$(version***REMOVED***/$${platform}"; \
	file="terraform-provider-ocm$${ext}"; \
	mkdir -p "$${dir}"; \
	go build -o "$${dir}/$${file}"

.PHONY: test tests
test tests: build
	ginkgo -r tests

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
	rm -rf .terraform.d
