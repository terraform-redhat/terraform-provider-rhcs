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
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	semver "github.com/hashicorp/go-version"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/registry_config"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	idputils "github.com/openshift-online/ocm-common/pkg/idp/utils"
	"github.com/openshift-online/ocm-common/pkg/ocm/consts"
	ocmConsts "github.com/openshift-online/ocm-common/pkg/ocm/consts"
	ocmUtils "github.com/openshift-online/ocm-common/pkg/ocm/utils"
	"github.com/openshift-online/ocm-common/pkg/rosa/oidcconfigs"
	commonutils "github.com/openshift-online/ocm-common/pkg/utils"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocm_errors "github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/identityprovider"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"

	ocmr "github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
	rosa "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/common"
	rosaTypes "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/common/types"
	sharedvpc "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/hcp/shared_vpc"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/hcp/upgrade"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/sts"
)

const (
	// FIXME: This should be coming from the API or only validate at the API level
	MinVersion = "4.12.0"
)

type ClusterRosaHcpResource struct {
	rosaTypes.BaseCluster
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
		Description: "OpenShift managed cluster using ROSA HCP.",
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
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: fmt.Sprintf("Name of the cluster. Cannot exceed %d characters in length. %s",
					rosa.MaxClusterNameLength,
					common.ValueCannotBeChangedStringDescription),
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(rosa.MaxClusterNameLength),
				},
			},
			"domain_prefix": schema.StringAttribute{
				Description: fmt.Sprintf("The domain prefix is optionally assigned by the user."+
					"It will appear in the Cluster's domain when the cluster is provisioned. "+
					"If not supplied, it will be auto generated. It cannot exceed %d characters in length. %s",
					rosa.MaxClusterDomainPrefixLength, common.ValueCannotBeChangedStringDescription),
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(rosa.MaxClusterDomainPrefixLength),
				},
			},
			"cloud_region": schema.StringAttribute{
				Description: "AWS region identifier, for example 'us-east-1'.",
				Required:    true,
			},
			"sts": schema.SingleNestedAttribute{
				Description: "STS configuration.",
				Attributes:  sts.HcpStsResource(),
				Required:    true,
			},
			"properties": schema.MapAttribute{
				Description: "User defined properties. It is essential to include property 'role_creator_arn' with the value of the user creating the cluster. Example: properties = {rosa_creator_arn = data.aws_caller_identity.current.arn}",
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
				//nolint:lll
				Description: "Base DNS domain name previously reserved, e.g. '1vo8.p3.openshiftapps.com'. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Description: "Number of worker/compute nodes to provision. " +
					"Requires that the number supplied be a multiple of the number of private subnets. " +
					rosaTypes.PoolMessage,
				Optional: true,
			},
			"compute_machine_type": schema.StringAttribute{
				Description: "Identifies the machine type used by the initial worker nodes, " +
					"for example `m5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values. " + rosaTypes.PoolMessage,
				Optional: true,
			},
			"aws_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d{12}$`), "aws account ID must be only digits and exactly 12 in length"),
				},
			},
			"aws_billing_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account for billing. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d{12}$`), "aws billing account ID must be only digits and exactly 12 in length"),
				},
			},
			"aws_subnet_ids": schema.ListAttribute{
				Description: "AWS subnet IDs. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Required:    true,
			},
			"kms_key_arn": schema.StringAttribute{
				Description: "Used to encrypt root volume of compute node pools. The key ARN is the Amazon Resource Name (ARN) of a AWS Key Management Service (KMS) Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID" +
					"(optional). " + common.ValueCannotBeChangedStringDescription,
				Optional: true,
			},
			"etcd_kms_key_arn": schema.StringAttribute{
				Description: "Used for etcd encryption. The key ARN is the Amazon Resource Name (ARN) of a AWS Key Management Service (KMS) Key. It is a unique, " +
					"fully qualified identifier for the AWS KMS Key. A key ARN includes the AWS account, Region, and the key ID" +
					"(optional). " + common.ValueCannotBeChangedStringDescription,
				Optional: true,
			},
			"private": schema.BoolAttribute{
				Description: "Provides private connectivity from your cluster's VPC to Red Hat SRE, without exposing traffic to the public internet. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zones": schema.ListAttribute{
				Description: "Availability zones. " + rosaTypes.PoolMessage,
				ElementType: types.StringType,
				Required:    true,
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
				Description: "Wait until the cluster is either in a ready state or in an error state. The waiter has a timeout of 45 minutes, with the default value set to false",
				Optional:    true,
			},
			"wait_for_std_compute_nodes_complete": schema.BoolAttribute{
				Description: "Wait until the cluster standard compute pools are created. The waiter has a timeout of 60 minutes, with the default value set to false. This can only be provided when also waiting for create completion.",
				Optional:    true,
			},
			"max_hcp_cluster_wait_timeout_in_minutes": schema.Int64Attribute{
				Description: "This value sets the maximum duration in minutes to wait for a HCP cluster to be in a ready state.",
				Optional:    true,
			},
			"max_machinepool_wait_timeout_in_minutes": schema.Int64Attribute{
				Description: "This value sets the maximum duration in minutes to wait for machine pools to be in a ready state.",
				Optional:    true,
			},
			"create_admin_user": schema.BoolAttribute{
				Description: "Indicates if create cluster admin user. Set it true to create cluster admin user with default username `cluster-admin` " +
					"and generated password. It will be ignored if `admin_credentials` is set." + common.ValueCannotBeChangedStringDescription,
				Optional: true,
			},
			"admin_credentials": schema.SingleNestedAttribute{
				Description: "Admin user credentials. " + common.ValueCannotBeChangedStringDescription,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "Admin username that will be created with the cluster.",
						Optional:    true,
						Computed:    true,
						Validators:  identityprovider.HTPasswdUsernameValidators,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"password": schema.StringAttribute{
						Description: "Admin password that will be created with the cluster.",
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
						Validators:  identityprovider.HTPasswdPasswordValidators,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"ec2_metadata_http_tokens": schema.StringAttribute{
				Description: "This value determines which EC2 Instance Metadata Service mode to use for EC2 instances in the cluster." +
					"This can be set as `optional` (IMDS v1 or v2) or `required` (IMDSv2 only)." + common.ValueCannotBeChangedStringDescription,
				Optional: true,
				Computed: true,
				Validators: []validator.String{attrvalidators.EnumValueValidator([]string{string(cmv1.Ec2MetadataHttpTokensOptional),
					string(cmv1.Ec2MetadataHttpTokensRequired)})},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registry_config": schema.SingleNestedAttribute{
				Description: "Registry configuration for this cluster.",
				Attributes:  registry_config.RegistryConfigResource(),
				Optional:    true,
			},
			"worker_disk_size": schema.Int64Attribute{
				Description: "Compute node root disk size, in GiB. " + rosaTypes.PoolMessage,
				Optional:    true,
			},
			"aws_additional_compute_security_group_ids": schema.ListAttribute{
				Description: "AWS additional compute security group ids.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"shared_vpc": schema.SingleNestedAttribute{
				Description: "Shared VPC configuration." + common.ValueCannotBeChangedStringDescription,
				Attributes:  sharedvpc.SharedVpcResource(),
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("base_dns_domain")),
					sharedvpc.HcpSharedVpcValidator,
				},
			},
			"aws_additional_allowed_principals": schema.ListAttribute{
				Description: "AWS additional allowed principals.",
				ElementType: types.StringType,
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
			fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.ClusterCollection = connection.ClustersMgmt().V1().Clusters()
	r.VersionCollection = connection.ClustersMgmt().V1().Versions()
	r.ClusterWait = common.NewClusterWait(r.ClusterCollection, connection)
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
	if len(clusterName) > rosa.MaxClusterNameLength {
		errDescription := fmt.Sprintf("Expected a valid value for 'name' maximum of %d characters in length. Provided Cluster name '%s' is of length '%d'",
			rosa.MaxClusterNameLength,
			clusterName, len(clusterName),
		)
		tflog.Error(ctx, errDescription)

		diags.AddError(
			errHeadline,
			errDescription,
		)
		return nil, errors.New(errHeadline + "\n" + errDescription)
	}

	builder.Name(clusterName)
	if common.HasValue(state.DomainPrefix) {
		builder.DomainPrefix(state.DomainPrefix.ValueString())
	}
	builder.CloudProvider(cmv1.NewCloudProvider().ID(string(rosaTypes.Aws)))
	builder.Product(cmv1.NewProduct().ID(string(rosaTypes.Rosa)))
	builder.Region(cmv1.NewCloudRegion().ID(state.CloudRegion.ValueString()))

	// Set default properties
	properties := make(map[string]string)
	for k, v := range rosa.OCMProperties {
		properties[k] = v
	}

	// TODO: refactor to common pkg in properties file
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

		creatorArn, ok := propertiesElements[rosa.PropertyRosaCreatorArn]
		if !ok {
			errDescription := fmt.Sprintf("Expected property '%s'. Please include it, for instance by supplying 'data.aws_caller_identity.current.arn'", rosa.PropertyRosaCreatorArn)
			diags.AddError(
				errHeadline,
				errDescription,
			)
			return nil, errors.New(errHeadline + "\n" + errDescription)
		}
		if !rosa.UserArnRE.MatchString(creatorArn) {
			errDescription := fmt.Sprintf("Property '%s' does not have a valid user arn. Please include it, for instance by supplying 'data.aws_caller_identity.current.arn'", rosa.PropertyRosaCreatorArn)
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

	//autoScalingEnabled := common.BoolWithFalseDefault(state.AutoScalingEnabled)

	replicas := common.OptionalInt64(state.Replicas)
	computeMachineType := common.OptionalString(state.ComputeMachineType)

	availabilityZones, err := common.StringListToArray(ctx, state.AvailabilityZones)
	if err != nil {
		return nil, err
	}

	workerDiskSize := common.OptionalInt64(state.WorkerDiskSize)

	if err := ocmClusterResource.CreateNodes(rosaTypes.Hcp, false, replicas, nil, nil,
		computeMachineType, nil, availabilityZones, true, workerDiskSize, nil); err != nil {
		return nil, err
	}

	// ccs should be enabled in ocm rosa clusters
	ccs := cmv1.NewCCS()
	ccs.Enabled(true)
	builder.CCS(ccs)

	ec2MetadataHttpTokens := common.OptionalString(state.Ec2MetadataHttpTokens)
	kmsKeyARN := common.OptionalString(state.KMSKeyArn)
	etcdKmsKeyArn := common.OptionalString(state.EtcdKmsKeyArn)
	awsAccountID := common.OptionalString(state.AWSAccountID)
	awsBillingAccountId := common.OptionalString(state.AWSBillingAccountID)

	isPrivate := common.BoolWithFalseDefault(state.Private)
	awsSubnetIDs, err := common.StringListToArray(ctx, state.AWSSubnetIDs)
	if err != nil {
		return nil, err
	}
	awsAdditionalComputeSecurityGroupIds, err := common.StringListToArray(ctx, state.AWSAdditionalComputeSecurityGroupIds)
	if err != nil {
		return nil, err
	}
	var stsBuilder *cmv1.STSBuilder
	if state.Sts != nil {
		stsBuilder = ocmr.CreateSTS(state.Sts.RoleARN.ValueString(), state.Sts.SupportRoleArn.ValueString(),
			nil, state.Sts.InstanceIAMRoles.WorkerRoleARN.ValueString(),
			state.Sts.OperatorRolePrefix.ValueString(), common.OptionalString(state.Sts.OIDCConfigID))
	}

	awsTags, err := common.OptionalMap(ctx, state.Tags)
	if err != nil {
		return nil, err
	}

	var ingressHostedZoneId, route53RoleArn, vpceRoleArn, internalCommunicationHostedZoneId *string
	if state.SharedVpc != nil &&
		!common.IsStringAttributeUnknownOrEmpty(state.SharedVpc.IngressPrivateHostedZoneId) &&
		!common.IsStringAttributeUnknownOrEmpty(state.SharedVpc.Route53RoleArn) &&
		!common.IsStringAttributeUnknownOrEmpty(state.SharedVpc.InternalCommunicationPrivateHostedZoneId) &&
		!common.IsStringAttributeUnknownOrEmpty(state.SharedVpc.VpceRoleArn) {
		route53RoleArn = state.SharedVpc.Route53RoleArn.ValueStringPointer()
		ingressHostedZoneId = state.SharedVpc.IngressPrivateHostedZoneId.ValueStringPointer()
		vpceRoleArn = state.SharedVpc.VpceRoleArn.ValueStringPointer()
		internalCommunicationHostedZoneId = state.SharedVpc.InternalCommunicationPrivateHostedZoneId.ValueStringPointer()
	}

	awsAdditionalAllowedPrincipals, err := common.StringListToArray(ctx, state.AWSAdditionalAllowedPrincipals)
	if err != nil {
		return nil, err
	}

	if err := ocmClusterResource.CreateAWSBuilder(rosaTypes.Hcp, awsTags, ec2MetadataHttpTokens,
		kmsKeyARN, etcdKmsKeyArn,
		isPrivate, awsAccountID, awsBillingAccountId, stsBuilder, awsSubnetIDs,
		ingressHostedZoneId, route53RoleArn, internalCommunicationHostedZoneId, vpceRoleArn,
		awsAdditionalComputeSecurityGroupIds, nil, nil, awsAdditionalAllowedPrincipals); err != nil {
		return nil, err
	}

	if !common.IsStringAttributeUnknownOrEmpty(state.BaseDNSDomain) {
		dnsBuilder := cmv1.NewDNS()
		dnsBuilder.BaseDomain(state.BaseDNSDomain.ValueString())
		builder.DNS(dnsBuilder)
	}

	if err := ocmClusterResource.SetAPIPrivacy(isPrivate, isPrivate, stsBuilder != nil); err != nil {
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

	registryConfigBuilder, err := registry_config.CreateRegistryConfigBuilder(ctx, state.RegistryConfig)
	if err != nil {
		return nil, err
	}
	if !registryConfigBuilder.Empty() {
		builder.RegistryConfig(registryConfigBuilder)
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

		vBuilder.ID(ocmUtils.CreateVersionId(state.Version.ValueString(), channelGroup))
		vBuilder.ChannelGroup(channelGroup)
		builder.Version(vBuilder)
	}

	username, password := rosaTypes.ExpandAdminCredentials(ctx, state.AdminCredentials, diags)
	if common.BoolWithFalseDefault(state.CreateAdminUser) || common.HasValue(state.AdminCredentials) {
		if username == "" {
			username = commonutils.ClusterAdminUsername
		}
		if password == "" {
			password, err = idputils.GenerateRandomPassword()
			if err != nil {
				tflog.Error(ctx, "Failed to generate random password")
				return nil, err
			}
		}

		hashedPwd, err := idputils.GenerateHTPasswdCompatibleHash(password)
		if err != nil {
			tflog.Error(ctx, "Failed to hash the password")
			return nil, err
		}
		if os.Getenv("IS_TEST") == "true" {
			hashedPwd = fmt.Sprintf("hash(%s)", password)
		}
		htpasswdUsers := []*cmv1.HTPasswdUserBuilder{
			cmv1.NewHTPasswdUser().Username(username).HashedPassword(hashedPwd),
		}
		htpassUserList := cmv1.NewHTPasswdUserList().Items(htpasswdUsers...)
		htPasswdIDP := cmv1.NewHTPasswdIdentityProvider().Users(htpassUserList)
		builder.Htpasswd(htPasswdIDP)
	}
	state.AdminCredentials = rosaTypes.FlattenAdminCredentials(username, password)

	builder, err = proxy.BuildProxy(state.Proxy, builder)
	if err != nil {
		tflog.Error(ctx, "Failed to build the Proxy's attributes")
		return nil, err
	}

	object, err := builder.Build()
	return object, err
}

// TODO: move to ocm commons
func getOcmVersionMinor(ver string) string {
	version, err := semver.NewVersion(ver)
	if err != nil {
		segments := strings.Split(ver, ".")
		return fmt.Sprintf("%s.%s", segments[0], segments[1])
	}
	segments := version.Segments()
	return fmt.Sprintf("%d.%d", segments[0], segments[1])
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

	shouldWaitCreationComplete := common.BoolWithFalseDefault(state.WaitForCreateComplete)
	shouldWaitComputeNodesComplete := common.BoolWithFalseDefault(state.WaitForStdComputeNodesComplete)
	if shouldWaitComputeNodesComplete && !shouldWaitCreationComplete {
		response.Diagnostics.AddError(
			summary,
			"When waiting for standard compute nodes to complete it is also required to wait for creation of the cluster",
		)
		return
	}

	hasEtcdEncrpytion := common.BoolWithFalseDefault(state.EtcdEncryption)
	hasEtcdKmsKeyArn := common.HasValue(state.EtcdKmsKeyArn) && state.EtcdKmsKeyArn.ValueString() != ""
	if (!hasEtcdEncrpytion && hasEtcdKmsKeyArn) || (hasEtcdEncrpytion && !hasEtcdKmsKeyArn) {
		response.Diagnostics.AddError(
			summary,
			"When utilizing etcd encryption an etcd kms key arn must also be supplied and vice versa",
		)
		return
	}

	// In case version with "openshift-v" prefix was used here,
	// Give a meaningful message to inform the user that it not supported any more
	if common.HasValue(state.Version) && strings.HasPrefix(state.Version.ValueString(), rosa.VersionPrefix) {
		response.Diagnostics.AddError(
			summary,
			"Openshift version must be provided without the \"openshift-v\" prefix",
		)
		return
	}

	channelGroup := consts.DefaultChannelGroup
	if common.HasValue(state.ChannelGroup) {
		channelGroup = state.ChannelGroup.ValueString()
	}
	desiredVersion := ""
	if common.HasValue(state.Version) {
		desiredVersion = state.Version.ValueString()
	}
	_, err := r.GetAndValidateVersionInChannelGroup(ctx, rosaTypes.Hcp, channelGroup, desiredVersion)
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

	add, err := r.ClusterCollection.Add().Body(object).SendContext(ctx)
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

	// Save initial state:
	err = populateRosaHcpClusterState(ctx, object, state)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}

	if shouldWaitCreationComplete {
		tflog.Info(ctx, "Waiting for cluster to get ready")
		timeOut := common.OptionalInt64(state.MaxHCPClusterWaitTimeoutInMinutes)
		timeOut, err = common.ValidateTimeout(timeOut, rosa.MaxHCPClusterWaitTimeoutInMinutes)
		if err != nil {
			response.Diagnostics.AddError(
				"Waiting for cluster creation finished with error",
				fmt.Sprintf("Waiting for cluster creation finished with the error %v", err),
			)
		}
		object, err = r.ClusterWait.WaitForClusterToBeReady(ctx, object.ID(), *timeOut)
		if err != nil {
			response.Diagnostics.AddError(
				"Waiting for cluster creation finished with error",
				fmt.Sprintf("Waiting for cluster creation finished with the error %v", err),
			)
			if object == nil {
				diags = response.State.Set(ctx, state)
				response.Diagnostics.Append(diags...)
				return
			}
		}
		if shouldWaitComputeNodesComplete {
			tflog.Info(ctx, "Waiting for standard compute nodes to get ready")
			timeOut := common.OptionalInt64(state.MaxMachinePoolWaitTimeoutInMinutes)
			timeOut, err = common.ValidateTimeout(timeOut, rosa.MaxMachinePoolWaitTimeoutInMinutes)
			if err != nil {
				response.Diagnostics.AddError(
					"Waiting for cluster creation finished with error",
					fmt.Sprintf("Waiting for cluster creation finished with the error %v", err),
				)
			}
			object, err = r.ClusterWait.WaitForStdComputeNodesToBeReady(ctx, object.ID(), *timeOut)
			if err != nil {
				response.Diagnostics.AddError(
					"Waiting for std compute nodes completion finished with error",
					fmt.Sprintf("Waiting for std compute nodes completion finished with the error %v", err),
				)
				if object == nil {
					diags = response.State.Set(ctx, state)
					response.Diagnostics.Append(diags...)
					return
				}
			}
		}
	}

	// Save the state post wait completion:
	err = populateRosaHcpClusterState(ctx, object, state)
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
	get, err := r.ClusterCollection.Cluster(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		if get.Status() == http.StatusNotFound {
			tflog.Warn(ctx, fmt.Sprintf("cluster (%s) not found, removing from state",
				state.ID.ValueString(),
			))
			response.State.RemoveResource(ctx)
			return
		}
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
	err = populateRosaHcpClusterState(ctx, object, state)
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
	common.ValidateStateAndPlanEquals(state.CloudRegion, plan.CloudRegion, "cloud_region", &diags)
	common.ValidateStateAndPlanEquals(state.DomainPrefix, plan.DomainPrefix, "domain_prefix", &diags)
	common.ValidateStateAndPlanEquals(state.ExternalID, plan.ExternalID, "external_id", &diags)
	common.ValidateStateAndPlanEquals(state.Tags, plan.Tags, "tags", &diags)
	common.ValidateStateAndPlanEquals(state.AWSAccountID, plan.AWSAccountID, "aws_account_id", &diags)
	common.ValidateStateAndPlanEquals(state.AWSSubnetIDs, plan.AWSSubnetIDs, "aws_subnet_ids", &diags)
	common.ValidateStateAndPlanEquals(state.EtcdEncryption, plan.EtcdEncryption, "etcd_encryption", &diags)
	common.ValidateStateAndPlanEquals(state.KMSKeyArn, plan.KMSKeyArn, "kms_key_arn", &diags)
	common.ValidateStateAndPlanEquals(state.EtcdKmsKeyArn, plan.EtcdKmsKeyArn, "etcd_kms_key_arn", &diags)
	common.ValidateStateAndPlanEquals(state.Private, plan.Private, "private", &diags)
	common.ValidateStateAndPlanEquals(state.MachineCIDR, plan.MachineCIDR, "machine_cidr", &diags)
	common.ValidateStateAndPlanEquals(state.ServiceCIDR, plan.ServiceCIDR, "service_cidr", &diags)
	common.ValidateStateAndPlanEquals(state.PodCIDR, plan.PodCIDR, "pod_cidr", &diags)
	common.ValidateStateAndPlanEquals(state.HostPrefix, plan.HostPrefix, "host_prefix", &diags)
	common.ValidateStateAndPlanEquals(state.ChannelGroup, plan.ChannelGroup, "channel_group", &diags)

	// STS field validations
	common.ValidateStateAndPlanEquals(state.Sts.RoleARN, plan.Sts.RoleARN, "sts.role_arn", &diags)
	common.ValidateStateAndPlanEquals(state.Sts.SupportRoleArn, plan.Sts.SupportRoleArn, "sts.support_role_arn", &diags)
	common.ValidateStateAndPlanEquals(state.Sts.InstanceIAMRoles.WorkerRoleARN, plan.Sts.InstanceIAMRoles.WorkerRoleARN, "sts.instance_iam_roles.worker_role_arn", &diags)
	common.ValidateStateAndPlanEquals(state.Sts.OIDCConfigID, plan.Sts.OIDCConfigID, "sts.oidc_config_id", &diags)
	common.ValidateStateAndPlanEquals(state.Sts.OperatorRolePrefix, plan.Sts.OperatorRolePrefix, "sts.operator_role_prefix", &diags)
	common.ValidateStateAndPlanEquals(state.AWSAdditionalComputeSecurityGroupIds, plan.AWSAdditionalComputeSecurityGroupIds, "aws_additional_compute_security_group_ids", &diags)

	// default node pool's attributes
	//common.ValidateStateAndPlanEquals(state.AutoScalingEnabled, plan.AutoScalingEnabled, "autoscaling_enabled", &diags)
	common.ValidateStateAndPlanEquals(state.Replicas, plan.Replicas, "replicas", &diags)
	common.ValidateStateAndPlanEquals(state.ComputeMachineType, plan.ComputeMachineType, "compute_machine_type", &diags)
	common.ValidateStateAndPlanEquals(state.AvailabilityZones, plan.AvailabilityZones, "availability_zones", &diags)
	common.ValidateStateAndPlanEquals(state.Ec2MetadataHttpTokens, plan.Ec2MetadataHttpTokens, "ec2_metadata_http_tokens", &diags)
	common.ValidateStateAndPlanEquals(state.WorkerDiskSize, plan.WorkerDiskSize, "worker_disk_size", &diags)

	// cluster admin attributes
	common.ValidateStateAndPlanEquals(state.CreateAdminUser, plan.CreateAdminUser, "create_admin_user", &diags)
	if !rosaTypes.AdminCredentialsEqual(state.AdminCredentials, plan.AdminCredentials) {
		diags.AddError(common.AssertionErrorSummaryMessage, fmt.Sprintf(common.AssertionErrorDetailsMessage, "admin_credentials", state.AdminCredentials, plan.AdminCredentials))
	}

	common.ValidateStateAndPlanEquals(state.BaseDNSDomain, plan.BaseDNSDomain, "base_dns_domain", &diags)
	if !reflect.DeepEqual(state.SharedVpc, plan.SharedVpc) {
		diags.AddError(common.AssertionErrorSummaryMessage, fmt.Sprintf(common.AssertionErrorDetailsMessage, "shared_vpc",
			common.GetJsonStringOrNullString(state.SharedVpc), common.GetJsonStringOrNullString(plan.SharedVpc)))
	}

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
		propertiesElements, err := rosa.ValidatePatchProperties(ctx, state.Properties, plan.Properties)
		if err != nil {
			response.Diagnostics.AddWarning(
				"Shouldn't patch cluster properties",
				fmt.Sprintf("Shouldn't patch cluster with identifier: '%s', %v", state.ID.ValueString(), err),
			)
		}
		if propertiesElements != nil {
			for k, v := range rosa.OCMProperties {
				propertiesElements[k] = v
			}
			clusterBuilder.Properties(propertiesElements)
		}
	}

	registryConfigBuilder, err := registry_config.UpdateRegistryConfigBuilder(ctx,
		state.RegistryConfig, plan.RegistryConfig)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't patch cluster",
			fmt.Sprintf("Can't patch registry config for cluster with identifier: '%s', %v", state.ID.ValueString(), err),
		)
		return
	}
	if !registryConfigBuilder.Empty() {
		clusterBuilder.RegistryConfig(registryConfigBuilder)
	}

	awsBuilder := cmv1.NewAWS()
	changesToAws := false

	if toPatch, shouldPatch := common.ShouldPatchList(state.AWSAdditionalAllowedPrincipals, plan.AWSAdditionalAllowedPrincipals); shouldPatch {
		additionalAllowedPrincipalsPatch, err := common.StringListToArray(ctx, toPatch)
		if err != nil {
			response.Diagnostics.AddError(
				"Can't patch cluster",
				fmt.Sprintf("Can't patch additional allowed principals for cluster with identifier: '%s', %v", state.ID.ValueString(), err),
			)
			return
		}
		awsBuilder.AdditionalAllowedPrincipals(additionalAllowedPrincipalsPatch...)
		changesToAws = shouldPatch
	}

	if newBillingAcc, shouldPatch := common.ShouldPatchString(state.AWSBillingAccountID, plan.AWSBillingAccountID); shouldPatch {
		awsBuilder.BillingAccountID(newBillingAcc)
		changesToAws = shouldPatch
	}

	if changesToAws {
		clusterBuilder.AWS(awsBuilder)
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

	update, err := r.ClusterCollection.Cluster(state.ID.ValueString()).Update().
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
	err = populateRosaHcpClusterState(ctx, object, plan)
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
	desiredVersion, err := semver.NewVersion(plan.Version.ValueString())
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
	upgrades, err := upgrade.GetScheduledUpgrades(ctx, r.ClusterCollection, state.ID.ValueString())
	if err != nil {
		return fmt.Errorf("failed to get upgrade policies: %v", err)
	}

	// Stop if an upgrade is already in progress
	correctUpgradePending, err := upgrade.CheckAndCancelUpgrades(
		ctx, r.ClusterCollection, upgrades, desiredVersion)
	if err != nil {
		return err
	}

	// Schedule a new upgrade
	if !correctUpgradePending && !cancelingUpgradeOnly {
		ackString := plan.UpgradeAcksFor.ValueString()
		if err = scheduleUpgrade(ctx, r.ClusterCollection, state.ID.ValueString(), desiredVersion, ackString); err != nil {
			return err
		}
	}

	state.Version = plan.Version
	state.UpgradeAcksFor = plan.UpgradeAcksFor
	return nil
}

func (r *ClusterRosaHcpResource) validateUpgrade(ctx context.Context, state, plan *ClusterRosaHcpState) error {
	availableVersions, err := upgrade.GetAvailableUpgradeVersions(
		ctx, r.ClusterCollection, r.VersionCollection, state.ID.ValueString())
	if err != nil {
		return fmt.Errorf("failed to get available upgrades: %v", err)
	}
	trimmedDesiredVersion := strings.TrimPrefix(plan.Version.ValueString(), rosa.VersionPrefix)
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
	upgradePoliciesClient := clusterClient.ControlPlane().UpgradePolicies()
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
	newPolicy, err := cmv1.NewControlPlaneUpgradePolicy().
		ScheduleType(cmv1.ScheduleTypeManual).
		Version(desiredVersion.String()).
		NextRun(tenMinFromNow).
		Build()
	if err != nil {
		return fmt.Errorf("failed to create upgrade policy: %v", err)
	}
	_, err = clusterClient.ControlPlane().UpgradePolicies().
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
		clusterBuilder, err = proxy.BuildProxy(plan.Proxy, clusterBuilder)
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
	resource := r.ClusterCollection.Cluster(state.ID.ValueString())
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
	// TODO: refactor into function so can be shared for both hcp and classic flows
	if common.HasValue(state.DisableWaitingInDestroy) && state.DisableWaitingInDestroy.ValueBool() {
		tflog.Info(ctx, "Waiting for destroy to be completed, is disabled")
	} else {
		timeout := rosa.MaxMachinePoolWaitTimeoutInMinutes
		if common.HasValue(state.DestroyTimeout) {
			if state.DestroyTimeout.ValueInt64() <= 0 {
				response.Diagnostics.AddWarning(rosa.NonPositiveTimeoutSummary, fmt.Sprintf(rosa.NonPositiveTimeoutFormat, state.ID.ValueString()))
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
				fmt.Sprintf("The cluster with identifier '%s' is not deleted yet, but the polling finished due to a timeout", state.ID.ValueString()),
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
func populateRosaHcpClusterState(ctx context.Context, object *cmv1.Cluster, state *ClusterRosaHcpState) error {
	state.ID = types.StringValue(object.ID())
	state.ExternalID = types.StringValue(object.ExternalID())
	object.API()
	state.Name = types.StringValue(object.Name())
	state.CloudRegion = types.StringValue(object.Region().ID())
	state.DomainPrefix = types.StringValue(object.DomainPrefix())

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
	state.Domain = types.StringValue(fmt.Sprintf("%s.%s", object.DomainPrefix(), object.DNS().BaseDomain()))
	state.BaseDNSDomain = types.StringValue(object.DNS().BaseDomain())

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
		if creatorARN, ok := object.Properties()[rosa.PropertyRosaCreatorArn]; ok {
			if arn, err := arn.Parse(creatorARN); err == nil {
				state.AWSAccountID = types.StringValue(arn.AccountID)
			}
		}

	}

	awsPrivateLink, ok := object.AWS().GetPrivateLink()
	if ok {
		state.Private = types.BoolValue(awsPrivateLink)
	} else {
		state.Private = types.BoolValue(true)
	}
	kmsKeyArn, ok := object.AWS().GetKMSKeyArn()
	if ok {
		state.KMSKeyArn = types.StringValue(kmsKeyArn)
	}
	if object.EtcdEncryption() {
		etcdKmsKeyArn, ok := object.AWS().EtcdEncryption().GetKMSKeyARN()
		if ok {
			state.EtcdKmsKeyArn = types.StringValue(etcdKmsKeyArn)
		}
	}

	httpTokensState, ok := object.AWS().GetEc2MetadataHttpTokens()
	if ok && httpTokensState != "" {
		state.Ec2MetadataHttpTokens = types.StringValue(string(httpTokensState))
	} else {
		state.Ec2MetadataHttpTokens = types.StringValue(ec2.HttpTokensStateOptional)
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
		thumbprint, err := oidcconfigs.FetchThumbprint(stsState.OIDCEndpointURL())
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
		if ok && httpProxy != "" {
			state.Proxy.HttpProxy = types.StringValue(httpProxy)
		}

		httpsProxy, ok := proxyObj.GetHTTPSProxy()
		if ok && httpsProxy != "" {
			state.Proxy.HttpsProxy = types.StringValue(httpsProxy)
		}

		noProxy, ok := proxyObj.GetNoProxy()
		if ok && noProxy != "" {
			state.Proxy.NoProxy = types.StringValue(noProxy)
		}
	} else {
		// We cannot set the proxy to nil because the attribute state.Proxy.AdditionalTrustBundle might contain a value.
		// Due to the sensitivity of this attribute, the backend returns the value `REDACTED` for a non-empty AdditionalTrustBundle
		// and if state.Proxy is null it will override the actual value.
		hasProxy = false
		if state.Proxy != nil {
			hasProxy = true
		}
	}

	trustBundle, ok := object.GetAdditionalTrustBundle()
	if ok {
		// If AdditionalTrustBundle is not empty, the ocm-backend always "REDACTED" (sensitive value)
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
	version = strings.TrimPrefix(version, rosa.VersionPrefix)
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
	if state.AdminCredentials.IsUnknown() {
		state.AdminCredentials = rosaTypes.AdminCredentialsNull()
	}

	err := registry_config.PopulateRegistryConfigState(object, state.RegistryConfig)
	if err != nil {
		return err
	}

	if awsObj, ok := object.GetAWS(); ok {
		ingressHostedZoneId := awsObj.PrivateHostedZoneID()
		route53RoleArn := awsObj.PrivateHostedZoneRoleARN()
		internalCommunicationHostedZoneId := awsObj.HcpInternalCommunicationHostedZoneId()
		vpceRoleArn := awsObj.VpcEndpointRoleArn()
		if len(ingressHostedZoneId) > 0 && len(route53RoleArn) > 0 &&
			len(internalCommunicationHostedZoneId) > 0 && len(vpceRoleArn) > 0 {
			state.SharedVpc = &sharedvpc.SharedVpc{
				IngressPrivateHostedZoneId:               types.StringValue(ingressHostedZoneId),
				InternalCommunicationPrivateHostedZoneId: types.StringValue(internalCommunicationHostedZoneId),
				Route53RoleArn:                           types.StringValue(route53RoleArn),
				VpceRoleArn:                              types.StringValue(vpceRoleArn),
			}
		}

		if additionalAllowedPrincipals, ok := awsObj.GetAdditionalAllowedPrincipals(); ok {
			awsAdditionalAllowedPrincipals, err := common.StringArrayToList(additionalAllowedPrincipals)
			if err != nil {
				return err
			}
			state.AWSAdditionalAllowedPrincipals = awsAdditionalAllowedPrincipals
		}
	}

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
		Interval(rosa.DefaultPollingIntervalInMinutes * time.Minute).
		Status(http.StatusNotFound).
		StartContext(pollCtx)
	sdkErr, ok := err.(*ocm_errors.Error)
	if ok && sdkErr.Status() == http.StatusNotFound {
		tflog.Info(ctx, "Cluster was removed")
		return true, nil
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Can't poll cluster deletion: %v", err))
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
