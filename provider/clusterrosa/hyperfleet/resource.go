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

package hyperfleet

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hfwrappers "github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/providerdata"
)

var _ resource.Resource = &ClusterHyperfleetResource{}
var _ resource.ResourceWithConfigure = &ClusterHyperfleetResource{}
var _ resource.ResourceWithImportState = &ClusterHyperfleetResource{}

type ClusterHyperfleetResource struct {
	client    hyperfleet.Interface
	accountID string
	callerARN string
}

func New() resource.Resource {
	return &ClusterHyperfleetResource{}
}

func (r *ClusterHyperfleetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_hyperfleet"
}

func (r *ClusterHyperfleetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		//nolint:lll
		Description: "Manages a ROSA HCP cluster through the Hyperfleet Platform API v2. Authentication uses ambient AWS credentials (environment variables, shared config, or instance profile) — no OCM token is required. The provider must be configured with `hyperfleet_url` and `aws_account_id` to enable this resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Cluster UID assigned by the Platform API on creation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable cluster name. Immutable after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operator_roles_prefix": schema.StringAttribute{
				//nolint:lll
				Description: "Prefix used to construct the seven operator IAM role ARNs (e.g. a prefix of `my-cluster` produces `my-cluster-ingress`, `my-cluster-ebs-csi`, etc.). Immutable after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_subnet_ids": schema.ListAttribute{
				//nolint:lll
				Description: "IDs of worker-node subnets. The first entry is used as the cluster subnet. Typically sourced from `data.aws_subnet`. Immutable after creation.",
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				//nolint:lll
				Description: "ID of the VPC that contains the worker subnet. Typically sourced from `data.aws_subnet.*.vpc_id`. Immutable after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zones": schema.ListAttribute{
				//nolint:lll
				Description: "Availability zones for worker nodes. The first entry determines the cluster AZ. Typically sourced from `data.aws_subnet.*.availability_zone`. Immutable after creation.",
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"expiration_timestamp": schema.StringAttribute{
				Description: "Optional RFC3339 timestamp after which the Platform API automatically deletes the cluster.",
				Optional:    true,
			},
			"aws_partition": schema.StringAttribute{
				Description: "AWS partition used when computing operator role ARNs. Defaults to `aws`. Set to `aws-us-gov` for GovCloud regions.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("aws"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"creator_arn": schema.StringAttribute{
				Description: "IAM ARN of the caller that created the cluster. Populated from the provider `aws_caller_arn` attribute.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cloud_region": schema.StringAttribute{
				Description: "AWS region identifier (e.g. `us-east-1`). Derived from `availability_zones` if not set. Immutable after creation.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"phase": schema.StringAttribute{
				Description: "Current lifecycle phase of the cluster (WaitingForPlacement, Provisioning, Ready, Deleting).",
				Computed:    true,
			},
			"api_url": schema.StringAttribute{
				Description: "Control-plane API endpoint host. Empty until the cluster is Ready.",
				Computed:    true,
			},
			"current_version": schema.StringAttribute{
				Description: "Running OpenShift version as reported by the operator.",
				Computed:    true,
			},
			"management_cluster": schema.StringAttribute{
				Description: "Management cluster ID assigned by the Platform API.",
				Computed:    true,
			},
			"oidc_issuer": schema.StringAttribute{
				//nolint:lll
				Description: "OIDC service-account issuer URL for the cluster. Use this to create the IAM OIDC provider and build operator role trust policies after cluster creation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ClusterHyperfleetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	shared, ok := req.ProviderData.(*providerdata.ProviderSharedData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *providerdata.ProviderSharedData, got: %T.", req.ProviderData),
		)
		return
	}

	if shared.HyperfleetClient == nil {
		resp.Diagnostics.AddError(
			"Hyperfleet not configured",
			"The rhcs_cluster_hyperfleet resource requires the provider to be configured "+
				"with 'hyperfleet_url' and 'aws_account_id'. Add these attributes to your "+
				"provider block.",
		)
		return
	}

	r.client = shared.HyperfleetClient
	r.accountID = shared.HyperfleetAccountID
	r.callerARN = shared.HyperfleetCallerARN
}

func (r *ClusterHyperfleetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClusterHyperfleetState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	partition := plan.AWSPartition.ValueString()
	rolesRef := computeRolesRef(plan.OperatorRolesPrefix.ValueString(), r.accountID, partition)

	var subnetIDs []string
	resp.Diagnostics.Append(plan.AWSSubnetIDs.ElementsAs(ctx, &subnetIDs, false)...)
	if resp.Diagnostics.HasError() || len(subnetIDs) == 0 {
		resp.Diagnostics.AddError("aws_subnet_ids must not be empty", "")
		return
	}
	subnetID := subnetIDs[0]

	var azs []string
	resp.Diagnostics.Append(plan.AvailabilityZones.ElementsAs(ctx, &azs, false)...)
	if resp.Diagnostics.HasError() || len(azs) == 0 {
		resp.Diagnostics.AddError("availability_zones must not be empty", "")
		return
	}
	az := azs[0]

	region := plan.CloudRegion.ValueString()
	if region == "" {
		region = regionFromAZ(az)
	}

	clusterSpec := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: plan.Name.ValueString()},
		Spec: v1alpha1.ClusterSpec{
			HostedCluster: v1alpha1.HostedClusterSpecPassthrough{
				Platform: hypershiftv1beta1.PlatformSpec{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS: &hypershiftv1beta1.AWSPlatformSpec{
						Region:   region,
						RolesRef: rolesRef,
						CloudProviderConfig: &hypershiftv1beta1.AWSCloudProviderConfig{
							VPC:  plan.VPCID.ValueString(),
							Zone: az,
							Subnet: &hypershiftv1beta1.AWSResourceReference{
								ID: &subnetID,
							},
						},
					},
				},
			},
		},
	}

	if !plan.ExpirationTimestamp.IsNull() && !plan.ExpirationTimestamp.IsUnknown() {
		t, err := time.Parse(time.RFC3339, plan.ExpirationTimestamp.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid expiration_timestamp",
				fmt.Sprintf("Must be RFC3339 format: %v", err),
			)
			return
		}
		mt := metav1.NewTime(t)
		clusterSpec.Spec.ExpirationTimestamp = &mt
	}

	cluster, err := r.client.HyperfleetV1alpha1().Clusters().Create(ctx, clusterSpec, hfwrappers.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cluster", err.Error())
		return
	}

	populateState(cluster, r.callerARN, region, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClusterHyperfleetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClusterHyperfleetState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.HyperfleetV1alpha1().Clusters().Get(ctx, state.ID.ValueString(), hfwrappers.GetOptions{})
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read cluster", err.Error())
		return
	}

	populateState(cluster, r.callerARN, state.CloudRegion.ValueString(), &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterHyperfleetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan ClusterHyperfleetState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only expiration_timestamp is mutable.
	cluster, err := r.client.HyperfleetV1alpha1().Clusters().Get(ctx, state.ID.ValueString(), hfwrappers.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read cluster before update", err.Error())
		return
	}

	if !plan.ExpirationTimestamp.IsNull() && !plan.ExpirationTimestamp.IsUnknown() {
		t, err := time.Parse(time.RFC3339, plan.ExpirationTimestamp.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid expiration_timestamp", fmt.Sprintf("Must be RFC3339 format: %v", err))
			return
		}
		mt := metav1.NewTime(t)
		cluster.Spec.ExpirationTimestamp = &mt
	} else {
		cluster.Spec.ExpirationTimestamp = nil
	}

	updated, err := r.client.HyperfleetV1alpha1().Clusters().Update(ctx, cluster, hfwrappers.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cluster", err.Error())
		return
	}

	populateState(updated, r.callerARN, state.CloudRegion.ValueString(), &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClusterHyperfleetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClusterHyperfleetState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.HyperfleetV1alpha1().Clusters().Delete(ctx, state.ID.ValueString(), hfwrappers.DeleteOptions{})
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete cluster", err.Error())
		return
	}
}

