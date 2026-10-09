// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

var _ = Describe("ValidateAdminCredentialsUnchanged", func() {
	const (
		statePassword = "state-password123456789$"
		planPassword  = "plan-password123456789$"
	)

	It("adds no error when admin_credentials are unchanged", func() {
		diags := diag.Diagnostics{}
		ValidateAdminCredentialsUnchanged(
			FlattenAdminCredentials("test-username", statePassword),
			FlattenAdminCredentials("test-username", statePassword),
			&diags,
		)
		Expect(diags.HasError()).To(BeFalse())
	})

	It("names admin_credentials without exposing the passwords when they are changed", func() {
		diags := diag.Diagnostics{}
		ValidateAdminCredentialsUnchanged(
			FlattenAdminCredentials("test-username", statePassword),
			FlattenAdminCredentials("test-username", planPassword),
			&diags,
		)
		Expect(diags.ErrorsCount()).To(Equal(1))
		Expect(diags.Errors()[0].Summary()).To(Equal(common.AssertionErrorSummaryMessage))
		Expect(diags.Errors()[0].Detail()).To(Equal("Attribute admin_credentials, cannot be changed"))
		Expect(diags.Errors()[0].Detail()).NotTo(ContainSubstring(statePassword))
		Expect(diags.Errors()[0].Detail()).NotTo(ContainSubstring(planPassword))
	})
})
