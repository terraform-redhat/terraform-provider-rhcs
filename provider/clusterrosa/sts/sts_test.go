// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package sts

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockSTSExternalIDSource struct {
	externalID string
	ok         bool
}

func (m *mockSTSExternalIDSource) GetExternalID() (string, bool) {
	return m.externalID, m.ok
}

var _ = Describe("STS schema and helpers", func() {
	Context("ClassicStsResource", func() {
		It("marks trust_policy_external_id as optional and computed", func() {
			schema := ClassicStsResource()

			attr, exists := schema["trust_policy_external_id"]
			Expect(exists).To(BeTrue())
			Expect(attr.IsOptional()).To(BeTrue())
			Expect(attr.IsComputed()).To(BeTrue())
		})
	})

	Context("HcpStsResource", func() {
		It("marks trust_policy_external_id as optional and computed", func() {
			schema := HcpStsResource()

			attr, exists := schema["trust_policy_external_id"]
			Expect(exists).To(BeTrue())
			Expect(attr.IsOptional()).To(BeTrue())
			Expect(attr.IsComputed()).To(BeTrue())
		})
	})

	Context("ClassicStsDatasource", func() {
		It("marks trust_policy_external_id as computed only", func() {
			schema := ClassicStsDatasource()

			attr, exists := schema["trust_policy_external_id"]
			Expect(exists).To(BeTrue())
			Expect(attr.IsComputed()).To(BeTrue())
			Expect(attr.IsOptional()).To(BeFalse())
		})
	})

	Context("HcpStsDatasource", func() {
		It("marks trust_policy_external_id as computed only", func() {
			schema := HcpStsDatasource()

			attr, exists := schema["trust_policy_external_id"]
			Expect(exists).To(BeTrue())
			Expect(attr.IsComputed()).To(BeTrue())
			Expect(attr.IsOptional()).To(BeFalse())
		})
	})

	Context("trustPolicyExternalIDValidators", func() {
		It("returns a non-empty list of validators", func() {
			validators := trustPolicyExternalIDValidators()
			Expect(validators).NotTo(BeEmpty())
		})

		It("accepts a valid external ID", func() {
			validators := trustPolicyExternalIDValidators()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue("valid-external-id-123"),
			}
			resp := &validator.StringResponse{}
			validators[0].ValidateString(context.Background(), req, resp)
			Expect(resp.Diagnostics.HasError()).To(BeFalse())
		})

		It("rejects an invalid external ID", func() {
			validators := trustPolicyExternalIDValidators()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue("x"),
			}
			resp := &validator.StringResponse{}
			validators[0].ValidateString(context.Background(), req, resp)
			Expect(resp.Diagnostics.HasError()).To(BeTrue())
		})

		It("skips validation for null values", func() {
			validators := trustPolicyExternalIDValidators()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringNull(),
			}
			resp := &validator.StringResponse{}
			validators[0].ValidateString(context.Background(), req, resp)
			Expect(resp.Diagnostics.HasError()).To(BeFalse())
		})

		It("skips validation for unknown values", func() {
			validators := trustPolicyExternalIDValidators()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringUnknown(),
			}
			resp := &validator.StringResponse{}
			validators[0].ValidateString(context.Background(), req, resp)
			Expect(resp.Diagnostics.HasError()).To(BeFalse())
		})
	})

	Context("ValidateTrustPolicyExternalIDFromConfig", func() {
		const (
			installerARN = "arn:aws:iam::123456789012:role/my-installer"
			supportARN   = "arn:aws:iam::123456789012:role/my-support"
			externalID   = "valid-external-id-123"
		)

		AfterEach(func() {
			TrustPolicyValidator = validateTrustPolicyExternalIDWithAWS
		})

		It("passes value when trust_policy_external_id is set", func() {
			TrustPolicyValidator = func(
				_ context.Context, entered, _, _, _ string,
			) error {
				Expect(entered).To(Equal(externalID))
				return nil
			}

			err := ValidateTrustPolicyExternalIDFromConfig(
				context.Background(),
				types.StringValue(externalID),
				types.StringValue(installerARN),
				types.StringValue(supportARN),
				"us-east-1",
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("passes empty when trust_policy_external_id is null", func() {
			TrustPolicyValidator = func(
				_ context.Context, entered, _, _, _ string,
			) error {
				Expect(entered).To(BeEmpty())
				return nil
			}

			err := ValidateTrustPolicyExternalIDFromConfig(
				context.Background(),
				types.StringNull(),
				types.StringValue(installerARN),
				types.StringValue(supportARN),
				"us-east-1",
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("passes empty when trust_policy_external_id is unknown", func() {
			TrustPolicyValidator = func(
				_ context.Context, entered, _, _, _ string,
			) error {
				Expect(entered).To(BeEmpty())
				return nil
			}

			err := ValidateTrustPolicyExternalIDFromConfig(
				context.Background(),
				types.StringUnknown(),
				types.StringValue(installerARN),
				types.StringValue(supportARN),
				"us-east-1",
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns validation errors", func() {
			err := ValidateTrustPolicyExternalIDFromConfig(
				context.Background(),
				types.StringValue("x"),
				types.StringValue(installerARN),
				types.StringValue(supportARN),
				"us-east-1",
			)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("PopulateTrustPolicyExternalIDFromSTS", func() {
		It("sets null when stsState is nil", func() {
			target := types.StringValue("prior")

			PopulateTrustPolicyExternalIDFromSTS(nil, &target)

			Expect(target.IsNull()).To(BeTrue())
		})

		It("sets value when external ID is present", func() {
			target := types.StringNull()
			source := &mockSTSExternalIDSource{externalID: "discovered-id", ok: true}

			PopulateTrustPolicyExternalIDFromSTS(source, &target)

			Expect(target.ValueString()).To(Equal("discovered-id"))
		})

		It("sets null when external ID is missing or empty", func() {
			target := types.StringValue("prior")
			source := &mockSTSExternalIDSource{externalID: "", ok: false}

			PopulateTrustPolicyExternalIDFromSTS(source, &target)

			Expect(target.IsNull()).To(BeTrue())
		})
	})
})
