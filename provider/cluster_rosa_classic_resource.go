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
package provider

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/openshift/rosa/pkg/helper"
	"github.com/openshift/rosa/pkg/properties"

	"github.com/openshift/rosa/pkg/ocm"
	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	ocmr "github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/idps"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/upgrade"

	semver "github.com/hashicorp/go-version"
	ver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfrschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocm_errors "github.com/openshift-online/ocm-sdk-go/errors"
)

const (
	awsCloudProvider      = "aws"
	rosaProduct           = "rosa"
	MinVersion            = "4.10.0"
	maxClusterNameLength  = 15
	tagsPrefix            = "rosa_"
	tagsOpenShiftVersion  = tagsPrefix + "openshift_version"
	lowestHttpTokensVer   = "4.11.0"
	propertyRosaTfVersion = tagsPrefix + "tf_version"
	propertyRosaTfCommit  = tagsPrefix + "tf_commit"
)

var OCMProperties = map[string]string{
	propertyRosaTfVersion: build.Version,
	propertyRosaTfCommit:  build.Commit,
}

type ClusterRosaClassicResource struct {
	clusterCollection *cmv1.ClustersClient
	versionCollection *cmv1.VersionsClient
}

var _ resource.ResourceWithConfigure = &ClusterRosaClassicResource{}
var _ resource.ResourceWithImportState = &ClusterRosaClassicResource{}

func NewClusterRosaClassicResource() resource.Resource {
	return &ClusterRosaClassicResource{}
}

func (r *ClusterRosaClassicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_rosa_classic"
}

func (r *ClusterRosaClassicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfrschema.Schema{
		Description: "OpenShift managed cluster using rosa sts.",
		Attributes: map[string]tfrschema.Attribute{
			"id": tfrschema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					// This passes the state through to the plan, preventing
					// "known after apply" since we know it won't change.
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_id": tfrschema.StringAttribute{
				Description: "Unique external identifier of the cluster.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"name": tfrschema.StringAttribute{
				Description: "Name of the cluster. Cannot exceed 15 characters in length.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					ValueCannotBeChangedModifier(),
				},
			},
			"cloud_region": tfrschema.StringAttribute{
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Required:    true,
			},
			"sts": tfrschema.SingleNestedAttribute{
				Description: "STS configuration.",
				Attributes:  stsResource(),
				Optional:    true,
			},
			"multi_az": tfrschema.BoolAttribute{
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"disable_workload_monitoring": tfrschema.BoolAttribute{
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE) platform metrics.",
				Optional: true,
			},
			"disable_scp_checks": tfrschema.BoolAttribute{
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE) platform metrics.",
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					ValueCannotBeChangedModifier(),
				},
			},
			"properties": tfrschema.MapAttribute{
				Description: "User defined properties.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators:  propertiesValidators(),
			},
			"ocm_properties": tfrschema.MapAttribute{
				Description: "Merged properties defined by OCM and the user defined 'properties'.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"tags": tfrschema.MapAttribute{
				Description: "Apply user defined tags to all resources created in AWS.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					ValueCannotBeChangedModifier(),
				},
			},
			"ccs_enabled": tfrschema.BoolAttribute{
				Description: "Enables customer cloud subscription.",
				Computed:    true,
			},
			"etcd_encryption": tfrschema.BoolAttribute{
				Description: "Encrypt etcd data.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"autoscaling_enabled": tfrschema.BoolAttribute{
				Description: "Enables autoscaling.",
				Optional:    true,
			},
			"min_replicas": tfrschema.Int64Attribute{
				Description: "Minimum replicas.",
				Optional:    true,
				Computed:    true,
			},
			"max_replicas": tfrschema.Int64Attribute{
				Description: "Maximum replicas.",
				Optional:    true,
				Computed:    true,
			},
			"api_url": tfrschema.StringAttribute{
				Description: "URL of the API server.",
				Computed:    true,
			},
			"console_url": tfrschema.StringAttribute{
				Description: "URL of the console.",
				Computed:    true,
			},
			"domain": tfrschema.StringAttribute{
				Description: "DNS domain of cluster.",
				Computed:    true,
			},
			"replicas": tfrschema.Int64Attribute{
				Description: "Number of worker nodes to provision. Single zone clusters need at least 2 nodes, " +
					"multizone clusters need at least 3 nodes.",
				Optional: true,
				Computed: true,
			},
			"compute_machine_type": tfrschema.StringAttribute{
				Description: "Identifies the machine type used by the compute nodes, " +
					"for example `r5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"default_mp_labels": tfrschema.MapAttribute{
				Description: "This value is the default machine pool labels. Format should be a comma-separated list of '{\"key1\"=\"value1\", \"key2\"=\"value2\"}'. " +
					"This list overwrites any modifications made to Node labels on an ongoing basis. ",
				ElementType: types.StringType,
				Optional:    true,
			},
			"aws_account_id": tfrschema.StringAttribute{
				Description: "Identifier of the AWS account.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					ValueCannotBeChangedModifier(),
				},
			},
			"aws_subnet_ids": tfrschema.ListAttribute{
				Description: "AWS subnet IDs.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					ValueCannotBeChangedModifier(),
				},
			},
			"kms_key_arn": tfrschema.StringAttribute{
				Description: "The key ARN is the Amazon Resource Name (ARN) of a AWS Key Management Service (KMS) Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					ValueCannotBeChangedModifier(),
				},
			},
			"fips": tfrschema.BoolAttribute{
				Description: "Create cluster that uses FIPS Validated / Modules in Process cryptographic libraries.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					ValueCannotBeChangedModifier(),
				},
			},
			"aws_private_link": tfrschema.BoolAttribute{
				Description: "Provides private connectivity between VPCs, AWS services, and your on-premises networks, without exposing your traffic to the public internet.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"private": tfrschema.BoolAttribute{
				Description: "Restrict master API endpoint and application routes to direct, private connectivity.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"availability_zones": tfrschema.ListAttribute{
				Description: "Availability zones.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"machine_cidr": tfrschema.StringAttribute{
				Description: "Block of IP addresses for nodes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"proxy": tfrschema.SingleNestedAttribute{
				Description: "proxy",
				Attributes: map[string]tfrschema.Attribute{
					"http_proxy": tfrschema.StringAttribute{
						Description: "HTTP proxy.",
						Optional:    true,
					},
					"https_proxy": tfrschema.StringAttribute{
						Description: "HTTPS proxy.",
						Optional:    true,
					},
					"no_proxy": tfrschema.StringAttribute{
						Description: "No proxy.",
						Optional:    true,
					},
					"additional_trust_bundle": tfrschema.StringAttribute{
						Description: "A string containing a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store.",
						Optional:    true,
					},
				},
				Optional:   true,
				Validators: proxyValidators(),
			},
			"service_cidr": tfrschema.StringAttribute{
				Description: "Block of IP addresses for services.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"pod_cidr": tfrschema.StringAttribute{
				Description: "Block of IP addresses for pods.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"host_prefix": tfrschema.Int64Attribute{
				Description: "Length of the prefix of the subnet assigned to each node.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"channel_group": tfrschema.StringAttribute{
				Description: "Name of the channel group where you select the OpenShift cluster version, for example 'stable'.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"version": tfrschema.StringAttribute{
				Description: "Desired version of OpenShift for the cluster, for example '4.1.0'. If version is greater than the currently running version, an upgrade will be scheduled.",
				Optional:    true,
			},
			"current_version": tfrschema.StringAttribute{
				Description: "The currently running version of OpenShift on the cluster, for example '4.1.0'.",
				Computed:    true,
			},
			"disable_waiting_in_destroy": tfrschema.BoolAttribute{
				Description: "Disable addressing cluster state in the destroy resource. Default value is false.",
				Optional:    true,
			},
			"destroy_timeout": tfrschema.Int64Attribute{
				Description: "This value sets the maximum duration in minutes to allow for destroying resources. Default value is 60 minutes.",
				Optional:    true,
			},
			"state": tfrschema.StringAttribute{
				Description: "State of the cluster.",
				Computed:    true,
			},
			"ec2_metadata_http_tokens": tfrschema.StringAttribute{
				Description: "This value determines which EC2 metadata mode to use for metadata service interaction " +
					"options for EC2 instances can be optional or required. Required is available from " +
					"OpenShift version 4.11.0 and newer.",
				Optional: true,
				Computed: true,
				Validators: EnumValueValidator([]string{string(cmv1.Ec2MetadataHttpTokensOptional),
					string(cmv1.Ec2MetadataHttpTokensRequired)}),
				PlanModifiers: []planmodifier.String{
					ValueCannotBeChangedModifier(),
				},
			},
			"upgrade_acknowledgements_for": tfrschema.StringAttribute{
				Description: "Indicates acknowledgement of agreements required to upgrade the cluster version between" +
					" minor versions (e.g. a value of \"4.12\" indicates acknowledgement of any agreements required to " +
					"upgrade to OpenShift 4.12.z from 4.11 or before).",
				Optional: true,
			},
			"admin_credentials": tfrschema.SingleNestedAttribute{
				Description: "Admin user credentials",
				Attributes: map[string]tfrschema.Attribute{
					"username": tfrschema.StringAttribute{
						Description: "Admin username that will be created with the cluster.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							ValueCannotBeChangedModifier(),
						},
					},
					"password": tfrschema.StringAttribute{
						Description: "Admin password that will be created with the cluster.",
						Required:    true,
						Sensitive:   true,
						PlanModifiers: []planmodifier.String{
							ValueCannotBeChangedModifier(),
						},
					},
				},
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					ValueCannotBeChangedModifier(),
				},
				Validators: adminCredsValidators(),
			},
		},
	}
	return
}

