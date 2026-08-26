// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hfwrappers "github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/providerdata"
)

var _ resource.Resource = &NodePoolHyperfleetResource{}
var _ resource.ResourceWithConfigure = &NodePoolHyperfleetResource{}
var _ resource.ResourceWithImportState = &NodePoolHyperfleetResource{}

// workerInstanceProfileSuffix is appended to the operator roles prefix to form
// the worker IAM instance profile name. It mirrors the naming convention in the
// IAM manifests (`${operator_roles_prefix}-ROSA-Worker-Role`, see
// tests/tf-manifests/aws/iam-roles/hyperfleet/main.tf).
const workerInstanceProfileSuffix = "-ROSA-Worker-Role"

type NodePoolHyperfleetResource struct {
	client hyperfleet.Interface
}

func NewNodePool() resource.Resource {
	return &NodePoolHyperfleetResource{}
}

func (r *NodePoolHyperfleetResource) Metadata(
	_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_nodepool_hyperfleet"
}

func (r *NodePoolHyperfleetResource) Schema(
	_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		//nolint:lll
		Description: "Manages a node pool for an rhcs_cluster_hyperfleet cluster through the Hyperfleet Platform API v2. The provider must be configured with `hyperfleet_url` and `aws_account_id`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Node pool UID assigned by the Platform API on creation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the node pool. Immutable after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster": schema.StringAttribute{
				Description: "Cluster UID (the `id` output of `rhcs_cluster_hyperfleet`). Immutable after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"replicas": schema.Int64Attribute{
				Description: "Fixed number of nodes.",
				Optional:    true,
			},
			"labels": schema.MapAttribute{
				Description: "Node labels applied on creation. Overwrites any out-of-band modifications.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
				},
			},
			"subnet_id": schema.StringAttribute{
				Description: "ID of the subnet where node instances are placed. Immutable after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_node_pool": schema.SingleNestedAttribute{
				Description: "AWS-specific settings for the node pool.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"instance_type": schema.StringAttribute{
						Description: "EC2 instance type for node instances (e.g. `m5.xlarge`). Immutable after creation.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"tags": schema.MapAttribute{
						Description: "Additional AWS resource tags applied to node instances.",
						ElementType: types.StringType,
						Optional:    true,
					},
					"disk_size": schema.Int64Attribute{
						Description: "Root volume size in GiB. Immutable after creation.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
							int64planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"auto_repair": schema.BoolAttribute{
				Description: "Enables automatic repair of unhealthy nodes.",
				Required:    true,
			},
			"phase": schema.StringAttribute{
				Description: "Current lifecycle phase of the node pool (WaitingForCluster, Provisioning, Ready, Deleting).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ignore_deletion_error": schema.BoolAttribute{
				//nolint:lll
				Description: "When true, errors during deletion are suppressed and the resource is removed from state regardless. Useful when destroying a node pool alongside its parent cluster.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *NodePoolHyperfleetResource) Configure(
	_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
) {
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
			"The rhcs_nodepool_hyperfleet resource requires the provider to be configured "+
				"with 'hyperfleet_url' and 'aws_account_id'. Add these attributes to your "+
				"provider block.",
		)
		return
	}

	r.client = shared.HyperfleetClient
}

func (r *NodePoolHyperfleetResource) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan NodePoolHyperfleetState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec, diags := buildNodePoolSpec(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnetID := plan.SubnetID.ValueString()
	spec.NodePool.Platform.AWS.Subnet = hypershiftv1beta1.AWSResourceReference{ID: &subnetID}

	// The worker IAM instance profile is not carried on the cluster object, so the
	// provider computes it from the operator roles prefix (recovered from the
	// parent cluster's RolesRef) following the IAM manifest naming convention.
	// Without it, CAPA defaults to a non-existent "<infra-id>-worker-profile" and
	// the create fails with "Invalid IAM Instance Profile name".
	//
	// NOTE: this value is derived, not user-supplied, and is intentionally not a
	// schema attribute nor persisted in state. If a future use case needs custom
	// per-nodepool instance profiles, promote it to an optional
	// `aws_node_pool.instance_profile` attribute and read it back from the API
	// response in populateNodePoolState instead of computing it here.
	cluster, err := r.client.HyperfleetV1alpha1().
		Clusters().
		Get(ctx, plan.Cluster.ValueString(), hfwrappers.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read parent cluster for node pool", err.Error())
		return
	}
	if cluster.Spec.HostedCluster.Platform.AWS == nil {
		resp.Diagnostics.AddError(
			"Cannot compute worker instance profile",
			fmt.Sprintf(
				"Parent cluster %q has no AWS platform configuration; cannot derive the operator roles prefix.",
				plan.Cluster.ValueString(),
			), //nolint:lll
		)
		return
	}
	prefix, _ := prefixAndPartitionFromRolesRef(cluster.Spec.HostedCluster.Platform.AWS.RolesRef)
	if prefix == "" {
		resp.Diagnostics.AddError(
			"Cannot compute worker instance profile",
			fmt.Sprintf("Could not derive the operator roles prefix from parent cluster %q RolesRef "+
				"(NodePoolManagementARN=%q). The worker instance profile is required; without it the "+
				"create fails with \"Invalid IAM Instance Profile name\".",
				plan.Cluster.ValueString(), cluster.Spec.HostedCluster.Platform.AWS.RolesRef.NodePoolManagementARN),
		)
		return
	}
	spec.NodePool.Platform.AWS.InstanceProfile = prefix + workerInstanceProfileSuffix

	obj := &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      plan.Name.ValueString(),
			Namespace: plan.Cluster.ValueString(),
		},
		Spec: spec,
	}

	nodePools := r.client.HyperfleetV1alpha1().NodePools(plan.Cluster.ValueString())
	created, err := nodePools.Create(ctx, obj, hfwrappers.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create node pool", err.Error())
		return
	}

	populateNodePoolState(ctx, created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NodePoolHyperfleetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NodePoolHyperfleetState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodePools := r.client.HyperfleetV1alpha1().NodePools(state.Cluster.ValueString())
	np, err := nodePools.Get(ctx, state.ID.ValueString(), hfwrappers.GetOptions{})
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read node pool", err.Error())
		return
	}

	populateNodePoolState(ctx, np, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NodePoolHyperfleetResource) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var state, plan NodePoolHyperfleetState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodePools := r.client.HyperfleetV1alpha1().NodePools(state.Cluster.ValueString())
	np, err := nodePools.Get(ctx, state.ID.ValueString(), hfwrappers.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read node pool before update", err.Error())
		return
	}

	spec, diags := buildNodePoolSpec(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the immutable subnet from the live object.
	spec.NodePool.Platform.AWS.Subnet = np.Spec.NodePool.Platform.AWS.Subnet
	// Preserve the computed worker instance profile from the live object (set on
	// create; see the note in Create).
	spec.NodePool.Platform.AWS.InstanceProfile = np.Spec.NodePool.Platform.AWS.InstanceProfile
	// Preserve the ClusterName set by the API.
	spec.NodePool.ClusterName = np.Spec.NodePool.ClusterName
	np.Spec = spec

	updated, err := nodePools.Update(ctx, np, hfwrappers.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update node pool", err.Error())
		return
	}

	populateNodePoolState(ctx, updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NodePoolHyperfleetResource) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state NodePoolHyperfleetState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodePools := r.client.HyperfleetV1alpha1().NodePools(state.Cluster.ValueString())
	err := nodePools.Delete(ctx, state.ID.ValueString(), hfwrappers.DeleteOptions{})
	if err != nil && !isNotFound(err) {
		if !state.IgnoreDeletionError.ValueBool() {
			resp.Diagnostics.AddError("Failed to delete node pool", err.Error())
		}
	}
}

