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
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/openshift/rosa/pkg/helper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/openshift/rosa/pkg/ocm"
	"github.com/terraform-redhat/terraform-provider-ocm/build"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	semver "github.com/hashicorp/go-version"
	ver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

var kmsArnRE = regexp.MustCompile(
	`^arn:aws[\w-]*:kms:[\w-]+:\d{12}:key\/mrk-[0-9a-f]{32}$|[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

var addTerraformProviderVersionToUserAgent = request.NamedHandler{
	Name: "ocmTerraformProvider.VersionUserAgentHandler",
	Fn:   request.MakeAddToUserAgentHandler("TERRAFORM_PROVIDER_OCM", build.Version),
}

type ClusterRosaClassicResourceType struct {
}

type ClusterRosaClassicResource struct {
	clusterCollection *cmv1.ClustersClient
	versionCollection *cmv1.VersionsClient
}

func (t *ClusterRosaClassicResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "OpenShift managed cluster using rosa sts.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Unique identifier of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"external_id": {
				Description: "Unique external identifier of the cluster.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"name": {
				Description: "Name of the cluster. Must be a maximum of 15 characters in length.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"cloud_region": {
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Type:        types.StringType,
				Required:    true,
			},
			"sts": {
				Description: "STS Configuration",
				Attributes:  stsResource(),
				Optional:    true,
			},
			"multi_az": {
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"disable_workload_monitoring": {
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE) platform metrics.",
				Type:     types.BoolType,
				Optional: true,
			},
			"disable_scp_checks": {
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE) platform metrics.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"properties": {
				Description: "User defined properties.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:   true,
				Computed:   true,
				Validators: propertiesValidators(),
			},
			"ocm_properties": {
				Description: "Merged properties defined by OCM and the user defined 'properties'",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Computed: true,
			},
			"tags": {
				Description: "Apply user defined tags to all resources created in AWS.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"ccs_enabled": {
				Description: "Enables customer cloud subscription.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"etcd_encryption": {
				Description: "Encrypt etcd data.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"autoscaling_enabled": {
				Description: "Enables autoscaling.",
				Type:        types.BoolType,
				Optional:    true,
			},
			"min_replicas": {
				Description: "Min replicas.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"max_replicas": {
				Description: "Max replicas.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"api_url": {
				Description: "URL of the API server.",
				Type:        types.StringType,
				Computed:    true,
			},
			"console_url": {
				Description: "URL of the console.",
				Type:        types.StringType,
				Computed:    true,
			},
			"domain": {
				Description: "DNS Domain of Cluster",
				Type:        types.StringType,
				Computed:    true,
			},
			"replicas": {
				Description: "Number of worker nodes to provision. Single zone clusters need at least 2 nodes, " +
					"multizone clusters need at least 3 nodes.",
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
			},
			"compute_machine_type": {
				Description: "Identifier of the machine type used by the compute nodes, " +
					"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
					"source to find the possible values.",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"default_mp_labels": {
				Description: "Labels for the default machine pool. Format should be a comma-separated list of '{\"key1\"=\"value1\", \"key2\"=\"value2\"}'. " +
					"This list will overwrite any modifications made to Node labels on an ongoing basis.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"aws_account_id": {
				Description: "Identifier of the AWS account.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"aws_subnet_ids": {
				Description: "aws subnet ids",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"kms_key_arn": {
				Description: "The key ARN is the Amazon Resource Name (ARN) of a AWS KMS (Key Management Service) Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"fips": {
				Description: "Create cluster that uses FIPS Validated / Modules in Process cryptographic libraries",
				Type:        types.BoolType,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"aws_private_link": {
				Description: "Provides private connectivity between VPCs, AWS services, and your on-premises networks, without exposing your traffic to the public internet.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"availability_zones": {
				Description: "availability zones",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"machine_cidr": {
				Description: "Block of IP addresses for nodes.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"proxy": {
				Description: "proxy",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"http_proxy": {
						Description: "http proxy",
						Type:        types.StringType,
						Optional:    true,
					},
					"https_proxy": {
						Description: "https proxy",
						Type:        types.StringType,
						Optional:    true,
					},
					"no_proxy": {
						Description: "no proxy",
						Type:        types.StringType,
						Optional:    true,
					},
					"additional_trust_bundle": {
						Description: "a string contains contains a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store.",
						Type:        types.StringType,
						Optional:    true,
					},
				}),
				Optional:   true,
				Validators: proxyValidators(),
			},
			"service_cidr": {
				Description: "Block of IP addresses for services.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"pod_cidr": {
				Description: "Block of IP addresses for pods.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"host_prefix": {
				Description: "Length of the prefix of the subnet assigned to each node.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"channel_group": {
				Description: "Name of the channel group from which to select the OpenShift cluster version, for example 'stable'.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"version": {
				Description: "Identifier of the version of OpenShift, for example 'openshift-v4.1.0'.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"disable_waiting_in_destroy": {
				Description: "Disable addressing cluster state in the destroy resource. Default value is false",
				Type:        types.BoolType,
				Optional:    true,
			},
			"destroy_timeout": {
				Description: "Timeout in minutes for addressing cluster state in destroy resource. Default value is 60 minutes.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"state": {
				Description: "State of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"ec2_metadata_http_tokens": {
				Description: "Which ec2 metadata mode to use for metadata service interaction options for EC2 instances" +
					"can be optional or required, available only from 4.11.0",
				Type:     types.StringType,
				Optional: true,
				Validators: EnumValueValidator([]string{string(cmv1.Ec2MetadataHttpTokensOptional),
					string(cmv1.Ec2MetadataHttpTokensRequired)}),
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
		},
	}
	return
}

func (t *ClusterRosaClassicResourceType) NewResource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.Resource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider)

	// Get the cluster collection:
	clusterCollection := parent.connection.ClustersMgmt().V1().Clusters()

	// Get the version collection
	versionCollection := parent.connection.ClustersMgmt().V1().Versions()

	// Create the resource:
	result = &ClusterRosaClassicResource{
		clusterCollection: clusterCollection,
		versionCollection: versionCollection,
	}

	return
}

const (
	errHeadline = "Can't build cluster"
)

func createClassicClusterObject(ctx context.Context,
	state *ClusterRosaClassicState, diags diag.Diagnostics) (*cmv1.Cluster, error) {

	builder := cmv1.NewCluster()
	clusterName := state.Name.Value
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

	builder.Name(state.Name.Value)
	builder.CloudProvider(cmv1.NewCloudProvider().ID(awsCloudProvider))
	builder.Product(cmv1.NewProduct().ID(rosaProduct))
	builder.Region(cmv1.NewCloudRegion().ID(state.CloudRegion.Value))
	if !state.MultiAZ.Unknown && !state.MultiAZ.Null {
		builder.MultiAZ(state.MultiAZ.Value)
	}
	// Set default properties
	properties := make(map[string]string)
	for k, v := range OCMProperties {
		properties[k] = v
	}
	if !state.Properties.Unknown && !state.Properties.Null {
		for k, v := range state.Properties.Elems {
			properties[k] = v.(types.String).Value
		}
	}
	builder.Properties(properties)

	if !state.EtcdEncryption.Unknown && !state.EtcdEncryption.Null {
		builder.EtcdEncryption(state.EtcdEncryption.Value)
	}

	if !state.ExternalID.Unknown && !state.ExternalID.Null {
		builder.ExternalID(state.ExternalID.Value)
	}

	if !state.DisableWorkloadMonitoring.Unknown && !state.DisableWorkloadMonitoring.Null {
		builder.DisableUserWorkloadMonitoring(state.DisableWorkloadMonitoring.Value)
	}

	nodes := cmv1.NewClusterNodes()
	if !state.Replicas.Unknown && !state.Replicas.Null {
		nodes.Compute(int(state.Replicas.Value))
	}
	if !state.ComputeMachineType.Unknown && !state.ComputeMachineType.Null {
		nodes.ComputeMachineType(
			cmv1.NewMachineType().ID(state.ComputeMachineType.Value),
		)
	}

	if !state.DefaultMPLabels.Unknown && !state.DefaultMPLabels.Null {
		labels := map[string]string{}
		for k, v := range state.DefaultMPLabels.Elems {
			labels[k] = v.(types.String).Value
		}
		nodes.ComputeLabels(labels)
	}

	if !state.AvailabilityZones.Unknown && !state.AvailabilityZones.Null {
		azs := make([]string, 0)
		for _, e := range state.AvailabilityZones.Elems {
			azs = append(azs, e.(types.String).Value)
		}
		nodes.AvailabilityZones(azs...)
	}

	if !state.AutoScalingEnabled.Unknown && !state.AutoScalingEnabled.Null && state.AutoScalingEnabled.Value {
		autoscaling := cmv1.NewMachinePoolAutoscaling()
		if !state.MaxReplicas.Unknown && !state.MaxReplicas.Null {
			autoscaling.MaxReplicas(int(state.MaxReplicas.Value))
		}
		if !state.MinReplicas.Unknown && !state.MinReplicas.Null {
			autoscaling.MinReplicas(int(state.MinReplicas.Value))
		}
		if !autoscaling.Empty() {
			nodes.AutoscaleCompute(autoscaling)
		}
	}

	if !nodes.Empty() {
		builder.Nodes(nodes)
	}

	// ccs should be enabled in ocm rosa clusters
	ccs := cmv1.NewCCS()
	ccs.Enabled(true)

	if !state.DisableSCPChecks.Unknown && !state.DisableSCPChecks.Null && state.DisableSCPChecks.Value {
		ccs.DisableSCPChecks(true)
	}
	builder.CCS(ccs)

	aws := cmv1.NewAWS()

	if !state.Tags.Unknown && !state.Tags.Null {
		tags := map[string]string{}
		for k, v := range state.Tags.Elems {
			if _, ok := tags[k]; ok {
				errDescription := fmt.Sprintf("Invalid tags, user tag keys must be unique, duplicate key '%s' found", k)
				tflog.Error(ctx, errDescription)

				diags.AddError(
					errHeadline,
					errDescription,
				)
				return nil, errors.New(errHeadline + "\n" + errDescription)
			}
			tags[k] = v.(types.String).Value
		}

		aws.Tags(tags)
	}

	if !common.IsStringAttributeEmpty(state.Ec2MetadataHttpTokens) {
		// value validation was done before
		aws.Ec2MetadataHttpTokens(cmv1.Ec2MetadataHttpTokens(state.Ec2MetadataHttpTokens.Value))
	}

	if !state.KMSKeyArn.Unknown && !state.KMSKeyArn.Null && state.KMSKeyArn.Value != "" {
		kmsKeyARN := state.KMSKeyArn.Value
		if !kmsArnRE.MatchString(kmsKeyARN) {
			errDescription := fmt.Sprintf("Expected a valid value for kms-key-arn matching %s", kmsArnRE)
			tflog.Error(ctx, errDescription)

			diags.AddError(
				errHeadline,
				errDescription,
			)
			return nil, errors.New(errHeadline + "\n" + errDescription)
		}
		aws.KMSKeyArn(kmsKeyARN)
	}

	if !state.AWSAccountID.Unknown && !state.AWSAccountID.Null {
		aws.AccountID(state.AWSAccountID.Value)
	}

	if !state.AWSPrivateLink.Unknown && !state.AWSPrivateLink.Null {
		aws.PrivateLink((state.AWSPrivateLink.Value))
		api := cmv1.NewClusterAPI()
		if state.AWSPrivateLink.Value {
			api.Listening(cmv1.ListeningMethodInternal)
		}
		builder.API(api)
	}

	if !state.FIPS.Unknown && !state.FIPS.Null && state.FIPS.Value {
		builder.FIPS(true)
	}

	sts := cmv1.NewSTS()
	var err error
	if state.Sts != nil {
		sts.RoleARN(state.Sts.RoleARN.Value)
		sts.SupportRoleARN(state.Sts.SupportRoleArn.Value)
		instanceIamRoles := cmv1.NewInstanceIAMRoles()
		instanceIamRoles.MasterRoleARN(state.Sts.InstanceIAMRoles.MasterRoleARN.Value)
		instanceIamRoles.WorkerRoleARN(state.Sts.InstanceIAMRoles.WorkerRoleARN.Value)
		sts.InstanceIAMRoles(instanceIamRoles)

		// set OIDC config ID
		if !state.Sts.OIDCConfigID.Unknown && !state.Sts.OIDCConfigID.Null && state.Sts.OIDCConfigID.Value != "" {
			sts.OidcConfig(cmv1.NewOidcConfig().ID(state.Sts.OIDCConfigID.Value))
		}

		sts.OperatorRolePrefix(state.Sts.OperatorRolePrefix.Value)
		aws.STS(sts)
	}

	if !state.AWSSubnetIDs.Unknown && !state.AWSSubnetIDs.Null {
		subnetIds := make([]string, 0)
		for _, e := range state.AWSSubnetIDs.Elems {
			subnetIds = append(subnetIds, e.(types.String).Value)
		}
		aws.SubnetIDs(subnetIds...)
	}

	if !aws.Empty() {
		builder.AWS(aws)
	}
	network := cmv1.NewNetwork()
	if !state.MachineCIDR.Unknown && !state.MachineCIDR.Null {
		network.MachineCIDR(state.MachineCIDR.Value)
	}
	if !state.ServiceCIDR.Unknown && !state.ServiceCIDR.Null {
		network.ServiceCIDR(state.ServiceCIDR.Value)
	}
	if !state.PodCIDR.Unknown && !state.PodCIDR.Null {
		network.PodCIDR(state.PodCIDR.Value)
	}
	if !state.HostPrefix.Unknown && !state.HostPrefix.Null {
		network.HostPrefix(int(state.HostPrefix.Value))
	}
	if !network.Empty() {
		builder.Network(network)
	}

	channelGroup := ocm.DefaultChannelGroup
	if !state.ChannelGroup.Unknown && !state.ChannelGroup.Null {
		channelGroup = state.ChannelGroup.Value
	}

	if !state.Version.Unknown && !state.Version.Null {
		// TODO: update it to support all cluster versions
		isSupported, err := common.IsGreaterThanOrEqual(state.Version.Value, MinVersion)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error validating required cluster version %s", err))
			errDescription := fmt.Sprintf(
				"Can't check if cluster version is supported '%s': %v",
				state.Version.Value, err,
			)
			diags.AddError(
				errHeadline,
				errDescription,
			)
			return nil, errors.New(errHeadline + "\n" + errDescription)
		}
		if !isSupported {
			tflog.Error(ctx, fmt.Sprintf("Cluster version %s is not supported", state.Version.Value))
			errDescription := fmt.Sprintf(
				"Can't check if cluster version is supported '%s': %v",
				state.Version.Value, err,
			)
			diags.AddError(
				errHeadline,
				errDescription,
			)
			return nil, errors.New(errHeadline + "\n" + errDescription)
		}
		vBuilder := cmv1.NewVersion()
		versionID := state.Version.Value
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

	builder, err = buildProxy(state, builder)
	if err != nil {
		tflog.Error(ctx, "Failed to build the Proxy's attributes")
		return nil, err
	}

	object, err := builder.Build()
	return object, err
}

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
		version = strings.Replace(state.Version.Value, "openshift-v", "", 1)
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
	if common.IsStringAttributeEmpty(state.Ec2MetadataHttpTokens) {
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

func (r *ClusterRosaClassicResource) validateAccountRoles(ctx context.Context, state *ClusterRosaClassicState, version string) error {
	tflog.Debug(ctx, "Validating if cluster version is compatible to account roles' version")
	region := state.CloudRegion.Value

	tflog.Debug(ctx, fmt.Sprintf("Cluster version is %s", version))
	roleARNs := []string{
		state.Sts.RoleARN.Value,
		state.Sts.SupportRoleArn.Value,
		state.Sts.InstanceIAMRoles.MasterRoleARN.Value,
		state.Sts.InstanceIAMRoles.WorkerRoleARN.Value,
	}

	for _, ARN := range roleARNs {
		if ARN == "" {
			continue
		}
		// get role from arn
		role, err := getRoleByARN(ARN, region)
		if err != nil {
			return fmt.Errorf("Could not get Role '%s' : %v", ARN, err)
		}

		validVersion, err := r.hasCompatibleVersionTags(ctx, role.Tags, getOcmVersionMinor(version))
		if err != nil {
			return fmt.Errorf("Could not validate Role '%s' : %v", ARN, err)
		}
		if !validVersion {
			return fmt.Errorf("account role '%s' is not compatible with version %s. "+
				"Run 'rosa create account-roles' to create compatible roles and try again",
				ARN, version)
		}
	}

	return nil
}
func (r *ClusterRosaClassicResource) hasCompatibleVersionTags(ctx context.Context, iamTags []*iam.Tag, version string) (bool, error) {
	if len(iamTags) == 0 {
		return false, nil
	}
	for _, tag := range iamTags {
		if aws.StringValue(tag.Key) == tagsOpenShiftVersion {
			tflog.Debug(ctx, fmt.Sprintf("role version is %s", aws.StringValue(tag.Value)))
			if version == aws.StringValue(tag.Value) {
				return true, nil
			}
			wantedVersion, err := semver.NewVersion(version)
			if err != nil {
				return false, err
			}
			currentVersion, err := semver.NewVersion(aws.StringValue(tag.Value))
			if err != nil {
				return false, err
			}
			return currentVersion.GreaterThanOrEqual(wantedVersion), nil
		}
	}
	return false, nil
}

func getRoleByARN(roleARN, region string) (*iam.Role, error) {
	// validate arn
	parsedARN, err := arn.Parse(roleARN)
	if err != nil {
		return nil, fmt.Errorf("expected a valid IAM role ARN: %s", err)
	}
	// validate arn is for a role resource
	resource := parsedARN.Resource
	isRole := strings.Contains(resource, "role/")
	if !isRole {
		return nil, fmt.Errorf("expected ARN '%s' to be IAM role resource", roleARN)
	}

	// get resource name
	m := strings.LastIndex(resource, "/")
	roleName := resource[m+1:]

	sess, err := buildSession(region)
	if err != nil {
		return nil, err
	}
	iamClient := iam.New(sess)
	roleOutput, err := iamClient.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})

	if err != nil {
		return nil, err
	}
	return roleOutput.Role, nil
}

func buildSession(region string) (*session.Session, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           "",
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        &region,
			Retryer:                       buildCustomRetryer(),
			HTTPClient: &http.Client{
				Transport: http.DefaultTransport,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create session. Check your AWS configuration and try again")
	}

	sess.Handlers.Build.PushBackNamed(addTerraformProviderVersionToUserAgent)

	if _, err = sess.Config.Credentials.Get(); err != nil {
		return nil, fmt.Errorf("Failed to find credentials. Check your AWS configuration and try again")
	}

	return sess, nil
}

func buildCustomRetryer() client.DefaultRetryer {
	return client.DefaultRetryer{
		NumMaxRetries:    12,
		MinRetryDelay:    1 * time.Second,
		MinThrottleDelay: 5 * time.Second,
		MaxThrottleDelay: 5 * time.Second,
	}

}

func getOcmVersionMinor(ver string) string {
	rawID := strings.Replace(ver, "openshift-v", "", 1)
	version, err := semver.NewVersion(rawID)
	if err != nil {
		segments := strings.Split(rawID, ".")
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
	// Get the plan:
	state := &ClusterRosaClassicState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	summary := "Can't build cluster"

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

	err = r.validateAccountRoles(ctx, state, version)
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s', failed while validating account roles: %v",
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
	// Get the current state:
	state := &ClusterRosaClassicState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the cluster:
	get, err := r.clusterCollection.Cluster(state.ID.Value).Get().SendContext(ctx)
	if err != nil {
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

	clusterBuilder := cmv1.NewCluster()

	clusterBuilder, shouldUpdateNodes, err := updateNodes(state, plan, clusterBuilder)
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

	clusterBuilder, shouldUpdateProxy, err := updateProxy(state, plan, clusterBuilder)
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

	if !shouldUpdateProxy && !shouldUpdateNodes && !shouldPatchDisableWorkloadMonitoring {
		return
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
	// Try to retrieve the object:
	get, err := r.clusterCollection.Cluster(request.ID).Get().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				request.ID, err,
			),
		)
		return
	}
	object := get.Body()

	// Save the state:
	state := &ClusterRosaClassicState{}
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

	diags := response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
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
			if k == propertyRosaTfCommit || k == propertyRosaTfVersion {
				state.OCMProperties.Elems[k] = types.String{
					Value: v,
				}
			} else {
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
		state.AvailabilityZones.Elems = make([]attr.Value, 0)
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

	//The API does not return account id
	awsAccountID, ok := object.AWS().GetAccountID()
	if ok {
		state.AWSAccountID = types.String{
			Value: awsAccountID,
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
		oidc_endpoint_url := sts.OIDCEndpointURL()
		if strings.HasPrefix(oidc_endpoint_url, "https://") {
			oidc_endpoint_url = strings.TrimPrefix(oidc_endpoint_url, "https://")
		}

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
		if state.Sts.OperatorRolePrefix.Unknown || state.Sts.OperatorRolePrefix.Null {
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
	if ok {
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
	if ok {
		state.Version = types.String{
			Value: version,
		}
	} else {
		state.Version = types.String{
			Null: true,
		}
	}
	state.State = types.String{
		Value: string(object.State()),
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
				if state == nil {
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
						if k == propertyRosaTfVersion || k == propertyRosaTfCommit {
							errHead := "Invalid property key."
							errDesc := fmt.Sprintf("Can not override reserved properties keys. Reserved keys: '%s'/'%s'", propertyRosaTfVersion, propertyRosaTfCommit)
							resp.Diagnostics.AddError(errHead, errDesc)
							return
						}
					}
				}
			},
		},
	}
}