func (r *ClusterRosaClassicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
}

const (
	errHeadline = "Can't build cluster"
)

func createClassicClusterObject(ctx context.Context,
	state *ClusterRosaClassicState, diags diag.Diagnostics) (*cmv1.Cluster, error) {

	ocmClusterResource := ocmr.NewCluster()
	builder := ocmClusterResource.GetClusterBuilder()
	clusterName := state.Name.ValueString()
	if len(clusterName) > maxClusterNameLength {
		errDescription := fmt.Sprintf("Expected a valid value for 'name' maximum of 15 characters in length. Provided Cluster name '%s' is of length '%d'",
			clusterName, len(clusterName),
		)
		tflog.Error(ctx, errDescription)

		diags.AddError(
			errHeadline,
			errDescription,
		)
		return nil, errors.New(errHeadline + "\n" + errDescription)
	}

	builder.Name(state.Name.ValueString())
	builder.CloudProvider(cmv1.NewCloudProvider().ID(awsCloudProvider))
	builder.Product(cmv1.NewProduct().ID(rosaProduct))
	builder.Region(cmv1.NewCloudRegion().ID(state.CloudRegion.ValueString()))
	multiAZ := common.Bool(state.MultiAZ)
	builder.MultiAZ(multiAZ)
	// Set default properties
	properties := make(map[string]string)
	for k, v := range OCMProperties {
		properties[k] = v
	}
	if !state.Properties.IsUnknown() && !state.Properties.IsNull() {
		for k, v := range state.Properties.Elements() {
			properties[k] = v.(types.String).ValueString()
		}
	}
	builder.Properties(properties)

	if !state.EtcdEncryption.IsUnknown() && !state.EtcdEncryption.IsNull() {
		builder.EtcdEncryption(state.EtcdEncryption.ValueBool())
	}

	if !state.ExternalID.IsUnknown() && !state.ExternalID.IsNull() {
		builder.ExternalID(state.ExternalID.ValueString())
	}

	if !state.DisableWorkloadMonitoring.IsUnknown() && !state.DisableWorkloadMonitoring.IsNull() {
		builder.DisableUserWorkloadMonitoring(state.DisableWorkloadMonitoring.ValueBool())
	}

	autoScalingEnabled := common.Bool(state.AutoScalingEnabled)
	replicas := common.OptionalInt64(state.Replicas)
	minReplicas := common.OptionalInt64(state.MinReplicas)
	maxReplicas := common.OptionalInt64(state.MaxReplicas)
	computeMachineType := common.OptionalString(state.ComputeMachineType)
	labels := common.OptionalMap(state.DefaultMPLabels)
	availabilityZones := common.OptionalList(state.AvailabilityZones)

	if err := ocmClusterResource.CreateNodes(autoScalingEnabled, replicas, minReplicas, maxReplicas,
		computeMachineType, labels, availabilityZones, multiAZ); err != nil {
		return nil, err
	}

	// ccs should be enabled in ocm rosa clusters
	ccs := cmv1.NewCCS()
	ccs.Enabled(true)

	if !state.DisableSCPChecks.IsUnknown() && !state.DisableSCPChecks.IsNull() && state.DisableSCPChecks.ValueBool() {
		ccs.DisableSCPChecks(true)
	}
	builder.CCS(ccs)

	awsTags := common.OptionalMap(state.Tags)
	ec2MetadataHttpTokens := common.OptionalString(state.Ec2MetadataHttpTokens)
	kmsKeyARN := common.OptionalString(state.KMSKeyArn)
	awsAccountID := common.OptionalString(state.AWSAccountID)
	isPrivateLink := common.Bool(state.AWSPrivateLink)
	isPrivate := common.Bool(state.Private)
	awsSubnetIDs := common.OptionalList(state.AWSSubnetIDs)
	var stsBuilder *cmv1.STSBuilder
	if state.Sts != nil {
		stsBuilder = ocmr.CreateSTS(state.Sts.RoleARN.ValueString(), state.Sts.SupportRoleArn.ValueString(),
			state.Sts.InstanceIAMRoles.MasterRoleARN.ValueString(), state.Sts.InstanceIAMRoles.WorkerRoleARN.ValueString(),
			state.Sts.OperatorRolePrefix.ValueString(), common.OptionalString(state.Sts.OIDCConfigID))
	}

	if err := ocmClusterResource.CreateAWSBuilder(awsTags, ec2MetadataHttpTokens, kmsKeyARN,
		isPrivateLink, awsAccountID, stsBuilder, awsSubnetIDs); err != nil {
		return nil, err
	}

	if err := ocmClusterResource.SetAPIPrivacy(isPrivate, isPrivateLink, stsBuilder != nil); err != nil {
		return nil, err
	}

	if !state.FIPS.IsUnknown() && !state.FIPS.IsNull() && state.FIPS.ValueBool() {
		builder.FIPS(true)
	}

	network := cmv1.NewNetwork()
	if !state.MachineCIDR.IsUnknown() && !state.MachineCIDR.IsNull() {
		network.MachineCIDR(state.MachineCIDR.ValueString())
	}
	if !state.ServiceCIDR.IsUnknown() && !state.ServiceCIDR.IsNull() {
		network.ServiceCIDR(state.ServiceCIDR.ValueString())
	}
	if !state.PodCIDR.IsUnknown() && !state.PodCIDR.IsNull() {
		network.PodCIDR(state.PodCIDR.ValueString())
	}
	if !state.HostPrefix.IsUnknown() && !state.HostPrefix.IsNull() {
		network.HostPrefix(int(state.HostPrefix.ValueInt64()))
	}
	if !network.Empty() {
		builder.Network(network)
	}

	channelGroup := ocm.DefaultChannelGroup
	if !state.ChannelGroup.IsUnknown() && !state.ChannelGroup.IsNull() {
		channelGroup = state.ChannelGroup.ValueString()
	}

	if !state.Version.IsUnknown() && !state.Version.IsNull() {
		// TODO: update it to support all cluster versions
		isSupported, err := common.IsGreaterThanOrEqual(state.Version.ValueString(), MinVersion)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error validating required cluster version %s", err))
			errDescription := fmt.Sprintf(
				"Can't check if cluster version is supported '%s': %v",
				state.Version.ValueString(), err,
			)
			diags.AddError(
				errHeadline,
				errDescription,
			)
			return nil, errors.New(errHeadline + "\n" + errDescription)
		}
		if !isSupported {
			description := fmt.Sprintf("Cluster version %s is not supported (minimal supported version is %s)", state.Version.ValueString(), MinVersion)
			tflog.Error(ctx, description)
			diags.AddError(
				errHeadline,
				description,
			)
			return nil, errors.New(errHeadline + "\n" + description)
		}
		vBuilder := cmv1.NewVersion()
		versionID := fmt.Sprintf("openshift-v%s", state.Version.ValueString())
		// When using a channel group other than the default, the channel name
		// must be appended to the version ID or the API server will return an
		// error stating unexpected channel group.
		if channelGroup != ocm.DefaultChannelGroup {
			versionID = versionID + "-" + channelGroup
		}
		vBuilder.ID(versionID)
		vBuilder.ChannelGroup(channelGroup)
		builder.Version(vBuilder)
	}

	if state.AdminCredentials != nil {
		htpasswdUsers := []*cmv1.HTPasswdUserBuilder{}
		htpasswdUsers = append(htpasswdUsers, cmv1.NewHTPasswdUser().
			Username(state.AdminCredentials.Username.ValueString()).Password(state.AdminCredentials.Password.ValueString()))
		htpassUserList := cmv1.NewHTPasswdUserList().Items(htpasswdUsers...)
		htPasswdIDP := cmv1.NewHTPasswdIdentityProvider().Users(htpassUserList)
		builder.Htpasswd(htPasswdIDP)
	}

	builder, err := buildProxy(state, builder)
	if err != nil {
		tflog.Error(ctx, "Failed to build the Proxy's attributes")
		return nil, err
	}

	object, err := builder.Build()
	return object, err
}

