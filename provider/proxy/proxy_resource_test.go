// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"  // nolint
	. "github.com/onsi/ginkgo/v2/dsl/table" // nolint
	. "github.com/onsi/gomega"              // nolint
)

var _ = Describe("Proxy schema Sensitive", func() {
	DescribeTable("resource attributes",
		func(name string, sensitive bool) {
			attrs := ProxyResource()
			Expect(attrs).To(HaveKey(name))
			Expect(attrs[name].IsSensitive()).To(Equal(sensitive))
		},
		Entry("http_proxy", "http_proxy", true),
		Entry("https_proxy", "https_proxy", true),
		Entry("additional_trust_bundle", "additional_trust_bundle", true),
		Entry("no_proxy", "no_proxy", false),
	)

	DescribeTable("data source attributes",
		func(name string, sensitive bool) {
			attrs := ProxyDatasource()
			Expect(attrs).To(HaveKey(name))
			Expect(attrs[name].IsSensitive()).To(Equal(sensitive))
		},
		Entry("http_proxy", "http_proxy", true),
		Entry("https_proxy", "https_proxy", true),
		Entry("additional_trust_bundle", "additional_trust_bundle", true),
		Entry("no_proxy", "no_proxy", false),
	)
})
