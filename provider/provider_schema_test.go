// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provider schema Sensitive", func() {
	It("marks trusted_cas as sensitive", func() {
		response := &tfprovider.SchemaResponse{}
		New().Schema(context.Background(), tfprovider.SchemaRequest{}, response)

		Expect(response.Schema.Attributes).To(HaveKey("trusted_cas"))
		Expect(response.Schema.Attributes["trusted_cas"].IsSensitive()).To(BeTrue())
	})
})