// =====================================================
// XXX: This is as far as I've gotten with the refactor
// =====================================================

func buildProxy(state *ClusterRosaClassicState, builder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, error) {
	proxy := cmv1.NewProxy()
	if state.Proxy != nil {
		httpsProxy := ""
		httpProxy := ""
		additionalTrustBundle := ""

		if !common.IsStringAttributeEmpty(state.Proxy.HttpProxy) {
			httpProxy = state.Proxy.HttpProxy.Value
			proxy.HTTPProxy(httpProxy)
		}
		if !common.IsStringAttributeEmpty(state.Proxy.HttpsProxy) {
			httpsProxy = state.Proxy.HttpsProxy.Value
			proxy.HTTPSProxy(httpsProxy)
		}
		if !common.IsStringAttributeEmpty(state.Proxy.NoProxy) {
			proxy.NoProxy(state.Proxy.NoProxy.Value)
		}

		if !common.IsStringAttributeEmpty(state.Proxy.AdditionalTrustBundle) {
			additionalTrustBundle = state.Proxy.AdditionalTrustBundle.Value
			builder.AdditionalTrustBundle(additionalTrustBundle)
		}

		builder.Proxy(proxy)
	}

	return builder, nil
}

// getAndValidateVersionInChannelGroup ensures that the cluster version is
// available in the channel group
func (r *ClusterRosaClassicResource) getAndValidateVersionInChannelGroup(ctx context.Context, state *ClusterRosaClassicState) (string, error) {
	channelGroup := ocm.DefaultChannelGroup
	if !state.ChannelGroup.Unknown && !state.ChannelGroup.Null {
		channelGroup = state.ChannelGroup.Value
	}

	versionList, err := r.getVersionList(ctx, channelGroup)
	if err != nil {
		return "", err
	}

	version := versionList[0]
	if !state.Version.Unknown && !state.Version.Null {
		version = state.Version.Value
	}

	tflog.Debug(ctx, fmt.Sprintf("Validating if cluster version %s is in the list of supported versions: %v", version, versionList))
	for _, v := range versionList {
		if v == version {
			return version, nil
		}
	}

	return "", fmt.Errorf("version %s is not in the list of supported versions: %v", version, versionList)
}

func validateHttpTokensVersion(ctx context.Context, state *ClusterRosaClassicState, version string) error {
	if common.IsStringAttributeEmpty(state.Ec2MetadataHttpTokens) ||
		cmv1.Ec2MetadataHttpTokens(state.Ec2MetadataHttpTokens.Value) == cmv1.Ec2MetadataHttpTokensOptional {
		return nil
	}

	greater, err := common.IsGreaterThanOrEqual(version, lowestHttpTokensVer)
	if err != nil {
		return fmt.Errorf("version '%s' is not supported: %v", version, err)
	}
	if !greater {
		msg := fmt.Sprintf("version '%s' is not supported with ec2_metadata_http_tokens, "+
			"minimum supported version is %s", version, lowestHttpTokensVer)
		tflog.Error(ctx, msg)
		return fmt.Errorf(msg)
	}
	return nil
}

func getOcmVersionMinor(ver string) string {
	version, err := semver.NewVersion(ver)
	if err != nil {
		segments := strings.Split(ver, ".")
		return fmt.Sprintf("%s.%s", segments[0], segments[1])
	}
	segments := version.Segments()
	return fmt.Sprintf("%d.%d", segments[0], segments[1])
}

