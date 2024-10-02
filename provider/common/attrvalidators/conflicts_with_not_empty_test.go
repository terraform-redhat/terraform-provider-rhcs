package attrvalidators

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

var _ = Describe("Conflict with not empty validator", func() {

	DescribeTable("should validate correctly",
		func(request validator.ListRequest,
			expectedErr bool) {
			response := validator.ListResponse{}
			ConflictsWithNotEmpty(path.MatchRelative().AtParent().AtName("other")).
				ValidateList(context.Background(), request, &response)
			if expectedErr {
				Expect(response.Diagnostics.HasError()).To(BeTrue())
			}
		},
		Entry("2 not empty lists -> error",
			validator.ListRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Config: buildConfig(
					[]attr.Value{types.StringValue("test")},
					[]attr.Value{types.StringValue("test")}),
				ConfigValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			true,
		),
		Entry("only 1 not empty list -> ok",
			validator.ListRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Config: buildConfig(
					[]attr.Value{types.StringValue("")},
					[]attr.Value{types.StringValue("test")}),
				ConfigValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			false,
		),
		Entry("only 1 not empty list, other null -> ok",
			validator.ListRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				Config: buildConfig(
					[]attr.Value{types.StringNull()},
					[]attr.Value{types.StringValue("test")}),
				ConfigValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			false,
		),
	)
})

func buildConfig(attrValue1 []attr.Value, attrValue2 []attr.Value) tfsdk.Config {
	val, _ := types.ListValueMust(types.StringType, attrValue1).ToTerraformValue(context.Background())
	val2, _ := types.ListValueMust(types.StringType, attrValue2).ToTerraformValue(context.Background())
	return tfsdk.Config{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"test": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"other": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.List{ElementType: tftypes.String},
					"other": tftypes.List{ElementType: tftypes.String},
				},
			},
			map[string]tftypes.Value{
				"test":  val,
				"other": val2,
			},
		),
	}
}
