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
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	semver "github.com/hashicorp/go-version"
	ver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	ocmConsts "github.com/openshift-online/ocm-common/pkg/ocm/consts"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocm_errors "github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"

	ocmr "github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/classic/upgrade"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/rosa"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/sts"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

const (
	defaultTimeoutInMinutes   = int64(60)
	nonPositiveTimeoutSummary = "Can't poll cluster state with a non-positive timeout"
	nonPositiveTimeoutFormat  = "Can't poll state of cluster with identifier '%s', the timeout that was set is not a positive number"
	pollingIntervalInMinutes  = 2

	awsCloudProvider      = "aws"
	rosaProduct           = "rosa"
	MinVersion            = "4.10.0"
	maxClusterNameLength  = 15
	tagsPrefix            = "rosa_"
	tagsOpenShiftVersion  = tagsPrefix + "openshift_version"
	lowestHttpTokensVer   = "4.11.0"
	propertyRosaTfVersion = tagsPrefix + "tf_version"
	propertyRosaTfCommit  = tagsPrefix + "tf_commit"
	waitTimeoutInMinutes  = 60
)

type ClusterRosaHcpResource struct {
	clusterCollection *cmv1.ClustersClient
	versionCollection *cmv1.VersionsClient
	clusterWait       common.ClusterWait
}

var _ resource.ResourceWithConfigure = &ClusterRosaHcpResource{}
var _ resource.ResourceWithImportState = &ClusterRosaHcpResource{}

func New() resource.Resource {
	return &ClusterRosaHcpResource{}
}

func (r *ClusterRosaHcpResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_rosa_hcp"
}

func (r *ClusterRosaHcpResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "OpenShift managed cluster using rosa sts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					// This passes the state through to the plan, preventing
					// "known after apply" since we know it won't change.
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_id": schema.StringAttribute{
				Description: "Unique external identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster. Cannot exceed 15 characters in length. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(15),
				},
			},
			"cloud_region": schema.StringAttribute{
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Required:    true,
			},
			"sts": schema.SingleNestedAttribute{
				Description: "STS configuration.",
				Attributes:  sts.HcpStsResource(),
				Optional:    true,
			},
			"properties": schema.MapAttribute{
				Description: "User defined properties.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Map{rosa.PropertiesValidator},
			},
			"ocm_properties": schema.MapAttribute{
				Description: "Merged properties defined by OCM and the user defined 'properties'.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				Description: "Apply user defined tags to all cluster resources created in AWS. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Optional:    true,
			},
			"etcd_encryption": schema.BoolAttribute{
				Description: "Encrypt etcd data. Note that all AWS storage is already encrypted. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"autoscaling_enabled": schema.BoolAttribute{
				Description: "Enable autoscaling for the initial worker pool. " + rosa.GeneratePoolMessage(rosa.Classic),
				Optional:    true,
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
			"replicas": schema.Int64Attribute{
				Description: "Number of worker/compute nodes to provision. Single zone clusters need at least 2 nodes, " +
					"multizone clusters need at least 3 nodes. " + rosa.GeneratePoolMessage(rosa.Classic),
				Optional: true,
			},
			"compute_machine_type": schema.StringAttribute{
				Description: "Identifies the machine type used by the initial worker nodes, " +
					"for example `m5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values. " + rosa.GeneratePoolMessage(rosa.Classic),
				Optional: true,
			},
			"aws_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"aws_billing_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account for billing. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"aws_subnet_ids": schema.ListAttribute{
				Description: "AWS subnet IDs. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Optional:    true,
			},
			"kms_key_arn": schema.StringAttribute{
				Description: "The key ARN is the Amazon Resource Name (ARN) of a AWS Key Management Service (KMS) Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID" +
					"(optional). " + common.ValueCannotBeChangedStringDescription,
				Optional: true,
			},
			"aws_private_link": schema.BoolAttribute{
				Description: "Provides private connectivity from your cluster's VPC to Red Hat SRE, without exposing traffic to the public internet. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zones": schema.ListAttribute{
				Description: "Availability zones. " + rosa.GeneratePoolMessage(rosa.Hcp),
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(rosa.AvailabilityZoneValidator),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"machine_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for nodes. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"proxy": schema.SingleNestedAttribute{
				Description: "proxy",
				Attributes:  proxy.ProxyResource(),
				Optional:    true,
				Validators:  []validator.Object{proxy.ProxyValidator()},
			},
			"service_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for the cluster service network. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pod_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for pods. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host_prefix": schema.Int64Attribute{
				Description: "Length of the prefix of the subnet assigned to each node. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"channel_group": schema.StringAttribute{
				Description: "Name of the channel group where you select the OpenShift cluster version, for example 'stable'. " +
					"For ROSA, only 'stable' is supported. " + common.ValueCannotBeChangedStringDescription,
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(ocmConsts.DefaultChannelGroup),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: "Desired version of OpenShift for the cluster, for example '4.11.0'. If version is greater than the currently running version, an upgrade will be scheduled.",
				Optional:    true,
			},
			"current_version": schema.StringAttribute{
				Description: "The currently running version of OpenShift on the cluster, for example '4.11.0'.",
				Computed:    true,
			},
			"disable_waiting_in_destroy": schema.BoolAttribute{
				Description: "Disable addressing cluster state in the destroy resource. Default value is false, and so a `destroy` will wait for the cluster to be deleted.",
				Optional:    true,
			},
			"destroy_timeout": schema.Int64Attribute{
				Description: "This value sets the maximum duration in minutes to allow for destroying resources. Default value is 60 minutes.",
				Optional:    true,
			},
			"state": schema.StringAttribute{
				Description: "State of the cluster.",
				Computed:    true,
			},
			"upgrade_acknowledgements_for": schema.StringAttribute{
				Description: "Indicates acknowledgement of agreements required to upgrade the cluster version between" +
					" minor versions (e.g. a value of \"4.12\" indicates acknowledgement of any agreements required to " +
					"upgrade to OpenShift 4.12.z from 4.11 or before).",
				Optional: true,
			},
			"wait_for_create_complete": schema.BoolAttribute{
				Description: "Wait until the cluster is either in a ready state or in an error state. The waiter has a timeout of 60 minutes, with the default value set to false",
				Optional:    true,
			},
		},
	}
}