// getVersionList returns a list of versions for the given channel group, sorted by
// descending semver
func (r *ClusterRosaClassicResource) getVersionList(ctx context.Context, channelGroup string) (versionList []string, err error) {
	vs, err := r.getVersions(ctx, channelGroup)
	if err != nil {
		err = fmt.Errorf("Failed to retrieve versions: %s", err)
		return
	}

	for _, v := range vs {
		versionList = append(versionList, v.RawID())
	}

	if len(versionList) == 0 {
		err = fmt.Errorf("Could not find versions")
		return
	}

	return
}
func (r *ClusterRosaClassicResource) getVersions(ctx context.Context, channelGroup string) (versions []*cmv1.Version, err error) {
	page := 1
	size := 100
	filter := strings.Join([]string{
		"enabled = 'true'",
		"rosa_enabled = 'true'",
		fmt.Sprintf("channel_group = '%s'", channelGroup),
	}, " AND ")
	for {
		var response *cmv1.VersionsListResponse
		response, err = r.versionCollection.List().
			Search(filter).
			Order("default desc, id desc").
			Page(page).
			Size(size).
			Send()
		if err != nil {
			tflog.Debug(ctx, err.Error())
			return nil, err
		}
		versions = append(versions, response.Items().Slice()...)
		if response.Size() < size {
			break
		}
		page++
	}

	// Sort list in descending order
	sort.Slice(versions, func(i, j int) bool {
		a, erra := ver.NewVersion(versions[i].RawID())
		b, errb := ver.NewVersion(versions[j].RawID())
		if erra != nil || errb != nil {
			return false
		}
		return a.GreaterThan(b)
	})

	return
}

func (r *ClusterRosaClassicResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	tflog.Debug(ctx, "begin create()")
	// Get the plan:
	state := &ClusterRosaClassicState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	summary := "Can't build cluster"

	// In case version with "openshift-v" prefix was used here,
	// Give a meaningful message to inform the user that it not supported any more
	if !state.Version.Unknown && !state.Version.Null && strings.HasPrefix(state.Version.Value, "openshift-v") {
		response.Diagnostics.AddError(
			summary,
			"Openshift version must be provided without the \"openshift-v\" prefix",
		)
		return
	}

	version, err := r.getAndValidateVersionInChannelGroup(ctx, state)
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.Value, err,
			),
		)
		return
	}

	err = validateHttpTokensVersion(ctx, state, version)
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.Value, err,
			),
		)
		return
	}

	object, err := createClassicClusterObject(ctx, state, diags)
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.Value, err,
			),
		)
		return
	}

	add, err := r.clusterCollection.Add().Body(object).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't create cluster with name '%s': %v",
				state.Name.Value, err,
			),
		)
		return
	}
	object = add.Body()

	// Save the state:
	err = populateRosaClassicClusterState(ctx, object, state, DefaultHttpClient{})
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

func (r *ClusterRosaClassicResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse) {
	tflog.Debug(ctx, "begin Read()")
	// Get the current state:
	state := &ClusterRosaClassicState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the cluster:
	get, err := r.clusterCollection.Cluster(state.ID.Value).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("cluster (%s) not found, removing from state",
			state.ID.Value,
		))
		response.State.RemoveResource(ctx)
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.ID.Value, err,
			),
		)
		return
	}

	object := get.Body()

	// Save the state:
	err = populateRosaClassicClusterState(ctx, object, state, DefaultHttpClient{})
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

func (r *ClusterRosaClassicResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse) {
	var diags diag.Diagnostics

	tflog.Debug(ctx, "begin update()")

	// Get the state:
	state := &ClusterRosaClassicState{}
	diags = request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &ClusterRosaClassicState{}
	diags = request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Schedule a cluster upgrade if a newer version is requested
	if err := r.upgradeClusterIfNeeded(ctx, state, plan); err != nil {
		response.Diagnostics.AddError(
			"Can't upgrade cluster",
			fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, %v", state.ID.Value, err),
		)
		return
	}

	clusterBuilder := cmv1.NewCluster()

	clusterBuilder, _, err := updateNodes(state, plan, clusterBuilder)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster nodes for cluster with identifier: `%s`, %v",
				state.ID.Value, err,
			),
		)
		return
	}

	clusterBuilder, _, err = updateProxy(state, plan, clusterBuilder)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update proxy's configuration for cluster with identifier: `%s`, %v",
				state.ID.Value, err,
			),
		)
		return
	}

	_, shouldPatchDisableWorkloadMonitoring := common.ShouldPatchBool(state.DisableWorkloadMonitoring, plan.DisableWorkloadMonitoring)
	if shouldPatchDisableWorkloadMonitoring {
		clusterBuilder.DisableUserWorkloadMonitoring(plan.DisableWorkloadMonitoring.Value)
	}

	shouldPatchProperties := shouldPatchProperties(state, plan)

	if shouldPatchProperties {
		properties := make(map[string]string)
		for k, v := range OCMProperties {
			properties[k] = v
		}
		if !plan.Properties.Unknown && !plan.Properties.Null {
			for k, v := range plan.Properties.Elems {
				properties[k] = v.(types.String).Value
			}
		}
		clusterBuilder.Properties(properties)
	}

	clusterSpec, err := clusterBuilder.Build()
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster patch",
			fmt.Sprintf(
				"Can't build patch for cluster with identifier '%s': %v",
				state.ID.Value, err,
			),
		)
		return
	}

	update, err := r.clusterCollection.Cluster(state.ID.Value).Update().
		Body(clusterSpec).
		SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster with identifier '%s': %v",
				state.ID.Value, err,
			),
		)
		return
	}

	// update the autoscaling enabled with the plan value (important for nil and false cases)
	state.AutoScalingEnabled = plan.AutoScalingEnabled
	// update the Replicas with the plan value (important for nil and zero value cases)
	state.Replicas = plan.Replicas

	object := update.Body()

	// Update the state:
	err = populateRosaClassicClusterState(ctx, object, plan, DefaultHttpClient{})
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}
	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