func (r *NodePoolHyperfleetResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	// Import format: <cluster_uuid>/<nodepool_name>
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected format: <cluster_uuid>/<nodepool_name>",
		)
		return
	}
	clusterID, name := parts[0], parts[1]

	np, err := r.client.HyperfleetV1alpha1().NodePools(clusterID).Get(ctx, name, hfwrappers.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to import node pool", err.Error())
		return
	}

	var state NodePoolHyperfleetState
	state.Cluster = types.StringValue(clusterID)
	populateNodePoolState(ctx, np, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// buildNodePoolSpec constructs a NodePoolSpec from Terraform plan state.
// The caller must set spec.NodePool.Platform.AWS.Subnet separately (immutable after creation).
func buildNodePoolSpec(ctx context.Context, plan *NodePoolHyperfleetState) (v1alpha1.NodePoolSpec, diag.Diagnostics) {
	var diags diag.Diagnostics

	awsPool := &hypershiftv1beta1.AWSNodePoolPlatform{
		InstanceType: plan.AWSNodePool.InstanceType.ValueString(),
	}

	if !plan.AWSNodePool.DiskSize.IsNull() && !plan.AWSNodePool.DiskSize.IsUnknown() {
		awsPool.RootVolume = &hypershiftv1beta1.Volume{
			Size: plan.AWSNodePool.DiskSize.ValueInt64(),
		}
	}

	if !plan.AWSNodePool.Tags.IsNull() && !plan.AWSNodePool.Tags.IsUnknown() {
		var rawTags map[string]string
		diags.Append(plan.AWSNodePool.Tags.ElementsAs(ctx, &rawTags, false)...)
		for k, v := range rawTags {
			awsPool.ResourceTags = append(awsPool.ResourceTags, hypershiftv1beta1.AWSResourceTag{Key: k, Value: v})
		}
	}

	autoRepair := plan.AutoRepair.ValueBool()
	spec := v1alpha1.NodePoolSpec{
		AutoRepair: &autoRepair,
		NodePool: v1alpha1.NodePoolSpecPassthrough{
			Platform: hypershiftv1beta1.NodePoolPlatform{Type: hypershiftv1beta1.AWSPlatform, AWS: awsPool},
		},
	}

	if !plan.Replicas.IsNull() && !plan.Replicas.IsUnknown() {
		r := int32(plan.Replicas.ValueInt64())
		spec.NodePool.Replicas = &r
	}

	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		var rawLabels map[string]string
		diags.Append(plan.Labels.ElementsAs(ctx, &rawLabels, false)...)
		spec.Labels = rawLabels
	}

	return spec, diags
}

