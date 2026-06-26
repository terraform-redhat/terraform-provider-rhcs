// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package sharedvpc

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
)

func TestSharedVpc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shared VPC Suite")
}

var sharedVpcAttrTypes = map[string]attr.Type{
	"ingress_private_hosted_zone_id":                types.StringType,
	"internal_communication_private_hosted_zone_id": types.StringType,
	"route53_role_arn":                              types.StringType,
	"vpce_role_arn":                                 types.StringType,
}

func validSharedVpcValues() map[string]attr.Value {
	return map[string]attr.Value{
		"ingress_private_hosted_zone_id":                types.StringValue("Z05646003S02O1ENCDCSN"),
		"internal_communication_private_hosted_zone_id": types.StringValue("Z05646003S02O1ENCDCSN"),
		"route53_role_arn":                              types.StringValue("arn:aws:iam::123456789012:role/route53"),
		"vpce_role_arn":                                 types.StringValue("arn:aws:iam::123456789012:role/vpce"),
	}
}

var _ = Describe("HCP shared VPC validator", func() {
	DescribeTable("should validate correctly",
		func(request validator.ObjectRequest, expectedErr bool) {
			response := validator.ObjectResponse{}
			HcpSharedVpcValidator.ValidateObject(context.Background(), request, &response)
			Expect(response.Diagnostics.HasError()).To(Equal(expectedErr))
		},
		Entry("all attributes set -> ok",
			validator.ObjectRequest{
				Path:           path.Root("shared_vpc"),
				PathExpression: path.MatchRoot("shared_vpc"),
				ConfigValue:    types.ObjectValueMust(sharedVpcAttrTypes, validSharedVpcValues()),
			},
			false,
		),
		Entry("empty ingress_private_hosted_zone_id -> error",
			validator.ObjectRequest{
				Path:           path.Root("shared_vpc"),
				PathExpression: path.MatchRoot("shared_vpc"),
				ConfigValue: func() types.Object {
					values := validSharedVpcValues()
					values["ingress_private_hosted_zone_id"] = types.StringValue("")
					return types.ObjectValueMust(sharedVpcAttrTypes, values)
				}(),
			},
			true,
		),
		Entry("empty internal_communication_private_hosted_zone_id -> error",
			validator.ObjectRequest{
				Path:           path.Root("shared_vpc"),
				PathExpression: path.MatchRoot("shared_vpc"),
				ConfigValue: func() types.Object {
					values := validSharedVpcValues()
					values["internal_communication_private_hosted_zone_id"] = types.StringValue("")
					return types.ObjectValueMust(sharedVpcAttrTypes, values)
				}(),
			},
			true,
		),
		Entry("empty route53_role_arn -> error",
			validator.ObjectRequest{
				Path:           path.Root("shared_vpc"),
				PathExpression: path.MatchRoot("shared_vpc"),
				ConfigValue: func() types.Object {
					values := validSharedVpcValues()
					values["route53_role_arn"] = types.StringValue("")
					return types.ObjectValueMust(sharedVpcAttrTypes, values)
				}(),
			},
			true,
		),
		Entry("empty vpce_role_arn -> error",
			validator.ObjectRequest{
				Path:           path.Root("shared_vpc"),
				PathExpression: path.MatchRoot("shared_vpc"),
				ConfigValue: func() types.Object {
					values := validSharedVpcValues()
					values["vpce_role_arn"] = types.StringValue("")
					return types.ObjectValueMust(sharedVpcAttrTypes, values)
				}(),
			},
			true,
		),
		Entry("null object -> ok",
			validator.ObjectRequest{
				Path:           path.Root("shared_vpc"),
				PathExpression: path.MatchRoot("shared_vpc"),
				ConfigValue:    types.ObjectNull(sharedVpcAttrTypes),
			},
			false,
		),
		Entry("unknown object -> ok",
			validator.ObjectRequest{
				Path:           path.Root("shared_vpc"),
				PathExpression: path.MatchRoot("shared_vpc"),
				ConfigValue:    types.ObjectUnknown(sharedVpcAttrTypes),
			},
			false,
		),
	)
})
