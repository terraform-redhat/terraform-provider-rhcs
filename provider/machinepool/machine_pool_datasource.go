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

package machinepool

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
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
				Description: "Identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the machine pool.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine pool. Must consist of lower-case alphanumeric characters or '-', start and end with an alphanumeric character. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"machine_type": schema.StringAttribute{
				Description: "Identifier of the machine type used by the nodes, " +
					"for example `m5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values. " + common.ValueCannotBeChangedStringDescription,
				Computed: true,
			},
			"replicas": schema.Int64Attribute{
				Description: "The number of machines of the pool",
				Computed:    true,
			},
			"use_spot_instances": schema.BoolAttribute{
				Description: "Use Amazon EC2 Spot Instances. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"max_spot_price": schema.Float64Attribute{
				Description: "Max Spot price. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"autoscaling_enabled": schema.BoolAttribute{
				Description: "Enables autoscaling. If `true`, this variable requires you to set a maximum and minimum replicas range using the `max_replicas` and `min_replicas` variables.",
				Computed:    true,
			},
			"min_replicas": schema.Int64Attribute{
				Description: "The minimum number of replicas for autoscaling functionality.",
				Computed:    true,
			},
			"max_replicas": schema.Int64Attribute{
				Description: "The maximum number of replicas for autoscaling functionality.",
				Computed:    true,
			},
			"taints": schema.ListNestedAttribute{
				Description: "Taints for a machine pool. Format should be a comma-separated " +
					"list of 'key=value'. This list will overwrite any modifications " +
					"made to node taints on an ongoing basis.\n",
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
				Description: "Labels for the machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"multi_availability_zone": schema.BoolAttribute{
				Description: "Create a multi-AZ machine pool for a multi-AZ cluster (default is `true`). " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"availability_zone": schema.StringAttribute{
				Description: "Select the availability zone in which to create a single AZ machine pool for a multi-AZ cluster. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"availability_zones": schema.ListAttribute{
				Description: "Availability zones. ",
				ElementType: types.StringType,
				Computed:    true,
			},
			"subnet_id": schema.StringAttribute{
				Description: "Select the subnet in which to create a single AZ machine pool for BYO-VPC cluster. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"aws_subnet_ids": schema.ListAttribute{
				Description: "AWS subnet IDs. ",
				ElementType: types.StringType,
				Computed:    true,
			},
			"disk_size": schema.Int64Attribute{
				Description: "Root disk size, in GiB. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"aws_additional_security_group_ids": schema.ListAttribute{
				Description: "AWS additional security group ids. " + common.ValueCannotBeChangedStringDescription,
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
		// If we can't find the machine pool, it was deleted. Remove if from the
		// state and don't return an error so the TF apply() will automatically
		// recreate it.
		tflog.Warn(ctx, fmt.Sprintf("machine pool (%s) of cluster (%s) not found, removing from state",
			state.ID.ValueString(), state.Cluster.ValueString(),
		))
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