// Upgrades the cluster if the desired (plan) version is greater than the
// current version
func (r *ClusterRosaClassicResource) upgradeClusterIfNeeded(ctx context.Context, state, plan *ClusterRosaClassicState) error {
	if common.IsStringAttributeEmpty(plan.Version) || common.IsStringAttributeEmpty(state.CurrentVersion) {
		// No version information, nothing to do
		tflog.Debug(ctx, "Insufficient cluster version information to determine if upgrade should be performed.")
		return nil
	}

	tflog.Debug(ctx, "Cluster versions",
		"current_version", state.CurrentVersion.Value,
		"plan-version", plan.Version.Value,
		"state-version", state.Version.Value)

	// See if the user has changed the requested version for this run
	requestedVersionChanged := true
	if !common.IsStringAttributeEmpty(plan.Version) && !common.IsStringAttributeEmpty(state.Version) {
		if plan.Version.Value == state.Version.Value {
			requestedVersionChanged = false
		}
	}

	// Check the versions to see if we need to upgrade
	currentVersion, err := semver.NewVersion(state.CurrentVersion.Value)
	if err != nil {
		return fmt.Errorf("failed to parse current cluster version: %v", err)
	}
	// For backward compatibility
	// In case version format with "openshift-v" was already used
	// remove the prefix to adapt the right format and avoid failure
	fixedVersion := strings.TrimPrefix(plan.Version.Value, "openshift-v")
	desiredVersion, err := semver.NewVersion(fixedVersion)
	if err != nil {
		return fmt.Errorf("failed to parse desired cluster version: %v", err)
	}
	if currentVersion.GreaterThan(desiredVersion) {
		tflog.Debug(ctx, "No cluster version upgrade needed.")
		if requestedVersionChanged {
			// User changed the version they want, but actual is higher. We
			// don't support downgrades.
			return fmt.Errorf("cluster version is already above the requested version")
		}
		return nil
	}
	cancelingUpgradeOnly := desiredVersion.Equal(currentVersion)

	if !cancelingUpgradeOnly {
		if err = r.validateUpgrade(ctx, state, plan); err != nil {
			return err
		}
	}

	// Fetch existing upgrade policies
	upgrades, err := upgrade.GetScheduledUpgrades(ctx, r.clusterCollection, state.ID.Value)
	if err != nil {
		return fmt.Errorf("failed to get upgrade policies: %v", err)
	}

	// Stop if an upgrade is already in progress
	correctUpgradePending, err := upgrade.CheckAndCancelUpgrades(ctx, r.clusterCollection, upgrades, desiredVersion)
	if err != nil {
		return err
	}

	// Schedule a new upgrade
	if !correctUpgradePending && !cancelingUpgradeOnly {
		ackString := plan.UpgradeAcksFor.Value
		if err = scheduleUpgrade(ctx, r.clusterCollection, state.ID.Value, desiredVersion, ackString); err != nil {
			return err
		}
	}

	state.Version = plan.Version
	state.UpgradeAcksFor = plan.UpgradeAcksFor
	return nil
}

func (r *ClusterRosaClassicResource) validateUpgrade(ctx context.Context, state, plan *ClusterRosaClassicState) error {
	// Make sure the desired version is available
	versionId := fmt.Sprintf("openshift-v%s", state.CurrentVersion.Value)
	if !state.ChannelGroup.Unknown && !state.ChannelGroup.Null && state.ChannelGroup.Value != ocm.DefaultChannelGroup {
		versionId += "-" + state.ChannelGroup.Value
	}
	availableVersions, err := upgrade.GetAvailableUpgradeVersions(ctx, r.versionCollection, versionId)
	if err != nil {
		return fmt.Errorf("failed to get available upgrades: %v", err)
	}
	trimmedDesiredVersion := strings.TrimPrefix(plan.Version.Value, "openshift-v")
	desiredVersion, err := semver.NewVersion(trimmedDesiredVersion)
	if err != nil {
		return fmt.Errorf("failed to parse desired version: %v", err)
	}
	found := false
	for _, v := range availableVersions {
		sem, err := semver.NewVersion(v.RawID())
		if err != nil {
			return fmt.Errorf("failed to parse available upgrade version: %v", err)
		}
		if desiredVersion.Equal(sem) {
			found = true
			break
		}
	}
	if !found {
		avail := []string{}
		for _, v := range availableVersions {
			avail = append(avail, v.RawID())
		}
		return fmt.Errorf("desired version (%s) is not in the list of available upgrades (%v)", desiredVersion, avail)
	}

	return nil
}

// Ensure user has acked upgrade gates and schedule the upgrade
func scheduleUpgrade(ctx context.Context, client *cmv1.ClustersClient, clusterID string, desiredVersion *semver.Version, userAckString string) error {
	// Gate agreements are checked when the upgrade is scheduled, resulting
	// in an error return. ROSA cli does this by scheduling once w/ dryRun
	// to look for un-acked agreements.
	clusterClient := client.Cluster(clusterID)
	upgradePoliciesClient := clusterClient.UpgradePolicies()
	gates, description, err := upgrade.CheckMissingAgreements(desiredVersion.String(), clusterID, upgradePoliciesClient)
	if err != nil {
		return fmt.Errorf("failed to check for missing upgrade agreements: %v", err)
	}
	// User ack is required if we have any non-STS-only gates
	userAckRequired := false
	for _, gate := range gates {
		if !gate.STSOnly() {
			userAckRequired = true
		}
	}
	targetMinorVersion := getOcmVersionMinor(desiredVersion.String())
	if userAckRequired && userAckString != targetMinorVersion { // User has not acknowledged mandatory gates, stop here.
		return fmt.Errorf("%s\nTo acknowledge these items, please add \"upgrade_acknowledgements_for = %s\""+
			" and re-apply the changes", description, targetMinorVersion)
	}

	// Ack all gates to OCM
	for _, gate := range gates {
		gateID := gate.ID()
		tflog.Debug(ctx, "Acknowledging version gate", "gateID", gateID)
		gateAgreementsClient := clusterClient.GateAgreements()
		err := upgrade.AckVersionGate(gateAgreementsClient, gateID)
		if err != nil {
			return fmt.Errorf("failed to acknowledge version gate '%s' for cluster '%s': %v",
				gateID, clusterID, err)
		}
	}

	// Schedule an upgrade
	tenMinFromNow := time.Now().UTC().Add(10 * time.Minute)
	newPolicy, err := cmv1.NewUpgradePolicy().
		ScheduleType("manual").
		Version(desiredVersion.String()).
		NextRun(tenMinFromNow).
		Build()
	if err != nil {
		return fmt.Errorf("failed to create upgrade policy: %v", err)
	}
	_, err = clusterClient.UpgradePolicies().
		Add().
		Body(newPolicy).
		SendContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to schedule upgrade: %v", err)
	}
	return nil
}

