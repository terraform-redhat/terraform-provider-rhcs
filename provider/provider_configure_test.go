// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provider Configure", func() {
	It("rejects an unknown allow_tls12 value with a diagnostic error", func() {
		ctx := context.Background()
		p := New()

		// Obtain the provider schema so the raw tftypes.Value matches it.
		schemaResp := &tfprovider.SchemaResponse{}
		p.Schema(ctx, tfprovider.SchemaRequest{}, schemaResp)
		Expect(schemaResp.Diagnostics).To(BeEmpty())

		// Build a raw config where allow_tls12 is unknown and every other
		// attribute is null. This mirrors what Terraform sends when a
		// provider attribute depends on an unresolved value.
		rawValues := map[string]tftypes.Value{
			"url":           tftypes.NewValue(tftypes.String, nil),
			"token_url":     tftypes.NewValue(tftypes.String, nil),
			"token":         tftypes.NewValue(tftypes.String, nil),
			"refresh_token": tftypes.NewValue(tftypes.String, nil),
			"client_id":     tftypes.NewValue(tftypes.String, nil),
			"client_secret": tftypes.NewValue(tftypes.String, nil),
			"trusted_cas":   tftypes.NewValue(tftypes.String, nil),
			"insecure":      tftypes.NewValue(tftypes.Bool, nil),
			"allow_tls12":   tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
		}

		schemaType := schemaResp.Schema.Type().TerraformType(ctx)
		configReq := tfprovider.ConfigureRequest{
			Config: tfsdk.Config{
				Raw:    tftypes.NewValue(schemaType, rawValues),
				Schema: schemaResp.Schema,
			},
		}
		configResp := &tfprovider.ConfigureResponse{
			Diagnostics: diag.Diagnostics{},
		}

		p.Configure(ctx, configReq, configResp)

		Expect(configResp.Diagnostics.HasError()).To(BeTrue())
		Expect(configResp.Diagnostics.Errors()).To(HaveLen(1))
		Expect(configResp.Diagnostics.Errors()[0].Summary()).To(
			Equal("the value of 'allow_tls12' is not yet known"),
		)
		Expect(configResp.Diagnostics.Errors()[0].Detail()).To(
			ContainSubstring("must be known during provider configuration"),
		)

		// Confirm the early return: no connection should have been set.
		Expect(configResp.DataSourceData).To(BeNil())
		Expect(configResp.ResourceData).To(BeNil())
	})
})
