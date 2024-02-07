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
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/sts"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type ClusterRosaHcpDatasource struct {
	clusterCollection *cmv1.ClustersClient
	versionCollection *cmv1.VersionsClient
	clusterWait       common.ClusterWait
}

var _ datasource.DataSource = &ClusterRosaHcpDatasource{}
var _ datasource.DataSourceWithConfigure = &ClusterRosaHcpDatasource{}

const deprecatedMessage = "This attribute is not support for cluster data source. Therefore, it will not be displayed as an output of the datasource"

func NewDataSource() datasource.DataSource {
	return &ClusterRosaHcpDatasource{}
}

func (r *ClusterRosaHcpDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_rosa_hcp"
}

func (r *ClusterRosaHcpDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "OpenShift managed cluster using rosa sts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Required:    true,
			},
			"external_id": schema.StringAttribute{
				Description: "Unique external identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster. Cannot exceed 15 characters in length. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"cloud_region": schema.StringAttribute{
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Computed:    true,
			},
			"sts": schema.SingleNestedAttribute{
				Description: "STS configuration.",
				Attributes:  sts.HcpStsDatasource(),
				Computed:    true,
			},
			"properties": schema.MapAttribute{
				Description: "User defined properties.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"ocm_properties": schema.MapAttribute{
				Description: "Merged properties defined by OCM and the user defined 'properties'.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				Description: "Apply user defined tags to all cluster resources created in AWS. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Computed:    true,
			},
			"etcd_encryption": schema.BoolAttribute{
				Description: "Encrypt etcd data. Note that all AWS storage is already encrypted. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "DNS domain of cluster.",
				Computed:    true,
			},
			"infra_id": schema.StringAttribute{
				Description: "The ROSA cluster infrastructure ID.",
				Computed:    true,
			},
			"api_url": schema.StringAttribute{
				Description: "URL of the API server.",
				Computed:    true,
			},
			"console_url": schema.StringAttribute{
				Description: "URL of the console.",
				Computed:    true,
			},
			"aws_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"aws_subnet_ids": schema.ListAttribute{
				Description: "AWS subnet IDs. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Computed:    true,
			},
			"kms_key_arn": schema.StringAttribute{
				Description: "The key ARN is the Amazon Resource Name (ARN) of a AWS Key Management Service (KMS) Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID" +
					"(optional). " + common.ValueCannotBeChangedStringDescription,
				Optional: true,
			},
			"aws_private_link": schema.BoolAttribute{
				Description: "Provides private connectivity from your cluster's VPC to Red Hat SRE, without exposing traffic to the public internet. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"proxy": schema.SingleNestedAttribute{
				Description: "proxy",
				Attributes:  proxy.ProxyDatasource(),
				Computed:    true,
			},
			"pod_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for pods. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"machine_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for nodes. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"service_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for the cluster service network. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"host_prefix": schema.Int64Attribute{
				Description: "Length of the prefix of the subnet assigned to each node. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"current_version": schema.StringAttribute{
				Description: "The currently running version of OpenShift on the cluster, for example '4.11.0'.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "State of the cluster.",
				Computed:    true,
			},
			"channel_group": schema.StringAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"upgrade_acknowledgements_for": schema.StringAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},

			// Deprecated Attributes:
			// Those attributes were copied from cluster_rosa_clasic resource in order to use the same state struct.
			// We can't change the rosa_classic struct to include Embedded Structs due to that issue: https://github.com/hashicorp/terraform-plugin-framework/issues/242
			// If we will remove those attributes from the schema we will get a parsing error in the Read function
			"disable_waiting_in_destroy": schema.BoolAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"destroy_timeout": schema.Int64Attribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"wait_for_create_complete": schema.BoolAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			// "autoscaling_enabled": schema.BoolAttribute{
			// 	Description: deprecatedMessage,
			// 	Computed:    true,
			// },
			"replicas": schema.Int64Attribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"compute_machine_type": schema.StringAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
		},
	}
}

func (r *ClusterRosaHcpDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

	r.clusterCollection = connection.ClustersMgmt().V1().Clusters()
	r.versionCollection = connection.ClustersMgmt().V1().Versions()
	r.clusterWait = common.NewClusterWait(r.clusterCollection)
}

func (r *ClusterRosaHcpDatasource) Read(ctx context.Context, request datasource.ReadRequest,
	response *datasource.ReadResponse) {
	tflog.Debug(ctx, "begin Read()")
	// Get the current state:
	state := &ClusterRosaHcpState{}
	diags := request.Config.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the cluster:
	get, err := r.clusterCollection.Cluster(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("cluster (%s) not found, removing from state",
			state.ID.ValueString(),
		))
		response.State.RemoveResource(ctx)
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}

	object := get.Body()

	// Save the state:
	err = populateRosaHcpClusterState(ctx, object, state, common.DefaultHttpClient{})
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}

	// set deprecated attributes to null:
	state.DisableWaitingInDestroy = types.BoolNull()
	state.ChannelGroup = types.StringNull()
	state.Version = types.StringNull()
	state.DestroyTimeout = types.Int64Null()
	state.UpgradeAcksFor = types.StringNull()
	state.WaitForCreateComplete = types.BoolNull()
	//state.AutoScalingEnabled = types.BoolNull()
	state.Replicas = types.Int64Null()
	state.ComputeMachineType = types.StringNull()

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}
