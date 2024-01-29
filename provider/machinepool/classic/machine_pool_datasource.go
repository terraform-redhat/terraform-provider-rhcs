/*
Copyright (c) 2021 Red Hat, Inc.

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

package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type MachinePoolDatasource struct {
	collection *cmv1.ClustersClient
}

var _ datasource.DataSource = &MachinePoolDatasource{}
var _ datasource.DataSourceWithConfigure = &MachinePoolDatasource{}

func NewDatasource() datasource.DataSource {
	return &MachinePoolDatasource{}
}

func (r *MachinePoolDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_pool"
}

func (r *MachinePoolDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.collection = connection.ClustersMgmt().V1().Clusters()
}

func (r *MachinePoolDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Machine pool.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster of the machine pool. ",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the machine pool.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the machine pool",
				Computed:    true,
			},
			"machine_type": schema.StringAttribute{
				Description: "Identifier of the machine type used by the nodes, for example `m5.xlarge`. ",
				Computed:    true,
			},
			"replicas": schema.Int64Attribute{
				Description: "The machines number in the machine pool. relevant only in case of 'autoscaling_enabled = false'",
				Computed:    true,
			},
			"use_spot_instances": schema.BoolAttribute{
				Description: "Indicates if Amazon EC2 Spot Instances used in this machine pool.",
				Computed:    true,
			},
			"max_spot_price": schema.Float64Attribute{
				Description: "Max Spot price.",
				Computed:    true,
			},
			"autoscaling_enabled": schema.BoolAttribute{
				Description: "Specifies whether auto-scaling is activated for this machine pool.",
				Computed:    true,
			},
			"min_replicas": schema.Int64Attribute{
				Description: "The minimum number of replicas for autos-caling functionality. relevant only in case of 'autoscaling_enabled = true",
				Computed:    true,
			},
			"max_replicas": schema.Int64Attribute{
				Description: "The maximum number of replicas for auto-scaling functionality. relevant only in case of 'autoscaling_enabled = true'",
				Computed:    true,
			},
			"taints": schema.ListNestedAttribute{
				Description: "The list of the Taints of this machine pool.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Taints key",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Taints value",
							Computed:    true,
						},
						"schedule_type": schema.StringAttribute{
							Description: "Taints schedule type",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
			"labels": schema.MapAttribute{
				Description: "The list of the Labels of this machine pool.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"multi_availability_zone": schema.BoolAttribute{
				Description: "Specifies whether this machine pool is a multi-AZ machine pool. Relevant only in case of multi-AZ cluster",
				Computed:    true,
			},
			"availability_zone": schema.StringAttribute{
				Description: "A single availability zone in which the machines of this machine pool are created. Relevant only for a single availability zone machine pool. For multiple availability zones check \"availability_zones\" attribute",
				Computed:    true,
			},
			"availability_zones": schema.ListAttribute{
				Description: "A list of Availability Zones. Relevant only for multiple availability zones machine pool. For single availability zone check \"availability_zone\" attribute.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"subnet_id": schema.StringAttribute{
				Description: "An ID of single subnet in which the machines of this machine pool are created. Relevant only for a machine pool with single subnet. For machine pool with multiple subnets check \"subnet_ids\" attribute",
				Computed:    true,
			},
			"subnet_ids": schema.ListAttribute{
				Description: "A list of IDs of subnets in which the machines of this machine pool are created. Relevant only for a machine pool with multiple subnets. For machine pool with single subnet check \"subnet_id\" attribute",
				ElementType: types.StringType,
				Computed:    true,
			},
			"disk_size": schema.Int64Attribute{
				Description: "The root disk size, in GiB.",
				Computed:    true,
			},
			"aws_additional_security_group_ids": schema.ListAttribute{
				Description: "AWS additional security group ids.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (r *MachinePoolDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the current state:
	state := &MachinePoolState{}
	diags := req.Config.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notFound, diags := readState(ctx, state, r.collection)
	if notFound {
		diags.AddError(
			"Failed to find machine pool",
			fmt.Sprintf(
				"Failed to find machine pool with identifier %s for cluster %s.",
				state.ID.ValueString(), state.Cluster.ValueString(),
			),
		)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