func updateProxy(state, plan *ClusterRosaClassicState, clusterBuilder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, bool, error) {
	shouldUpdateProxy := false
	if (state.Proxy == nil && plan.Proxy != nil) || (state.Proxy != nil && plan.Proxy == nil) {
		shouldUpdateProxy = true
	} else if state.Proxy != nil && plan.Proxy != nil {
		_, patchNoProxy := common.ShouldPatchString(state.Proxy.NoProxy, plan.Proxy.NoProxy)
		_, patchHttpProxy := common.ShouldPatchString(state.Proxy.HttpProxy, plan.Proxy.HttpProxy)
		_, patchHttpsProxy := common.ShouldPatchString(state.Proxy.HttpsProxy, plan.Proxy.HttpsProxy)
		_, patchAdditionalTrustBundle := common.ShouldPatchString(state.Proxy.AdditionalTrustBundle, plan.Proxy.AdditionalTrustBundle)
		if patchNoProxy || patchHttpProxy || patchHttpsProxy || patchAdditionalTrustBundle {
			shouldUpdateProxy = true
		}
	}

	if shouldUpdateProxy {
		var err error
		clusterBuilder, err = buildProxy(plan, clusterBuilder)
		if err != nil {
			return nil, false, err
		}
	}

	return clusterBuilder, shouldUpdateProxy, nil
}
func updateNodes(state, plan *ClusterRosaClassicState, clusterBuilder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, bool, error) {
	// Send request to update the cluster:
	shouldUpdateNodes := false
	clusterNodesBuilder := cmv1.NewClusterNodes()
	compute, ok := common.ShouldPatchInt(state.Replicas, plan.Replicas)
	if ok {
		clusterNodesBuilder = clusterNodesBuilder.Compute(int(compute))
		shouldUpdateNodes = true
	}

	if !plan.AutoScalingEnabled.Unknown && !plan.AutoScalingEnabled.Null && plan.AutoScalingEnabled.Value {
		// autoscaling enabled
		autoscaling := cmv1.NewMachinePoolAutoscaling()

		if !plan.MaxReplicas.Unknown && !plan.MaxReplicas.Null {
			autoscaling = autoscaling.MaxReplicas(int(plan.MaxReplicas.Value))
		}
		if !plan.MinReplicas.Unknown && !plan.MinReplicas.Null {
			autoscaling = autoscaling.MinReplicas(int(plan.MinReplicas.Value))
		}

		clusterNodesBuilder = clusterNodesBuilder.AutoscaleCompute(autoscaling)
		shouldUpdateNodes = true

	} else {
		if (!plan.MaxReplicas.Unknown && !plan.MaxReplicas.Null) || (!plan.MinReplicas.Unknown && !plan.MinReplicas.Null) {
			return nil, false, fmt.Errorf("Can't update MaxReplica and/or MinReplica of cluster when autoscaling is not enabled")
		}
	}

	// MP labels update
	if !plan.DefaultMPLabels.Unknown && !plan.DefaultMPLabels.Null {
		if labelsPlan, ok := common.ShouldPatchMap(state.DefaultMPLabels, plan.DefaultMPLabels); ok {
			labels := map[string]string{}
			for k, v := range labelsPlan.Elems {
				labels[k] = v.(types.String).Value
			}
			clusterNodesBuilder.ComputeLabels(labels)
			shouldUpdateNodes = true
		}
	}

	if shouldUpdateNodes {
		clusterBuilder = clusterBuilder.Nodes(clusterNodesBuilder)
	}

	return clusterBuilder, shouldUpdateNodes, nil
}

func (r *ClusterRosaClassicResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse) {
	tflog.Debug(ctx, "begin delete()")

	// Get the state:
	state := &ClusterRosaClassicState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send the request to delete the cluster:
	resource := r.clusterCollection.Cluster(state.ID.Value)
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete cluster",
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				state.ID.Value, err,
			),
		)
		return
	}
	if !state.DisableWaitingInDestroy.Unknown && !state.DisableWaitingInDestroy.Null && state.DisableWaitingInDestroy.Value {
		tflog.Info(ctx, "Waiting for destroy to be completed, is disabled")
	} else {
		timeout := defaultTimeoutInMinutes
		if !state.DestroyTimeout.Unknown && !state.DestroyTimeout.Null {
			if state.DestroyTimeout.Value <= 0 {
				response.Diagnostics.AddWarning(nonPositiveTimeoutSummary, fmt.Sprintf(nonPositiveTimeoutFormat, state.ID.Value))
			} else {
				timeout = state.DestroyTimeout.Value
			}
		}
		isNotFound, err := r.retryClusterNotFoundWithTimeout(3, 1*time.Minute, ctx, timeout, resource)
		if err != nil {
			response.Diagnostics.AddError(
				"Can't poll cluster state",
				fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					state.ID.Value, err,
				),
			)
			return
		}

		if !isNotFound {
			response.Diagnostics.AddWarning(
				"Cluster wasn't deleted yet",
				fmt.Sprintf("The cluster with identifier '%s' is not deleted yet, but the polling finisehd due to a timeout", state.ID.Value),
			)
		}

	}
	// Remove the state:
	response.State.RemoveResource(ctx)
}

func (r *ClusterRosaClassicResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	tflog.Debug(ctx, "begin importstate()")

	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), request, response)
}

