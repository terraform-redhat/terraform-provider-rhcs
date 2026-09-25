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

	It("adds a user-focused warning with a nicely formatted date for an endpoint deprecation header", func() {
		diags := diag.Diagnostics{}
		header := http.Header{}
		header.Set("Deprecation", "2027-01-01T00:00:00Z")
		header.Set("X-OCM-Deprecation-Message", "To continue with OpenShift v5, create a new ROSA HCP cluster")

		AddDeprecationWarning(&diags, header)

		Expect(diags).To(HaveLen(1))
		warning := diags[0]
		Expect(warning.Severity()).To(Equal(diag.SeverityWarning))
		Expect(warning.Summary()).To(Equal("OCM API deprecation notice"))
		Expect(warning.Detail()).To(Equal(
			"Deprecation warning: To continue with OpenShift v5, create a new ROSA HCP cluster. " +
				"Deprecation date: January 1, 2027.",
		))
	})

	It("falls back to the raw date string when it can't be parsed", func() {
		diags := diag.Diagnostics{}
		header := http.Header{}
		header.Set("Deprecation", "not-a-real-date")
		header.Set("X-OCM-Deprecation-Message", "This will be removed soon")

		AddDeprecationWarning(&diags, header)

		Expect(diags).To(HaveLen(1))
		Expect(diags[0].Detail()).To(ContainSubstring("Deprecation date: not-a-real-date."))
	})

	It("adds a clean, user-focused warning for a field deprecation header, without exposing the raw field name", func() {
		diags := diag.Diagnostics{}
		header := http.Header{}
		header.Set("X-OCM-Field-Deprecation", `{"ocp_v4_eol":"To continue with OpenShift v5, create a new ROSA HCP cluster"}`)

		AddDeprecationWarning(&diags, header)

		Expect(diags).To(HaveLen(1))
		Expect(diags[0].Detail()).To(Equal(
			"Deprecation warning: To continue with OpenShift v5, create a new ROSA HCP cluster.",
		))
		Expect(diags[0].Detail()).NotTo(ContainSubstring("ocp_v4_eol"))
	})

	It("de-duplicates identical messages when both endpoint and field deprecation headers are present", func() {
		diags := diag.Diagnostics{}
		header := http.Header{}
		header.Set("Deprecation", "2027-01-01T00:00:00Z")
		header.Set("X-OCM-Deprecation-Message", "To continue with OpenShift v5, create a new ROSA HCP cluster")
		header.Set("X-OCM-Field-Deprecation", `{"ocp_v4_eol":"To continue with OpenShift v5, create a new ROSA HCP cluster"}`)

		AddDeprecationWarning(&diags, header)

		Expect(diags).To(HaveLen(1))
		detail := diags[0].Detail()
		Expect(detail).To(Equal(
			"Deprecation warning: To continue with OpenShift v5, create a new ROSA HCP cluster. " +
				"Deprecation date: January 1, 2027.",
		))
		Expect(detail).NotTo(ContainSubstring("ocp_v4_eol"))
	})

	It("includes distinct messages from both headers when they differ", func() {
		diags := diag.Diagnostics{}
		header := http.Header{}
		header.Set("X-OCM-Deprecation-Message", "This endpoint is going away")
		header.Set("X-OCM-Field-Deprecation", `{"some_field":"Use another_field instead"}`)

		AddDeprecationWarning(&diags, header)

		Expect(diags).To(HaveLen(1))
		detail := diags[0].Detail()
		Expect(detail).To(ContainSubstring("This endpoint is going away."))
		Expect(detail).To(ContainSubstring("Use another_field instead."))
	})
})