func (r *ClusterRosaHcpResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

const (
	errHeadline = "Can't build cluster"
)

func createHcpClusterObject(ctx context.Context,
	state *ClusterRosaHcpState, diags diag.Diagnostics) (*cmv1.Cluster, error) {

	ocmClusterResource := ocmr.NewCluster()
	builder := ocmClusterResource.GetClusterBuilder()
	builder.Hypershift(cmv1.NewHypershift().Enabled(true))
	builder.MultiAZ(true)
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

	// Set default properties
	properties := make(map[string]string)
	for k, v := range rosa.OCMProperties {
		properties[k] = v
	}
	if common.HasValue(state.Properties) {
		propertiesElements, err := common.OptionalMap(ctx, state.Properties)
		if err != nil {
			errDescription := fmt.Sprintf("Expected a valid Map for 'properties' '%v'",
				diags.Errors()[0].Detail(),
			)
			tflog.Error(ctx, errDescription)

			diags.AddError(
				errHeadline,
				errDescription,
			)
			return nil, errors.New(errHeadline + "\n" + errDescription)
		}

		for k, v := range propertiesElements {
			properties[k] = v
		}
	}
	builder.Properties(properties)

	if common.HasValue(state.EtcdEncryption) {
		builder.EtcdEncryption(state.EtcdEncryption.ValueBool())
	}

	if common.HasValue(state.ExternalID) {
		builder.ExternalID(state.ExternalID.ValueString())
	}

	autoScalingEnabled := common.BoolWithFalseDefault(state.AutoScalingEnabled)

	replicas := common.OptionalInt64(state.Replicas)
	computeMachineType := common.OptionalString(state.ComputeMachineType)

	availabilityZones, err := common.StringListToArray(ctx, state.AvailabilityZones)
	if err != nil {
		return nil, err
	}

	if err := ocmClusterResource.CreateNodes(rosa.Hcp, autoScalingEnabled, replicas, nil, nil,
		computeMachineType, nil, availabilityZones, true, nil); err != nil {
		return nil, err
	}

	// ccs should be enabled in ocm rosa clusters
	ccs := cmv1.NewCCS()
	ccs.Enabled(true)
	builder.CCS(ccs)

	kmsKeyARN := common.OptionalString(state.KMSKeyArn)
	awsAccountID := common.OptionalString(state.AWSAccountID)
	awsBillingAccountId := common.OptionalString(state.AWSBillingAccountID)

	isPrivateLink := common.BoolWithFalseDefault(state.AWSPrivateLink)
	awsSubnetIDs, err := common.StringListToArray(ctx, state.AWSSubnetIDs)
	if err != nil {
		return nil, err
	}
	var stsBuilder *cmv1.STSBuilder
	if state.Sts != nil {
		stsBuilder = ocmr.CreateSTS(state.Sts.RoleARN.ValueString(), state.Sts.SupportRoleArn.ValueString(),
			"", state.Sts.InstanceIAMRoles.WorkerRoleARN.ValueString(),
			state.Sts.OperatorRolePrefix.ValueString(), common.OptionalString(state.Sts.OIDCConfigID))
	}

	awsTags, err := common.OptionalMap(ctx, state.Tags)
	if err != nil {
		return nil, err
	}
	if err := ocmClusterResource.CreateAWSBuilder(rosa.Hcp, awsTags, nil, kmsKeyARN,
		isPrivateLink, awsAccountID, awsBillingAccountId, stsBuilder, awsSubnetIDs, nil, nil,
		nil, nil, nil); err != nil {
		return nil, err
	}

	if err := ocmClusterResource.SetAPIPrivacy(isPrivateLink, isPrivateLink, stsBuilder != nil); err != nil {
		return nil, err
	}

	network := cmv1.NewNetwork()
	if common.HasValue(state.MachineCIDR) {
		network.MachineCIDR(state.MachineCIDR.ValueString())
	}
	if common.HasValue(state.ServiceCIDR) {
		network.ServiceCIDR(state.ServiceCIDR.ValueString())
	}
	if common.HasValue(state.PodCIDR) {
		network.PodCIDR(state.PodCIDR.ValueString())
	}
	if common.HasValue(state.HostPrefix) {
		network.HostPrefix(int(state.HostPrefix.ValueInt64()))
	}
	if !network.Empty() {
		builder.Network(network)
	}

	channelGroup := ocmConsts.DefaultChannelGroup
	if common.HasValue(state.ChannelGroup) {
		channelGroup = state.ChannelGroup.ValueString()
	}

	if common.HasValue(state.Version) {
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
		if channelGroup != ocmConsts.DefaultChannelGroup {
			versionID = versionID + "-" + channelGroup
		}
		vBuilder.ID(versionID)
		vBuilder.ChannelGroup(channelGroup)
		builder.Version(vBuilder)
	}

	builder, err = buildProxy(state, builder)
	if err != nil {
		tflog.Error(ctx, "Failed to build the Proxy's attributes")
		return nil, err
	}

	object, err := builder.Build()
	return object, err
}

func buildProxy(state *ClusterRosaHcpState, builder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, error) {
	if state.Proxy != nil {
		proxy := cmv1.NewProxy()
		proxyIsEmpty := true

		if !common.IsStringAttributeUnknownOrEmpty(state.Proxy.HttpProxy) {
			proxy.HTTPProxy(state.Proxy.HttpProxy.ValueString())
			proxyIsEmpty = false
		}
		if !common.IsStringAttributeUnknownOrEmpty(state.Proxy.HttpsProxy) {
			proxy.HTTPSProxy(state.Proxy.HttpsProxy.ValueString())
			proxyIsEmpty = false
		}
		if !common.IsStringAttributeUnknownOrEmpty(state.Proxy.NoProxy) {
			proxy.NoProxy(state.Proxy.NoProxy.ValueString())
			proxyIsEmpty = false
		}
		if !proxyIsEmpty {
			builder.Proxy(proxy)
		}

		if !common.IsStringAttributeUnknownOrEmpty(state.Proxy.AdditionalTrustBundle) {
			builder.AdditionalTrustBundle(state.Proxy.AdditionalTrustBundle.ValueString())
		}

	}

	return builder, nil
}

// getAndValidateVersionInChannelGroup ensures that the cluster version is
// available in the channel group
func (r *ClusterRosaHcpResource) getAndValidateVersionInChannelGroup(ctx context.Context, state *ClusterRosaHcpState) (string, error) {
	channelGroup := ocmConsts.DefaultChannelGroup
	if common.HasValue(state.ChannelGroup) {
		channelGroup = state.ChannelGroup.ValueString()
	}

	versionList, err := r.getVersionList(ctx, channelGroup)
	if err != nil {
		return "", err
	}

	version := versionList[0]
	if common.HasValue(state.Version) {
		version = state.Version.ValueString()
	}

	tflog.Debug(ctx, fmt.Sprintf("Validating if cluster version %s is in the list of supported versions: %v", version, versionList))
	for _, v := range versionList {
		if v == version {
			return version, nil
		}
	}

	return "", fmt.Errorf("version %s is not in the list of supported versions: %v", version, versionList)
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
func (r *ClusterRosaHcpResource) getVersionList(ctx context.Context, channelGroup string) (versionList []string, err error) {
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
func (r *ClusterRosaHcpResource) getVersions(ctx context.Context, channelGroup string) (versions []*cmv1.Version, err error) {
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

func (r *ClusterRosaHcpResource) Create(ctx context.Context, request resource.CreateRequest,
	response *resource.CreateResponse) {
	tflog.Debug(ctx, "begin create()")

	// Get the plan:
	state := &ClusterRosaHcpState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	summary := "Can't build cluster"

	// In case version with "openshift-v" prefix was used here,
	// Give a meaningful message to inform the user that it not supported any more
	if common.HasValue(state.Version) && strings.HasPrefix(state.Version.ValueString(), "openshift-v") {
		response.Diagnostics.AddError(
			summary,
			"Openshift version must be provided without the \"openshift-v\" prefix",
		)
		return
	}

	_, err := r.getAndValidateVersionInChannelGroup(ctx, state)
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.ValueString(), err,
			),
		)
		return
	}

	object, err := createHcpClusterObject(ctx, state, diags)
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.ValueString(), err,
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
				state.Name.ValueString(), err,
			),
		)
		return
	}
	object = add.Body()

	if common.HasValue(state.WaitForCreateComplete) && state.WaitForCreateComplete.ValueBool() {
		object, err = r.clusterWait.RetryClusterReadiness(ctx, object.ID(), 3, 30*time.Second, waitTimeoutInMinutes)
		if err != nil {
			response.Diagnostics.AddError(
				"Waiting for cluster creation finished with error",
				fmt.Sprintf("Waiting for cluster creation finished with the error %v", err),
			)
			if object == nil {
				return
			}
		}
	}

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

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterRosaHcpResource) Read(ctx context.Context, request resource.ReadRequest,
	response *resource.ReadResponse) {
	tflog.Debug(ctx, "begin Read()")
	// Get the current state:
	state := &ClusterRosaHcpState{}
	diags := request.State.Get(ctx, state)
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

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func validateNoImmutableAttChange(state, plan *ClusterRosaHcpState) diag.Diagnostics {
	diags := diag.Diagnostics{}
	common.ValidateStateAndPlanEquals(state.Name, plan.Name, "name", &diags)
	common.ValidateStateAndPlanEquals(state.ExternalID, plan.ExternalID, "external_id", &diags)
	common.ValidateStateAndPlanEquals(state.Tags, plan.Tags, "tags", &diags)
	common.ValidateStateAndPlanEquals(state.EtcdEncryption, plan.EtcdEncryption, "etcd_encryption", &diags)
	common.ValidateStateAndPlanEquals(state.AWSAccountID, plan.AWSAccountID, "aws_account_id", &diags)
	common.ValidateStateAndPlanEquals(state.AWSSubnetIDs, plan.AWSSubnetIDs, "aws_subnet_ids", &diags)
	common.ValidateStateAndPlanEquals(state.KMSKeyArn, plan.KMSKeyArn, "kms_key_arn", &diags)
	common.ValidateStateAndPlanEquals(state.AWSPrivateLink, plan.AWSPrivateLink, "aws_private_link", &diags)
	common.ValidateStateAndPlanEquals(state.MachineCIDR, plan.MachineCIDR, "machine_cidr", &diags)
	common.ValidateStateAndPlanEquals(state.ServiceCIDR, plan.ServiceCIDR, "service_cidr", &diags)
	common.ValidateStateAndPlanEquals(state.PodCIDR, plan.PodCIDR, "pod_cidr", &diags)
	common.ValidateStateAndPlanEquals(state.HostPrefix, plan.HostPrefix, "host_prefix", &diags)
	common.ValidateStateAndPlanEquals(state.ChannelGroup, plan.ChannelGroup, "channel_group", &diags)

	// default node pool's attributes
	common.ValidateStateAndPlanEquals(state.AutoScalingEnabled, plan.AutoScalingEnabled, "autoscaling_enabled", &diags)
	common.ValidateStateAndPlanEquals(state.Replicas, plan.Replicas, "replicas", &diags)
	common.ValidateStateAndPlanEquals(state.ComputeMachineType, plan.ComputeMachineType, "compute_machine_type", &diags)
	common.ValidateStateAndPlanEquals(state.AvailabilityZones, plan.AvailabilityZones, "availability_zones", &diags)

	return diags

}

func (r *ClusterRosaHcpResource) Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse) {
	var diags diag.Diagnostics

	tflog.Debug(ctx, "begin update()")

	// Get the state:
	state := &ClusterRosaHcpState{}
	diags = request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &ClusterRosaHcpState{}
	diags = request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	//assert no changes on specific attributes
	diags = validateNoImmutableAttChange(state, plan)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	clusterState := "Unknown"
	if common.HasValue(state.State) && state.State.ValueString() != "" {
		clusterState = state.State.ValueString()
	}
	if clusterState != string(cmv1.ClusterStateReady) {
		response.Diagnostics.AddError(
			"Update cluster operation is only supported while cluster is ready",
			fmt.Sprintf(
				"Update cluster operation is only supported while cluster is ready, cluster state is %s", clusterState,
			),
		)
		return
	}

	// Schedule a cluster upgrade if a newer version is requested
	if err := r.upgradeClusterIfNeeded(ctx, state, plan); err != nil {
		response.Diagnostics.AddError(
			"Can't upgrade cluster",
			fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, %v", state.ID.ValueString(), err),
		)
		return
	}

	clusterBuilder := cmv1.NewCluster()

	clusterBuilder, err := updateProxy(state, plan, clusterBuilder)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update proxy's configuration for cluster with identifier: `%s`, %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}

	patchProperties := shouldPatchProperties(state, plan)
	if patchProperties {
		propertiesElements, err := common.OptionalMap(ctx, plan.Properties)
		if err != nil {
			response.Diagnostics.AddError(
				"Can't upgrade cluster",
				fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, %v", state.ID.ValueString(), err),
			)
			return
		}
		if propertiesElements != nil {
			for k, v := range rosa.OCMProperties {
				propertiesElements[k] = v
			}
			clusterBuilder.Properties(propertiesElements)
		}
	}

	clusterSpec, err := clusterBuilder.Build()
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster patch",
			fmt.Sprintf(
				"Can't build patch for cluster with identifier '%s': %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}

	update, err := r.clusterCollection.Cluster(state.ID.ValueString()).Update().
		Body(clusterSpec).
		SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster with identifier '%s': %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}

	object := update.Body()

	// Update the state:
	err = populateRosaHcpClusterState(ctx, object, plan, common.DefaultHttpClient{})
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
func (r *ClusterRosaHcpResource) upgradeClusterIfNeeded(ctx context.Context, state, plan *ClusterRosaHcpState) error {
	if common.IsStringAttributeUnknownOrEmpty(plan.Version) || common.IsStringAttributeUnknownOrEmpty(state.CurrentVersion) {
		// No version information, nothing to do
		tflog.Debug(ctx, "Insufficient cluster version information to determine if upgrade should be performed.")
		return nil
	}

	tflog.Debug(ctx, "Cluster versions",
		map[string]interface{}{
			"current_version": state.CurrentVersion.ValueString(),
			"plan-version":    plan.Version.ValueString(),
			"state-version":   state.Version.ValueString(),
		})

	// See if the user has changed the requested version for this run
	requestedVersionChanged := true
	if !common.IsStringAttributeUnknownOrEmpty(plan.Version) && !common.IsStringAttributeUnknownOrEmpty(state.Version) {
		if plan.Version.ValueString() == state.Version.ValueString() {
			requestedVersionChanged = false
		}
	}

	// Check the versions to see if we need to upgrade
	currentVersion, err := semver.NewVersion(state.CurrentVersion.ValueString())
	if err != nil {
		return fmt.Errorf("failed to parse current cluster version: %v", err)
	}
	// For backward compatibility
	// In case version format with "openshift-v" was already used
	// remove the prefix to adapt the right format and avoid failure
	fixedVersion := strings.TrimPrefix(plan.Version.ValueString(), "openshift-v")
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
	upgrades, err := upgrade.GetScheduledUpgrades(ctx, r.clusterCollection, state.ID.ValueString())
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
		ackString := plan.UpgradeAcksFor.ValueString()
		if err = scheduleUpgrade(ctx, r.clusterCollection, state.ID.ValueString(), desiredVersion, ackString); err != nil {
			return err
		}
	}

	state.Version = plan.Version
	state.UpgradeAcksFor = plan.UpgradeAcksFor
	return nil
}

