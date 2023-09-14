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
package clusterrosaclassic

***REMOVED***
	"context"
	"errors"
***REMOVED***
***REMOVED***
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"
	semver "github.com/hashicorp/go-version"
	ver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	_ "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocm_errors "github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift/rosa/pkg/ocm"
	"github.com/openshift/rosa/pkg/properties"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	ocmr "github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosaclassic/upgrade"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/identityprovider"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"
***REMOVED***

const (
	defaultTimeoutInMinutes   = int64(60***REMOVED***
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
***REMOVED***

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

func New(***REMOVED*** resource.Resource {
	return &ClusterRosaClassicResource{}
}

func (r *ClusterRosaClassicResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_cluster_rosa_classic"
}

func (r *ClusterRosaClassicResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "OpenShift managed cluster using rosa sts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					// This passes the state through to the plan, preventing
					// "known after apply" since we know it won't change.
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"external_id": schema.StringAttribute{
				Description: "Unique external identifier of the cluster.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"name": schema.StringAttribute{
				Description: "Name of the cluster. Cannot exceed 15 characters in length.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"cloud_region": schema.StringAttribute{
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Required:    true,
	***REMOVED***,
			"sts": schema.SingleNestedAttribute{
				Description: "STS configuration.",
				Attributes:  stsResource(***REMOVED***,
				Optional:    true,
	***REMOVED***,
			"multi_az": schema.BoolAttribute{
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(***REMOVED***,
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"disable_workload_monitoring": schema.BoolAttribute{
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE***REMOVED*** platform metrics.",
				Optional: true,
	***REMOVED***,
			"disable_scp_checks": schema.BoolAttribute{
				Description: "Enables you to monitor your own projects in isolation from Red Hat " +
					"Site Reliability Engineer (SRE***REMOVED*** platform metrics.",
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"properties": schema.MapAttribute{
				Description: "User defined properties.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Map{PropertiesValidator(***REMOVED***},
	***REMOVED***,
			"ocm_properties": schema.MapAttribute{
				Description: "Merged properties defined by OCM and the user defined 'properties'.",
				ElementType: types.StringType,
				Computed:    true,
	***REMOVED***,
			"tags": schema.MapAttribute{
				Description: "Apply user defined tags to all resources created in AWS.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"ccs_enabled": schema.BoolAttribute{
				Description: "Enables customer cloud subscription.",
				Computed:    true,
	***REMOVED***,
			"etcd_encryption": schema.BoolAttribute{
				Description: "Encrypt etcd data.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(***REMOVED***,
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"autoscaling_enabled": schema.BoolAttribute{
				Description: "Enables autoscaling.",
				Optional:    true,
	***REMOVED***,
			"min_replicas": schema.Int64Attribute{
				Description: "Minimum replicas.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"max_replicas": schema.Int64Attribute{
				Description: "Maximum replicas.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"api_url": schema.StringAttribute{
				Description: "URL of the API server.",
				Computed:    true,
	***REMOVED***,
			"console_url": schema.StringAttribute{
				Description: "URL of the console.",
				Computed:    true,
	***REMOVED***,
			"domain": schema.StringAttribute{
				Description: "DNS domain of cluster.",
				Computed:    true,
	***REMOVED***,
			"base_dns_domain": schema.StringAttribute{
				Description: "Base DNS domain name previously reserved and matching the hosted " +
					"zone name of the private Route 53 hosted zone associated with intended shared " +
					"VPC, e.g., '1vo8.p1.openshiftapps.com'.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"replicas": schema.Int64Attribute{
				Description: "Number of worker nodes to provision. Single zone clusters need at least 2 nodes, " +
					"multizone clusters need at least 3 nodes.",
				Optional: true,
				Computed: true,
	***REMOVED***,
			"compute_machine_type": schema.StringAttribute{
				Description: "Identifies the machine type used by the compute nodes, " +
					"for example `r5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"default_mp_labels": schema.MapAttribute{
				Description: "This value is the default machine pool labels. Format should be a comma-separated list of '{\"key1\"=\"value1\", \"key2\"=\"value2\"}'. " +
					"This list overwrites any modifications made to Node labels on an ongoing basis. ",
				ElementType: types.StringType,
				Optional:    true,
	***REMOVED***,
			"aws_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"aws_subnet_ids": schema.ListAttribute{
				Description: "AWS subnet IDs.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"kms_key_arn": schema.StringAttribute{
				Description: "The key ARN is the Amazon Resource Name (ARN***REMOVED*** of a AWS Key Management Service (KMS***REMOVED*** Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"fips": schema.BoolAttribute{
				Description: "Create cluster that uses FIPS Validated / Modules in Process cryptographic libraries.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"aws_private_link": schema.BoolAttribute{
				Description: "Provides private connectivity between VPCs, AWS services, and your on-premises networks, without exposing your traffic to the public internet.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(***REMOVED***,
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"private": schema.BoolAttribute{
				Description: "Restrict master API endpoint and application routes to direct, private connectivity.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(***REMOVED***,
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"availability_zones": schema.ListAttribute{
				Description: "Availability zones.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(***REMOVED***,
					listplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"machine_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for nodes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"proxy": schema.SingleNestedAttribute{
				Description: "proxy",
				Attributes:  proxy.ProxyResource(***REMOVED***,
				Optional:    true,
				Validators:  []validator.Object{proxy.ProxyValidator(***REMOVED***},
	***REMOVED***,
			"service_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for services.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"pod_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for pods.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"host_prefix": schema.Int64Attribute{
				Description: "Length of the prefix of the subnet assigned to each node.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(***REMOVED***,
					int64planmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"channel_group": schema.StringAttribute{
				Description: "Name of the channel group where you select the OpenShift cluster version, for example 'stable'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(ocm.DefaultChannelGroup***REMOVED***,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"version": schema.StringAttribute{
				Description: "Desired version of OpenShift for the cluster, for example '4.1.0'. If version is greater than the currently running version, an upgrade will be scheduled.",
				Optional:    true,
	***REMOVED***,
			"current_version": schema.StringAttribute{
				Description: "The currently running version of OpenShift on the cluster, for example '4.1.0'.",
				Computed:    true,
	***REMOVED***,
			"disable_waiting_in_destroy": schema.BoolAttribute{
				Description: "Disable addressing cluster state in the destroy resource. Default value is false.",
				Optional:    true,
	***REMOVED***,
			"destroy_timeout": schema.Int64Attribute{
				Description: "This value sets the maximum duration in minutes to allow for destroying resources. Default value is 60 minutes.",
				Optional:    true,
	***REMOVED***,
			"state": schema.StringAttribute{
				Description: "State of the cluster.",
				Computed:    true,
	***REMOVED***,
			"ec2_metadata_http_tokens": schema.StringAttribute{
				Description: "This value determines which EC2 metadata mode to use for metadata service interaction " +
					"options for EC2 instances can be optional or required. This feature is available from " +
					"OpenShift version 4.11.0 and newer.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{attrvalidators.EnumValueValidator([]string{string(cmv1.Ec2MetadataHttpTokensOptional***REMOVED***,
					string(cmv1.Ec2MetadataHttpTokensRequired***REMOVED***}***REMOVED***},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"upgrade_acknowledgements_for": schema.StringAttribute{
				Description: "Indicates acknowledgement of agreements required to upgrade the cluster version between" +
					" minor versions (e.g. a value of \"4.12\" indicates acknowledgement of any agreements required to " +
					"upgrade to OpenShift 4.12.z from 4.11 or before***REMOVED***.",
				Optional: true,
	***REMOVED***,
			"admin_credentials": schema.SingleNestedAttribute{
				Description: "Admin user credentials",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "Admin username that will be created with the cluster.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(***REMOVED***,
				***REMOVED***,
						Validators: identityprovider.HTPasswdUsernameValidators,
			***REMOVED***,
					"password": schema.StringAttribute{
						Description: "Admin password that will be created with the cluster.",
						Required:    true,
						Sensitive:   true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(***REMOVED***,
				***REMOVED***,
						Validators: identityprovider.HTPasswdPasswordValidators,
			***REMOVED***,
		***REMOVED***,
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"private_hosted_zone": schema.SingleNestedAttribute{
				Description: "Used in a shared VPC typology. HostedZone attributes",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "ID assigned by AWS to private Route 53 hosted zone associated with intended shared VPC, " +
							"e.g. 'Z05646003S02O1ENCDCSN'.",
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(***REMOVED***,
				***REMOVED***,
			***REMOVED***,
					"role_arn": schema.StringAttribute{
						Description: "AWS IAM role ARN with a policy attached, granting permissions necessary to " +
							"create and manage Route 53 DNS records in private Route 53 hosted zone associated with " +
							"intended shared VPC.",
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(***REMOVED***,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
				Validators: []validator.Object{
					objectvalidator.AlsoRequires(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("sts"***REMOVED******REMOVED***,
					objectvalidator.AlsoRequires(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("base_dns_domain"***REMOVED******REMOVED***,
					objectvalidator.AlsoRequires(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("availability_zones"***REMOVED******REMOVED***,
					objectvalidator.AlsoRequires(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("aws_subnet_ids"***REMOVED******REMOVED***,
					PrivateHZValidator(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (r *ClusterRosaClassicResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection***REMOVED***
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData***REMOVED***,
		***REMOVED***
		return
	}

	r.clusterCollection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
	r.versionCollection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***
}

const (
	errHeadline = "Can't build cluster"
***REMOVED***

func createClassicClusterObject(ctx context.Context,
	state *ClusterRosaClassicState, diags diag.Diagnostics***REMOVED*** (*cmv1.Cluster, error***REMOVED*** {

	ocmClusterResource := ocmr.NewCluster(***REMOVED***
	builder := ocmClusterResource.GetClusterBuilder(***REMOVED***
	clusterName := state.Name.ValueString(***REMOVED***
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

	builder.Name(state.Name.ValueString(***REMOVED******REMOVED***
	builder.CloudProvider(cmv1.NewCloudProvider(***REMOVED***.ID(awsCloudProvider***REMOVED******REMOVED***
	builder.Product(cmv1.NewProduct(***REMOVED***.ID(rosaProduct***REMOVED******REMOVED***
	builder.Region(cmv1.NewCloudRegion(***REMOVED***.ID(state.CloudRegion.ValueString(***REMOVED******REMOVED******REMOVED***
	multiAZ := common.BoolWithFalseDefault(state.MultiAZ***REMOVED***
	builder.MultiAZ(multiAZ***REMOVED***

	// Set default properties
	properties := make(map[string]string***REMOVED***
	for k, v := range OCMProperties {
		properties[k] = v
	}
	if common.HasValue(state.Properties***REMOVED*** {
		propertiesElements, err := common.OptionalMap(ctx, state.Properties***REMOVED***
		if err != nil {
			errDescription := fmt.Sprintf("Expected a valid Map for 'properties' '%v'",
				diags.Errors(***REMOVED***[0].Detail(***REMOVED***,
			***REMOVED***
			tflog.Error(ctx, errDescription***REMOVED***

			diags.AddError(
				errHeadline,
				errDescription,
			***REMOVED***
			return nil, errors.New(errHeadline + "\n" + errDescription***REMOVED***
***REMOVED***

		for k, v := range propertiesElements {
			properties[k] = v
***REMOVED***
	}
	builder.Properties(properties***REMOVED***

	if common.HasValue(state.EtcdEncryption***REMOVED*** {
		builder.EtcdEncryption(state.EtcdEncryption.ValueBool(***REMOVED******REMOVED***
	}

	if common.HasValue(state.ExternalID***REMOVED*** {
		builder.ExternalID(state.ExternalID.ValueString(***REMOVED******REMOVED***
	}

	if common.HasValue(state.DisableWorkloadMonitoring***REMOVED*** {
		builder.DisableUserWorkloadMonitoring(state.DisableWorkloadMonitoring.ValueBool(***REMOVED******REMOVED***
	}

	if !common.IsStringAttributeEmpty(state.BaseDNSDomain***REMOVED*** {
		dnsBuilder := cmv1.NewDNS(***REMOVED***
		dnsBuilder.BaseDomain(state.BaseDNSDomain.Value***REMOVED***
		builder.DNS(dnsBuilder***REMOVED***
	}

	autoScalingEnabled := common.BoolWithFalseDefault(state.AutoScalingEnabled***REMOVED***

	replicas := common.OptionalInt64(state.Replicas***REMOVED***
	minReplicas := common.OptionalInt64(state.MinReplicas***REMOVED***
	maxReplicas := common.OptionalInt64(state.MaxReplicas***REMOVED***
	computeMachineType := common.OptionalString(state.ComputeMachineType***REMOVED***
	labels, err := common.OptionalMap(ctx, state.DefaultMPLabels***REMOVED***
	if err != nil {
		return nil, err
	}
	availabilityZones, err := common.StringListToArray(ctx, state.AvailabilityZones***REMOVED***
	if err != nil {
		return nil, err
	}

	if err = ocmClusterResource.CreateNodes(autoScalingEnabled, replicas, minReplicas, maxReplicas,
		computeMachineType, labels, availabilityZones, multiAZ***REMOVED***; err != nil {
		return nil, err
	}

	// ccs should be enabled in ocm rosa clusters
	ccs := cmv1.NewCCS(***REMOVED***
	ccs.Enabled(true***REMOVED***

	if common.HasValue(state.DisableSCPChecks***REMOVED*** && state.DisableSCPChecks.ValueBool(***REMOVED*** {
		ccs.DisableSCPChecks(true***REMOVED***
	}
	builder.CCS(ccs***REMOVED***

	ec2MetadataHttpTokens := common.OptionalString(state.Ec2MetadataHttpTokens***REMOVED***
	kmsKeyARN := common.OptionalString(state.KMSKeyArn***REMOVED***
	awsAccountID := common.OptionalString(state.AWSAccountID***REMOVED***

	var privateHostedZoneID, privateHostedZoneRoleARN *string = nil, nil
	if state.PrivateHostedZone != nil &&
		!common.IsStringAttributeEmpty(state.PrivateHostedZone.ID***REMOVED*** &&
		!common.IsStringAttributeEmpty(state.PrivateHostedZone.RoleARN***REMOVED*** {
		privateHostedZoneRoleARN = &state.PrivateHostedZone.RoleARN.Value
		privateHostedZoneID = &state.PrivateHostedZone.ID.Value

	isPrivateLink := common.BoolWithFalseDefault(state.AWSPrivateLink***REMOVED***
	isPrivate := common.BoolWithFalseDefault(state.Private***REMOVED***
	awsSubnetIDs, err := common.StringListToArray(ctx, state.AWSSubnetIDs***REMOVED***
	if err != nil {
		return nil, err
	}
	var stsBuilder *cmv1.STSBuilder
	if state.Sts != nil {
		stsBuilder = ocmr.CreateSTS(state.Sts.RoleARN.ValueString(***REMOVED***, state.Sts.SupportRoleArn.ValueString(***REMOVED***,
			state.Sts.InstanceIAMRoles.MasterRoleARN.ValueString(***REMOVED***, state.Sts.InstanceIAMRoles.WorkerRoleARN.ValueString(***REMOVED***,
			state.Sts.OperatorRolePrefix.ValueString(***REMOVED***, common.OptionalString(state.Sts.OIDCConfigID***REMOVED******REMOVED***
	}

	awsTags, err := common.OptionalMap(ctx, state.Tags***REMOVED***
	if err != nil {
		return nil, err
	}
	if err := ocmClusterResource.CreateAWSBuilder(awsTags, ec2MetadataHttpTokens, kmsKeyARN,
		isPrivateLink, awsAccountID, stsBuilder, awsSubnetIDs, privateHostedZoneID, privateHostedZoneRoleARN***REMOVED***; err != nil {
		return nil, err
	}

	if err := ocmClusterResource.SetAPIPrivacy(isPrivate, isPrivateLink, stsBuilder != nil***REMOVED***; err != nil {
		return nil, err
	}

	if common.HasValue(state.FIPS***REMOVED*** && state.FIPS.ValueBool(***REMOVED*** {
		builder.FIPS(true***REMOVED***
	}

	network := cmv1.NewNetwork(***REMOVED***
	if common.HasValue(state.MachineCIDR***REMOVED*** {
		network.MachineCIDR(state.MachineCIDR.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.ServiceCIDR***REMOVED*** {
		network.ServiceCIDR(state.ServiceCIDR.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.PodCIDR***REMOVED*** {
		network.PodCIDR(state.PodCIDR.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.HostPrefix***REMOVED*** {
		network.HostPrefix(int(state.HostPrefix.ValueInt64(***REMOVED******REMOVED******REMOVED***
	}
	if !network.Empty(***REMOVED*** {
		builder.Network(network***REMOVED***
	}

	channelGroup := ocm.DefaultChannelGroup
	if common.HasValue(state.ChannelGroup***REMOVED*** {
		channelGroup = state.ChannelGroup.ValueString(***REMOVED***
	}

	if common.HasValue(state.Version***REMOVED*** {
		// TODO: update it to support all cluster versions
		isSupported, err := common.IsGreaterThanOrEqual(state.Version.ValueString(***REMOVED***, MinVersion***REMOVED***
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error validating required cluster version %s", err***REMOVED******REMOVED***
			errDescription := fmt.Sprintf(
				"Can't check if cluster version is supported '%s': %v",
				state.Version.ValueString(***REMOVED***, err,
			***REMOVED***
			diags.AddError(
				errHeadline,
				errDescription,
			***REMOVED***
			return nil, errors.New(errHeadline + "\n" + errDescription***REMOVED***
***REMOVED***
		if !isSupported {
			description := fmt.Sprintf("Cluster version %s is not supported (minimal supported version is %s***REMOVED***", state.Version.ValueString(***REMOVED***, MinVersion***REMOVED***
			tflog.Error(ctx, description***REMOVED***
			diags.AddError(
				errHeadline,
				description,
			***REMOVED***
			return nil, errors.New(errHeadline + "\n" + description***REMOVED***
***REMOVED***
		vBuilder := cmv1.NewVersion(***REMOVED***
		versionID := fmt.Sprintf("openshift-v%s", state.Version.ValueString(***REMOVED******REMOVED***
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

	if state.AdminCredentials != nil {
		htpasswdUsers := []*cmv1.HTPasswdUserBuilder{}
		htpasswdUsers = append(htpasswdUsers, cmv1.NewHTPasswdUser(***REMOVED***.
			Username(state.AdminCredentials.Username.ValueString(***REMOVED******REMOVED***.Password(state.AdminCredentials.Password.ValueString(***REMOVED******REMOVED******REMOVED***
		htpassUserList := cmv1.NewHTPasswdUserList(***REMOVED***.Items(htpasswdUsers...***REMOVED***
		htPasswdIDP := cmv1.NewHTPasswdIdentityProvider(***REMOVED***.Users(htpassUserList***REMOVED***
		builder.Htpasswd(htPasswdIDP***REMOVED***
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
			httpProxy = state.Proxy.HttpProxy.ValueString(***REMOVED***
			proxy.HTTPProxy(httpProxy***REMOVED***
***REMOVED***
		if !common.IsStringAttributeEmpty(state.Proxy.HttpsProxy***REMOVED*** {
			httpsProxy = state.Proxy.HttpsProxy.ValueString(***REMOVED***
			proxy.HTTPSProxy(httpsProxy***REMOVED***
***REMOVED***
		if !common.IsStringAttributeEmpty(state.Proxy.NoProxy***REMOVED*** {
			proxy.NoProxy(state.Proxy.NoProxy.ValueString(***REMOVED******REMOVED***
***REMOVED***

		if !common.IsStringAttributeEmpty(state.Proxy.AdditionalTrustBundle***REMOVED*** {
			additionalTrustBundle = state.Proxy.AdditionalTrustBundle.ValueString(***REMOVED***
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
	if common.HasValue(state.ChannelGroup***REMOVED*** {
		channelGroup = state.ChannelGroup.ValueString(***REMOVED***
	}

	versionList, err := r.getVersionList(ctx, channelGroup***REMOVED***
	if err != nil {
		return "", err
	}

	version := versionList[0]
	if common.HasValue(state.Version***REMOVED*** {
		version = state.Version.ValueString(***REMOVED***
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
	if common.IsStringAttributeEmpty(state.Ec2MetadataHttpTokens***REMOVED*** ||
		cmv1.Ec2MetadataHttpTokens(state.Ec2MetadataHttpTokens.ValueString(***REMOVED******REMOVED*** == cmv1.Ec2MetadataHttpTokensOptional {
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

func (r *ClusterRosaClassicResource***REMOVED*** Create(ctx context.Context, request resource.CreateRequest,
	response *resource.CreateResponse***REMOVED*** {
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
	if common.HasValue(state.Version***REMOVED*** && strings.HasPrefix(state.Version.ValueString(***REMOVED***, "openshift-v"***REMOVED*** {
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
				state.Name.ValueString(***REMOVED***, err,
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
				state.Name.ValueString(***REMOVED***, err,
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
				state.Name.ValueString(***REMOVED***, err,
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
				state.Name.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	// Save the state:
	err = populateRosaClassicClusterState(ctx, object, state, common.DefaultHttpClient{}***REMOVED***
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

func (r *ClusterRosaClassicResource***REMOVED*** Read(ctx context.Context, request resource.ReadRequest,
	response *resource.ReadResponse***REMOVED*** {
	tflog.Debug(ctx, "begin Read(***REMOVED***"***REMOVED***
	// Get the current state:
	state := &ClusterRosaClassicState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the cluster:
	get, err := r.clusterCollection.Cluster(state.ID.ValueString(***REMOVED******REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil && get.Status(***REMOVED*** == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("cluster (%s***REMOVED*** not found, removing from state",
			state.ID.ValueString(***REMOVED***,
		***REMOVED******REMOVED***
		response.State.RemoveResource(ctx***REMOVED***
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object := get.Body(***REMOVED***

	// Save the state:
	err = populateRosaClassicClusterState(ctx, object, state, common.DefaultHttpClient{}***REMOVED***
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

func (r *ClusterRosaClassicResource***REMOVED*** Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse***REMOVED*** {
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
			fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, %v", state.ID.ValueString(***REMOVED***, err***REMOVED***,
		***REMOVED***
		return
	}

	clusterBuilder := cmv1.NewCluster(***REMOVED***
	clusterBuilder, _, err := updateNodes(ctx, state, plan, clusterBuilder***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster nodes for cluster with identifier: `%s`, %v",
				state.ID.ValueString(***REMOVED***, err,
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
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	_, shouldPatchDisableWorkloadMonitoring := common.ShouldPatchBool(state.DisableWorkloadMonitoring, plan.DisableWorkloadMonitoring***REMOVED***
	if shouldPatchDisableWorkloadMonitoring {
		clusterBuilder.DisableUserWorkloadMonitoring(plan.DisableWorkloadMonitoring.ValueBool(***REMOVED******REMOVED***
	}

	patchProperties := shouldPatchProperties(state, plan***REMOVED***
	if patchProperties {
		propertiesElements, err := common.OptionalMap(ctx, plan.Properties***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
				"Can't upgrade cluster",
				fmt.Sprintf("Can't upgrade cluster version with identifier: `%s`, %v", state.ID.ValueString(***REMOVED***, err***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		if propertiesElements != nil {
			for k, v := range OCMProperties {
				propertiesElements[k] = v
	***REMOVED***
			clusterBuilder.Properties(propertiesElements***REMOVED***
***REMOVED***
	}

	clusterSpec, err := clusterBuilder.Build(***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster patch",
			fmt.Sprintf(
				"Can't build patch for cluster with identifier '%s': %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	update, err := r.clusterCollection.Cluster(state.ID.ValueString(***REMOVED******REMOVED***.Update(***REMOVED***.
		Body(clusterSpec***REMOVED***.
		SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster with identifier '%s': %v",
				state.ID.ValueString(***REMOVED***, err,
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
	err = populateRosaClassicClusterState(ctx, object, plan, common.DefaultHttpClient{}***REMOVED***
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

	tflog.Debug(ctx, "Cluster versions",
		map[string]interface{}{
			"current_version": state.CurrentVersion.ValueString(***REMOVED***,
			"plan-version":    plan.Version.ValueString(***REMOVED***,
			"state-version":   state.Version.ValueString(***REMOVED***,
***REMOVED******REMOVED***

	// See if the user has changed the requested version for this run
	requestedVersionChanged := true
	if !common.IsStringAttributeEmpty(plan.Version***REMOVED*** && !common.IsStringAttributeEmpty(state.Version***REMOVED*** {
		if plan.Version.ValueString(***REMOVED*** == state.Version.ValueString(***REMOVED*** {
			requestedVersionChanged = false
***REMOVED***
	}

	// Check the versions to see if we need to upgrade
	currentVersion, err := semver.NewVersion(state.CurrentVersion.ValueString(***REMOVED******REMOVED***
	if err != nil {
		return fmt.Errorf("failed to parse current cluster version: %v", err***REMOVED***
	}
	// For backward compatibility
	// In case version format with "openshift-v" was already used
	// remove the prefix to adapt the right format and avoid failure
	fixedVersion := strings.TrimPrefix(plan.Version.ValueString(***REMOVED***, "openshift-v"***REMOVED***
	desiredVersion, err := semver.NewVersion(fixedVersion***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to parse desired cluster version: %v", err***REMOVED***
	}
	if currentVersion.GreaterThan(desiredVersion***REMOVED*** {
		tflog.Debug(ctx, "No cluster version upgrade needed."***REMOVED***
		if requestedVersionChanged {
			// User changed the version they want, but actual is higher. We
			// don't support downgrades.
			return fmt.Errorf("cluster version is already above the requested version"***REMOVED***
***REMOVED***
		return nil
	}
	cancelingUpgradeOnly := desiredVersion.Equal(currentVersion***REMOVED***

	if !cancelingUpgradeOnly {
		if err = r.validateUpgrade(ctx, state, plan***REMOVED***; err != nil {
			return err
***REMOVED***
	}

	// Fetch existing upgrade policies
	upgrades, err := upgrade.GetScheduledUpgrades(ctx, r.clusterCollection, state.ID.ValueString(***REMOVED******REMOVED***
	if err != nil {
		return fmt.Errorf("failed to get upgrade policies: %v", err***REMOVED***
	}

	// Stop if an upgrade is already in progress
	correctUpgradePending, err := upgrade.CheckAndCancelUpgrades(ctx, r.clusterCollection, upgrades, desiredVersion***REMOVED***
	if err != nil {
		return err
	}

	// Schedule a new upgrade
	if !correctUpgradePending && !cancelingUpgradeOnly {
		ackString := plan.UpgradeAcksFor.ValueString(***REMOVED***
		if err = scheduleUpgrade(ctx, r.clusterCollection, state.ID.ValueString(***REMOVED***, desiredVersion, ackString***REMOVED***; err != nil {
			return err
***REMOVED***
	}

	state.Version = plan.Version
	state.UpgradeAcksFor = plan.UpgradeAcksFor
	return nil
}

func (r *ClusterRosaClassicResource***REMOVED*** validateUpgrade(ctx context.Context, state, plan *ClusterRosaClassicState***REMOVED*** error {
	// Make sure the desired version is available
	versionId := fmt.Sprintf("openshift-v%s", state.CurrentVersion.ValueString(***REMOVED******REMOVED***
	if common.HasValue(state.ChannelGroup***REMOVED*** && state.ChannelGroup.ValueString(***REMOVED*** != ocm.DefaultChannelGroup {
		versionId += "-" + state.ChannelGroup.ValueString(***REMOVED***
	}
	availableVersions, err := upgrade.GetAvailableUpgradeVersions(ctx, r.versionCollection, versionId***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to get available upgrades: %v", err***REMOVED***
	}
	trimmedDesiredVersion := strings.TrimPrefix(plan.Version.ValueString(***REMOVED***, "openshift-v"***REMOVED***
	desiredVersion, err := semver.NewVersion(trimmedDesiredVersion***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to parse desired version: %v", err***REMOVED***
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

	return nil
}

// Ensure user has acked upgrade gates and schedule the upgrade
func scheduleUpgrade(ctx context.Context, client *cmv1.ClustersClient, clusterID string, desiredVersion *semver.Version, userAckString string***REMOVED*** error {
	// Gate agreements are checked when the upgrade is scheduled, resulting
	// in an error return. ROSA cli does this by scheduling once w/ dryRun
	// to look for un-acked agreements.
	clusterClient := client.Cluster(clusterID***REMOVED***
	upgradePoliciesClient := clusterClient.UpgradePolicies(***REMOVED***
	gates, description, err := upgrade.CheckMissingAgreements(desiredVersion.String(***REMOVED***, clusterID, upgradePoliciesClient***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to check for missing upgrade agreements: %v", err***REMOVED***
	}
	// User ack is required if we have any non-STS-only gates
	userAckRequired := false
	for _, gate := range gates {
		if !gate.STSOnly(***REMOVED*** {
			userAckRequired = true
***REMOVED***
	}
	targetMinorVersion := getOcmVersionMinor(desiredVersion.String(***REMOVED******REMOVED***
	if userAckRequired && userAckString != targetMinorVersion { // User has not acknowledged mandatory gates, stop here.
		return fmt.Errorf("%s\nTo acknowledge these items, please add \"upgrade_acknowledgements_for = %s\""+
			" and re-apply the changes", description, targetMinorVersion***REMOVED***
	}

	// Ack all gates to OCM
	for _, gate := range gates {
		gateID := gate.ID(***REMOVED***
		tflog.Debug(ctx, "Acknowledging version gate", map[string]interface{}{"gateID": gateID}***REMOVED***
		gateAgreementsClient := clusterClient.GateAgreements(***REMOVED***
		err := upgrade.AckVersionGate(gateAgreementsClient, gateID***REMOVED***
		if err != nil {
			return fmt.Errorf("failed to acknowledge version gate '%s' for cluster '%s': %v",
				gateID, clusterID, err***REMOVED***
***REMOVED***
	}

	// Schedule an upgrade
	tenMinFromNow := time.Now(***REMOVED***.UTC(***REMOVED***.Add(10 * time.Minute***REMOVED***
	newPolicy, err := cmv1.NewUpgradePolicy(***REMOVED***.
		ScheduleType("manual"***REMOVED***.
		Version(desiredVersion.String(***REMOVED******REMOVED***.
		NextRun(tenMinFromNow***REMOVED***.
		Build(***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to create upgrade policy: %v", err***REMOVED***
	}
	_, err = clusterClient.UpgradePolicies(***REMOVED***.
		Add(***REMOVED***.
		Body(newPolicy***REMOVED***.
		SendContext(ctx***REMOVED***
	if err != nil {
		return fmt.Errorf("failed to schedule upgrade: %v", err***REMOVED***
	}
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
func updateNodes(ctx context.Context, state, plan *ClusterRosaClassicState, clusterBuilder *cmv1.ClusterBuilder***REMOVED*** (*cmv1.ClusterBuilder, bool, error***REMOVED*** {
	// Send request to update the cluster:
	shouldUpdateNodes := false
	clusterNodesBuilder := cmv1.NewClusterNodes(***REMOVED***
	compute, ok := common.ShouldPatchInt(state.Replicas, plan.Replicas***REMOVED***
	if ok {
		clusterNodesBuilder = clusterNodesBuilder.Compute(int(compute***REMOVED******REMOVED***
		shouldUpdateNodes = true
	}
	if common.HasValue(plan.AutoScalingEnabled***REMOVED*** && common.HasValue(plan.AutoScalingEnabled***REMOVED*** {
		// autoscaling enabled
		autoscaling := cmv1.NewMachinePoolAutoscaling(***REMOVED***

		if common.HasValue(plan.MaxReplicas***REMOVED*** {
			autoscaling = autoscaling.MaxReplicas(int(plan.MaxReplicas.ValueInt64(***REMOVED******REMOVED******REMOVED***
***REMOVED***
		if common.HasValue(plan.MinReplicas***REMOVED*** {
			autoscaling = autoscaling.MinReplicas(int(plan.MinReplicas.ValueInt64(***REMOVED******REMOVED******REMOVED***
***REMOVED***

		clusterNodesBuilder = clusterNodesBuilder.AutoscaleCompute(autoscaling***REMOVED***
		shouldUpdateNodes = true

	} else {
		if common.HasValue(plan.MaxReplicas***REMOVED*** || common.HasValue(plan.MinReplicas***REMOVED*** {
			return nil, false, fmt.Errorf("Can't update MaxReplica and/or MinReplica of cluster when autoscaling is not enabled"***REMOVED***
***REMOVED***
	}

	// MP labels update
	if common.HasValue(plan.DefaultMPLabels***REMOVED*** {
		if labelsPlan, ok := common.ShouldPatchMap(state.DefaultMPLabels, plan.DefaultMPLabels***REMOVED***; ok {
			labels := map[string]string{}
			for k, v := range labelsPlan.Elements(***REMOVED*** {
				labels[k] = v.(types.String***REMOVED***.ValueString(***REMOVED***
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

func (r *ClusterRosaClassicResource***REMOVED*** Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse***REMOVED*** {
	tflog.Debug(ctx, "begin delete(***REMOVED***"***REMOVED***

	// Get the state:
	state := &ClusterRosaClassicState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete the cluster:
	resource := r.clusterCollection.Cluster(state.ID.ValueString(***REMOVED******REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete cluster",
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	if common.HasValue(state.DisableWaitingInDestroy***REMOVED*** && state.DisableWaitingInDestroy.ValueBool(***REMOVED*** {
		tflog.Info(ctx, "Waiting for destroy to be completed, is disabled"***REMOVED***
	} else {
		timeout := defaultTimeoutInMinutes
		if common.HasValue(state.DestroyTimeout***REMOVED*** {
			if state.DestroyTimeout.ValueInt64(***REMOVED*** <= 0 {
				response.Diagnostics.AddWarning(nonPositiveTimeoutSummary, fmt.Sprintf(nonPositiveTimeoutFormat, state.ID.ValueString(***REMOVED******REMOVED******REMOVED***
	***REMOVED*** else {
				timeout = state.DestroyTimeout.ValueInt64(***REMOVED***
	***REMOVED***
***REMOVED***
		isNotFound, err := r.retryClusterNotFoundWithTimeout(3, 1*time.Minute, ctx, timeout, resource***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
				"Can't poll cluster state",
				fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					state.ID.ValueString(***REMOVED***, err,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***

		if !isNotFound {
			response.Diagnostics.AddWarning(
				"Cluster wasn't deleted yet",
				fmt.Sprintf("The cluster with identifier '%s' is not deleted yet, but the polling finisehd due to a timeout", state.ID.ValueString(***REMOVED******REMOVED***,
			***REMOVED***
***REMOVED***

	}
	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *ClusterRosaClassicResource***REMOVED*** ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse***REMOVED*** {
	tflog.Debug(ctx, "begin importstate(***REMOVED***"***REMOVED***

	resource.ImportStatePassthroughID(ctx, path.Root("id"***REMOVED***, request, response***REMOVED***
}

// populateRosaClassicClusterState copies the data from the API object to the Terraform state.
func populateRosaClassicClusterState(ctx context.Context, object *cmv1.Cluster, state *ClusterRosaClassicState, httpClient common.HttpClient***REMOVED*** error {
	state.ID = types.StringValue(object.ID(***REMOVED******REMOVED***
	state.ExternalID = types.StringValue(object.ExternalID(***REMOVED******REMOVED***
	object.API(***REMOVED***
	state.Name = types.StringValue(object.Name(***REMOVED******REMOVED***
	state.CloudRegion = types.StringValue(object.Region(***REMOVED***.ID(***REMOVED******REMOVED***
	state.MultiAZ = types.BoolValue(object.MultiAZ(***REMOVED******REMOVED***
	if props, ok := object.GetProperties(***REMOVED***; ok {
		propertiesMap := map[string]string{}
		ocmPropertiesMap := map[string]string{}
		for k, v := range props {
			ocmPropertiesMap[k] = v
			if _, isDefault := OCMProperties[k]; !isDefault {
				propertiesMap[k] = v
	***REMOVED***
***REMOVED***
		mapValue, err := common.ConvertStringMapToMapType(propertiesMap***REMOVED***
		if err != nil {
			return err
***REMOVED*** else {
			state.Properties = mapValue
***REMOVED***
		mapValue, err = common.ConvertStringMapToMapType(ocmPropertiesMap***REMOVED***
		if err != nil {
			return err
***REMOVED*** else {
			state.OCMProperties = mapValue
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
	state.BaseDNSDomain = types.String{
		Value: object.DNS(***REMOVED***.BaseDomain(***REMOVED***,
	}
	state.APIURL = types.StringValue(object.API(***REMOVED***.URL(***REMOVED******REMOVED***
	state.ConsoleURL = types.StringValue(object.Console(***REMOVED***.URL(***REMOVED******REMOVED***
	state.Domain = types.StringValue(fmt.Sprintf("%s.%s", object.Name(***REMOVED***, object.DNS(***REMOVED***.BaseDomain(***REMOVED******REMOVED******REMOVED***
	state.Replicas = types.Int64Value(int64(object.Nodes(***REMOVED***.Compute(***REMOVED******REMOVED******REMOVED***
	state.ComputeMachineType = types.StringValue(object.Nodes(***REMOVED***.ComputeMachineType(***REMOVED***.ID(***REMOVED******REMOVED***
	state.BaseDNSDomain =types.StringValue(object.DNS(***REMOVED***.BaseDomain(***REMOVED******REMOVED***
	labels, ok := object.Nodes(***REMOVED***.GetComputeLabels(***REMOVED***
	if ok {
		mapValue, err := common.ConvertStringMapToMapType(labels***REMOVED***
		if err != nil {
			return err
***REMOVED*** else {
			state.DefaultMPLabels = mapValue
***REMOVED***
	}

	disableUserWorkload, ok := object.GetDisableUserWorkloadMonitoring(***REMOVED***
	if ok && disableUserWorkload {
		state.DisableWorkloadMonitoring = types.BoolValue(true***REMOVED***
	}

	isFips, ok := object.GetFIPS(***REMOVED***
	if ok && isFips {
		state.FIPS = types.BoolValue(true***REMOVED***
	}
	autoScaleCompute, ok := object.Nodes(***REMOVED***.GetAutoscaleCompute(***REMOVED***
	if ok {
		var maxReplicas, minReplicas int
		state.AutoScalingEnabled = types.BoolValue(true***REMOVED***

		maxReplicas, ok = autoScaleCompute.GetMaxReplicas(***REMOVED***
		if ok {
			state.MaxReplicas = types.Int64Value(int64(maxReplicas***REMOVED******REMOVED***
***REMOVED***

		minReplicas, ok = autoScaleCompute.GetMinReplicas(***REMOVED***
		if ok {
			state.MinReplicas = types.Int64Value(int64(minReplicas***REMOVED******REMOVED***
***REMOVED***
	} else {
		// autoscaling not enabled - initialize the MaxReplica and MinReplica
		state.MaxReplicas = types.Int64Null(***REMOVED***
		state.MinReplicas = types.Int64Null(***REMOVED***
	}

	azs, ok := object.Nodes(***REMOVED***.GetAvailabilityZones(***REMOVED***
	if ok {
		listValue, err := common.StringArrayToList(azs***REMOVED***
		if err != nil {
			return err
***REMOVED*** else {
			state.AvailabilityZones = listValue
***REMOVED***
	}

	state.CCSEnabled = types.BoolValue(object.CCS(***REMOVED***.Enabled(***REMOVED******REMOVED***

	disableSCPChecks, ok := object.CCS(***REMOVED***.GetDisableSCPChecks(***REMOVED***
	if ok && disableSCPChecks {
		state.DisableSCPChecks = types.BoolValue(true***REMOVED***
	}

	state.EtcdEncryption = types.BoolValue(object.EtcdEncryption(***REMOVED******REMOVED***

	// Note: The API does not currently return account id, but we try to get it
	// anyway. Failing that, we fetch the creator ARN from the properties like
	// rosa cli does.
	awsAccountID, ok := object.AWS(***REMOVED***.GetAccountID(***REMOVED***
	if ok {
		state.AWSAccountID = types.StringValue(awsAccountID***REMOVED***
	} else {
		// rosa cli gets it from the properties, so we do the same
		if creatorARN, ok := object.Properties(***REMOVED***[properties.CreatorARN]; ok {
			if arn, err := arn.Parse(creatorARN***REMOVED***; err == nil {
				state.AWSAccountID = types.StringValue(arn.AccountID***REMOVED***
	***REMOVED***
***REMOVED***

	}

	awsPrivateLink, ok := object.AWS(***REMOVED***.GetPrivateLink(***REMOVED***
	if ok {
		state.AWSPrivateLink = types.BoolValue(awsPrivateLink***REMOVED***
	} else {
		state.AWSPrivateLink = types.BoolValue(true***REMOVED***
	}
	listeningMethod, ok := object.API(***REMOVED***.GetListening(***REMOVED***
	if ok {
		state.Private = types.BoolValue(listeningMethod == cmv1.ListeningMethodInternal***REMOVED***
	} else {
		state.Private = types.BoolValue(true***REMOVED***
	}
	kmsKeyArn, ok := object.AWS(***REMOVED***.GetKMSKeyArn(***REMOVED***
	if ok {
		state.KMSKeyArn = types.StringValue(kmsKeyArn***REMOVED***
	}

	httpTokensState, ok := object.AWS(***REMOVED***.GetEc2MetadataHttpTokens(***REMOVED***
	if ok && httpTokensState != "" {
		state.Ec2MetadataHttpTokens = types.StringValue(string(httpTokensState***REMOVED******REMOVED***
	}

	sts, ok := object.AWS(***REMOVED***.GetSTS(***REMOVED***
	if ok {
		if state.Sts == nil {
			state.Sts = &Sts{}
***REMOVED***
		oidc_endpoint_url := strings.TrimPrefix(sts.OIDCEndpointURL(***REMOVED***, "https://"***REMOVED***

		state.Sts.OIDCEndpointURL = types.StringValue(oidc_endpoint_url***REMOVED***
		state.Sts.RoleARN = types.StringValue(sts.RoleARN(***REMOVED******REMOVED***
		state.Sts.SupportRoleArn = types.StringValue(sts.SupportRoleARN(***REMOVED******REMOVED***
		instanceIAMRoles := sts.InstanceIAMRoles(***REMOVED***
		if instanceIAMRoles != nil {
			state.Sts.InstanceIAMRoles.MasterRoleARN = types.StringValue(instanceIAMRoles.MasterRoleARN(***REMOVED******REMOVED***
			state.Sts.InstanceIAMRoles.WorkerRoleARN = types.StringValue(instanceIAMRoles.WorkerRoleARN(***REMOVED******REMOVED***
***REMOVED***
		// TODO: fix a bug in uhc-cluster-services
		if common.IsStringAttributeEmpty(state.Sts.OperatorRolePrefix***REMOVED*** {
			operatorRolePrefix, ok := sts.GetOperatorRolePrefix(***REMOVED***
			if ok {
				state.Sts.OperatorRolePrefix = types.StringValue(operatorRolePrefix***REMOVED***
	***REMOVED***
***REMOVED***
		thumbprint, err := common.GetThumbprint(sts.OIDCEndpointURL(***REMOVED***, httpClient***REMOVED***
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("cannot get thumbprint %v", err***REMOVED******REMOVED***
			state.Sts.Thumbprint = types.StringValue(""***REMOVED***
***REMOVED*** else {
			state.Sts.Thumbprint = types.StringValue(thumbprint***REMOVED***
***REMOVED***
		oidcConfig, ok := sts.GetOidcConfig(***REMOVED***
		if ok && oidcConfig != nil {
			state.Sts.OIDCConfigID = types.StringValue(oidcConfig.ID(***REMOVED******REMOVED***
***REMOVED***
	}

	subnetIds, ok := object.AWS(***REMOVED***.GetSubnetIDs(***REMOVED***
	if ok {
		awsSubnetIds, err := common.StringArrayToList(subnetIds***REMOVED***
		if err != nil {
			return err
***REMOVED***
		state.AWSSubnetIDs = awsSubnetIds
	}

	proxyObj, ok := object.GetProxy(***REMOVED***
	if ok {
		if state.Proxy == nil {
			state.Proxy = &proxy.Proxy{}
***REMOVED***
		httpProxy, ok := proxyObj.GetHTTPProxy(***REMOVED***
		if ok {
			state.Proxy.HttpProxy = types.StringValue(httpProxy***REMOVED***
***REMOVED***

		httpsProxy, ok := proxyObj.GetHTTPSProxy(***REMOVED***
		if ok {
			state.Proxy.HttpsProxy = types.StringValue(httpsProxy***REMOVED***
***REMOVED***

		noProxy, ok := proxyObj.GetNoProxy(***REMOVED***
		if ok {
			state.Proxy.NoProxy = types.StringValue(noProxy***REMOVED***
***REMOVED***
	}

	trustBundle, ok := object.GetAdditionalTrustBundle(***REMOVED***
	if ok && common.IsStringAttributeEmpty(state.Proxy.AdditionalTrustBundle***REMOVED*** {
		if state.Proxy == nil {
			state.Proxy = &proxy.Proxy{}
***REMOVED***
		state.Proxy.AdditionalTrustBundle = types.StringValue(trustBundle***REMOVED***
	}

	machineCIDR, ok := object.Network(***REMOVED***.GetMachineCIDR(***REMOVED***
	if ok {
		state.MachineCIDR = types.StringValue(machineCIDR***REMOVED***
	} else {
		state.MachineCIDR = types.StringNull(***REMOVED***
	}
	serviceCIDR, ok := object.Network(***REMOVED***.GetServiceCIDR(***REMOVED***
	if ok {
		state.ServiceCIDR = types.StringValue(serviceCIDR***REMOVED***
	} else {
		state.ServiceCIDR = types.StringNull(***REMOVED***
	}
	podCIDR, ok := object.Network(***REMOVED***.GetPodCIDR(***REMOVED***
	if ok {
		state.PodCIDR = types.StringValue(podCIDR***REMOVED***
	} else {
		state.PodCIDR = types.StringNull(***REMOVED***
	}
	hostPrefix, ok := object.Network(***REMOVED***.GetHostPrefix(***REMOVED***
	if ok {
		state.HostPrefix = types.Int64Value(int64(hostPrefix***REMOVED******REMOVED***
	} else {
		state.HostPrefix = types.Int64Null(***REMOVED***
	}
	channel_group, ok := object.Version(***REMOVED***.GetChannelGroup(***REMOVED***
	if ok {
		state.ChannelGroup = types.StringValue(channel_group***REMOVED***
	}

	if awsObj, ok := object.GetAWS(***REMOVED***; ok {
		id := awsObj.PrivateHostedZoneID(***REMOVED***
		arn := awsObj.PrivateHostedZoneRoleARN(***REMOVED***

		if len(id***REMOVED*** > 0 && len(arn***REMOVED*** > 0 {
			state.PrivateHostedZone = &PrivateHostedZone{
				RoleARN: types.String{
					Value: arn,
		***REMOVED***,
				ID: types.String{
					Value: id,
		***REMOVED***,
	***REMOVED***
***REMOVED***
	}

	version, ok := object.Version(***REMOVED***.GetID(***REMOVED***
	// If we're using a non-default channel group, it will have been appended to
	// the version ID. Remove it before saving state.
	version = strings.TrimSuffix(version, fmt.Sprintf("-%s", channel_group***REMOVED******REMOVED***
	version = strings.TrimPrefix(version, "openshift-v"***REMOVED***
	if ok {
		tflog.Debug(ctx, fmt.Sprintf("actual cluster version: %v", version***REMOVED******REMOVED***
		state.CurrentVersion = types.StringValue(version***REMOVED***
	} else {
		tflog.Debug(ctx, "Unknown cluster version"***REMOVED***
		state.CurrentVersion = types.StringNull(***REMOVED***

	}
	state.State = types.StringValue(string(object.State(***REMOVED******REMOVED******REMOVED***
	state.Name = types.StringValue(object.Name(***REMOVED******REMOVED***
	state.CloudRegion = types.StringValue(object.Region(***REMOVED***.ID(***REMOVED******REMOVED***

	return nil
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

func shouldPatchProperties(state, plan *ClusterRosaClassicState***REMOVED*** bool {
	// User defined properties needs update
	if _, should := common.ShouldPatchMap(state.Properties, plan.Properties***REMOVED***; should {
		return true
	}

	extractedDefaults := map[string]string{}
	for k, v := range state.OCMProperties.Elements(***REMOVED*** {
		if _, ok := state.Properties.Elements(***REMOVED***[k]; !ok {
			extractedDefaults[k] = v.(types.String***REMOVED***.ValueString(***REMOVED***
***REMOVED***
	}

	if len(extractedDefaults***REMOVED*** != len(OCMProperties***REMOVED*** {
		return true
	}

	for k, v := range OCMProperties {
		if _, ok := extractedDefaults[k]; !ok {
			return true
***REMOVED*** else if extractedDefaults[k] != v {
			return true
***REMOVED***

	}

	return false

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
						if _, isDefaultKey := OCMProperties[k]; isDefaultKey {
							errHead := "Invalid property key."
							errDesc := fmt.Sprintf("Can not override reserved properties keys. %s is a reserved property key", k***REMOVED***
							resp.Diagnostics.AddError(errHead, errDesc***REMOVED***
							return
				***REMOVED***
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func adminCredsValidators(***REMOVED*** []tfsdk.AttributeValidator {
	errSumm := "Invalid admin_creedntials"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate admin username",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				var creds *AdminCredentials
				diag := req.Config.GetAttribute(ctx, req.AttributePath, creds***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				if creds != nil {
					if common.IsStringAttributeEmpty(creds.Username***REMOVED*** {
						diag.AddError(errSumm, "Usename can't be empty"***REMOVED***
						return
			***REMOVED***
					if err := idps.ValidateHTPasswdUsername(creds.Username.Value***REMOVED***; err != nil {
						diag.AddError(errSumm, err.Error(***REMOVED******REMOVED***
						return
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
		&common.AttributeValidator{
			Desc: "Validate admin password",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				var creds *AdminCredentials
				diag := req.Config.GetAttribute(ctx, req.AttributePath, creds***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				if creds != nil {
					if common.IsStringAttributeEmpty(creds.Password***REMOVED*** {
						diag.AddError(errSumm, "Usename can't be empty"***REMOVED***
						return
			***REMOVED***
					if err := idps.ValidateHTPasswdPassword(creds.Password.Value***REMOVED***; err != nil {
						diag.AddError(errSumm, err.Error(***REMOVED******REMOVED***
						return
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func validatePrivateHostedZone(clusterState *ClusterRosaClassicState***REMOVED*** error {
	if clusterState.PrivateHostedZone == nil {
		// Nothing to validate.
		return nil
	}
	// validate ID and ARN are not empty
	if common.IsStringAttributeEmpty(clusterState.PrivateHostedZone.ID***REMOVED*** || common.IsStringAttributeEmpty(clusterState.PrivateHostedZone.RoleARN***REMOVED*** {
		return fmt.Errorf("Invalid configuration. 'private_hosted_zone.id' and 'private_hosted_zone.arn' are required"***REMOVED***
	}
	// Validate running in STS mode
	if clusterState.Sts == nil {
		return fmt.Errorf("Invalid configuration. 'private_hosted_zone' requires 'sts' configueration"***REMOVED***
	}
	// Validate subnets exists
	if len(clusterState.AWSSubnetIDs.Elems***REMOVED*** <= 0 {
		return fmt.Errorf("Invalid configuration. 'private_hosted_zone' requires 'aws_subnet_ids' configueration"***REMOVED***
	}
	// Validate availabilityZones exists
	if len(clusterState.AvailabilityZones.Elems***REMOVED*** <= 0 {
		return fmt.Errorf("Invalid configuration. 'private_hosted_zone' requires 'aws_subnet_ids' configueration"***REMOVED***
	}
	// Validate BaseDomain
	if common.IsStringAttributeEmpty(clusterState.BaseDNSDomain***REMOVED*** {
		return fmt.Errorf("Invalid configuration. 'private_hosted_zone' requires 'base_dns_domain' configueration"***REMOVED***
	}
	return nil
}

// Place holder until the v2 refactoring
func privateHZValidators(***REMOVED*** tfsdk.AttributeValidator {
	return &common.AttributeValidator{
		Desc: "Validate private_hosted_zone",
		Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
			var clusterState *ClusterRosaClassicState
			diag := req.Config.Get(ctx, clusterState***REMOVED***
			if diag.HasError(***REMOVED*** {
				// No attribute to validate
				return
	***REMOVED***
			// Validate
			if err := validatePrivateHostedZone(clusterState***REMOVED***; err != nil {
				diag.AddError("Invalid private_hosted_zone configuration", err.Error(***REMOVED******REMOVED***
	***REMOVED***
***REMOVED***,
	}
}