// populateRosaClassicClusterState copies the data from the API object to the Terraform state.
func populateRosaClassicClusterState(ctx context.Context, object *cmv1.Cluster, state *ClusterRosaClassicState, httpClient HttpClient) error {
	state.ID = types.String{
		Value: object.ID(),
	}
	state.ExternalID = types.String{
		Value: object.ExternalID(),
	}
	object.API()
	state.Name = types.String{
		Value: object.Name(),
	}
	state.CloudRegion = types.String{
		Value: object.Region().ID(),
	}
	state.MultiAZ = types.Bool{
		Value: object.MultiAZ(),
	}

	state.Properties = types.Map{
		ElemType: types.StringType,
		Elems:    map[string]attr.Value{},
	}
	state.OCMProperties = types.Map{
		ElemType: types.StringType,
		Elems:    map[string]attr.Value{},
	}
	if props, ok := object.GetProperties(); ok {
		for k, v := range props {
			state.OCMProperties.Elems[k] = types.String{
				Value: v,
			}
			if _, isDefault := OCMProperties[k]; !isDefault {
				state.Properties.Elems[k] = types.String{
					Value: v,
				}
			}
		}
	}

	state.APIURL = types.String{
		Value: object.API().URL(),
	}
	state.ConsoleURL = types.String{
		Value: object.Console().URL(),
	}
	state.Domain = types.String{
		Value: fmt.Sprintf("%s.%s", object.Name(), object.DNS().BaseDomain()),
	}
	state.Replicas = types.Int64{
		Value: int64(object.Nodes().Compute()),
	}
	state.ComputeMachineType = types.String{
		Value: object.Nodes().ComputeMachineType().ID(),
	}

	labels, ok := object.Nodes().GetComputeLabels()
	if ok {
		state.DefaultMPLabels = types.Map{
			ElemType: types.StringType,
			Elems:    map[string]attr.Value{},
		}
		for k, v := range labels {
			state.DefaultMPLabels.Elems[k] = types.String{
				Value: v,
			}
		}
	}

	disableUserWorkload, ok := object.GetDisableUserWorkloadMonitoring()
	if ok && disableUserWorkload {
		state.DisableWorkloadMonitoring = types.Bool{
			Value: true,
		}
	}

	isFips, ok := object.GetFIPS()
	if ok && isFips {
		state.FIPS = types.Bool{
			Value: true,
		}
	}
	autoScaleCompute, ok := object.Nodes().GetAutoscaleCompute()
	if ok {
		var maxReplicas, minReplicas int
		state.AutoScalingEnabled = types.Bool{
			Value: true,
		}

		maxReplicas, ok = autoScaleCompute.GetMaxReplicas()
		if ok {
			state.MaxReplicas = types.Int64{
				Value: int64(maxReplicas),
			}
		}

		minReplicas, ok = autoScaleCompute.GetMinReplicas()
		if ok {
			state.MinReplicas = types.Int64{
				Value: int64(minReplicas),
			}
		}
	} else {
		// autoscaling not enabled - initialize the MaxReplica and MinReplica
		state.MaxReplicas.Null = true
		state.MinReplicas.Null = true
	}

	azs, ok := object.Nodes().GetAvailabilityZones()
	if ok {
		state.AvailabilityZones = types.List{
			ElemType: types.StringType,
			Elems:    []attr.Value{},
		}
		for _, az := range azs {
			state.AvailabilityZones.Elems = append(state.AvailabilityZones.Elems, types.String{
				Value: az,
			})
		}
	}

	state.CCSEnabled = types.Bool{
		Value: object.CCS().Enabled(),
	}

	disableSCPChecks, ok := object.CCS().GetDisableSCPChecks()
	if ok && disableSCPChecks {
		state.DisableSCPChecks = types.Bool{
			Value: true,
		}
	}

	state.EtcdEncryption = types.Bool{
		Value: object.EtcdEncryption(),
	}

	// Note: The API does not currently return account id, but we try to get it
	// anyway. Failing that, we fetch the creator ARN from the properties like
	// rosa cli does.
	awsAccountID, ok := object.AWS().GetAccountID()
	if ok {
		state.AWSAccountID = types.String{
			Value: awsAccountID,
		}
	} else {
		// rosa cli gets it from the properties, so we do the same
		if creatorARN, ok := object.Properties()[properties.CreatorARN]; ok {
			if arn, err := arn.Parse(creatorARN); err == nil {
				state.AWSAccountID = types.String{
					Value: arn.AccountID,
				}
			}
		}

	}

	awsPrivateLink, ok := object.AWS().GetPrivateLink()
	if ok {
		state.AWSPrivateLink = types.Bool{
			Value: awsPrivateLink,
		}
	} else {
		state.AWSPrivateLink = types.Bool{
			Null: true,
		}
	}
	listeningMethod, ok := object.API().GetListening()
	if ok {
		state.Private = types.Bool{
			Value: listeningMethod == cmv1.ListeningMethodInternal,
		}
	} else {
		state.Private = types.Bool{
			Null: true,
		}
	}
	kmsKeyArn, ok := object.AWS().GetKMSKeyArn()
	if ok {
		state.KMSKeyArn = types.String{
			Value: kmsKeyArn,
		}
	}

	httpTokensState, ok := object.AWS().GetEc2MetadataHttpTokens()
	if ok && httpTokensState != "" {
		state.Ec2MetadataHttpTokens = types.String{
			Value: string(httpTokensState),
		}
	}

	sts, ok := object.AWS().GetSTS()
	if ok {
		if state.Sts == nil {
			state.Sts = &Sts{}
		}
		oidc_endpoint_url := strings.TrimPrefix(sts.OIDCEndpointURL(), "https://")

		state.Sts.OIDCEndpointURL = types.String{
			Value: oidc_endpoint_url,
		}
		state.Sts.RoleARN = types.String{
			Value: sts.RoleARN(),
		}
		state.Sts.SupportRoleArn = types.String{
			Value: sts.SupportRoleARN(),
		}
		instanceIAMRoles := sts.InstanceIAMRoles()
		if instanceIAMRoles != nil {
			state.Sts.InstanceIAMRoles.MasterRoleARN = types.String{
				Value: instanceIAMRoles.MasterRoleARN(),
			}
			state.Sts.InstanceIAMRoles.WorkerRoleARN = types.String{
				Value: instanceIAMRoles.WorkerRoleARN(),
			}
		}
		// TODO: fix a bug in uhc-cluster-services
		if common.IsStringAttributeEmpty(state.Sts.OperatorRolePrefix) {
			operatorRolePrefix, ok := sts.GetOperatorRolePrefix()
			if ok {
				state.Sts.OperatorRolePrefix = types.String{
					Value: operatorRolePrefix,
				}
			}
		}
		thumbprint, err := getThumbprint(sts.OIDCEndpointURL(), httpClient)
		if err != nil {
			tflog.Error(ctx, "cannot get thumbprint", err)
			state.Sts.Thumbprint = types.String{
				Value: "",
			}
		} else {
			state.Sts.Thumbprint = types.String{
				Value: thumbprint,
			}
		}
		oidcConfig, ok := sts.GetOidcConfig()
		if ok && oidcConfig != nil {
			state.Sts.OIDCConfigID = types.String{
				Value: oidcConfig.ID(),
			}
		}
	}

	subnetIds, ok := object.AWS().GetSubnetIDs()
	if ok {
		state.AWSSubnetIDs.Elems = make([]attr.Value, 0)
		for _, subnetId := range subnetIds {
			state.AWSSubnetIDs.Elems = append(state.AWSSubnetIDs.Elems, types.String{
				Value: subnetId,
			})
		}
	}

	proxy, ok := object.GetProxy()
	if ok {
		if state.Proxy == nil {
			state.Proxy = &Proxy{}
		}
		httpProxy, ok := proxy.GetHTTPProxy()
		if ok {
			state.Proxy.HttpProxy = types.String{
				Value: httpProxy,
			}
		}

		httpsProxy, ok := proxy.GetHTTPSProxy()
		if ok {
			state.Proxy.HttpsProxy = types.String{
				Value: httpsProxy,
			}
		}

		noProxy, ok := proxy.GetNoProxy()
		if ok {
			state.Proxy.NoProxy = types.String{
				Value: noProxy,
			}
		}
	}

	trustBundle, ok := object.GetAdditionalTrustBundle()
	if ok && common.IsStringAttributeEmpty(state.Proxy.AdditionalTrustBundle) {
		if state.Proxy == nil {
			state.Proxy = &Proxy{}
		}
		state.Proxy.AdditionalTrustBundle = types.String{
			Value: trustBundle,
		}
	}

	machineCIDR, ok := object.Network().GetMachineCIDR()
	if ok {
		state.MachineCIDR = types.String{
			Value: machineCIDR,
		}
	} else {
		state.MachineCIDR = types.String{
			Null: true,
		}
	}
	serviceCIDR, ok := object.Network().GetServiceCIDR()
	if ok {
		state.ServiceCIDR = types.String{
			Value: serviceCIDR,
		}
	} else {
		state.ServiceCIDR = types.String{
			Null: true,
		}
	}
	podCIDR, ok := object.Network().GetPodCIDR()
	if ok {
		state.PodCIDR = types.String{
			Value: podCIDR,
		}
	} else {
		state.PodCIDR = types.String{
			Null: true,
		}
	}
	hostPrefix, ok := object.Network().GetHostPrefix()
	if ok {
		state.HostPrefix = types.Int64{
			Value: int64(hostPrefix),
		}
	} else {
		state.HostPrefix = types.Int64{
			Null: true,
		}
	}
	channel_group, ok := object.Version().GetChannelGroup()
	if ok {
		state.ChannelGroup = types.String{
			Value: channel_group,
		}
	} else {
		state.ChannelGroup = types.String{
			Null: true,
		}
	}
	version, ok := object.Version().GetID()
	// If we're using a non-default channel group, it will have been appended to
	// the version ID. Remove it before saving state.
	version = strings.TrimSuffix(version, fmt.Sprintf("-%s", channel_group))
	version = strings.TrimPrefix(version, "openshift-v")
	if ok {
		tflog.Debug(ctx, "actual cluster version: %v", version)
		state.CurrentVersion = types.String{
			Value: version,
		}
	} else {
		tflog.Debug(ctx, "unknown cluster version")
		state.CurrentVersion = types.String{
			Null: true,
		}
	}
	state.State = types.String{
		Value: string(object.State()),
	}
	state.Name = types.String{
		Value: object.Name(),
	}
	state.CloudRegion = types.String{
		Value: object.Region().ID(),
	}

	return nil
}

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type DefaultHttpClient struct {
}