func (r *ClusterRosaHcpResource) validateUpgrade(ctx context.Context, state, plan *ClusterRosaHcpState) error {
	// Make sure the desired version is available
	versionId := fmt.Sprintf("openshift-v%s", state.CurrentVersion.ValueString())
	if common.HasValue(state.ChannelGroup) && state.ChannelGroup.ValueString() != ocmConsts.DefaultChannelGroup {
		versionId += "-" + state.ChannelGroup.ValueString()
	}
	availableVersions, err := upgrade.GetAvailableUpgradeVersions(ctx, r.versionCollection, versionId)
	if err != nil {
		return fmt.Errorf("failed to get available upgrades: %v", err)
	}
	trimmedDesiredVersion := strings.TrimPrefix(plan.Version.ValueString(), "openshift-v")
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
		tflog.Debug(ctx, "Acknowledging version gate", map[string]interface{}{"gateID": gateID})
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

func updateProxy(state, plan *ClusterRosaHcpState, clusterBuilder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, error) {
	if !reflect.DeepEqual(state.Proxy, plan.Proxy) {
		var err error
		if plan.Proxy == nil {
			plan.Proxy = &proxy.Proxy{}
		}
		clusterBuilder, err = buildProxy(plan, clusterBuilder)
		if err != nil {
			return nil, err
		}
	}

	return clusterBuilder, nil
}

func (r *ClusterRosaHcpResource) Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse) {
	tflog.Debug(ctx, "begin delete()")

	// Get the state:
	state := &ClusterRosaHcpState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send the request to delete the cluster:
	resource := r.clusterCollection.Cluster(state.ID.ValueString())
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete cluster",
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}
	if common.HasValue(state.DisableWaitingInDestroy) && state.DisableWaitingInDestroy.ValueBool() {
		tflog.Info(ctx, "Waiting for destroy to be completed, is disabled")
	} else {
		timeout := defaultTimeoutInMinutes
		if common.HasValue(state.DestroyTimeout) {
			if state.DestroyTimeout.ValueInt64() <= 0 {
				response.Diagnostics.AddWarning(nonPositiveTimeoutSummary, fmt.Sprintf(nonPositiveTimeoutFormat, state.ID.ValueString()))
			} else {
				timeout = state.DestroyTimeout.ValueInt64()
			}
		}
		isNotFound, err := r.retryClusterNotFoundWithTimeout(3, 1*time.Minute, ctx, timeout, resource)
		if err != nil {
			response.Diagnostics.AddError(
				"Can't poll cluster state",
				fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					state.ID.ValueString(), err,
				),
			)
			return
		}

		if !isNotFound {
			response.Diagnostics.AddWarning(
				"Cluster wasn't deleted yet",
				fmt.Sprintf("The cluster with identifier '%s' is not deleted yet, but the polling finisehd due to a timeout", state.ID.ValueString()),
			)
		}

	}
	// Remove the state:
	response.State.RemoveResource(ctx)
}

