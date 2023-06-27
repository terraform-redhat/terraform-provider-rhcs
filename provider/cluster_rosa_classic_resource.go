/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
***REMOVED***
***REMOVED***
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
	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/upgrade"

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
***REMOVED***

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
***REMOVED***

var OCMProperties = map[string]string{
	propertyRosaTfVersion: build.Version,
	propertyRosaTfCommit:  build.Commit,
}

var kmsArnRE = regexp.MustCompile(
	`^arn:aws[\w-]*:kms:[\w-]+:\d{12}:key\/mrk-[0-9a-f]{32}$|[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
***REMOVED***

var addTerraformProviderVersionToUserAgent = request.NamedHandler{
	Name: "ocmTerraformProvider.VersionUserAgentHandler",
	Fn:   request.MakeAddToUserAgentHandler("TERRAFORM_PROVIDER_OCM", build.Version***REMOVED***,
}

type ClusterRosaClassicResourceType struct {
}

type ClusterRosaClassicResource struct {
	clusterCollection *cmv1.ClustersClient
	versionCollection *cmv1.VersionsClient
}

func (t *ClusterRosaClassicResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "OpenShift managed cluster using rosa sts.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Unique identifier of the cluster.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					// This passes the state through to the plan, preventing
					// "known after apply" since we know it won't change.
					tfsdk.UseStateForUnknown(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"external_id": {
				Description: "Unique external identifier of the cluster.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"name": {
				Description: "Name of the cluster. Cannot exceed 15 characters in length.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"cloud_region": {
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"sts": {
				Description: "STS configuration.",
				Attributes:  stsResource(***REMOVED***,
				Optional:    true,
	***REMOVED***,
			"multi_az": {
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"disable_workload_monitoring": {
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE***REMOVED*** platform metrics.",
				Type:     types.BoolType,
				Optional: true,
	***REMOVED***,
			"disable_scp_checks": {
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE***REMOVED*** platform metrics.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"properties": {
				Description: "User defined properties.",
				Type: types.MapType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional:   true,
				Computed:   true,
				Validators: propertiesValidators(***REMOVED***,
	***REMOVED***,
			"ocm_properties": {
				Description: "Merged properties defined by OCM and the user defined 'properties'.",
				Type: types.MapType{
					ElemType: types.StringType,
		***REMOVED***,
				Computed: true,
	***REMOVED***,
			"tags": {
				Description: "Apply user defined tags to all resources created in AWS.",
				Type: types.MapType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"ccs_enabled": {
				Description: "Enables customer cloud subscription.",
				Type:        types.BoolType,
				Computed:    true,
	***REMOVED***,
			"etcd_encryption": {
				Description: "Encrypt etcd data.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"autoscaling_enabled": {
				Description: "Enables autoscaling.",
				Type:        types.BoolType,
				Optional:    true,
	***REMOVED***,
			"min_replicas": {
				Description: "Minimum replicas.",
				Type:        types.Int64Type,
				Optional:    true,
	***REMOVED***,
			"max_replicas": {
				Description: "Maximum replicas.",
				Type:        types.Int64Type,
				Optional:    true,
	***REMOVED***,
			"api_url": {
				Description: "URL of the API server.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"console_url": {
				Description: "URL of the console.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"domain": {
				Description: "DNS domain of cluster.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"replicas": {
				Description: "Number of worker nodes to provision. Single zone clusters need at least 2 nodes, " +
					"multizone clusters need at least 3 nodes.",
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
	***REMOVED***,
			"compute_machine_type": {
				Description: "Identifies the machine type used by the compute nodes, " +
					"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
					"source to find the possible values.",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"default_mp_labels": {
				Description: "This value is the default machine pool labels. Format should be a comma-separated list of '{\"key1\"=\"value1\", \"key2\"=\"value2\"}'. " +
					"This list overwrites any modifications made to Node labels on an ongoing basis. ",
				Type: types.MapType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
	***REMOVED***,
			"aws_account_id": {
				Description: "Identifier of the AWS account.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"aws_subnet_ids": {
				Description: "AWS subnet IDs.",
				Type: types.ListType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"kms_key_arn": {
				Description: "The key ARN is the Amazon Resource Name (ARN***REMOVED*** of a AWS Key Management Service (KMS***REMOVED*** Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"fips": {
				Description: "Create cluster that uses FIPS Validated / Modules in Process cryptographic libraries.",
				Type:        types.BoolType,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"aws_private_link": {
				Description: "Provides private connectivity between VPCs, AWS services, and your on-premises networks, without exposing your traffic to the public internet.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"availability_zones": {
				Description: "Availability zones.",
				Type: types.ListType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"machine_cidr": {
				Description: "Block of IP addresses for nodes.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"proxy": {
				Description: "proxy",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"http_proxy": {
						Description: "HTTP proxy.",
						Type:        types.StringType,
						Optional:    true,
			***REMOVED***,
					"https_proxy": {
						Description: "HTTPS proxy.",
						Type:        types.StringType,
						Optional:    true,
			***REMOVED***,
					"no_proxy": {
						Description: "No proxy.",
						Type:        types.StringType,
						Optional:    true,
			***REMOVED***,
					"additional_trust_bundle": {
						Description: "A string containing a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store.",
						Type:        types.StringType,
						Optional:    true,
			***REMOVED***,
		***REMOVED******REMOVED***,
				Optional:   true,
				Validators: proxyValidators(***REMOVED***,
	***REMOVED***,
			"service_cidr": {
				Description: "Block of IP addresses for services.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"pod_cidr": {
				Description: "Block of IP addresses for pods.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"host_prefix": {
				Description: "Length of the prefix of the subnet assigned to each node.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"channel_group": {
				Description: "Name of the channel group where you select the OpenShift cluster version, for example 'stable'.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"version": {
				Description: "Desired version of OpenShift for the cluster, for example '4.1.0'. If version is greater than the currently running version, an upgrade will be scheduled.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"current_version": {
				Description: "The currently running version of OpenShift on the cluster, for example '4.1.0'.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"disable_waiting_in_destroy": {
				Description: "Disable addressing cluster state in the destroy resource. Default value is false.",
				Type:        types.BoolType,
				Optional:    true,
	***REMOVED***,
			"destroy_timeout": {
				Description: "This value sets the maximum duration in minutes to allow for destroying resources. Default value is 60 minutes.",
				Type:        types.Int64Type,
				Optional:    true,
	***REMOVED***,
			"state": {
				Description: "State of the cluster.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"ec2_metadata_http_tokens": {
				Description: "This value determines which EC2 metadata mode to use for metadata service interaction " +
					"options for EC2 instances can be optional or required. This feature is available from " +
					"OpenShift version 4.11.0 and newer.",
				Type:     types.StringType,
				Optional: true,
				Validators: EnumValueValidator([]string{string(cmv1.Ec2MetadataHttpTokensOptional***REMOVED***,
					string(cmv1.Ec2MetadataHttpTokensRequired***REMOVED***}***REMOVED***,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"upgrade_acknowledgements_for": {
				Description: "Indicates acknowledgement of agreements required to upgrade the cluster version between" +
					" minor versions (e.g. a value of \"4.12\" indicates acknowledgement of any agreements required to " +
					"upgrade to OpenShift 4.12.z from 4.11 or before***REMOVED***.",
				Type:     types.StringType,
				Optional: true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *ClusterRosaClassicResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Get the cluster collection:
	clusterCollection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Get the version collection
	versionCollection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***

	// Create the resource:
	result = &ClusterRosaClassicResource{
		clusterCollection: clusterCollection,
		versionCollection: versionCollection,
	}

	return
}

const (
	errHeadline = "Can't build cluster"
***REMOVED***

func createClassicClusterObject(ctx context.Context,
	state *ClusterRosaClassicState, diags diag.Diagnostics***REMOVED*** (*cmv1.Cluster, error***REMOVED*** {

	builder := cmv1.NewCluster(***REMOVED***
	clusterName := state.Name.Value
	if len(clusterName***REMOVED*** > maxClusterNameLength {
		errDescription := fmt.Sprintf("Expected a valid value for 'name' maximum of 15 characters in length. Provided Cluster name '%s' is of length '%d'",
			clusterName, len(clusterName***REMOVED***,
		***REMOVED***
		tflog.Error(ctx, errDescription***REMOVED***

		diags.AddError(
			errHeadline,
			errDescription,
		***REMOVED***
		return nil, errors.New(errHeadline + "\n" + errDescription***REMOVED***
	}

	builder.Name(state.Name.Value***REMOVED***
	builder.CloudProvider(cmv1.NewCloudProvider(***REMOVED***.ID(awsCloudProvider***REMOVED******REMOVED***
	builder.Product(cmv1.NewProduct(***REMOVED***.ID(rosaProduct***REMOVED******REMOVED***
	builder.Region(cmv1.NewCloudRegion(***REMOVED***.ID(state.CloudRegion.Value***REMOVED******REMOVED***
	if !state.MultiAZ.Unknown && !state.MultiAZ.Null {
		builder.MultiAZ(state.MultiAZ.Value***REMOVED***
	}
	// Set default properties
	properties := make(map[string]string***REMOVED***
	for k, v := range OCMProperties {
		properties[k] = v
	}
	if !state.Properties.Unknown && !state.Properties.Null {
		for k, v := range state.Properties.Elems {
			properties[k] = v.(types.String***REMOVED***.Value
***REMOVED***
	}
	builder.Properties(properties***REMOVED***

	if !state.EtcdEncryption.Unknown && !state.EtcdEncryption.Null {
		builder.EtcdEncryption(state.EtcdEncryption.Value***REMOVED***
	}

	if !state.ExternalID.Unknown && !state.ExternalID.Null {
		builder.ExternalID(state.ExternalID.Value***REMOVED***
	}

	if !state.DisableWorkloadMonitoring.Unknown && !state.DisableWorkloadMonitoring.Null {
		builder.DisableUserWorkloadMonitoring(state.DisableWorkloadMonitoring.Value***REMOVED***
	}

	nodes := cmv1.NewClusterNodes(***REMOVED***
	if !state.Replicas.Unknown && !state.Replicas.Null {
		nodes.Compute(int(state.Replicas.Value***REMOVED******REMOVED***
	}
	if !state.ComputeMachineType.Unknown && !state.ComputeMachineType.Null {
		nodes.ComputeMachineType(
			cmv1.NewMachineType(***REMOVED***.ID(state.ComputeMachineType.Value***REMOVED***,
		***REMOVED***
	}

	if !state.DefaultMPLabels.Unknown && !state.DefaultMPLabels.Null {
		labels := map[string]string{}
		for k, v := range state.DefaultMPLabels.Elems {
			labels[k] = v.(types.String***REMOVED***.Value
***REMOVED***
		nodes.ComputeLabels(labels***REMOVED***
	}

	if !state.AvailabilityZones.Unknown && !state.AvailabilityZones.Null {
		azs := make([]string, 0***REMOVED***
		for _, e := range state.AvailabilityZones.Elems {
			azs = append(azs, e.(types.String***REMOVED***.Value***REMOVED***
***REMOVED***
		nodes.AvailabilityZones(azs...***REMOVED***
	}

	if !state.AutoScalingEnabled.Unknown && !state.AutoScalingEnabled.Null && state.AutoScalingEnabled.Value {
		autoscaling := cmv1.NewMachinePoolAutoscaling(***REMOVED***
		if !state.MaxReplicas.Unknown && !state.MaxReplicas.Null {
			autoscaling.MaxReplicas(int(state.MaxReplicas.Value***REMOVED******REMOVED***
***REMOVED***
		if !state.MinReplicas.Unknown && !state.MinReplicas.Null {
			autoscaling.MinReplicas(int(state.MinReplicas.Value***REMOVED******REMOVED***
***REMOVED***
		if !autoscaling.Empty(***REMOVED*** {
			nodes.AutoscaleCompute(autoscaling***REMOVED***
***REMOVED***
	}

	if !nodes.Empty(***REMOVED*** {
		builder.Nodes(nodes***REMOVED***
	}

	// ccs should be enabled in ocm rosa clusters
	ccs := cmv1.NewCCS(***REMOVED***
	ccs.Enabled(true***REMOVED***

	if !state.DisableSCPChecks.Unknown && !state.DisableSCPChecks.Null && state.DisableSCPChecks.Value {
		ccs.DisableSCPChecks(true***REMOVED***
	}
	builder.CCS(ccs***REMOVED***

	aws := cmv1.NewAWS(***REMOVED***

	if !state.Tags.Unknown && !state.Tags.Null {
		tags := map[string]string{}
		for k, v := range state.Tags.Elems {
			if _, ok := tags[k]; ok {
				errDescription := fmt.Sprintf("Invalid tags, user tag keys must be unique, duplicate key '%s' found", k***REMOVED***
				tflog.Error(ctx, errDescription***REMOVED***

				diags.AddError(
					errHeadline,
					errDescription,
				***REMOVED***
				return nil, errors.New(errHeadline + "\n" + errDescription***REMOVED***
	***REMOVED***
			tags[k] = v.(types.String***REMOVED***.Value
***REMOVED***

		aws.Tags(tags***REMOVED***
	}

	if !common.IsStringAttributeEmpty(state.Ec2MetadataHttpTokens***REMOVED*** {
		// value validation was done before
		aws.Ec2MetadataHttpTokens(cmv1.Ec2MetadataHttpTokens(state.Ec2MetadataHttpTokens.Value***REMOVED******REMOVED***
	}

	if !state.KMSKeyArn.Unknown && !state.KMSKeyArn.Null && state.KMSKeyArn.Value != "" {
		kmsKeyARN := state.KMSKeyArn.Value
		if !kmsArnRE.MatchString(kmsKeyARN***REMOVED*** {
			errDescription := fmt.Sprintf("Expected a valid value for kms-key-arn matching %s", kmsArnRE***REMOVED***
			tflog.Error(ctx, errDescription***REMOVED***

			diags.AddError(
				errHeadline,
				errDescription,
			***REMOVED***
			return nil, errors.New(errHeadline + "\n" + errDescription***REMOVED***
***REMOVED***
		aws.KMSKeyArn(kmsKeyARN***REMOVED***
	}

	if !state.AWSAccountID.Unknown && !state.AWSAccountID.Null {
		aws.AccountID(state.AWSAccountID.Value***REMOVED***
	}

	if !state.AWSPrivateLink.Unknown && !state.AWSPrivateLink.Null {
		aws.PrivateLink((state.AWSPrivateLink.Value***REMOVED******REMOVED***
		api := cmv1.NewClusterAPI(***REMOVED***
		if state.AWSPrivateLink.Value {
			api.Listening(cmv1.ListeningMethodInternal***REMOVED***
***REMOVED***
		builder.API(api***REMOVED***
	}

	if !state.FIPS.Unknown && !state.FIPS.Null && state.FIPS.Value {
		builder.FIPS(true***REMOVED***
	}

	sts := cmv1.NewSTS(***REMOVED***
	var err error
	if state.Sts != nil {
		sts.RoleARN(state.Sts.RoleARN.Value***REMOVED***
		sts.SupportRoleARN(state.Sts.SupportRoleArn.Value***REMOVED***
		instanceIamRoles := cmv1.NewInstanceIAMRoles(***REMOVED***
		instanceIamRoles.MasterRoleARN(state.Sts.InstanceIAMRoles.MasterRoleARN.Value***REMOVED***
		instanceIamRoles.WorkerRoleARN(state.Sts.InstanceIAMRoles.WorkerRoleARN.Value***REMOVED***
		sts.InstanceIAMRoles(instanceIamRoles***REMOVED***

		// set OIDC config ID
		if !state.Sts.OIDCConfigID.Unknown && !state.Sts.OIDCConfigID.Null && state.Sts.OIDCConfigID.Value != "" {
			sts.OidcConfig(cmv1.NewOidcConfig(***REMOVED***.ID(state.Sts.OIDCConfigID.Value***REMOVED******REMOVED***
***REMOVED***

		sts.OperatorRolePrefix(state.Sts.OperatorRolePrefix.Value***REMOVED***
		aws.STS(sts***REMOVED***
	}

	if !state.AWSSubnetIDs.Unknown && !state.AWSSubnetIDs.Null {
		subnetIds := make([]string, 0***REMOVED***
		for _, e := range state.AWSSubnetIDs.Elems {
			subnetIds = append(subnetIds, e.(types.String***REMOVED***.Value***REMOVED***
***REMOVED***
		aws.SubnetIDs(subnetIds...***REMOVED***
	}

	if !aws.Empty(***REMOVED*** {
		builder.AWS(aws***REMOVED***
	}
	network := cmv1.NewNetwork(***REMOVED***
	if !state.MachineCIDR.Unknown && !state.MachineCIDR.Null {
		network.MachineCIDR(state.MachineCIDR.Value***REMOVED***
	}
	if !state.ServiceCIDR.Unknown && !state.ServiceCIDR.Null {
		network.ServiceCIDR(state.ServiceCIDR.Value***REMOVED***
	}
	if !state.PodCIDR.Unknown && !state.PodCIDR.Null {
		network.PodCIDR(state.PodCIDR.Value***REMOVED***
	}
	if !state.HostPrefix.Unknown && !state.HostPrefix.Null {
		network.HostPrefix(int(state.HostPrefix.Value***REMOVED******REMOVED***
	}
	if !network.Empty(***REMOVED*** {
		builder.Network(network***REMOVED***
	}

	channelGroup := ocm.DefaultChannelGroup
	if !state.ChannelGroup.Unknown && !state.ChannelGroup.Null {
		channelGroup = state.ChannelGroup.Value
	}

	if !state.Version.Unknown && !state.Version.Null {
		// TODO: update it to support all cluster versions
		isSupported, err := common.IsGreaterThanOrEqual(state.Version.Value, MinVersion***REMOVED***
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error validating required cluster version %s", err***REMOVED******REMOVED***
			errDescription := fmt.Sprintf(
				"Can't check if cluster version is supported '%s': %v",
				state.Version.Value, err,
			***REMOVED***
			diags.AddError(
				errHeadline,
				errDescription,
			***REMOVED***
			return nil, errors.New(errHeadline + "\n" + errDescription***REMOVED***
***REMOVED***
		if !isSupported {
			description := fmt.Sprintf("Cluster version %s is not supported (minimal supported version is %s***REMOVED***", state.Version.Value, MinVersion***REMOVED***
			tflog.Error(ctx, description***REMOVED***
			diags.AddError(
				errHeadline,
				description,
			***REMOVED***
			return nil, errors.New(errHeadline + "\n" + description***REMOVED***
***REMOVED***
		vBuilder := cmv1.NewVersion(***REMOVED***
		versionID := fmt.Sprintf("openshift-v%s", state.Version.Value***REMOVED***
		// When using a channel group other than the default, the channel name
		// must be appended to the version ID or the API server will return an
		// error stating unexpected channel group.
		if channelGroup != ocm.DefaultChannelGroup {
			versionID = versionID + "-" + channelGroup
***REMOVED***
		vBuilder.ID(versionID***REMOVED***
		vBuilder.ChannelGroup(channelGroup***REMOVED***
		builder.Version(vBuilder***REMOVED***
	}

	builder, err = buildProxy(state, builder***REMOVED***
	if err != nil {
		tflog.Error(ctx, "Failed to build the Proxy's attributes"***REMOVED***
		return nil, err
	}

	object, err := builder.Build(***REMOVED***
	return object, err
}

func buildProxy(state *ClusterRosaClassicState, builder *cmv1.ClusterBuilder***REMOVED*** (*cmv1.ClusterBuilder, error***REMOVED*** {
	proxy := cmv1.NewProxy(***REMOVED***
	if state.Proxy != nil {
		httpsProxy := ""
		httpProxy := ""
		additionalTrustBundle := ""

		if !common.IsStringAttributeEmpty(state.Proxy.HttpProxy***REMOVED*** {
			httpProxy = state.Proxy.HttpProxy.Value
			proxy.HTTPProxy(httpProxy***REMOVED***
***REMOVED***
		if !common.IsStringAttributeEmpty(state.Proxy.HttpsProxy***REMOVED*** {
			httpsProxy = state.Proxy.HttpsProxy.Value
			proxy.HTTPSProxy(httpsProxy***REMOVED***
***REMOVED***
		if !common.IsStringAttributeEmpty(state.Proxy.NoProxy***REMOVED*** {
			proxy.NoProxy(state.Proxy.NoProxy.Value***REMOVED***
***REMOVED***

		if !common.IsStringAttributeEmpty(state.Proxy.AdditionalTrustBundle***REMOVED*** {
			additionalTrustBundle = state.Proxy.AdditionalTrustBundle.Value
			builder.AdditionalTrustBundle(additionalTrustBundle***REMOVED***
***REMOVED***

		builder.Proxy(proxy***REMOVED***
	}

	return builder, nil
}

// getAndValidateVersionInChannelGroup ensures that the cluster version is
// available in the channel group
func (r *ClusterRosaClassicResource***REMOVED*** getAndValidateVersionInChannelGroup(ctx context.Context, state *ClusterRosaClassicState***REMOVED*** (string, error***REMOVED*** {
	channelGroup := ocm.DefaultChannelGroup
	if !state.ChannelGroup.Unknown && !state.ChannelGroup.Null {
		channelGroup = state.ChannelGroup.Value
	}

	versionList, err := r.getVersionList(ctx, channelGroup***REMOVED***
	if err != nil {
		return "", err
	}

	version := versionList[0]
	if !state.Version.Unknown && !state.Version.Null {
		version = state.Version.Value
	}

	tflog.Debug(ctx, fmt.Sprintf("Validating if cluster version %s is in the list of supported versions: %v", version, versionList***REMOVED******REMOVED***
	for _, v := range versionList {
		if v == version {
			return version, nil
***REMOVED***
	}

	return "", fmt.Errorf("version %s is not in the list of supported versions: %v", version, versionList***REMOVED***
}

func validateHttpTokensVersion(ctx context.Context, state *ClusterRosaClassicState, version string***REMOVED*** error {
	if common.IsStringAttributeEmpty(state.Ec2MetadataHttpTokens***REMOVED*** {
		return nil
	}

	greater, err := common.IsGreaterThanOrEqual(version, lowestHttpTokensVer***REMOVED***
	if err != nil {
		return fmt.Errorf("version '%s' is not supported: %v", version, err***REMOVED***
	}
	if !greater {
		msg := fmt.Sprintf("version '%s' is not supported with ec2_metadata_http_tokens, "+
			"minimum supported version is %s", version, lowestHttpTokensVer***REMOVED***
		tflog.Error(ctx, msg***REMOVED***
		return fmt.Errorf(msg***REMOVED***
	}
	return nil
}

func (r *ClusterRosaClassicResource***REMOVED*** validateAccountRoles(ctx context.Context, state *ClusterRosaClassicState, version string***REMOVED*** error {
	tflog.Debug(ctx, "Validating if cluster version is compatible to account roles' version"***REMOVED***
	region := state.CloudRegion.Value

	tflog.Debug(ctx, fmt.Sprintf("Cluster version is %s", version***REMOVED******REMOVED***
	roleARNs := []string{
		state.Sts.RoleARN.Value,
		state.Sts.SupportRoleArn.Value,
		state.Sts.InstanceIAMRoles.MasterRoleARN.Value,
		state.Sts.InstanceIAMRoles.WorkerRoleARN.Value,
	}

	for _, ARN := range roleARNs {
		if ARN == "" {
			continue
***REMOVED***
		// get role from arn
		role, err := getRoleByARN(ARN, region***REMOVED***
		if err != nil {
			return fmt.Errorf("Could not get Role '%s' : %v", ARN, err***REMOVED***
***REMOVED***

		validVersion, err := r.hasCompatibleVersionTags(ctx, role.Tags, getOcmVersionMinor(version***REMOVED******REMOVED***
		if err != nil {
			return fmt.Errorf("Could not validate Role '%s' : %v", ARN, err***REMOVED***
***REMOVED***
		if !validVersion {
			return fmt.Errorf("account role '%s' is not compatible with version %s. "+
				"Run 'rosa create account-roles' to create compatible roles and try again",
				ARN, version***REMOVED***
***REMOVED***
	}

	return nil
}

// validateOperatorRolePolicies ensures that the operator role policies are
// compatible with the requested cluster version
func (r *ClusterRosaClassicResource***REMOVED*** validateOperatorRolePolicies(ctx context.Context, state *ClusterRosaClassicState, version string***REMOVED*** error {
	tflog.Debug(ctx, "Validating if cluster version is compatible with the operator role policies"***REMOVED***

	operRoles := []*cmv1.OperatorIAMRole{}
	operRoleClient := r.clusterCollection.Cluster(state.ID.Value***REMOVED***.STSOperatorRoles(***REMOVED***
	page := 1
	size := 100
	for {
		resp, err := operRoleClient.List(***REMOVED***.Page(page***REMOVED***.Size(size***REMOVED***.SendContext(ctx***REMOVED***
		if err != nil {
			return fmt.Errorf("Could not list operator roles: %v", err***REMOVED***
***REMOVED***
		operRoles = append(operRoles, resp.Items(***REMOVED***.Slice(***REMOVED***...***REMOVED***
		if resp.Size(***REMOVED*** < size {
			break
***REMOVED***
		page++
	}

	region := state.CloudRegion.Value
	var session *session.Session
	var iamClient *iam.IAM
	for _, operRole := range operRoles {
		roleARN := operRole.RoleARN(***REMOVED***
		if roleARN == "" {
			continue
***REMOVED***
		if session == nil {
			var err error
			session, err = buildSession(region***REMOVED***
			if err != nil {
				return fmt.Errorf("Could not build session: %v", err***REMOVED***
	***REMOVED***
***REMOVED***
		if iamClient == nil {
			iamClient = iam.New(session***REMOVED***
***REMOVED***
		role, err := getRoleByARN(roleARN, state.CloudRegion.Value***REMOVED***
		if err != nil {
			return fmt.Errorf("Could not get Role '%s' : %v", roleARN, err***REMOVED***
***REMOVED***
		attachedPolicies, err := iamClient.ListAttachedRolePoliciesWithContext(ctx, &iam.ListAttachedRolePoliciesInput{
			MaxItems: aws.Int64(100***REMOVED***,
			RoleName: role.RoleName,
***REMOVED******REMOVED***
		if err != nil {
			return fmt.Errorf("Could not list attached policies for role '%s' : %v", roleARN, err***REMOVED***
***REMOVED***
		for _, policy := range attachedPolicies.AttachedPolicies {
			policyARN := policy.PolicyArn
			policyOut, err := iamClient.GetPolicyWithContext(ctx, &iam.GetPolicyInput{
				PolicyArn: policyARN,
	***REMOVED******REMOVED***
			if err != nil {
				return fmt.Errorf("Could not get policy '%s' : %v", aws.StringValue(policyARN***REMOVED***, err***REMOVED***
	***REMOVED***
			tags := policyOut.Policy.Tags
			validVersion, err := r.hasCompatibleVersionTags(ctx, tags, getOcmVersionMinor(version***REMOVED******REMOVED***
			if err != nil {
				return fmt.Errorf("Could not validate policy '%s' : %v", aws.StringValue(policyARN***REMOVED***, err***REMOVED***
	***REMOVED***
			if !validVersion {
				return fmt.Errorf("operator role policy '%s' is not compatible with version %s. "+
					"Upgrade operator roles and try again",
					aws.StringValue(policyARN***REMOVED***, version***REMOVED***
	***REMOVED***
***REMOVED***
	}
	return nil
}

// Check whether the list of tags contains a tag indicating the version of
// OpenShift it was creted for, and whether that version is at lest as new as
// the provided version.
func (r *ClusterRosaClassicResource***REMOVED*** hasCompatibleVersionTags(ctx context.Context, iamTags []*iam.Tag, version string***REMOVED*** (bool, error***REMOVED*** {
	if len(iamTags***REMOVED*** == 0 {
		return false, nil
	}
	for _, tag := range iamTags {
		if aws.StringValue(tag.Key***REMOVED*** == tagsOpenShiftVersion {
			tflog.Debug(ctx, fmt.Sprintf("tag version is %s", aws.StringValue(tag.Value***REMOVED******REMOVED******REMOVED***
			if version == aws.StringValue(tag.Value***REMOVED*** {
				return true, nil
	***REMOVED***
			wantedVersion, err := semver.NewVersion(version***REMOVED***
			if err != nil {
				return false, err
	***REMOVED***
			currentVersion, err := semver.NewVersion(aws.StringValue(tag.Value***REMOVED******REMOVED***
			if err != nil {
				return false, err
	***REMOVED***
			return currentVersion.GreaterThanOrEqual(wantedVersion***REMOVED***, nil
***REMOVED***
	}
	return false, nil
}

func getRoleByARN(roleARN, region string***REMOVED*** (*iam.Role, error***REMOVED*** {
	// validate arn
	parsedARN, err := arn.Parse(roleARN***REMOVED***
	if err != nil {
		return nil, fmt.Errorf("expected a valid IAM role ARN: %s", err***REMOVED***
	}
	// validate arn is for a role resource
	resource := parsedARN.Resource
	isRole := strings.Contains(resource, "role/"***REMOVED***
	if !isRole {
		return nil, fmt.Errorf("expected ARN '%s' to be IAM role resource", roleARN***REMOVED***
	}

	// get resource name
	m := strings.LastIndex(resource, "/"***REMOVED***
	roleName := resource[m+1:]

	sess, err := buildSession(region***REMOVED***
	if err != nil {
		return nil, err
	}
	iamClient := iam.New(sess***REMOVED***
	roleOutput, err := iamClient.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(roleName***REMOVED***,
	}***REMOVED***

	if err != nil {
		return nil, err
	}
	return roleOutput.Role, nil
}

func buildSession(region string***REMOVED*** (*session.Session, error***REMOVED*** {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           "",
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true***REMOVED***,
			Region:                        &region,
			Retryer:                       buildCustomRetryer(***REMOVED***,
			HTTPClient: &http.Client{
				Transport: http.DefaultTransport,
	***REMOVED***,
***REMOVED***,
	}***REMOVED***
	if err != nil {
		return nil, fmt.Errorf("Failed to create session. Check your AWS configuration and try again"***REMOVED***
	}

	sess.Handlers.Build.PushBackNamed(addTerraformProviderVersionToUserAgent***REMOVED***

	if _, err = sess.Config.Credentials.Get(***REMOVED***; err != nil {
		return nil, fmt.Errorf("Failed to find credentials. Check your AWS configuration and try again"***REMOVED***
	}

	return sess, nil
}

func buildCustomRetryer(***REMOVED*** client.DefaultRetryer {
	return client.DefaultRetryer{
		NumMaxRetries:    12,
		MinRetryDelay:    1 * time.Second,
		MinThrottleDelay: 5 * time.Second,
		MaxThrottleDelay: 5 * time.Second,
	}

}

func getOcmVersionMinor(ver string***REMOVED*** string {
	version, err := semver.NewVersion(ver***REMOVED***
	if err != nil {
		segments := strings.Split(ver, "."***REMOVED***
		return fmt.Sprintf("%s.%s", segments[0], segments[1]***REMOVED***
	}
	segments := version.Segments(***REMOVED***
	return fmt.Sprintf("%d.%d", segments[0], segments[1]***REMOVED***
}

// getVersionList returns a list of versions for the given channel group, sorted by
// descending semver
func (r *ClusterRosaClassicResource***REMOVED*** getVersionList(ctx context.Context, channelGroup string***REMOVED*** (versionList []string, err error***REMOVED*** {
	vs, err := r.getVersions(ctx, channelGroup***REMOVED***
	if err != nil {
		err = fmt.Errorf("Failed to retrieve versions: %s", err***REMOVED***
		return
	}

	for _, v := range vs {
		versionList = append(versionList, v.RawID(***REMOVED******REMOVED***
	}

	if len(versionList***REMOVED*** == 0 {
		err = fmt.Errorf("Could not find versions"***REMOVED***
		return
	}

	return
}
func (r *ClusterRosaClassicResource***REMOVED*** getVersions(ctx context.Context, channelGroup string***REMOVED*** (versions []*cmv1.Version, err error***REMOVED*** {
	page := 1
	size := 100
	filter := strings.Join([]string{
		"enabled = 'true'",
		"rosa_enabled = 'true'",
		fmt.Sprintf("channel_group = '%s'", channelGroup***REMOVED***,
	}, " AND "***REMOVED***
	for {
		var response *cmv1.VersionsListResponse
		response, err = r.versionCollection.List(***REMOVED***.
			Search(filter***REMOVED***.
			Order("default desc, id desc"***REMOVED***.
			Page(page***REMOVED***.
			Size(size***REMOVED***.
			Send(***REMOVED***
		if err != nil {
			tflog.Debug(ctx, err.Error(***REMOVED******REMOVED***
			return nil, err
***REMOVED***
		versions = append(versions, response.Items(***REMOVED***.Slice(***REMOVED***...***REMOVED***
		if response.Size(***REMOVED*** < size {
			break
***REMOVED***
		page++
	}

	// Sort list in descending order
	sort.Slice(versions, func(i, j int***REMOVED*** bool {
		a, erra := ver.NewVersion(versions[i].RawID(***REMOVED******REMOVED***
		b, errb := ver.NewVersion(versions[j].RawID(***REMOVED******REMOVED***
		if erra != nil || errb != nil {
			return false
***REMOVED***
		return a.GreaterThan(b***REMOVED***
	}***REMOVED***

	return
}

func (r *ClusterRosaClassicResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	tflog.Debug(ctx, "begin create(***REMOVED***"***REMOVED***
	// Get the plan:
	state := &ClusterRosaClassicState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}
	summary := "Can't build cluster"

	// In case version with "openshift-v" prefix was used here,
	// Give a meaningful message to inform the user that it not supported any more
	if !state.Version.Unknown && !state.Version.Null && strings.HasPrefix(state.Version.Value, "openshift-v"***REMOVED*** {
		response.Diagnostics.AddError(
			summary,
			"Openshift version must be provided without the \"openshift-v\" prefix",
		***REMOVED***
		return
	}

	version, err := r.getAndValidateVersionInChannelGroup(ctx, state***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	err = r.validateAccountRoles(ctx, state, version***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s', failed while validating account roles: %v",
				state.Name.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	err = validateHttpTokensVersion(ctx, state, version***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object, err := createClassicClusterObject(ctx, state, diags***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	add, err := r.clusterCollection.Add(***REMOVED***.Body(object***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"Can't create cluster with name '%s': %v",
				state.Name.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	// Save the state:
	err = populateRosaClassicClusterState(ctx, object, state, DefaultHttpClient{}***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			***REMOVED***,
		***REMOVED***
		return
	}
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterRosaClassicResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	// Get the current state:
	state := &ClusterRosaClassicState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the cluster:
	get, err := r.clusterCollection.Cluster(state.ID.Value***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Save the state:
	err = populateRosaClassicClusterState(ctx, object, state, DefaultHttpClient{}***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			***REMOVED***,
		***REMOVED***
		return
	}
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterRosaClassicResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
	var diags diag.Diagnostics

	tflog.Debug(ctx, "begin update(***REMOVED***"***REMOVED***

	// Get the state:
	state := &ClusterRosaClassicState{}
	diags = request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Get the plan:
	plan := &ClusterRosaClassicState{}
	diags = request.Plan.Get(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Schedule a cluster upgrade if a newer version is requested
	if err := r.upgradeClusterIfNeeded(ctx, state, plan***REMOVED***; err != nil {
		response.Diagnostics.AddError(
			"Can't upgrade cluster",
			fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, %v", state.ID.Value, err***REMOVED***,
		***REMOVED***
		return
	}

	clusterBuilder := cmv1.NewCluster(***REMOVED***

	clusterBuilder, _, err := updateNodes(state, plan, clusterBuilder***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster nodes for cluster with identifier: `%s`, %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	clusterBuilder, _, err = updateProxy(state, plan, clusterBuilder***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update proxy's configuration for cluster with identifier: `%s`, %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	_, shouldPatchDisableWorkloadMonitoring := common.ShouldPatchBool(state.DisableWorkloadMonitoring, plan.DisableWorkloadMonitoring***REMOVED***
	if shouldPatchDisableWorkloadMonitoring {
		clusterBuilder.DisableUserWorkloadMonitoring(plan.DisableWorkloadMonitoring.Value***REMOVED***
	}

	clusterSpec, err := clusterBuilder.Build(***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster patch",
			fmt.Sprintf(
				"Can't build patch for cluster with identifier '%s': %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	update, err := r.clusterCollection.Cluster(state.ID.Value***REMOVED***.Update(***REMOVED***.
		Body(clusterSpec***REMOVED***.
		SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster with identifier '%s': %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// update the autoscaling enabled with the plan value (important for nil and false cases***REMOVED***
	state.AutoScalingEnabled = plan.AutoScalingEnabled
	// update the Replicas with the plan value (important for nil and zero value cases***REMOVED***
	state.Replicas = plan.Replicas

	object := update.Body(***REMOVED***

	// Update the state:
	err = populateRosaClassicClusterState(ctx, object, plan, DefaultHttpClient{}***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			***REMOVED***,
		***REMOVED***
		return
	}
	diags = response.State.Set(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

// Upgrades the cluster if the desired (plan***REMOVED*** version is greater than the
// current version
func (r *ClusterRosaClassicResource***REMOVED*** upgradeClusterIfNeeded(ctx context.Context, state, plan *ClusterRosaClassicState***REMOVED*** error {
	if common.IsStringAttributeEmpty(plan.Version***REMOVED*** || common.IsStringAttributeEmpty(state.CurrentVersion***REMOVED*** {
		// No version information, nothing to do
		tflog.Debug(ctx, "Insufficient cluster version information to determine if upgrade should be performed."***REMOVED***
		return nil
	}

	// Check the versions to see if we need to upgrade
	currentVersion, err := semver.NewVersion(state.CurrentVersion.Value***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to parse current cluster version: %v", err***REMOVED***
	}
	// For backward compatibility
	// In case version format with "openshift-v" was already used
	// remove the prefix to adapt the right format and avoid failure
	fixedVersion := strings.TrimPrefix(plan.Version.Value, "openshift-v"***REMOVED***
	desiredVersion, err := semver.NewVersion(fixedVersion***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to parse desired cluster version: %v", err***REMOVED***
	}
	if currentVersion.GreaterThanOrEqual(desiredVersion***REMOVED*** {
		tflog.Debug(ctx, "No cluster version upgrade needed."***REMOVED***
		return nil
	}

	// Make sure the desired version is available
	versionId := fmt.Sprintf("openshift-v%s", state.CurrentVersion.Value***REMOVED***
	if !state.ChannelGroup.Unknown && !state.ChannelGroup.Null && state.ChannelGroup.Value != ocm.DefaultChannelGroup {
		versionId += "-" + state.ChannelGroup.Value
	}
	availableVersions, err := upgrade.GetAvailableUpgradeVersions(ctx, r.versionCollection, versionId***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to get available upgrades: %v", err***REMOVED***
	}
	found := false
	for _, v := range availableVersions {
		sem, err := semver.NewVersion(v.RawID(***REMOVED******REMOVED***
		if err != nil {
			return fmt.Errorf("failed to parse available upgrade version: %v", err***REMOVED***
***REMOVED***
		if desiredVersion.Equal(sem***REMOVED*** {
			found = true
			break
***REMOVED***
	}
	if !found {
		avail := []string{}
		for _, v := range availableVersions {
			avail = append(avail, v.RawID(***REMOVED******REMOVED***
***REMOVED***
		return fmt.Errorf("desired version (%s***REMOVED*** is not in the list of available upgrades (%v***REMOVED***", desiredVersion, avail***REMOVED***
	}

	// Make sure the account roles have been upgraded
	if err := r.validateAccountRoles(ctx, plan, desiredVersion.String(***REMOVED******REMOVED***; err != nil {
		return fmt.Errorf("failed to validate account roles: %v", err***REMOVED***
	}

	// Make sure the operator role policies have been upgraded
	if err := r.validateOperatorRolePolicies(ctx, plan, desiredVersion.String(***REMOVED******REMOVED***; err != nil {
		return fmt.Errorf("failed to validate operator role policies: %v", err***REMOVED***
	}

	// Fetch existing upgrade policies
	upgrades, err := upgrade.GetScheduledUpgrades(ctx, r.clusterCollection, state.ID.Value***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to get upgrade policies: %v", err***REMOVED***
	}

	// Stop if an upgrade is already in progress
	correctUpgradePending, err := upgrade.CheckAndCancelUpgrades(ctx, r.clusterCollection, upgrades, desiredVersion***REMOVED***
	if err != nil {
		return err
	}

	if !correctUpgradePending {
		// Gate agreements are checked when the upgrade is scheduled, resulting
		// in an error return. ROSA cli does this by scheduling once w/ dryRun
		// to look for un-acked agreements.
		clusterClient := r.clusterCollection.Cluster(state.ID.Value***REMOVED***
		upgradePoliciesClient := clusterClient.UpgradePolicies(***REMOVED***
		gates, description, err := upgrade.CheckMissingAgreements(desiredVersion.String(***REMOVED***, state.ID.Value, upgradePoliciesClient***REMOVED***
		if err != nil {
			return fmt.Errorf("failed to check for missing upgrade agreements: %v", err***REMOVED***
***REMOVED***
		// User ack is required if we have any non-STS-only gates
		userAckRequired := false
		for _, gate := range gates {
			if !gate.STSOnly(***REMOVED*** {
				userAckRequired = true
	***REMOVED***
***REMOVED***
		targetMinorVersion := getOcmVersionMinor(desiredVersion.String(***REMOVED******REMOVED***
		userAcksGates := !plan.UpgradeAcksFor.Unknown && !plan.UpgradeAcksFor.Null && plan.UpgradeAcksFor.Value == targetMinorVersion
		if userAckRequired && !userAcksGates { // User has not acknowledged mandatory gates, stop here.
			return fmt.Errorf("%s\nTo acknowledge these items, please add \"upgrade_acknowledgements_for = %s\""+
				" and re-apply the changes", description, targetMinorVersion***REMOVED***
***REMOVED***

		// Ack all gates to OCM
		for _, gate := range gates {
			gateID := gate.ID(***REMOVED***
			tflog.Debug(ctx, "Acknowledging version gate", "gateID", gateID***REMOVED***
			gateAgreementsClient := clusterClient.GateAgreements(***REMOVED***
			err := upgrade.AckVersionGate(gateAgreementsClient, gateID***REMOVED***
			if err != nil {
				return fmt.Errorf("failed to acknowledge version gate '%s' for cluster '%s': %v",
					gateID, state.ID.Value, err***REMOVED***
	***REMOVED***
***REMOVED***

		// Schedule an upgrade
		tenMinFromNow := time.Now(***REMOVED***.UTC(***REMOVED***.Add(10 * time.Minute***REMOVED***
		newPolicy, err := cmv1.NewUpgradePolicy(***REMOVED***.
			ScheduleType("manual"***REMOVED***.
			Version(desiredVersion.String(***REMOVED******REMOVED***.
			NextRun(tenMinFromNow***REMOVED***.
			Build(***REMOVED***
		if err != nil {
			return fmt.Errorf("failed to create upgrade policy: %v", err***REMOVED***
***REMOVED***
		_, err = r.clusterCollection.Cluster(state.ID.Value***REMOVED***.
			UpgradePolicies(***REMOVED***.
			Add(***REMOVED***.
			Body(newPolicy***REMOVED***.
			SendContext(ctx***REMOVED***
		if err != nil {
			return fmt.Errorf("failed to schedule upgrade: %v", err***REMOVED***
***REMOVED***
	}

	state.Version = plan.Version
	state.UpgradeAcksFor = plan.UpgradeAcksFor
	return nil
}

func updateProxy(state, plan *ClusterRosaClassicState, clusterBuilder *cmv1.ClusterBuilder***REMOVED*** (*cmv1.ClusterBuilder, bool, error***REMOVED*** {
	shouldUpdateProxy := false
	if (state.Proxy == nil && plan.Proxy != nil***REMOVED*** || (state.Proxy != nil && plan.Proxy == nil***REMOVED*** {
		shouldUpdateProxy = true
	} else if state.Proxy != nil && plan.Proxy != nil {
		_, patchNoProxy := common.ShouldPatchString(state.Proxy.NoProxy, plan.Proxy.NoProxy***REMOVED***
		_, patchHttpProxy := common.ShouldPatchString(state.Proxy.HttpProxy, plan.Proxy.HttpProxy***REMOVED***
		_, patchHttpsProxy := common.ShouldPatchString(state.Proxy.HttpsProxy, plan.Proxy.HttpsProxy***REMOVED***
		_, patchAdditionalTrustBundle := common.ShouldPatchString(state.Proxy.AdditionalTrustBundle, plan.Proxy.AdditionalTrustBundle***REMOVED***
		if patchNoProxy || patchHttpProxy || patchHttpsProxy || patchAdditionalTrustBundle {
			shouldUpdateProxy = true
***REMOVED***
	}

	if shouldUpdateProxy {
		var err error
		clusterBuilder, err = buildProxy(plan, clusterBuilder***REMOVED***
		if err != nil {
			return nil, false, err
***REMOVED***
	}

	return clusterBuilder, shouldUpdateProxy, nil
}
func updateNodes(state, plan *ClusterRosaClassicState, clusterBuilder *cmv1.ClusterBuilder***REMOVED*** (*cmv1.ClusterBuilder, bool, error***REMOVED*** {
	// Send request to update the cluster:
	shouldUpdateNodes := false
	clusterNodesBuilder := cmv1.NewClusterNodes(***REMOVED***
	compute, ok := common.ShouldPatchInt(state.Replicas, plan.Replicas***REMOVED***
	if ok {
		clusterNodesBuilder = clusterNodesBuilder.Compute(int(compute***REMOVED******REMOVED***
		shouldUpdateNodes = true
	}

	if !plan.AutoScalingEnabled.Unknown && !plan.AutoScalingEnabled.Null && plan.AutoScalingEnabled.Value {
		// autoscaling enabled
		autoscaling := cmv1.NewMachinePoolAutoscaling(***REMOVED***

		if !plan.MaxReplicas.Unknown && !plan.MaxReplicas.Null {
			autoscaling = autoscaling.MaxReplicas(int(plan.MaxReplicas.Value***REMOVED******REMOVED***
***REMOVED***
		if !plan.MinReplicas.Unknown && !plan.MinReplicas.Null {
			autoscaling = autoscaling.MinReplicas(int(plan.MinReplicas.Value***REMOVED******REMOVED***
***REMOVED***

		clusterNodesBuilder = clusterNodesBuilder.AutoscaleCompute(autoscaling***REMOVED***
		shouldUpdateNodes = true

	} else {
		if (!plan.MaxReplicas.Unknown && !plan.MaxReplicas.Null***REMOVED*** || (!plan.MinReplicas.Unknown && !plan.MinReplicas.Null***REMOVED*** {
			return nil, false, fmt.Errorf("Can't update MaxReplica and/or MinReplica of cluster when autoscaling is not enabled"***REMOVED***
***REMOVED***
	}

	// MP labels update
	if !plan.DefaultMPLabels.Unknown && !plan.DefaultMPLabels.Null {
		if labelsPlan, ok := common.ShouldPatchMap(state.DefaultMPLabels, plan.DefaultMPLabels***REMOVED***; ok {
			labels := map[string]string{}
			for k, v := range labelsPlan.Elems {
				labels[k] = v.(types.String***REMOVED***.Value
	***REMOVED***
			clusterNodesBuilder.ComputeLabels(labels***REMOVED***
			shouldUpdateNodes = true
***REMOVED***
	}

	if shouldUpdateNodes {
		clusterBuilder = clusterBuilder.Nodes(clusterNodesBuilder***REMOVED***
	}

	return clusterBuilder, shouldUpdateNodes, nil
}

func (r *ClusterRosaClassicResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	tflog.Debug(ctx, "begin delete(***REMOVED***"***REMOVED***

	// Get the state:
	state := &ClusterRosaClassicState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete the cluster:
	resource := r.clusterCollection.Cluster(state.ID.Value***REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete cluster",
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	if !state.DisableWaitingInDestroy.Unknown && !state.DisableWaitingInDestroy.Null && state.DisableWaitingInDestroy.Value {
		tflog.Info(ctx, "Waiting for destroy to be completed, is disabled"***REMOVED***
	} else {
		timeout := defaultTimeoutInMinutes
		if !state.DestroyTimeout.Unknown && !state.DestroyTimeout.Null {
			if state.DestroyTimeout.Value <= 0 {
				response.Diagnostics.AddWarning(nonPositiveTimeoutSummary, fmt.Sprintf(nonPositiveTimeoutFormat, state.ID.Value***REMOVED******REMOVED***
	***REMOVED*** else {
				timeout = state.DestroyTimeout.Value
	***REMOVED***
***REMOVED***
		isNotFound, err := r.retryClusterNotFoundWithTimeout(3, 1*time.Minute, ctx, timeout, resource***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
				"Can't poll cluster state",
				fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					state.ID.Value, err,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***

		if !isNotFound {
			response.Diagnostics.AddWarning(
				"Cluster wasn't deleted yet",
				fmt.Sprintf("The cluster with identifier '%s' is not deleted yet, but the polling finisehd due to a timeout", state.ID.Value***REMOVED***,
			***REMOVED***
***REMOVED***

	}
	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *ClusterRosaClassicResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	tflog.Debug(ctx, "begin importstate(***REMOVED***"***REMOVED***

	// Try to retrieve the object:
	get, err := r.clusterCollection.Cluster(request.ID***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				request.ID, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Save the state:
	state := &ClusterRosaClassicState{}
	err = populateRosaClassicClusterState(ctx, object, state, DefaultHttpClient{}***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			***REMOVED***,
		***REMOVED***
		return
	}

	diags := response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

// populateRosaClassicClusterState copies the data from the API object to the Terraform state.
func populateRosaClassicClusterState(ctx context.Context, object *cmv1.Cluster, state *ClusterRosaClassicState, httpClient HttpClient***REMOVED*** error {
	state.ID = types.String{
		Value: object.ID(***REMOVED***,
	}
	state.ExternalID = types.String{
		Value: object.ExternalID(***REMOVED***,
	}
	object.API(***REMOVED***
	state.Name = types.String{
		Value: object.Name(***REMOVED***,
	}
	state.CloudRegion = types.String{
		Value: object.Region(***REMOVED***.ID(***REMOVED***,
	}
	state.MultiAZ = types.Bool{
		Value: object.MultiAZ(***REMOVED***,
	}

	state.Properties = types.Map{
		ElemType: types.StringType,
		Elems:    map[string]attr.Value{},
	}
	state.OCMProperties = types.Map{
		ElemType: types.StringType,
		Elems:    map[string]attr.Value{},
	}
	if props, ok := object.GetProperties(***REMOVED***; ok {
		for k, v := range props {
			if k == propertyRosaTfCommit || k == propertyRosaTfVersion {
				state.OCMProperties.Elems[k] = types.String{
					Value: v,
		***REMOVED***
	***REMOVED*** else {
				state.Properties.Elems[k] = types.String{
					Value: v,
		***REMOVED***
	***REMOVED***
***REMOVED***
	}

	state.APIURL = types.String{
		Value: object.API(***REMOVED***.URL(***REMOVED***,
	}
	state.ConsoleURL = types.String{
		Value: object.Console(***REMOVED***.URL(***REMOVED***,
	}
	state.Domain = types.String{
		Value: fmt.Sprintf("%s.%s", object.Name(***REMOVED***, object.DNS(***REMOVED***.BaseDomain(***REMOVED******REMOVED***,
	}
	state.Replicas = types.Int64{
		Value: int64(object.Nodes(***REMOVED***.Compute(***REMOVED******REMOVED***,
	}
	state.ComputeMachineType = types.String{
		Value: object.Nodes(***REMOVED***.ComputeMachineType(***REMOVED***.ID(***REMOVED***,
	}

	labels, ok := object.Nodes(***REMOVED***.GetComputeLabels(***REMOVED***
	if ok {
		state.DefaultMPLabels = types.Map{
			ElemType: types.StringType,
			Elems:    map[string]attr.Value{},
***REMOVED***
		for k, v := range labels {
			state.DefaultMPLabels.Elems[k] = types.String{
				Value: v,
	***REMOVED***
***REMOVED***
	}

	disableUserWorkload, ok := object.GetDisableUserWorkloadMonitoring(***REMOVED***
	if ok && disableUserWorkload {
		state.DisableWorkloadMonitoring = types.Bool{
			Value: true,
***REMOVED***
	}

	isFips, ok := object.GetFIPS(***REMOVED***
	if ok && isFips {
		state.FIPS = types.Bool{
			Value: true,
***REMOVED***
	}
	autoScaleCompute, ok := object.Nodes(***REMOVED***.GetAutoscaleCompute(***REMOVED***
	if ok {
		var maxReplicas, minReplicas int
		state.AutoScalingEnabled = types.Bool{
			Value: true,
***REMOVED***

		maxReplicas, ok = autoScaleCompute.GetMaxReplicas(***REMOVED***
		if ok {
			state.MaxReplicas = types.Int64{
				Value: int64(maxReplicas***REMOVED***,
	***REMOVED***
***REMOVED***

		minReplicas, ok = autoScaleCompute.GetMinReplicas(***REMOVED***
		if ok {
			state.MinReplicas = types.Int64{
				Value: int64(minReplicas***REMOVED***,
	***REMOVED***
***REMOVED***
	} else {
		// autoscaling not enabled - initialize the MaxReplica and MinReplica
		state.MaxReplicas.Null = true
		state.MinReplicas.Null = true
	}

	azs, ok := object.Nodes(***REMOVED***.GetAvailabilityZones(***REMOVED***
	if ok {
		state.AvailabilityZones.Elems = make([]attr.Value, 0***REMOVED***
		for _, az := range azs {
			state.AvailabilityZones.Elems = append(state.AvailabilityZones.Elems, types.String{
				Value: az,
	***REMOVED******REMOVED***
***REMOVED***
	}

	state.CCSEnabled = types.Bool{
		Value: object.CCS(***REMOVED***.Enabled(***REMOVED***,
	}

	disableSCPChecks, ok := object.CCS(***REMOVED***.GetDisableSCPChecks(***REMOVED***
	if ok && disableSCPChecks {
		state.DisableSCPChecks = types.Bool{
			Value: true,
***REMOVED***
	}

	state.EtcdEncryption = types.Bool{
		Value: object.EtcdEncryption(***REMOVED***,
	}

	//The API does not return account id
	awsAccountID, ok := object.AWS(***REMOVED***.GetAccountID(***REMOVED***
	if ok {
		state.AWSAccountID = types.String{
			Value: awsAccountID,
***REMOVED***
	}

	awsPrivateLink, ok := object.AWS(***REMOVED***.GetPrivateLink(***REMOVED***
	if ok {
		state.AWSPrivateLink = types.Bool{
			Value: awsPrivateLink,
***REMOVED***
	} else {
		state.AWSPrivateLink = types.Bool{
			Null: true,
***REMOVED***
	}
	kmsKeyArn, ok := object.AWS(***REMOVED***.GetKMSKeyArn(***REMOVED***
	if ok {
		state.KMSKeyArn = types.String{
			Value: kmsKeyArn,
***REMOVED***
	}

	httpTokensState, ok := object.AWS(***REMOVED***.GetEc2MetadataHttpTokens(***REMOVED***
	if ok && httpTokensState != "" {
		state.Ec2MetadataHttpTokens = types.String{
			Value: string(httpTokensState***REMOVED***,
***REMOVED***
	}

	sts, ok := object.AWS(***REMOVED***.GetSTS(***REMOVED***
	if ok {
		if state.Sts == nil {
			state.Sts = &Sts{}
***REMOVED***
		oidc_endpoint_url := strings.TrimPrefix(sts.OIDCEndpointURL(***REMOVED***, "https://"***REMOVED***

		state.Sts.OIDCEndpointURL = types.String{
			Value: oidc_endpoint_url,
***REMOVED***
		state.Sts.RoleARN = types.String{
			Value: sts.RoleARN(***REMOVED***,
***REMOVED***
		state.Sts.SupportRoleArn = types.String{
			Value: sts.SupportRoleARN(***REMOVED***,
***REMOVED***
		instanceIAMRoles := sts.InstanceIAMRoles(***REMOVED***
		if instanceIAMRoles != nil {
			state.Sts.InstanceIAMRoles.MasterRoleARN = types.String{
				Value: instanceIAMRoles.MasterRoleARN(***REMOVED***,
	***REMOVED***
			state.Sts.InstanceIAMRoles.WorkerRoleARN = types.String{
				Value: instanceIAMRoles.WorkerRoleARN(***REMOVED***,
	***REMOVED***
***REMOVED***
		// TODO: fix a bug in uhc-cluster-services
		if state.Sts.OperatorRolePrefix.Unknown || state.Sts.OperatorRolePrefix.Null {
			operatorRolePrefix, ok := sts.GetOperatorRolePrefix(***REMOVED***
			if ok {
				state.Sts.OperatorRolePrefix = types.String{
					Value: operatorRolePrefix,
		***REMOVED***
	***REMOVED***
***REMOVED***
		thumbprint, err := getThumbprint(sts.OIDCEndpointURL(***REMOVED***, httpClient***REMOVED***
		if err != nil {
			tflog.Error(ctx, "cannot get thumbprint", err***REMOVED***
			state.Sts.Thumbprint = types.String{
				Value: "",
	***REMOVED***
***REMOVED*** else {
			state.Sts.Thumbprint = types.String{
				Value: thumbprint,
	***REMOVED***
***REMOVED***
		oidcConfig, ok := sts.GetOidcConfig(***REMOVED***
		if ok && oidcConfig != nil {
			state.Sts.OIDCConfigID = types.String{
				Value: oidcConfig.ID(***REMOVED***,
	***REMOVED***
***REMOVED***
	}

	subnetIds, ok := object.AWS(***REMOVED***.GetSubnetIDs(***REMOVED***
	if ok {
		state.AWSSubnetIDs.Elems = make([]attr.Value, 0***REMOVED***
		for _, subnetId := range subnetIds {
			state.AWSSubnetIDs.Elems = append(state.AWSSubnetIDs.Elems, types.String{
				Value: subnetId,
	***REMOVED******REMOVED***
***REMOVED***
	}

	proxy, ok := object.GetProxy(***REMOVED***
	if ok {
		httpProxy, ok := proxy.GetHTTPProxy(***REMOVED***
		if ok {
			state.Proxy.HttpProxy = types.String{
				Value: httpProxy,
	***REMOVED***
***REMOVED***

		httpsProxy, ok := proxy.GetHTTPSProxy(***REMOVED***
		if ok {
			state.Proxy.HttpsProxy = types.String{
				Value: httpsProxy,
	***REMOVED***
***REMOVED***

		noProxy, ok := proxy.GetNoProxy(***REMOVED***
		if ok {
			state.Proxy.NoProxy = types.String{
				Value: noProxy,
	***REMOVED***
***REMOVED***
	}

	trustBundle, ok := object.GetAdditionalTrustBundle(***REMOVED***
	if ok {
		state.Proxy.AdditionalTrustBundle = types.String{
			Value: trustBundle,
***REMOVED***
	}

	machineCIDR, ok := object.Network(***REMOVED***.GetMachineCIDR(***REMOVED***
	if ok {
		state.MachineCIDR = types.String{
			Value: machineCIDR,
***REMOVED***
	} else {
		state.MachineCIDR = types.String{
			Null: true,
***REMOVED***
	}
	serviceCIDR, ok := object.Network(***REMOVED***.GetServiceCIDR(***REMOVED***
	if ok {
		state.ServiceCIDR = types.String{
			Value: serviceCIDR,
***REMOVED***
	} else {
		state.ServiceCIDR = types.String{
			Null: true,
***REMOVED***
	}
	podCIDR, ok := object.Network(***REMOVED***.GetPodCIDR(***REMOVED***
	if ok {
		state.PodCIDR = types.String{
			Value: podCIDR,
***REMOVED***
	} else {
		state.PodCIDR = types.String{
			Null: true,
***REMOVED***
	}
	hostPrefix, ok := object.Network(***REMOVED***.GetHostPrefix(***REMOVED***
	if ok {
		state.HostPrefix = types.Int64{
			Value: int64(hostPrefix***REMOVED***,
***REMOVED***
	} else {
		state.HostPrefix = types.Int64{
			Null: true,
***REMOVED***
	}
	channel_group, ok := object.Version(***REMOVED***.GetChannelGroup(***REMOVED***
	if ok {
		state.ChannelGroup = types.String{
			Value: channel_group,
***REMOVED***
	} else {
		state.ChannelGroup = types.String{
			Null: true,
***REMOVED***
	}
	version, ok := object.Version(***REMOVED***.GetID(***REMOVED***
	// If we're using a non-default channel group, it will have been appended to
	// the version ID. Remove it before saving state.
	version = strings.TrimSuffix(version, fmt.Sprintf("-%s", channel_group***REMOVED******REMOVED***
	version = strings.TrimPrefix(version, "openshift-v"***REMOVED***
	if ok {
		tflog.Debug(ctx, "actual cluster version: %v", version***REMOVED***
		state.CurrentVersion = types.String{
			Value: version,
***REMOVED***
	} else {
		tflog.Debug(ctx, "unknown cluster version"***REMOVED***
		state.CurrentVersion = types.String{
			Null: true,
***REMOVED***
	}
	state.State = types.String{
		Value: string(object.State(***REMOVED******REMOVED***,
	}

	return nil
}

type HttpClient interface {
	Get(url string***REMOVED*** (resp *http.Response, err error***REMOVED***
}

type DefaultHttpClient struct {
}

func (c DefaultHttpClient***REMOVED*** Get(url string***REMOVED*** (resp *http.Response, err error***REMOVED*** {
	return http.Get(url***REMOVED***
}

func getThumbprint(oidcEndpointURL string, httpClient HttpClient***REMOVED*** (thumbprint string, err error***REMOVED*** {
	defer func(***REMOVED*** {
		if panicErr := recover(***REMOVED***; panicErr != nil {
			fmt.Fprintf(os.Stderr, "recovering from: %q\n", panicErr***REMOVED***
			thumbprint = ""
			err = fmt.Errorf("recovering from: %q", panicErr***REMOVED***
***REMOVED***
	}(***REMOVED***

	connect, err := url.ParseRequestURI(oidcEndpointURL***REMOVED***
	if err != nil {
		return "", err
	}

	response, err := httpClient.Get(fmt.Sprintf("https://%s:443", connect.Host***REMOVED******REMOVED***
	if err != nil {
		return "", err
	}

	certChain := response.TLS.PeerCertificates

	// Grab the CA in the chain
	for _, cert := range certChain {
		if cert.IsCA {
			if bytes.Equal(cert.RawIssuer, cert.RawSubject***REMOVED*** {
				hash, err := sha1Hash(cert.Raw***REMOVED***
				if err != nil {
					return "", err
		***REMOVED***
				return hash, nil
	***REMOVED***
***REMOVED***
	}

	// Fall back to using the last certficiate in the chain
	cert := certChain[len(certChain***REMOVED***-1]
	return sha1Hash(cert.Raw***REMOVED***
}

// sha1Hash computes the SHA1 of the byte array and returns the hex encoding as a string.
func sha1Hash(data []byte***REMOVED*** (string, error***REMOVED*** {
	// nolint:gosec
	hasher := sha1.New(***REMOVED***
	_, err := hasher.Write(data***REMOVED***
	if err != nil {
		return "", fmt.Errorf("Couldn't calculate hash:\n %v", err***REMOVED***
	}
	hashed := hasher.Sum(nil***REMOVED***
	return hex.EncodeToString(hashed***REMOVED***, nil
}

func (r *ClusterRosaClassicResource***REMOVED*** retryClusterNotFoundWithTimeout(attempts int, sleep time.Duration, ctx context.Context, timeout int64,
	resource *cmv1.ClusterClient***REMOVED*** (bool, error***REMOVED*** {
	isNotFound, err := r.waitTillClusterIsNotFoundWithTimeout(ctx, timeout, resource***REMOVED***
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep***REMOVED***
			return r.retryClusterNotFoundWithTimeout(attempts, 2*sleep, ctx, timeout, resource***REMOVED***
***REMOVED***
		return isNotFound, err
	}

	return isNotFound, nil
}

func (r *ClusterRosaClassicResource***REMOVED*** waitTillClusterIsNotFoundWithTimeout(ctx context.Context, timeout int64,
	resource *cmv1.ClusterClient***REMOVED*** (bool, error***REMOVED*** {
	timeoutInMinutes := time.Duration(timeout***REMOVED*** * time.Minute
	pollCtx, cancel := context.WithTimeout(ctx, timeoutInMinutes***REMOVED***
	defer cancel(***REMOVED***
	_, err := resource.Poll(***REMOVED***.
		Interval(pollingIntervalInMinutes * time.Minute***REMOVED***.
		Status(http.StatusNotFound***REMOVED***.
		StartContext(pollCtx***REMOVED***
	sdkErr, ok := err.(*ocm_errors.Error***REMOVED***
	if ok && sdkErr.Status(***REMOVED*** == http.StatusNotFound {
		tflog.Info(ctx, "Cluster was removed"***REMOVED***
		return true, nil
	}
	if err != nil {
		tflog.Error(ctx, "Can't poll cluster deletion"***REMOVED***
		return false, err
	}

	return false, nil
}

func proxyValidators(***REMOVED*** []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate proxy's attributes",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &Proxy{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				errSum := "Invalid proxy's attribute assignment"
				httpsProxy := ""
				httpProxy := ""
				additionalTrustBundle := ""
				var noProxySlice []string

				if !common.IsStringAttributeEmpty(state.HttpProxy***REMOVED*** {
					httpProxy = state.HttpProxy.Value
		***REMOVED***
				if !common.IsStringAttributeEmpty(state.HttpsProxy***REMOVED*** {
					httpsProxy = state.HttpsProxy.Value
		***REMOVED***
				if !common.IsStringAttributeEmpty(state.NoProxy***REMOVED*** {
					noProxySlice = helper.HandleEmptyStringOnSlice(strings.Split(state.NoProxy.Value, ","***REMOVED******REMOVED***
		***REMOVED***

				if !common.IsStringAttributeEmpty(state.AdditionalTrustBundle***REMOVED*** {
					additionalTrustBundle = state.AdditionalTrustBundle.Value
		***REMOVED***

				if httpProxy == "" && httpsProxy == "" && noProxySlice != nil && len(noProxySlice***REMOVED*** > 0 {
					resp.Diagnostics.AddError(errSum, "Expected at least one of the following: http-proxy, https-proxy"***REMOVED***
					return
		***REMOVED***

				if httpProxy == "" && httpsProxy == "" && additionalTrustBundle == "" {
					resp.Diagnostics.AddError(errSum, "Expected at least one of the following: http-proxy, https-proxy, additional-trust-bundle"***REMOVED***
					return
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func propertiesValidators(***REMOVED*** []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate property key override",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				propertiesState := &types.Map{
					ElemType: types.StringType,
		***REMOVED***
				diag := req.Config.GetAttribute(ctx, req.AttributePath, propertiesState***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				if !propertiesState.Null && !propertiesState.Unknown {
					for k := range propertiesState.Elems {
						if k == propertyRosaTfVersion || k == propertyRosaTfCommit {
							errHead := "Invalid property key."
							errDesc := fmt.Sprintf("Can not override reserved properties keys. Reserved keys: '%s'/'%s'", propertyRosaTfVersion, propertyRosaTfCommit***REMOVED***
							resp.Diagnostics.AddError(errHead, errDesc***REMOVED***
							return
				***REMOVED***
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}
