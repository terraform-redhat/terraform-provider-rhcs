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

package clusterrosaclassic

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"
	"net/http"
	"strings"
)

type ClusterDataSource struct {
	collection *cmv1.ClustersClient
}

var _ datasource.DataSource = &ClusterDataSource{}
var _ datasource.DataSourceWithConfigure = &ClusterDataSource{}

func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{}
}

func (c *ClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_rosa_classic"
}

func (c *ClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Show details of a cluster",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster. Cannot exceed 15 characters in length.",
				Computed:    true,
			},
			"external_id": schema.StringAttribute{
				Description: "Unique external identifier of the cluster.",
				Computed:    true,
			},
			"cloud_region": schema.StringAttribute{
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Computed:    true,
			},
			"sts": schema.SingleNestedAttribute{
				Description: "STS configuration.",
				Attributes: map[string]schema.Attribute{
					"oidc_endpoint_url": schema.StringAttribute{
						Description: "OIDC Endpoint URL",
						Computed:    true,
					},
					"oidc_config_id": schema.StringAttribute{
						Description: "OIDC Configuration ID",
						Computed:    true,
					},
					"thumbprint": schema.StringAttribute{
						Description: "SHA1-hash value of the root CA of the issuer URL",
						Computed:    true,
					},
					"role_arn": schema.StringAttribute{
						Description: "Installer Role",
						Computed:    true,
					},
					"support_role_arn": schema.StringAttribute{
						Description: "Support Role",
						Computed:    true,
					},
					"instance_iam_roles": schema.SingleNestedAttribute{
						Description: "Instance IAM Roles",
						Attributes: map[string]schema.Attribute{
							"master_role_arn": schema.StringAttribute{
								Description: "Master/Control Plane Node Role ARN",
								Computed:    true,
							},
							"worker_role_arn": schema.StringAttribute{
								Description: "Worker/Compute Node Role ARN",
								Computed:    true,
							},
						},
						Computed: true,
					},
					"operator_role_prefix": schema.StringAttribute{
						Description: "Operator IAM Role prefix",
						Computed:    true,
					},
					"operator_iam_role_arns": schema.ListAttribute{
						Description: "Operator IAM Role ARNs",
						ElementType: types.StringType,
						Computed:    true,
					},
				},
				Computed: true,
			},
			"multi_az": schema.BoolAttribute{
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Computed: true,
			},
			"disable_workload_monitoring": schema.BoolAttribute{
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE) platform metrics.",
				Computed: true,
			},
			"disable_scp_checks": schema.BoolAttribute{
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE) platform metrics.",
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
				Description: "Apply user defined tags to all cluster resources created in AWS.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"ccs_enabled": schema.BoolAttribute{
				Description: "Enables customer cloud subscription (Immutable with ROSA)",
				Computed:    true,
			},
			"etcd_encryption": schema.BoolAttribute{
				Description: "Encrypt etcd data. Note that all AWS storage is already encrypted.",
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
			"domain": schema.StringAttribute{
				Description: "DNS domain of cluster.",
				Computed:    true,
			},
			"base_dns_domain": schema.StringAttribute{
				Description: "Base DNS domain name previously reserved and matching the hosted " +
					"zone name of the private Route 53 hosted zone associated with intended shared " +
					"VPC, e.g., '1vo8.p1.openshiftapps.com'.",
				Computed: true,
			},
			"aws_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account.",
				Computed:    true,
			},
			"aws_subnet_ids": schema.ListAttribute{
				Description: "AWS subnet IDs.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"kms_key_arn": schema.StringAttribute{
				Description: "The key ARN is the Amazon Resource Name (ARN) of a AWS Key Management Service (KMS) Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID" +
					"(optional).",
				Computed: true,
			},
			"fips": schema.BoolAttribute{
				Description: "Create cluster that uses FIPS Validated / Modules in Process cryptographic libraries.",
				Computed:    true,
			},
			"aws_private_link": schema.BoolAttribute{
				Description: "Provides private connectivity from your cluster's VPC to Red Hat SRE, without exposing traffic to the public internet.",
				Computed:    true,
			},
			"private": schema.BoolAttribute{
				Description: "Restrict cluster API endpoint and application routes to, private connectivity. This requires that PrivateLink be enabled and by extension, your own VPC.",
				Computed:    true,
			},
			"availability_zones": schema.ListAttribute{
				Description: "Availability zones.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"machine_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for nodes.",
				Computed:    true,
			},
			"proxy": schema.SingleNestedAttribute{
				Description: "proxy",
				Attributes: map[string]schema.Attribute{
					"http_proxy": schema.StringAttribute{
						Description: "HTTP proxy.",
						Computed:    true,
					},
					"https_proxy": schema.StringAttribute{
						Description: "HTTPS proxy.",
						Computed:    true,
					},
					"no_proxy": schema.StringAttribute{
						Description: "No proxy.",
						Computed:    true,
					},
					"additional_trust_bundle": schema.StringAttribute{
						Description: "A string containing a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store.",
						Computed:    true,
					},
				},
				Computed: true,
			},
			"service_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for the cluster service network.",
				Computed:    true,
			},
			"pod_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for pods.",
				Computed:    true,
			},
			"host_prefix": schema.Int64Attribute{
				Description: "Length of the prefix of the subnet assigned to each node.",
				Computed:    true,
			},
			"channel_group": schema.StringAttribute{
				Description: "Name of the channel group where you select the OpenShift cluster version, for example 'stable'. For ROSA, only 'stable' is supported.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Desired version of OpenShift for the cluster, for example '4.11.0'. If version is greater than the currently running version, an upgrade will be scheduled.",
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
					"OpenShift version 4.11.0 and newer.",
				Computed: true,
			},
			"upgrade_acknowledgements_for": schema.StringAttribute{
				Description: "Indicates acknowledgement of agreements required to upgrade the cluster version between" +
					" minor versions (e.g. a value of \"4.12\" indicates acknowledgement of any agreements required to " +
					"upgrade to OpenShift 4.12.z from 4.11 or before).",
				Computed: true,
			},
			"admin_credentials": schema.SingleNestedAttribute{
				Description: "Admin user credentials",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "Admin username that will be created with the cluster.",
						Computed:    true,
					},
					"password": schema.StringAttribute{
						Description: "Admin password that will be created with the cluster.",
						Computed:    true,
					},
				},
				Computed: true,
			},
			"private_hosted_zone": schema.SingleNestedAttribute{
				Description: "Used in a shared VPC topology. HostedZone attributes",
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
			"aws_additional_compute_security_group_ids": schema.ListAttribute{
				Description: "AWS additional compute security group ids.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"aws_additional_infra_security_group_ids": schema.ListAttribute{
				Description: "AWS additional infra security group ids.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"aws_additional_control_plane_security_group_ids": schema.ListAttribute{
				Description: "AWS additional control plane security group ids.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"infra_id": schema.StringAttribute{
				Description: "The ROSA cluster infrastructure ID.",
				Computed:    true,
			},
		},
	}

}

