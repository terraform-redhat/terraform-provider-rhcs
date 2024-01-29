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
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/rosa"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/sts"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

var _ datasource.DataSource = &ClusterRosaClassicDatasource{}
var _ datasource.DataSourceWithConfigure = &ClusterRosaClassicDatasource{}

const deprecatedMessage = "This attribute is not support for cluster data source. Therefore, it will not be displayed as an output of the datasource"

type ClusterRosaClassicDatasource struct {
	clusterCollection *cmv1.ClustersClient
	versionCollection *cmv1.VersionsClient
	clusterWait       common.ClusterWait
}

func NewDataSource() datasource.DataSource {
	return &ClusterRosaClassicDatasource{}
}

func (r *ClusterRosaClassicDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_rosa_classic"
}

func (r *ClusterRosaClassicDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
				Attributes:  sts.ClassicStsDatasource(),
				Computed:    true,
			},
			"multi_az": schema.BoolAttribute{
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'. " + rosa.GeneratePoolMessage(rosa.Classic),
				Computed: true,
			},
			"disable_workload_monitoring": schema.BoolAttribute{
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE) platform metrics.",
				Computed: true,
			},
			"disable_scp_checks": schema.BoolAttribute{
				Description: "Indicates if cloud permission checks are disabled when attempting installation of the cluster. " +
					common.ValueCannotBeChangedStringDescription,
				Computed: true,
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
			"ccs_enabled": schema.BoolAttribute{
				Description: "Enables customer cloud subscription (Immutable with ROSA)",
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
			"base_dns_domain": schema.StringAttribute{
				Description: "Base DNS domain name previously reserved and matching the hosted " +
					"zone name of the private Route 53 hosted zone associated with intended shared " +
					"VPC, e.g., '1vo8.p1.openshiftapps.com'. " + common.ValueCannotBeChangedStringDescription,
				Computed: true,
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
			"aws_additional_compute_security_group_ids": schema.ListAttribute{
				Description: "AWS additional compute security group ids. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Computed:    true,
			},
			"aws_additional_infra_security_group_ids": schema.ListAttribute{
				Description: "AWS additional infra security group ids. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Computed:    true,
			},
			"aws_additional_control_plane_security_group_ids": schema.ListAttribute{
				Description: "AWS additional control plane security group ids. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Computed:    true,
			},
			"kms_key_arn": schema.StringAttribute{
				Description: "The key ARN is the Amazon Resource Name (ARN) of a AWS Key Management Service (KMS) Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID" +
					"(optional). " + common.ValueCannotBeChangedStringDescription,
				Optional: true,
			},
			"fips": schema.BoolAttribute{
				Description: "Create cluster that uses FIPS Validated / Modules in Process cryptographic libraries. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"aws_private_link": schema.BoolAttribute{
				Description: "Provides private connectivity from your cluster's VPC to Red Hat SRE, without exposing traffic to the public internet. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"private": schema.BoolAttribute{
				Description: "Restrict cluster API endpoint and application routes to, private connectivity. This requires that PrivateLink be enabled and by extension, your own VPC. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"availability_zones": schema.ListAttribute{
				Description: "Availability zones. " + rosa.GeneratePoolMessage(rosa.Classic),
				ElementType: types.StringType,
				Computed:    true,
			},
			"machine_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for nodes. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"proxy": schema.SingleNestedAttribute{
				Description: "proxy",
				Attributes:  proxy.ProxyDatasource(),
				Computed:    true,
			},
			"service_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for the cluster service network. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
			},
			"pod_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for pods. " + common.ValueCannotBeChangedStringDescription,
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
			"ec2_metadata_http_tokens": schema.StringAttribute{
				Description: "This value determines which EC2 Instance Metadata Service mode to use for EC2 instances in the cluster." +
					"This can be set as `optional` (IMDS v1 or v2) or `required` (IMDSv2 only). This feature is available from " +
					"OpenShift version 4.11.0 and newer. " + common.ValueCannotBeChangedStringDescription,
				Computed: true,
			},
			"private_hosted_zone": schema.SingleNestedAttribute{
				Description: "Used in a shared VPC topology. HostedZone attributes. " + common.ValueCannotBeChangedStringDescription,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "ID assigned by AWS to private Route 53 hosted zone associated with intended shared VPC, " +
							"e.g. 'Z05646003S02O1ENCDCSN'.",
						Computed: true,
					},
					"role_arn": schema.StringAttribute{
						Description: "AWS IAM role ARN with a policy attached, granting permissions necessary to " +
							"create and manage Route 53 DNS records in private Route 53 hosted zone associated with " +
							"intended shared VPC.",
						Computed: true,
					},
				},
				Computed: true,
			},

			// Deprecated Attributes:
			// Those attributes were copied from cluster_rosa_clasic resource in order to use the same state struct.
			// We can't change the rosa_classic struct to include Embedded Structs due to that issue: https://github.com/hashicorp/terraform-plugin-framework/issues/242
			// If we will remove those attributes from the schema we will get a parsing error in the Read function
			"disable_waiting_in_destroy": schema.BoolAttribute{
				Description: deprecatedMessage,
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
			"destroy_timeout": schema.Int64Attribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"upgrade_acknowledgements_for": schema.StringAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"admin_credentials": schema.SingleNestedAttribute{
				Description: deprecatedMessage,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "Admin username that will be created with the cluster.",
						Computed:    true,
					},
					"password": schema.StringAttribute{
						Description: "Admin password that will be created with the cluster.",
						Computed:    true,
						Sensitive:   true,
					},
				},
				Computed: true,
			},
			"wait_for_create_complete": schema.BoolAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"autoscaling_enabled": schema.BoolAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"min_replicas": schema.Int64Attribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"max_replicas": schema.Int64Attribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"replicas": schema.Int64Attribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"compute_machine_type": schema.StringAttribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"worker_disk_size": schema.Int64Attribute{
				Description: deprecatedMessage,
				Computed:    true,
			},
			"default_mp_labels": schema.MapAttribute{
				Description: deprecatedMessage,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (r *ClusterRosaClassicDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *ClusterRosaClassicDatasource) Read(ctx context.Context, request datasource.ReadRequest,
	response *datasource.ReadResponse) {
	tflog.Debug(ctx, "begin Read()")
	// Get the current state:
	state := &ClusterRosaClassicState{}
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
	err = populateRosaClassicClusterState(ctx, object, state, common.DefaultHttpClient{})
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
	state.AdminCredentials = nil
	state.WaitForCreateComplete = types.BoolNull()
	state.AutoScalingEnabled = types.BoolNull()
	state.MinReplicas = types.Int64Null()
	state.MaxReplicas = types.Int64Null()
	state.Replicas = types.Int64Null()
	state.ComputeMachineType = types.StringNull()
	state.WorkerDiskSize = types.Int64Null()
	state.DefaultMPLabels = types.MapNull(types.StringType)

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}
