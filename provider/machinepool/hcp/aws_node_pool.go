package hcp

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

const MaxAdditionalSecurityGroupHcp = 10

type AWSNodePool struct {
	InstanceType                  types.String `tfsdk:"instance_type"`
	InstanceProfile               types.String `tfsdk:"instance_profile"`
	Tags                          types.Map    `tfsdk:"tags"`
	AdditionalSecurityGroupIds    types.List   `tfsdk:"additional_security_group_ids"`
	Ec2MetadataHttpTokens         types.String `tfsdk:"ec2_metadata_http_tokens"`
	DiskSize                      types.Int64  `tfsdk:"disk_size"`
	CapacityReservationId         types.String `tfsdk:"capacity_reservation_id"`
	CapacityReservationPreference types.String `tfsdk:"capacity_reservation_preference"`
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
			Description: "Apply user defined tags to all machine pool resources created in AWS." +
				common.ValueCannotBeChangedStringDescription,
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
		"ec2_metadata_http_tokens": schema.StringAttribute{
			Description: "This value determines which EC2 Instance Metadata Service mode to use for EC2 instances in the nodes." +
				"This can be set as `optional` (IMDS v1 or v2) or `required` (IMDSv2 only). This feature is available from " + common.ValueCannotBeChangedStringDescription,
			Optional: true,
			Computed: true,
			Validators: []validator.String{attrvalidators.EnumValueValidator([]string{string(cmv1.Ec2MetadataHttpTokensOptional),
				string(cmv1.Ec2MetadataHttpTokensRequired)})},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"disk_size": schema.Int64Attribute{
			Description: "Root disk size, in GiB. " + common.ValueCannotBeChangedStringDescription,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"capacity_reservation_id": schema.StringAttribute{
			Description: "The ID of the AWS Capacity Reservation to use for the node pool. " + common.ValueCannotBeChangedStringDescription,
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"capacity_reservation_preference": schema.StringAttribute{
			Description: "The preference for using AWS Capacity Reservations. Valid values are 'none', 'open', or" +
				" 'capacity-reservations-only'. The preference controls how the node pool utilizes available " +
				"capacity reservations. " + common.ValueCannotBeChangedStringDescription,
			Optional: true,
			Computed: true,
			Validators: []validator.String{attrvalidators.EnumValueValidator([]string{
				"none",
				"open",
				"capacity-reservations-only",
			})},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
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
		"ec2_metadata_http_tokens": schema.StringAttribute{
			Description: "This value determines which EC2 Instance Metadata Service mode to use for EC2 instances in the nodes." +
				"This can be set as `optional` (IMDS v1 or v2) or `required` (IMDSv2 only). This feature is available from " + common.ValueCannotBeChangedStringDescription,
			Optional: true,
			Computed: true,
		},
		"disk_size": schema.Int64Attribute{
			Description: "The root disk size, in GiB.",
			Optional:    true,
			Computed:    true,
		},
		"capacity_reservation_id": schema.StringAttribute{
			Description: "The ID of the AWS Capacity Reservation used for the node pool.",
			Optional:    true,
			Computed:    true,
			Default:     nil,
		},
		"capacity_reservation_preference": schema.StringAttribute{
			Description: "The preference for using AWS Capacity Reservations. Valid values are 'none', 'open', or 'capacity-reservations-only'.",
			Optional:    true,
			Computed:    true,
			Default:     nil,
		},
	}
}
