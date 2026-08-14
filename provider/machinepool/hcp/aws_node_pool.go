// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	commonplanmodifiers "github.com/terraform-redhat/terraform-provider-rhcs/provider/common/planmodifiers"
)

const MaxAdditionalSecurityGroupHcp = 10

// MaxNodeDrainGracePeriodMinutes is the maximum node drain grace period in whole minutes (ROSA CLI limit).
const MaxNodeDrainGracePeriodMinutes = 10080

type AWSNodePool struct {
	InstanceType                  types.String  `tfsdk:"instance_type"`
	InstanceProfile               types.String  `tfsdk:"instance_profile"`
	Tags                          types.Map     `tfsdk:"tags"`
	AdditionalSecurityGroupIds    types.List    `tfsdk:"additional_security_group_ids"`
	Ec2MetadataHttpTokens         types.String  `tfsdk:"ec2_metadata_http_tokens"`
	DiskSize                      types.Int64   `tfsdk:"disk_size"`
	CapacityReservationId         types.String  `tfsdk:"capacity_reservation_id"`
	CapacityReservationPreference types.String  `tfsdk:"capacity_reservation_preference"`
	UseSpotInstances              types.Bool    `tfsdk:"use_spot_instances"`
	MaxSpotPrice                  types.Float64 `tfsdk:"max_spot_price"`
	ImageType                     types.String  `tfsdk:"image_type"`
	NodeDrainGracePeriod          types.Int64   `tfsdk:"node_drain_grace_period"`
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
				commonplanmodifiers.ImmutableString(),
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
		// use_spot_instances: IgnoreFalse normalizes false→null so false is treated
		// as "not set" and never stored in state, avoiding false→null drift on refresh.
		"use_spot_instances": schema.BoolAttribute{
			Description: "Use Amazon EC2 Spot Instances. When enabled, max_spot_price can be set " +
				"to control the maximum hourly price. " +
				"Cannot be used with capacity_reservation_id or capacity_reservation_preference. " +
				common.ValueCannotBeChangedStringDescription,
			Optional: true,
			PlanModifiers: []planmodifier.Bool{
				commonplanmodifiers.IgnoreFalse(),
				commonplanmodifiers.ImmutableBool(),
			},
		},
		"max_spot_price": schema.Float64Attribute{
			Description: "Maximum hourly price for Spot Instances in USD. Requires use_spot_instances to be true. " +
				"Must be a positive value (> 0). If not specified, the on-demand price is used as the maximum. " +
				common.ValueCannotBeChangedStringDescription,
			Optional: true,
			Validators: []validator.Float64{
				attrvalidators.PositiveFloat64(),
			},
			PlanModifiers: []planmodifier.Float64{
				commonplanmodifiers.ImmutableFloat64(),
			},
		},
		"image_type": schema.StringAttribute{
			Description: "The image type to use for the node pool. Valid values are 'Default' or 'Windows'. " +
				common.ValueCannotBeChangedStringDescription,
			Optional: true,
			Computed: true,
			Validators: []validator.String{attrvalidators.EnumValueValidator([]string{
				string(cmv1.ImageTypeDefault),
				string(cmv1.ImageTypeWindows),
			})},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node_drain_grace_period": schema.Int64Attribute{
			Description: "Grace period in whole minutes before nodes are forcibly drained during upgrade or replacement. " +
				"This value is stored on the NodePool in OpenShift Cluster Manager but is grouped under `aws_node_pool` for " +
				"consistency with other pool settings. Valid range is 0–10080 minutes (one week).",
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
				int64validator.AtMost(MaxNodeDrainGracePeriodMinutes),
			},
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
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
		"use_spot_instances": schema.BoolAttribute{
			Description: "Use Amazon EC2 Spot Instances.",
			Computed:    true,
		},
		"max_spot_price": schema.Float64Attribute{
			Description: "Max Spot price.",
			Computed:    true,
		},
		"image_type": schema.StringAttribute{
			Description: "The image type used for the node pool. Valid values are 'Default' or 'Windows'.",
			Optional:    true,
			Computed:    true,
		},
		"node_drain_grace_period": schema.Int64Attribute{
			Description: "Grace period in whole minutes before nodes are forcibly drained during upgrade or replacement.",
			Optional:    true,
			Computed:    true,
		},
	}
}