func (r *ClusterRosaHcpResource) ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse) {
	tflog.Debug(ctx, "begin importstate()")

	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// populateRosaHcpClusterState copies the data from the API object to the Terraform state.
func populateRosaHcpClusterState(ctx context.Context, object *cmv1.Cluster, state *ClusterRosaHcpState, httpClient common.HttpClient) error {
	state.ID = types.StringValue(object.ID())
	state.ExternalID = types.StringValue(object.ExternalID())
	object.API()
	state.Name = types.StringValue(object.Name())
	state.CloudRegion = types.StringValue(object.Region().ID())
	if props, ok := object.GetProperties(); ok {
		propertiesMap := map[string]string{}
		ocmPropertiesMap := map[string]string{}
		for k, v := range props {
			ocmPropertiesMap[k] = v
			if _, isDefault := rosa.OCMProperties[k]; !isDefault {
				propertiesMap[k] = v
			}
		}
		mapValue, err := common.ConvertStringMapToMapType(propertiesMap)
		if err != nil {
			return err
		} else {
			state.Properties = mapValue
		}
		mapValue, err = common.ConvertStringMapToMapType(ocmPropertiesMap)
		if err != nil {
			return err
		} else {
			state.OCMProperties = mapValue
		}
	}
	state.APIURL = types.StringValue(object.API().URL())
	state.ConsoleURL = types.StringValue(object.Console().URL())
	state.Domain = types.StringValue(fmt.Sprintf("%s.%s", object.Name(), object.DNS().BaseDomain()))

	if azs, ok := object.Nodes().GetAvailabilityZones(); ok {
		listValue, err := common.StringArrayToList(azs)
		if err != nil {
			return err
		}
		state.AvailabilityZones = listValue
	} else {
		state.AvailabilityZones = types.ListNull(types.StringType)
	}

	state.EtcdEncryption = types.BoolValue(object.EtcdEncryption())

	// Note: The API does not currently return account id, but we try to get it
	// anyway. Failing that, we fetch the creator ARN from the properties like
	// rosa cli does.
	awsAccountID, ok := object.AWS().GetAccountID()
	if ok {
		state.AWSAccountID = types.StringValue(awsAccountID)
	} else {
		// rosa cli gets it from the properties, so we do the same
		if creatorARN, ok := object.Properties()[ocmConsts.CreatorArn]; ok {
			if arn, err := arn.Parse(creatorARN); err == nil {
				state.AWSAccountID = types.StringValue(arn.AccountID)
			}
		}

	}

	awsPrivateLink, ok := object.AWS().GetPrivateLink()
	if ok {
		state.AWSPrivateLink = types.BoolValue(awsPrivateLink)
	} else {
		state.AWSPrivateLink = types.BoolValue(true)
	}
	kmsKeyArn, ok := object.AWS().GetKMSKeyArn()
	if ok {
		state.KMSKeyArn = types.StringValue(kmsKeyArn)
	}

	stsState, ok := object.AWS().GetSTS()
	if ok {
		if state.Sts == nil {
			state.Sts = &sts.HcpSts{}
		}
		oidcEndpointUrl := strings.TrimPrefix(stsState.OIDCEndpointURL(), "https://")

		state.Sts.OIDCEndpointURL = types.StringValue(oidcEndpointUrl)
		state.Sts.RoleARN = types.StringValue(stsState.RoleARN())
		state.Sts.SupportRoleArn = types.StringValue(stsState.SupportRoleARN())
		instanceIAMRoles := stsState.InstanceIAMRoles()
		if instanceIAMRoles != nil {
			state.Sts.InstanceIAMRoles.WorkerRoleARN = types.StringValue(instanceIAMRoles.WorkerRoleARN())
		}
		if common.IsStringAttributeUnknownOrEmpty(state.Sts.OperatorRolePrefix) {
			operatorRolePrefix, ok := stsState.GetOperatorRolePrefix()
			if ok {
				state.Sts.OperatorRolePrefix = types.StringValue(operatorRolePrefix)
			}
		}
		thumbprint, err := common.GetThumbprint(stsState.OIDCEndpointURL(), httpClient)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("cannot get thumbprint %v", err))
			state.Sts.Thumbprint = types.StringValue("")
		} else {
			state.Sts.Thumbprint = types.StringValue(thumbprint)
		}
		oidcConfig, ok := stsState.GetOidcConfig()
		if ok && oidcConfig != nil {
			state.Sts.OIDCConfigID = types.StringValue(oidcConfig.ID())
		}
	}

	subnetIds, ok := object.AWS().GetSubnetIDs()
	if ok {
		awsSubnetIds, err := common.StringArrayToList(subnetIds)
		if err != nil {
			return err
		}
		state.AWSSubnetIDs = awsSubnetIds
	}

	hasProxy := true
	hasAdditionalTrustBundle := true

	proxyObj, ok := object.GetProxy()
	if ok {
		if state.Proxy == nil {
			state.Proxy = &proxy.Proxy{}
		}
		httpProxy, ok := proxyObj.GetHTTPProxy()
		if ok {
			state.Proxy.HttpProxy = types.StringValue(httpProxy)
		}

		httpsProxy, ok := proxyObj.GetHTTPSProxy()
		if ok {
			state.Proxy.HttpsProxy = types.StringValue(httpsProxy)
		}

		noProxy, ok := proxyObj.GetNoProxy()
		if ok {
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

	machineCIDR, ok := object.Network().GetMachineCIDR()
	if ok {
		state.MachineCIDR = types.StringValue(machineCIDR)
	} else {
		state.MachineCIDR = types.StringNull()
	}
	serviceCIDR, ok := object.Network().GetServiceCIDR()
	if ok {
		state.ServiceCIDR = types.StringValue(serviceCIDR)
	} else {
		state.ServiceCIDR = types.StringNull()
	}
	podCIDR, ok := object.Network().GetPodCIDR()
	if ok {
		state.PodCIDR = types.StringValue(podCIDR)
	} else {
		state.PodCIDR = types.StringNull()
	}
	hostPrefix, ok := object.Network().GetHostPrefix()
	if ok {
		state.HostPrefix = types.Int64Value(int64(hostPrefix))
	} else {
		state.HostPrefix = types.Int64Null()
	}
	channel_group, ok := object.Version().GetChannelGroup()
	if ok {
		state.ChannelGroup = types.StringValue(channel_group)
	}

	version, ok := object.Version().GetID()
	// If we're using a non-default channel group, it will have been appended to
	// the version ID. Remove it before saving state.
	version = strings.TrimSuffix(version, fmt.Sprintf("-%s", channel_group))
	version = strings.TrimPrefix(version, "openshift-v")
	if ok {
		tflog.Debug(ctx, fmt.Sprintf("actual cluster version: %v", version))
		state.CurrentVersion = types.StringValue(version)
	} else {
		tflog.Debug(ctx, "Unknown cluster version")
		state.CurrentVersion = types.StringNull()

	}
	state.State = types.StringValue(string(object.State()))
	state.Name = types.StringValue(object.Name())
	state.CloudRegion = types.StringValue(object.Region().ID())

	return nil
}

func (r *ClusterRosaHcpResource) retryClusterNotFoundWithTimeout(attempts int, sleep time.Duration, ctx context.Context, timeout int64,
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

func (r *ClusterRosaHcpResource) waitTillClusterIsNotFoundWithTimeout(ctx context.Context, timeout int64,
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

func shouldPatchProperties(state, plan *ClusterRosaHcpState) bool {
	// User defined properties needs update
	if _, should := common.ShouldPatchMap(state.Properties, plan.Properties); should {
		return true
	}

	extractedDefaults := map[string]string{}
	for k, v := range state.OCMProperties.Elements() {
		if _, ok := state.Properties.Elements()[k]; !ok {
			extractedDefaults[k] = v.(types.String).ValueString()
		}
	}

	if len(extractedDefaults) != len(rosa.OCMProperties) {
		return true
	}

	for k, v := range rosa.OCMProperties {
		if _, ok := extractedDefaults[k]; !ok {
			return true
		} else if extractedDefaults[k] != v {
			return true
		}

	}

	return false

}
