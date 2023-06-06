package provider

***REMOVED***
	"context"
***REMOVED***
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type Waiter interface {
	Ready(***REMOVED*** bool
}

type ClusterWaiterResourceType struct {
}

type ClusterWaiterResource struct {
	collection *cmv1.ClustersClient
}

const (
	defaultTimeoutInMinutes   = int64(60***REMOVED***
	nonPositiveTimeoutSummary = "Can't poll cluster state with a non-positive timeout"
	nonPositiveTimeoutFormat  = "Can't poll state of cluster with identifier '%s', the timeout that was set is not a positive number"
	pollingIntervalInMinutes  = 2
***REMOVED***

func (t *ClusterWaiterResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "Wait Cluster Resource To be Ready",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"timeout": {
				Description: "An optional timeout till the cluster is ready. The timeout value should be in minutes." +
					" the default value is 60 minutes",
				Type:     types.Int64Type,
				Optional: true,
	***REMOVED***,
			"ready": {
				Description: "Whether the cluster is ready",
				Type:        types.BoolType,
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *ClusterWaiterResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation: use it directly when needed.
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Create the resource:
	result = &ClusterWaiterResource{
		collection: collection,
	}
	return
}

func (r *ClusterWaiterResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	// Get the plan:
	state := &ClusterWaiterState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	state.Ready = types.Bool{
		Value: false,
	}

	timeout := defaultTimeoutInMinutes
	if !state.Timeout.Unknown && !state.Timeout.Null {
		if state.Timeout.Value <= 0 {
			response.Diagnostics.AddWarning(nonPositiveTimeoutSummary, fmt.Sprintf(nonPositiveTimeoutFormat, state.Cluster.Value***REMOVED******REMOVED***
***REMOVED*** else {
			timeout = state.Timeout.Value
***REMOVED***
	}

	// Wait till the cluster is ready:
	object, err := r.retryClusterReadiness(3, 30*time.Second, state.Cluster.Value, ctx, timeout***REMOVED***
	if err != nil {

		response.Diagnostics.AddError(
			"Can't poll cluster state",
			fmt.Sprintf(
				"Can't poll state of cluster with identifier '%s': %v",
				state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	isClusterReady := false
	if object.State(***REMOVED*** == cmv1.ClusterStateReady {
		isClusterReady = true
	}

	state.Ready = types.Bool{
		Value: isClusterReady,
	}

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterWaiterResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	// Do Nothing
}

func (r *ClusterWaiterResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
	// Do Nothing
}

func (r *ClusterWaiterResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *ClusterWaiterResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	// Do Nothing
}

func (r *ClusterWaiterResource***REMOVED*** isClusterReady(clusterId string, ctx context.Context, timeout int64***REMOVED*** (*cmv1.Cluster, error***REMOVED*** {
	resource := r.collection.Cluster(clusterId***REMOVED***
	var object *cmv1.Cluster
	pollCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout***REMOVED****time.Minute***REMOVED***
	defer cancel(***REMOVED***
	_, err := resource.Poll(***REMOVED***.
		Interval(pollingIntervalInMinutes * time.Minute***REMOVED***.
		Predicate(func(getClusterResponse *cmv1.ClusterGetResponse***REMOVED*** bool {
			object = getClusterResponse.Body(***REMOVED***
			tflog.Debug(ctx, fmt.Sprintf("cluster state is %s", object.State(***REMOVED******REMOVED******REMOVED***
			switch object.State(***REMOVED*** {
			case cmv1.ClusterStateReady,
				cmv1.ClusterStateError:
				return true
	***REMOVED***
			return false
***REMOVED******REMOVED***.
		StartContext(pollCtx***REMOVED***
	if err != nil {
		tflog.Error(ctx, "Can't  poll cluster state"***REMOVED***
		return nil, err
	}

	return object, err
}

func (r *ClusterWaiterResource***REMOVED*** retryClusterReadiness(attempts int, sleep time.Duration, clusterId string, ctx context.Context, timeout int64***REMOVED*** (*cmv1.Cluster, error***REMOVED*** {
	object, err := r.isClusterReady(clusterId, ctx, timeout***REMOVED***
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep***REMOVED***
			return r.retryClusterReadiness(attempts, 2*sleep, clusterId, ctx, timeout***REMOVED***
***REMOVED***
		return object, err
	}

	return object, nil
}
