/*
Copyright (c) 2024 Red Hat, Inc.

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
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type HcpMachinePoolDatasource struct {
	collection *cmv1.ClustersClient
}

var _ datasource.DataSource = &HcpMachinePoolDatasource{}
var _ datasource.DataSourceWithConfigure = &HcpMachinePoolDatasource{}

func NewDatasource() datasource.DataSource {
	return &HcpMachinePoolDatasource{}
}

func (r *HcpMachinePoolDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hcp_machine_pool"
}

func (r *HcpMachinePoolDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *HcpMachinePoolDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Machine pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the machine pool.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine pool. Must consist of lower-case alphanumeric characters or '-', start and end with an alphanumeric character. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
				},
			},
			"replicas": schema.Int64Attribute{
				Description: "The number of machines in the pool. " +
					"Single zone clusters need at least 2 nodes, " +
					"multizone clusters need at least 3 nodes. " +
					"The maximum is 250 for cluster versions prior to 4.14.0-0.a, " +
					"and 500 for cluster versions 4.14.0-0.a and later.",
				Computed: true,
			},
			"autoscaling": schema.SingleNestedAttribute{
				Description: "Basic autoscaling options",
				Attributes:  AutoscalingDatasource(),
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
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Taints value",
							Required:    true,
						},
						"schedule_type": schema.StringAttribute{
							Description: "Taints schedule type",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("NoSchedule", "PreferNoSchedule", "NoExecute"),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				Computed: true,
			},
			"labels": schema.MapAttribute{
				Description: "Labels for the machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"availability_zone": schema.StringAttribute{
				Description: "Select the availability zone in which to create a single AZ machine pool for a multi-AZ cluster. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"subnet_id": schema.StringAttribute{
				Description: "Select the subnet in which to create a single AZ machine pool for BYO-VPC cluster. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"status": schema.SingleNestedAttribute{
				Description: "HCP replica status",
				Attributes:  NodePoolStatusDatasource(),
				Computed:    true,
			},
			"aws_node_pool": schema.SingleNestedAttribute{
				Description: "AWS settings for node pool",
				Attributes:  AwsNodePoolDatasource(),
				Computed:    true,
			},
			"tuning_configs": schema.ListAttribute{
				Description: "A list of tuning configs attached to the replica.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"kubelet_configs": schema.StringAttribute{
				Description: "Name of the kubelet config applied to the machine pool.",
				Computed:    true,
			},
			"auto_repair": schema.BoolAttribute{
				Description: "Indicates use of autor repair for replica",
				Optional:    true,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Desired version of OpenShift for the machine pool, for example '4.11.0'. If version is greater than the currently running version, an upgrade will be scheduled.",
				Optional:    true,
			},
			"current_version": schema.StringAttribute{
				Description: "The currently running version of OpenShift on the machine pool, for example '4.11.0'.",
				Computed:    true,
			},
			"upgrade_acknowledgements_for": schema.StringAttribute{
				Description: "Indicates acknowledgement of agreements required to upgrade the cluster version between" +
					" minor versions (e.g. a value of \"4.12\" indicates acknowledgement of any agreements required to " +
					"upgrade to OpenShift 4.12.z from 4.11 or before).",
				Computed: true,
			},
			"ignore_deletion_error": schema.BoolAttribute{
				Description: "Indicates to the provider to disregard API errors when deleting the machine pool." +
					" This will remove the resource from the management file, but not necessirely delete the underlying pool in case it errors." +
					" Setting this to true can bypass issues when destroying the cluster resource alongside the pool resource in the same management file." +
					" This is not recommended to be set in other use cases",
				Computed: true,
			},
		},
	}
}

func (r *HcpMachinePoolDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the current state:
	state := &HcpMachinePoolState{}
	diags := req.Config.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = state.Name

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

	state.UpgradeAcksFor = types.StringNull()
	state.Version = types.StringNull()
	state.IgnoreDeletionError = types.BoolNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
