// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
)

var _ = Describe("AddDeprecationWarning", func() {
	It("adds no warning when the header is nil", func() {
		diags := diag.Diagnostics{}
		AddDeprecationWarning(&diags, nil)
		Expect(diags.HasError()).To(BeFalse())
		Expect(diags).To(BeEmpty())
	})

	It("adds no warning when no deprecation headers are set", func() {
		diags := diag.Diagnostics{}
		AddDeprecationWarning(&diags, http.Header{})
		Expect(diags).To(BeEmpty())
	})

	It("adds a warning for an endpoint deprecation header", func() {
		diags := diag.Diagnostics{}
		header := http.Header{}
		header.Set("Deprecation", "2027-01-01T00:00:00Z")
		header.Set("X-OCM-Deprecation-Message", "To continue with OpenShift v5, create a new ROSA HCP cluster")

		AddDeprecationWarning(&diags, header)

		Expect(diags).To(HaveLen(1))
		warning := diags[0]
		Expect(warning.Severity()).To(Equal(diag.SeverityWarning))
		Expect(warning.Summary()).To(Equal("OCM API deprecation notice"))
		Expect(warning.Detail()).To(ContainSubstring("2027-01-01T00:00:00Z"))
		Expect(warning.Detail()).To(ContainSubstring("To continue with OpenShift v5, create a new ROSA HCP cluster"))
	})

	It("adds a warning for a field deprecation header", func() {
		diags := diag.Diagnostics{}
		header := http.Header{}
		header.Set("X-OCM-Field-Deprecation", `{"ocp_v4_eol":"To continue with OpenShift v5, create a new ROSA HCP cluster"}`)

		AddDeprecationWarning(&diags, header)

		Expect(diags).To(HaveLen(1))
		Expect(diags[0].Detail()).To(ContainSubstring("ocp_v4_eol"))
	})
})
