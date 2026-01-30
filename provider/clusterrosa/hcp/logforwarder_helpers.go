/*
Copyright (c) 2026 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hcp

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// LogForwarder represents a log forwarder configuration for building API requests
type LogForwarder struct {
	S3           types.Object `tfsdk:"s3"`
	CloudWatch   types.Object `tfsdk:"cloudwatch"`
	Applications types.List   `tfsdk:"applications"`
	Groups       types.List   `tfsdk:"groups"`
}

// S3Config represents S3 configuration for log forwarding
type S3Config struct {
	BucketName   types.String `tfsdk:"bucket_name"`
	BucketPrefix types.String `tfsdk:"bucket_prefix"`
}

// CloudWatchConfig represents CloudWatch configuration for log forwarding
type CloudWatchConfig struct {
	LogGroupName           types.String `tfsdk:"log_group_name"`
	LogDistributionRoleArn types.String `tfsdk:"log_distribution_role_arn"`
}

// LogForwarderGroup represents a log forwarder group
type LogForwarderGroup struct {
	ID      types.String `tfsdk:"id"`
	Version types.String `tfsdk:"version"`
}

// buildLogForwarders converts a Terraform types.List of log forwarders into an OCM SDK LogForwarderListBuilder.
// This is used during cluster creation (Day 1) to configure log forwarders as part of the control plane.
func buildLogForwarders(ctx context.Context, logForwardersList types.List) (*cmv1.LogForwarderListBuilder, error) {
	var logForwarders []LogForwarder
	diags := logForwardersList.ElementsAs(ctx, &logForwarders, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse log forwarders: %v", diags.Errors())
	}

	listBuilder := cmv1.NewLogForwarderList()
	builders := []*cmv1.LogForwarderBuilder{}

	for _, lfState := range logForwarders {
		builder := cmv1.NewLogForwarder()

		if !lfState.S3.IsNull() && !lfState.S3.IsUnknown() {
			var s3Config S3Config
			diags := lfState.S3.As(ctx, &s3Config, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, fmt.Errorf("failed to parse s3 config: %v", diags.Errors())
			}
			s3Builder := cmv1.NewLogForwarderS3Config().
				BucketName(s3Config.BucketName.ValueString())
			if !s3Config.BucketPrefix.IsNull() && !s3Config.BucketPrefix.IsUnknown() {
				s3Builder.BucketPrefix(s3Config.BucketPrefix.ValueString())
			}
			builder.S3(s3Builder)
		}

		if !lfState.CloudWatch.IsNull() && !lfState.CloudWatch.IsUnknown() {
			var cwConfig CloudWatchConfig
			diags := lfState.CloudWatch.As(ctx, &cwConfig, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, fmt.Errorf("failed to parse cloudwatch config: %v", diags.Errors())
			}
			cwBuilder := cmv1.NewLogForwarderCloudWatchConfig().
				LogGroupName(cwConfig.LogGroupName.ValueString()).
				LogDistributionRoleArn(cwConfig.LogDistributionRoleArn.ValueString())
			builder.Cloudwatch(cwBuilder)
		}

		if !lfState.Applications.IsNull() && !lfState.Applications.IsUnknown() && len(lfState.Applications.Elements()) > 0 {
			var applications []string
			diags := lfState.Applications.ElementsAs(ctx, &applications, false)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to parse applications: %v", diags.Errors())
			}
			if len(applications) > 0 {
				builder.Applications(applications...)
			}
		}

		if !lfState.Groups.IsNull() && !lfState.Groups.IsUnknown() && len(lfState.Groups.Elements()) > 0 {
			var groups []LogForwarderGroup
			diags := lfState.Groups.ElementsAs(ctx, &groups, false)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to parse groups: %v", diags.Errors())
			}
			groupBuilders := make([]*cmv1.LogForwarderGroupBuilder, 0, len(groups))
			for _, group := range groups {
				groupBuilder := cmv1.NewLogForwarderGroup().
					ID(group.ID.ValueString()).
					Version(group.Version.ValueString())
				groupBuilders = append(groupBuilders, groupBuilder)
			}
			if len(groupBuilders) > 0 {
				builder.Groups(groupBuilders...)
			}
		}

		builders = append(builders, builder)
	}

	listBuilder.Items(builders...)
	return listBuilder, nil
}

// flattenLogForwarders converts an OCM SDK LogForwarderList into a Terraform types.List.
// This is used during cluster reads to populate the state with log forwarder configuration
// from the API response.
func flattenLogForwarders(ctx context.Context, logForwardersList *cmv1.LogForwarderList) (types.List, error) {
	lfObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"s3": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bucket_name":   types.StringType,
					"bucket_prefix": types.StringType,
				},
			},
			"cloudwatch": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"log_group_name":            types.StringType,
					"log_distribution_role_arn": types.StringType,
				},
			},
			"applications": types.ListType{ElemType: types.StringType},
			"groups": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":      types.StringType,
						"version": types.StringType,
					},
				},
			},
		},
	}

	if logForwardersList == nil || logForwardersList.Len() == 0 {
		return types.ListNull(lfObjType), nil
	}

	var listItems []*cmv1.LogForwarder
	logForwardersList.Each(func(lf *cmv1.LogForwarder) bool {
		listItems = append(listItems, lf)
		return true
	})

	logForwarders := make([]attr.Value, 0, len(listItems))

	for _, lf := range listItems {
		lfObj := map[string]attr.Value{
			"s3": types.ObjectNull(map[string]attr.Type{
				"bucket_name":   types.StringType,
				"bucket_prefix": types.StringType,
			}),
			"cloudwatch": types.ObjectNull(map[string]attr.Type{
				"log_group_name":            types.StringType,
				"log_distribution_role_arn": types.StringType,
			}),
			"applications": types.ListNull(types.StringType),
			"groups": types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":      types.StringType,
					"version": types.StringType,
				},
			}),
		}

		if s3, ok := lf.GetS3(); ok {
			s3Attrs := map[string]attr.Value{
				"bucket_name":   types.StringValue(s3.BucketName()),
				"bucket_prefix": types.StringValue(s3.BucketPrefix()),
			}
			s3Obj, diags := types.ObjectValue(map[string]attr.Type{
				"bucket_name":   types.StringType,
				"bucket_prefix": types.StringType,
			}, s3Attrs)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), fmt.Errorf("failed to create s3 object: %v", diags.Errors())
			}
			lfObj["s3"] = s3Obj
		}

		if cw, ok := lf.GetCloudwatch(); ok {
			cwAttrs := map[string]attr.Value{
				"log_group_name":            types.StringValue(cw.LogGroupName()),
				"log_distribution_role_arn": types.StringValue(cw.LogDistributionRoleArn()),
			}
			cwObj, diags := types.ObjectValue(map[string]attr.Type{
				"log_group_name":            types.StringType,
				"log_distribution_role_arn": types.StringType,
			}, cwAttrs)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), fmt.Errorf("failed to create cloudwatch object: %v", diags.Errors())
			}
			lfObj["cloudwatch"] = cwObj
		}

		if apps, ok := lf.GetApplications(); ok && len(apps) > 0 {
			appsElements := make([]attr.Value, 0, len(apps))
			for _, app := range apps {
				appsElements = append(appsElements, types.StringValue(app))
			}
			appsList, diags := types.ListValue(types.StringType, appsElements)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), fmt.Errorf("failed to create applications list: %v", diags.Errors())
			}
			lfObj["applications"] = appsList
		}

		if groups, ok := lf.GetGroups(); ok && len(groups) > 0 {
			groupsElements := make([]attr.Value, 0, len(groups))
			for _, grp := range groups {
				grpAttrs := map[string]attr.Value{
					"id":      types.StringValue(grp.ID()),
					"version": types.StringValue(grp.Version()),
				}
				grpObj, diags := types.ObjectValue(map[string]attr.Type{
					"id":      types.StringType,
					"version": types.StringType,
				}, grpAttrs)
				if diags.HasError() {
					return types.ListNull(types.ObjectType{}), fmt.Errorf("failed to create group object: %v", diags.Errors())
				}
				groupsElements = append(groupsElements, grpObj)
			}
			groupsList, diags := types.ListValue(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":      types.StringType,
					"version": types.StringType,
				},
			}, groupsElements)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), fmt.Errorf("failed to create groups list: %v", diags.Errors())
			}
			lfObj["groups"] = groupsList
		}

		lfValue, diags := types.ObjectValue(map[string]attr.Type{
			"s3": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bucket_name":   types.StringType,
					"bucket_prefix": types.StringType,
				},
			},
			"cloudwatch": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"log_group_name":            types.StringType,
					"log_distribution_role_arn": types.StringType,
				},
			},
			"applications": types.ListType{ElemType: types.StringType},
			"groups": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":      types.StringType,
						"version": types.StringType,
					},
				},
			},
		}, lfObj)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{}), fmt.Errorf("failed to create log forwarder object: %v", diags.Errors())
		}

		logForwarders = append(logForwarders, lfValue)
	}

	logForwardersListValue, diags := types.ListValue(lfObjType, logForwarders)
	if diags.HasError() {
		return types.ListNull(lfObjType), fmt.Errorf("failed to create log forwarders list: %v", diags.Errors())
	}

	return logForwardersListValue, nil
}
