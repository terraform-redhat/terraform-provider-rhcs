// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Availability zone validator", func() {
	DescribeTable("should validate correctly",
		func(request validator.StringRequest, expectedErr bool) {
			response := validator.StringResponse{}
			AvailabilityZoneValidator.ValidateString(context.Background(), request, &response)
			Expect(response.Diagnostics.HasError()).To(Equal(expectedErr))
		},
		Entry("AZ contains cloud_region -> ok",
			validator.StringRequest{
				Path:           path.Root("availability_zone"),
				PathExpression: path.MatchRoot("availability_zone"),
				Config:         buildCloudRegionConfig("us-east-1"),
				ConfigValue:    types.StringValue("us-east-1a"),
			},
			false,
		),
		Entry("AZ does not contain cloud_region -> error",
			validator.StringRequest{
				Path:           path.Root("availability_zone"),
				PathExpression: path.MatchRoot("availability_zone"),
				Config:         buildCloudRegionConfig("us-east-1"),
				ConfigValue:    types.StringValue("us-west-2a"),
			},
			true,
		),
	)
})

var _ = Describe("Private hosted zone validator", func() {
	privateHZAttrTypes := map[string]attr.Type{
		"id":       types.StringType,
		"role_arn": types.StringType,
	}

	DescribeTable("should validate correctly",
		func(request validator.ObjectRequest, expectedErr bool) {
			response := validator.ObjectResponse{}
			PrivateHZValidator.ValidateObject(context.Background(), request, &response)
			Expect(response.Diagnostics.HasError()).To(Equal(expectedErr))
		},
		Entry("valid id and role_arn -> ok",
			validator.ObjectRequest{
				Path:           path.Root("private_hosted_zone"),
				PathExpression: path.MatchRoot("private_hosted_zone"),
				ConfigValue: types.ObjectValueMust(privateHZAttrTypes, map[string]attr.Value{
					"id":       types.StringValue("Z05646003S02O1ENCDCSN"),
					"role_arn": types.StringValue("arn:aws:iam::123456789012:role/route53"),
				}),
			},
			false,
		),
		Entry("empty id -> error",
			validator.ObjectRequest{
				Path:           path.Root("private_hosted_zone"),
				PathExpression: path.MatchRoot("private_hosted_zone"),
				ConfigValue: types.ObjectValueMust(privateHZAttrTypes, map[string]attr.Value{
					"id":       types.StringValue(""),
					"role_arn": types.StringValue("arn:aws:iam::123456789012:role/route53"),
				}),
			},
			true,
		),
		Entry("empty role_arn -> error",
			validator.ObjectRequest{
				Path:           path.Root("private_hosted_zone"),
				PathExpression: path.MatchRoot("private_hosted_zone"),
				ConfigValue: types.ObjectValueMust(privateHZAttrTypes, map[string]attr.Value{
					"id":       types.StringValue("Z05646003S02O1ENCDCSN"),
					"role_arn": types.StringValue(""),
				}),
			},
			true,
		),
		Entry("null object -> ok",
			validator.ObjectRequest{
				Path:           path.Root("private_hosted_zone"),
				PathExpression: path.MatchRoot("private_hosted_zone"),
				ConfigValue:    types.ObjectNull(privateHZAttrTypes),
			},
			false,
		),
		Entry("unknown object -> ok",
			validator.ObjectRequest{
				Path:           path.Root("private_hosted_zone"),
				PathExpression: path.MatchRoot("private_hosted_zone"),
				ConfigValue:    types.ObjectUnknown(privateHZAttrTypes),
			},
			false,
		),
	)
})

var _ = Describe("Properties validator", func() {
	DescribeTable("should validate correctly",
		func(request validator.MapRequest, expectedErr bool) {
			response := validator.MapResponse{}
			PropertiesValidator.ValidateMap(context.Background(), request, &response)
			Expect(response.Diagnostics.HasError()).To(Equal(expectedErr))
		},
		Entry("custom property key -> ok",
			validator.MapRequest{
				Path:           path.Root("properties"),
				PathExpression: path.MatchRoot("properties"),
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{
					"custom_key": types.StringValue("value"),
				}),
			},
			false,
		),
		Entry("reserved OCM property key -> error",
			validator.MapRequest{
				Path:           path.Root("properties"),
				PathExpression: path.MatchRoot("properties"),
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{
					PropertyRosaTfVersion: types.StringValue("override"),
				}),
			},
			true,
		),
		Entry("null map -> ok",
			validator.MapRequest{
				Path:           path.Root("properties"),
				PathExpression: path.MatchRoot("properties"),
				ConfigValue:    types.MapNull(types.StringType),
			},
			false,
		),
		Entry("unknown map -> ok",
			validator.MapRequest{
				Path:           path.Root("properties"),
				PathExpression: path.MatchRoot("properties"),
				ConfigValue:    types.MapUnknown(types.StringType),
			},
			false,
		),
	)
})

func buildCloudRegionConfig(region string) tfsdk.Config {
	regionVal, _ := types.StringValue(region).ToTerraformValue(context.Background())
	return tfsdk.Config{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"cloud_region": schema.StringAttribute{
					Optional: true,
				},
			},
		},
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"cloud_region": tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"cloud_region": regionVal,
			},
		),
	}
}