func (c *ClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection)

	// Get the collection of cloud providers:
	c.collection = connection.ClustersMgmt().V1().Clusters()
}

func (c *ClusterDataSource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(path.MatchRoot("name"), path.MatchRoot("id")),
	}
}

func (c *ClusterDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	tflog.Debug(ctx, "begin Read()")
	// Get the current state:
	state := &ClusterRosaClassicState{}
	diags := request.Config.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the cluster:
	get, err := c.collection.Cluster(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
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
	err = populateClusterRosaDataSourceState(ctx, object, state, common.DefaultHttpClient{})
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

// populateRosaClassicClusterState copies the data from the API object to the Terraform state.
func populateClusterRosaDataSourceState(ctx context.Context, object *cmv1.Cluster, state *ClusterRosaClassicState, httpClient common.HttpClient) error {
	if err := populateNonOptionalAttrs(ctx, object, state, httpClient); err != nil {
		return err
	}

	if azs, ok := object.Nodes().GetAvailabilityZones(); ok {
		listValue, err := common.StringArrayToList(azs)
		if err != nil {
			return err
		}
		state.AvailabilityZones = listValue
	}
	if awsPrivateLink, ok := object.AWS().GetPrivateLink(); ok {
		state.AWSPrivateLink = types.BoolValue(awsPrivateLink)
	}
	if listeningMethod, ok := object.API().GetListening(); ok {
		state.Private = types.BoolValue(listeningMethod == cmv1.ListeningMethodInternal)
	}
	if httpTokensState, ok := object.AWS().GetEc2MetadataHttpTokens(); ok {
		state.Ec2MetadataHttpTokens = types.StringValue(string(httpTokensState))
	}

	additionalComputeSecurityGroupIds, ok := object.AWS().GetAdditionalComputeSecurityGroupIds()
	if ok {
		awsAdditionalSecurityGroupIds, err := common.StringArrayToList(additionalComputeSecurityGroupIds)
		if err != nil {
			return err
		}
		state.AWSAdditionalComputeSecurityGroupIds = awsAdditionalSecurityGroupIds
	}

	additionalInfraSecurityGroupIds, ok := object.AWS().GetAdditionalInfraSecurityGroupIds()
	if ok {
		awsAdditionalSecurityGroupIds, err := common.StringArrayToList(additionalInfraSecurityGroupIds)
		if err != nil {
			return err
		}
		state.AWSAdditionalInfraSecurityGroupIds = awsAdditionalSecurityGroupIds
	}

	additionalControlPlaneSecurityGroupIds, ok := object.AWS().GetAdditionalControlPlaneSecurityGroupIds()
	if ok {
		awsAdditionalSecurityGroupIds, err := common.StringArrayToList(additionalControlPlaneSecurityGroupIds)
		if err != nil {
			return err
		}
		state.AWSAdditionalControlPlaneSecurityGroupIds = awsAdditionalSecurityGroupIds
	}

	hasProxy := true
	hasAdditionalTrustBundle := true

	proxyObj, ok := object.GetProxy()
	if ok {
		if state.Proxy == nil {
			state.Proxy = &proxy.Proxy{}
		}
		if httpProxy, ok := proxyObj.GetHTTPProxy(); ok {
			state.Proxy.HttpProxy = types.StringValue(httpProxy)
		}
		if httpsProxy, ok := proxyObj.GetHTTPSProxy(); ok {
			state.Proxy.HttpsProxy = types.StringValue(httpsProxy)
		}
		if noProxy, ok := proxyObj.GetNoProxy(); ok {
			state.Proxy.NoProxy = types.StringValue(noProxy)
		}
	} else {
		// We cannot set the proxy to nil because the attribute state.Proxy.AdditionalTrustBundle might contain a value.
		// Due to the sensitivity of this attribute, the backend returns the value `REDUCTED` for a non-empty AdditionalTrustBundle
		// and if state.Proxy is null it will override the actual value.
		hasProxy = false
	}

	trustBundle, ok := object.GetAdditionalTrustBundle()
	if ok {
		// If AdditionalTrustBundle is not empty, the ocm-backend always "REDUCTED" (sensitive value)
		// Therefore, we would like to update the state only if the current state is Null or Empty
		// it can happen after `import` command or when it was updated from a different cli tool
		if state.Proxy == nil || common.IsStringAttributeKnownAndEmpty(state.Proxy.AdditionalTrustBundle) {
			if state.Proxy == nil {
				state.Proxy = &proxy.Proxy{}
			}
			state.Proxy.AdditionalTrustBundle = types.StringValue(trustBundle)
		}
	} else {
		hasAdditionalTrustBundle = false
	}

	// Set state.Proxy to be null only if `object.Proxy()` and `object.AdditionalTrustBundle()` are empty
	if !hasProxy && !hasAdditionalTrustBundle {
		state.Proxy = nil
	}

	if machineCIDR, ok := object.Network().GetMachineCIDR(); ok {
		state.MachineCIDR = types.StringValue(machineCIDR)
	}
	if serviceCIDR, ok := object.Network().GetServiceCIDR(); ok {
		state.ServiceCIDR = types.StringValue(serviceCIDR)
	}
	if podCIDR, ok := object.Network().GetPodCIDR(); ok {
		state.PodCIDR = types.StringValue(podCIDR)
	}
	if hostPrefix, ok := object.Network().GetHostPrefix(); ok {
		state.HostPrefix = types.Int64Value(int64(hostPrefix))
	}
	channel_group, ok := object.Version().GetChannelGroup()
	if ok {
		state.ChannelGroup = types.StringValue(channel_group)
	}

	if version, ok := object.Version().GetID(); ok {
		// If we're using a non-default channel group, it will have been appended to
		// the version ID. Remove it before saving state.
		version = strings.TrimSuffix(version, fmt.Sprintf("-%s", channel_group))
		version = strings.TrimPrefix(version, "openshift-v")
		tflog.Debug(ctx, fmt.Sprintf("actual cluster version: %v", version))
		state.CurrentVersion = types.StringValue(version)
	}

	return nil
}
