package hcp

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

const MaxAdditionalSecurityGroupHcp = 10

type AWSNodePool struct {
	InstanceType               types.String `tfsdk:"instance_type"`
	InstanceProfile            types.String `tfsdk:"instance_profile"`
	Tags                       types.Map    `tfsdk:"tags"`
	AdditionalSecurityGroupIds types.List   `tfsdk:"additional_security_group_ids"`
}

func AwsNodePoolResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"instance_type": schema.StringAttribute{
			Description: "Identifier of the machine type used by the nodes, " +
				"for example `m5.xlarge`. Use the `rhcs_machine_types` data " +
				"source to find the possible values. " + common.ValueCannotBeChangedStringDescription,
			Required: true,
		},
		"instance_profile": schema.StringAttribute{
			Description: "Instance profile attached to the replica",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"tags": schema.MapAttribute{
			Description: "Apply user defined tags to all machine pool resources created in AWS. " + common.ValueCannotBeChangedStringDescription,
			ElementType: types.StringType,
			Optional:    true,
		},
		"additional_security_group_ids": schema.ListAttribute{
			Description: "Additional security group ids. " + common.ValueCannotBeChangedStringDescription,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtMost(MaxAdditionalSecurityGroupHcp),
			},
			Optional: true,
		},
	}
}

func AwsNodePoolDatasource() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"instance_type": schema.StringAttribute{
			Description: "Identifier of the machine type used by the nodes, " +
				"for example `m5.xlarge`. Use the `rhcs_machine_types` data " +
				"source to find the possible values. " + common.ValueCannotBeChangedStringDescription,
			Computed: true,
		},
		"instance_profile": schema.StringAttribute{
			Description: "Instance profile attached to the replica",
			Computed:    true,
		},
		"tags": schema.MapAttribute{
			Description: "Apply user defined tags to all machine pool resources created in AWS. " + common.ValueCannotBeChangedStringDescription,
			ElementType: types.StringType,
			Optional:    true,
		},
		"additional_security_group_ids": schema.ListAttribute{
			Description: "Additional security group ids. " + common.ValueCannotBeChangedStringDescription,
			ElementType: types.StringType,
			Optional:    true,
		},
	}
}
