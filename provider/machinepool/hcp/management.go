// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// Management holds the Terraform state for the management nested block.
type Management struct {
	Type           types.String `tfsdk:"type"`
	MaxSurge       types.String `tfsdk:"max_surge"`
	MaxUnavailable types.String `tfsdk:"max_unavailable"`
}

const (
	attrNameType           = "type"
	attrNameMaxSurge       = "max_surge"
	attrNameMaxUnavailable = "max_unavailable"
)

var intOrPercentageRE = regexp.MustCompile(`^[0-9]+%?$`)

var managementAttrTypes = map[string]attr.Type{
	attrNameType:           types.StringType,
	attrNameMaxSurge:       types.StringType,
	attrNameMaxUnavailable: types.StringType,
}

// ManagementResource returns the schema attributes for the management block on the resource.
func ManagementResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		attrNameType: schema.StringAttribute{
			Description: "Type of strategy for handling upgrades.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf("Replace"),
			},
		},
		attrNameMaxSurge: schema.StringAttribute{
			Description: "Maximum number of nodes that can be scheduled above the " +
				"desired number of nodes during the upgrade. " +
				"Can be an integer or a percentage (e.g., \"1\" or \"25%\").",
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(intOrPercentageRE,
					"must be a non-negative integer or a non-negative "+
						"integer followed by '%' (e.g., \"1\", \"25%\")"),
			},
		},
		attrNameMaxUnavailable: schema.StringAttribute{
			Description: "Maximum number of nodes that can be unavailable " +
				"during the upgrade. " +
				"Can be an integer or a percentage (e.g., \"0\" or \"10%\").",
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(intOrPercentageRE,
					"must be a non-negative integer or a non-negative "+
						"integer followed by '%' (e.g., \"0\", \"10%\")"),
			},
		},
	}
}

// ManagementDatasource returns the schema attributes for the management block on the data source.
func ManagementDatasource() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		attrNameType: dsschema.StringAttribute{
			Description: "Type of strategy for handling upgrades.",
			Computed:    true,
		},
		attrNameMaxSurge: dsschema.StringAttribute{
			Description: "Maximum number of nodes that can be scheduled " +
				"above the desired number of nodes during the upgrade.",
			Computed: true,
		},
		attrNameMaxUnavailable: dsschema.StringAttribute{
			Description: "Maximum number of nodes that can be unavailable " +
				"during the upgrade.",
			Computed: true,
		},
	}
}

// stringValueOrNull returns StringNull for empty strings, StringValue otherwise.
func stringValueOrNull(s string) basetypes.StringValue {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// flattenManagement converts API values into a Terraform types.Object for the management block.
func flattenManagement(upgradeType, maxSurge, maxUnavailable string) types.Object {
	attrs := map[string]attr.Value{
		attrNameType:           stringValueOrNull(upgradeType),
		attrNameMaxSurge:       stringValueOrNull(maxSurge),
		attrNameMaxUnavailable: stringValueOrNull(maxUnavailable),
	}
	return types.ObjectValueMust(managementAttrTypes, attrs)
}

// managementNull returns a typed-null object for the management block.
func managementNull() types.Object {
	return types.ObjectNull(managementAttrTypes)
}

// expandManagement converts a Terraform types.Object into a Management struct.
func expandManagement(
	ctx context.Context, object types.Object, diags *diag.Diagnostics,
) *Management {
	if object.IsNull() || object.IsUnknown() {
		return nil
	}
	var config Management
	diags.Append(object.As(ctx, &config, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return &config
}

// buildManagementUpgrade builds a NodePoolManagementUpgradeBuilder from the management block.
func buildManagementUpgrade(
	ctx context.Context, object types.Object, diags *diag.Diagnostics,
) *cmv1.NodePoolManagementUpgradeBuilder {
	config := expandManagement(ctx, object, diags)
	if config == nil {
		return nil
	}
	builder := cmv1.NewNodePoolManagementUpgrade()
	if !common.IsStringAttributeUnknownOrEmpty(config.Type) {
		builder.Type(config.Type.ValueString())
	}
	if !common.IsStringAttributeUnknownOrEmpty(config.MaxSurge) {
		builder.MaxSurge(config.MaxSurge.ValueString())
	}
	if !common.IsStringAttributeUnknownOrEmpty(config.MaxUnavailable) {
		builder.MaxUnavailable(config.MaxUnavailable.ValueString())
	}
	return builder
}
