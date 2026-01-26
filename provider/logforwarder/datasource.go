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

package logforwarder

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type LogForwardersDataSource struct {
	collection *cmv1.ClustersClient
}

var _ datasource.DataSource = &LogForwardersDataSource{}
var _ datasource.DataSourceWithConfigure = &LogForwardersDataSource{}

func NewDataSource() datasource.DataSource {
	return &LogForwardersDataSource{}
}

func (s *LogForwardersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_forwarders"
}

func (s *LogForwardersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of log forwarders for a cluster.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of log forwarders.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: s.itemAttributes(),
				},
				Computed: true,
			},
		},
	}
}

func (s *LogForwardersDataSource) itemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier of the log forwarder.",
			Computed:    true,
		},
		"cluster_id": schema.StringAttribute{
			Description: "Identifier of the cluster.",
			Computed:    true,
		},
		"s3": schema.SingleNestedAttribute{
			Description: "S3 configuration for log forwarding destination.",
			Computed:    true,
			Attributes: map[string]schema.Attribute{
				"bucket_name": schema.StringAttribute{
					Description: "The name of the S3 bucket.",
					Computed:    true,
				},
				"bucket_prefix": schema.StringAttribute{
					Description: "The prefix to use for objects stored in the S3 bucket.",
					Computed:    true,
				},
			},
		},
		"cloudwatch": schema.SingleNestedAttribute{
			Description: "CloudWatch configuration for log forwarding destination.",
			Computed:    true,
			Attributes: map[string]schema.Attribute{
				"log_group_name": schema.StringAttribute{
					Description: "The name of the CloudWatch log group.",
					Computed:    true,
				},
				"log_distribution_role_arn": schema.StringAttribute{
					Description: "The ARN of the IAM role for log distribution.",
					Computed:    true,
				},
			},
		},
		"applications": schema.ListAttribute{
			Description: "List of additional applications to forward logs for.",
			ElementType: types.StringType,
			Computed:    true,
		},
		"groups": schema.ListNestedAttribute{
			Description: "List of log forwarder groups.",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "The identifier of the log forwarder group.",
						Computed:    true,
					},
					"version": schema.StringAttribute{
						Description: "The version of the log forwarder group.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (s *LogForwardersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	s.collection = connection.ClustersMgmt().V1().Clusters()
}

func (s *LogForwardersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	state := &LogForwardersState{}
	diags := req.Config.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := state.Cluster.ValueString()

	logForwardersClient := s.collection.Cluster(clusterId).ControlPlane().LogForwarders()
	listResponse, err := logForwardersClient.List().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't list log forwarders",
			fmt.Sprintf("Can't list log forwarders for cluster '%s': %v", clusterId, err),
		)
		return
	}

	logForwardersList := listResponse.Items()
	state.Items = make([]*LogForwarderItem, 0, logForwardersList.Len())

	logForwardersList.Each(func(logForwarder *cmv1.LogForwarder) bool {
		lfState := &LogForwarderItem{
			ID:        types.StringValue(logForwarder.ID()),
			ClusterID: types.StringValue(clusterId),
		}

		if s3, ok := logForwarder.GetS3(); ok {
			lfState.S3 = &S3Config{
				BucketName:   types.StringValue(s3.BucketName()),
				BucketPrefix: types.StringValue(s3.BucketPrefix()),
			}
		}

		if cw, ok := logForwarder.GetCloudwatch(); ok {
			lfState.CloudWatch = &CloudWatchConfig{
				LogGroupName:           types.StringValue(cw.LogGroupName()),
				LogDistributionRoleArn: types.StringValue(cw.LogDistributionRoleArn()),
			}
		}

		if apps, ok := logForwarder.GetApplications(); ok && len(apps) > 0 {
			appList := make([]types.String, 0, len(apps))
			for _, app := range apps {
				appList = append(appList, types.StringValue(app))
			}
			listValue, _ := types.ListValueFrom(ctx, types.StringType, appList)
			lfState.Applications = listValue
		} else {
			lfState.Applications = types.ListNull(types.StringType)
		}

		if groups, ok := logForwarder.GetGroups(); ok && len(groups) > 0 {
			groupList := make([]*LogForwarderGroup, 0, len(groups))
			for _, group := range groups {
				groupList = append(groupList, &LogForwarderGroup{
					ID:      types.StringValue(group.ID()),
					Version: types.StringValue(group.Version()),
				})
			}
			lfState.Groups = groupList
		}

		state.Items = append(state.Items, lfState)
		return true
	})

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