func (r *ClusterHyperfleetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cluster, err := r.client.HyperfleetV1alpha1().Clusters().Get(ctx, req.ID, hfwrappers.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to import cluster", err.Error())
		return
	}

	var state ClusterHyperfleetState
	populateState(cluster, r.callerARN, "", &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// populateState maps a Platform API Cluster object into a ClusterHyperfleetState.
func populateState(cluster *v1alpha1.Cluster, callerARN, region string, state *ClusterHyperfleetState) {
	state.ID = types.StringValue(string(cluster.UID))
	state.Name = types.StringValue(cluster.Name)
	state.CreatorARN = types.StringValue(callerARN)
	state.CloudRegion = types.StringValue(region)
	state.Phase = types.StringValue(string(cluster.Status.Phase))
	state.CurrentVersion = types.StringValue(cluster.Status.Version)
	state.APIURL = types.StringValue(cluster.Status.ControlPlaneEndpoint.Host)

	if cluster.Status.PlacementRef != nil {
		state.ManagementCluster = types.StringValue(cluster.Status.PlacementRef.ManagementCluster)
	} else {
		state.ManagementCluster = types.StringValue("")
	}

	if cluster.Spec.ExpirationTimestamp != nil {
		state.ExpirationTimestamp = types.StringValue(cluster.Spec.ExpirationTimestamp.UTC().Format(time.RFC3339))
	} else {
		state.ExpirationTimestamp = types.StringNull()
	}

	// Propagate immutable fields from spec so they round-trip correctly.
	if cluster.Spec.HostedCluster.Platform.AWS != nil {
		aws := cluster.Spec.HostedCluster.Platform.AWS
		state.CloudRegion = types.StringValue(aws.Region)
		if aws.CloudProviderConfig != nil {
			state.VPCID = types.StringValue(aws.CloudProviderConfig.VPC)
			az := aws.CloudProviderConfig.Zone
			state.AvailabilityZones = types.ListValueMust(types.StringType, []attr.Value{types.StringValue(az)})
			if aws.CloudProviderConfig.Subnet != nil && aws.CloudProviderConfig.Subnet.ID != nil {
				state.AWSSubnetIDs = types.ListValueMust(types.StringType, []attr.Value{types.StringValue(*aws.CloudProviderConfig.Subnet.ID)})
			}
		}
		// Derive operator_roles_prefix and partition from RolesRef if not already set.
		if state.OperatorRolesPrefix.IsNull() || state.OperatorRolesPrefix.IsUnknown() || state.OperatorRolesPrefix.ValueString() == "" {
			prefix, partition := prefixAndPartitionFromRolesRef(aws.RolesRef)
			state.OperatorRolesPrefix = types.StringValue(prefix)
			state.AWSPartition = types.StringValue(partition)
		}
	}

	state.OIDCIssuer = types.StringValue(cluster.Spec.HostedCluster.IssuerURL)
}

// computeRolesRef builds the AWSRolesRef from an operator roles prefix, AWS
// account ID, and partition. Role names follow the convention established by
// the Platform API IAM CloudFormation template.
func computeRolesRef(prefix, accountID, partition string) hypershiftv1beta1.AWSRolesRef {
	arn := func(suffix string) string {
		return fmt.Sprintf("arn:%s:iam::%s:role/%s%s", partition, accountID, prefix, suffix)
	}
	return hypershiftv1beta1.AWSRolesRef{
		IngressARN:              arn("-ingress"),
		KubeCloudControllerARN:  arn("-cloud-controller-manager"),
		StorageARN:              arn("-ebs-csi"),
		ImageRegistryARN:        arn("-image-registry"),
		NetworkARN:              arn("-network-config"),
		ControlPlaneOperatorARN: arn("-control-plane-operator"),
		NodePoolManagementARN:   arn("-node-pool-management"),
	}
}

// prefixAndPartitionFromRolesRef derives the operator roles prefix and AWS
// partition from a cluster's RolesRef (used during import).
func prefixAndPartitionFromRolesRef(rolesRef hypershiftv1beta1.AWSRolesRef) (prefix, partition string) {
	arn := rolesRef.NodePoolManagementARN
	// ARN format: arn:<partition>:iam::<account>:role/<prefix>-node-pool-management
	parts := strings.SplitN(arn, ":", 6)
	if len(parts) < 6 {
		return "", "aws"
	}
	partition = parts[1]
	slash := strings.LastIndex(parts[5], "/")
	if slash < 0 {
		return "", partition
	}
	roleName := parts[5][slash+1:]
	prefix, _ = strings.CutSuffix(roleName, "-node-pool-management")
	return prefix, partition
}

// isNotFound returns true for HTTP 404-style errors from the Platform API.
func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "404")
}

// azRegionRE extracts the AWS region prefix from any AZ name, including
// standard (us-east-1a), GovCloud (us-gov-east-1a), Local Zone
// (us-east-1-bos-1a), and Wavelength Zone (us-east-1-wl1-bos-wlz-1).
var azRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

// regionFromAZ derives the AWS region from an availability zone name.
func regionFromAZ(az string) string {
	return azRegionRE.FindString(az)
}