func (c DefaultHttpClient) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func getThumbprint(oidcEndpointURL string, httpClient HttpClient) (thumbprint string, err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			fmt.Fprintf(os.Stderr, "recovering from: %q\n", panicErr)
			thumbprint = ""
			err = fmt.Errorf("recovering from: %q", panicErr)
		}
	}()

	connect, err := url.ParseRequestURI(oidcEndpointURL)
	if err != nil {
		return "", err
	}

	response, err := httpClient.Get(fmt.Sprintf("https://%s:443", connect.Host))
	if err != nil {
		return "", err
	}

	certChain := response.TLS.PeerCertificates

	// Grab the CA in the chain
	for _, cert := range certChain {
		if cert.IsCA {
			if bytes.Equal(cert.RawIssuer, cert.RawSubject) {
				hash, err := sha1Hash(cert.Raw)
				if err != nil {
					return "", err
				}
				return hash, nil
			}
		}
	}

	// Fall back to using the last certficiate in the chain
	cert := certChain[len(certChain)-1]
	return sha1Hash(cert.Raw)
}

// sha1Hash computes the SHA1 of the byte array and returns the hex encoding as a string.
func sha1Hash(data []byte) (string, error) {
	// nolint:gosec
	hasher := sha1.New()
	_, err := hasher.Write(data)
	if err != nil {
		return "", fmt.Errorf("Couldn't calculate hash:\n %v", err)
	}
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed), nil
}

func (r *ClusterRosaClassicResource) retryClusterNotFoundWithTimeout(attempts int, sleep time.Duration, ctx context.Context, timeout int64,
	resource *cmv1.ClusterClient) (bool, error) {
	isNotFound, err := r.waitTillClusterIsNotFoundWithTimeout(ctx, timeout, resource)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return r.retryClusterNotFoundWithTimeout(attempts, 2*sleep, ctx, timeout, resource)
		}
		return isNotFound, err
	}

	return isNotFound, nil
}

func (r *ClusterRosaClassicResource) waitTillClusterIsNotFoundWithTimeout(ctx context.Context, timeout int64,
	resource *cmv1.ClusterClient) (bool, error) {
	timeoutInMinutes := time.Duration(timeout) * time.Minute
	pollCtx, cancel := context.WithTimeout(ctx, timeoutInMinutes)
	defer cancel()
	_, err := resource.Poll().
		Interval(pollingIntervalInMinutes * time.Minute).
		Status(http.StatusNotFound).
		StartContext(pollCtx)
	sdkErr, ok := err.(*ocm_errors.Error)
	if ok && sdkErr.Status() == http.StatusNotFound {
		tflog.Info(ctx, "Cluster was removed")
		return true, nil
	}
	if err != nil {
		tflog.Error(ctx, "Can't poll cluster deletion")
		return false, err
	}

	return false, nil
}

func proxyValidators() []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate proxy's attributes",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &Proxy{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				errSum := "Invalid proxy's attribute assignment"
				httpsProxy := ""
				httpProxy := ""
				additionalTrustBundle := ""
				var noProxySlice []string

				if !common.IsStringAttributeEmpty(state.HttpProxy) {
					httpProxy = state.HttpProxy.Value
				}
				if !common.IsStringAttributeEmpty(state.HttpsProxy) {
					httpsProxy = state.HttpsProxy.Value
				}
				if !common.IsStringAttributeEmpty(state.NoProxy) {
					noProxySlice = helper.HandleEmptyStringOnSlice(strings.Split(state.NoProxy.Value, ","))
				}

				if !common.IsStringAttributeEmpty(state.AdditionalTrustBundle) {
					additionalTrustBundle = state.AdditionalTrustBundle.Value
				}

				if httpProxy == "" && httpsProxy == "" && noProxySlice != nil && len(noProxySlice) > 0 {
					resp.Diagnostics.AddError(errSum, "Expected at least one of the following: http-proxy, https-proxy")
					return
				}

				if httpProxy == "" && httpsProxy == "" && additionalTrustBundle == "" {
					resp.Diagnostics.AddError(errSum, "Expected at least one of the following: http-proxy, https-proxy, additional-trust-bundle")
					return
				}
			},
		},
	}
}

func shouldPatchProperties(state, plan *ClusterRosaClassicState) bool {
	// User defined properties needs update
	if _, should := common.ShouldPatchMap(state.Properties, plan.Properties); should {
		return true
	}

	extractedDefaults := map[string]string{}
	for k, v := range state.OCMProperties.Elems {
		if _, ok := state.Properties.Elems[k]; !ok {
			extractedDefaults[k] = v.(types.String).Value
		}
	}

	if len(extractedDefaults) != len(OCMProperties) {
		return true
	}

	for k, v := range OCMProperties {
		if _, ok := extractedDefaults[k]; !ok {
			return true
		} else if extractedDefaults[k] != v {
			return true
		}

	}

	return false

}

func propertiesValidators() []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate property key override",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				propertiesState := &types.Map{
					ElemType: types.StringType,
				}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, propertiesState)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				if !propertiesState.Null && !propertiesState.Unknown {
					for k := range propertiesState.Elems {
						if _, isDefaultKey := OCMProperties[k]; isDefaultKey {
							errHead := "Invalid property key."
							errDesc := fmt.Sprintf("Can not override reserved properties keys. %s is a reserved property key", k)
							resp.Diagnostics.AddError(errHead, errDesc)
							return
						}
					}
				}
			},
		},
	}
}

func adminCredsValidators() []tfsdk.AttributeValidator {
	errSumm := "Invalid admin_creedntials"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate admin username",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				var creds *AdminCredentials
				diag := req.Config.GetAttribute(ctx, req.AttributePath, creds)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				if creds != nil {
					if common.IsStringAttributeEmpty(creds.Username) {
						diag.AddError(errSumm, "Usename can't be empty")
						return
					}
					if err := idps.ValidateHTPasswdUsername(creds.Username.Value); err != nil {
						diag.AddError(errSumm, err.Error())
						return
					}
				}
			},
		},
		&common.AttributeValidator{
			Desc: "Validate admin password",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				var creds *AdminCredentials
				diag := req.Config.GetAttribute(ctx, req.AttributePath, creds)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				if creds != nil {
					if common.IsStringAttributeEmpty(creds.Password) {
						diag.AddError(errSumm, "Usename can't be empty")
						return
					}
					if err := idps.ValidateHTPasswdPassword(creds.Password.Value); err != nil {
						diag.AddError(errSumm, err.Error())
						return
					}
				}
			},
		},
	}
}
