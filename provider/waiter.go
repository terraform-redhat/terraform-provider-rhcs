package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
)

type Waiter interface {
	Ready() bool
}

type ClusterWaiterResourceType struct {
}

type ClusterWaiterState struct {
	Cluster types.String `tfsdk:"cluster"`
	Ready   types.Bool   `tfsdk:"ready"`
}

type ClusterWaiterResource struct {
	logger     logging.Logger
	collection *cmv1.ClustersClient
}

func (t *ClusterWaiterResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "Wait Cluster Resource To be Ready",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
			},
			"ready": {
				Description: "Whether the cluster is ready",
				Type:        types.BoolType,
				Computed:    true,
			},
		},
	}
	return
}

func (t *ClusterWaiterResourceType) NewResource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.Resource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation: use it directly when needed.
	parent := p.(*Provider)

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt().V1().Clusters()

	// Create the resource:
	result = &ClusterWaiterResource{
		logger:     parent.logger,
		collection: collection,
	}
	return
}

func (r *ClusterWaiterResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	// Get the plan:
	state := &ClusterWaiterState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	state.Ready = types.Bool{
		Value: false,
	}

	// Wait till the cluster is ready:
	err := r.isClusterReady(state.Cluster.Value, ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't poll cluster state",
			fmt.Sprintf(
				"Can't poll state of cluster with identifier '%s': %v",
				state.Cluster.Value, err,
			),
		)
		return
	} else {
		state.Ready = types.Bool{
			Value: true,
		}
	}

	// Save the state:
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterWaiterResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse) {
	// Do Nothing
}

func (r *ClusterWaiterResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse) {
	// Do Nothing
}

func (r *ClusterWaiterResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse) {
	response.State.RemoveResource(ctx)
}

func (r *ClusterWaiterResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	// Do Nothing
}

func (r *ClusterWaiterResource) isClusterReady(clusterId string, ctx context.Context) error {
	resource := r.collection.Cluster(clusterId)
	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()
	_, err := resource.Poll().
		Interval(30 * time.Second).
		Predicate(func(get *cmv1.ClusterGetResponse) bool {
			return get.Body().State() == cmv1.ClusterStateReady
		}).
		StartContext(pollCtx)
	return err
}
