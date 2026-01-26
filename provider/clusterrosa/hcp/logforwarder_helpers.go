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
