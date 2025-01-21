package clusterwaiter

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type ClusterWaiterResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
}

var _ resource.ResourceWithConfigure = &ClusterWaiterResource{}

const (
	defaultTimeoutInMinutes = int64(60)
)

func New() resource.Resource {
	return &ClusterWaiterResource{}
}

func (r *ClusterWaiterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_wait"
}

func (r *ClusterWaiterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Wait Cluster Resource To be Ready",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
				},
			},
			"timeout": schema.Int64Attribute{
				Description: "An optional timeout until the cluster is ready. The timeout value is set in minutes." +
					" The default value is 60 minutes.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1), // Timeout must be positive
				},
			},
			"ready": schema.BoolAttribute{
				Description: "Whether the cluster is ready." +
					"Note: this does not account for cluster operators still progressing to completion.",
				Computed: true,
			},
		},
	}
}

func (r *ClusterWaiterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.collection = connection.ClustersMgmt().V1().Clusters()
	r.clusterWait = common.NewClusterWait(r.collection, connection)
}

func (r *ClusterWaiterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get the plan:
	state := &ClusterWaiterState{}
	diags := req.Plan.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.startPolling(ctx, state)

	if err != nil {
		resp.Diagnostics.AddError(
			"Waiting for cluster creation finished with error",
			fmt.Sprintf("Waiting for cluster creation finished with the error %v", err),
		)
		if state == nil {
			return
		}
	}

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ClusterWaiterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Do Nothing
}

func (r *ClusterWaiterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &ClusterWaiterState{}
	diags := req.Plan.Get(ctx, plan)
	_ = req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, err := r.startPolling(ctx, plan)

	if err != nil {
		resp.Diagnostics.AddError("Can't poll cluster state (update resource)", err.Error())
		return
	}

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ClusterWaiterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *ClusterWaiterResource) startPolling(ctx context.Context, state *ClusterWaiterState) (*ClusterWaiterState, error) {
	state.Ready = types.BoolValue(false)

	timeout := defaultTimeoutInMinutes
	if !state.Timeout.IsUnknown() && !state.Timeout.IsNull() {
		timeout = state.Timeout.ValueInt64()
	}

	// Wait till the cluster is ready:
	object, err := r.clusterWait.WaitForClusterToBeReady(ctx, state.Cluster.ValueString(), timeout)
	if err != nil {
		return state, fmt.Errorf(
			"Can't poll state of cluster with identifier '%s': %v",
			state.Cluster.ValueString(), err,
		)
	}

	state.Ready = types.BoolValue(object.State() == cmv1.ClusterStateReady)
	return state, nil
}