// populateNodePoolState maps a Platform API NodePool object into a NodePoolHyperfleetState.
func populateNodePoolState(ctx context.Context, np *v1alpha1.NodePool, state *NodePoolHyperfleetState) {
	state.ID = types.StringValue(string(np.UID))
	state.Name = types.StringValue(np.Name)
	state.Cluster = types.StringValue(np.Namespace)
	state.Phase = types.StringValue(string(np.Status.Phase))

	// The server returns autoRepair under spec.nodePool.management.autoRepair
	// (the HyperShift internal path), not at spec.autoRepair where the public
	// API schema defines it. Until the server is updated, np.Spec.AutoRepair is
	// always nil; preserve the caller-supplied value (plan on Create/Update,
	// prior state on Read) to avoid a false inconsistency after apply.
	if np.Spec.AutoRepair != nil {
		state.AutoRepair = types.BoolValue(*np.Spec.AutoRepair)
	}

	if np.Spec.NodePool.Replicas != nil {
		state.Replicas = types.Int64Value(int64(*np.Spec.NodePool.Replicas))
	} else {
		state.Replicas = types.Int64Null()
	}

	if len(np.Spec.Labels) > 0 {
		labelsVal, _ := types.MapValueFrom(ctx, types.StringType, np.Spec.Labels)
		state.Labels = labelsVal
	} else {
		state.Labels = types.MapNull(types.StringType)
	}

	if aws := np.Spec.NodePool.Platform.AWS; aws != nil {
		awsPool := &NPAWSNodePool{
			InstanceType: types.StringValue(aws.InstanceType),
		}
		if aws.RootVolume != nil {
			awsPool.DiskSize = types.Int64Value(aws.RootVolume.Size)
		} else {
			awsPool.DiskSize = types.Int64Null()
		}
		if len(aws.ResourceTags) > 0 {
			tagMap := make(map[string]string, len(aws.ResourceTags))
			for _, tag := range aws.ResourceTags {
				tagMap[tag.Key] = tag.Value
			}
			tagsVal, _ := types.MapValueFrom(ctx, types.StringType, tagMap)
			awsPool.Tags = tagsVal
		} else {
			awsPool.Tags = types.MapNull(types.StringType)
		}
		state.AWSNodePool = awsPool

		if aws.Subnet.ID != nil {
			state.SubnetID = types.StringValue(*aws.Subnet.ID)
		}
	}
}
