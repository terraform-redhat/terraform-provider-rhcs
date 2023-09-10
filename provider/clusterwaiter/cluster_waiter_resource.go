package clusterwaiter

***REMOVED***
	"context"
***REMOVED***
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfrschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type Waiter interface {
	Ready(***REMOVED*** bool
}

type ClusterWaiterResource struct {
	collection *cmv1.ClustersClient
}

var _ resource.ResourceWithConfigure = &ClusterWaiterResource{}

const (
	defaultTimeoutInMinutes   = int64(60***REMOVED***
	nonPositiveTimeoutSummary = "Can't poll cluster state with a non-positive timeout"
	nonPositiveTimeoutFormat  = "Can't poll state of cluster with identifier '%s', the timeout that was set is not a positive number"
	pollingIntervalInMinutes  = 2
***REMOVED***

func NewClusterWaiterResource(***REMOVED*** resource.Resource {
	return &ClusterWaiterResource{}
}

func (r *ClusterWaiterResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_cluster_wait"
}

func (r *ClusterWaiterResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = tfrschema.Schema{
		Description: "Wait Cluster Resource To be Ready",
		Attributes: map[string]tfrschema.Attribute{
			"cluster": tfrschema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
	***REMOVED***,
			"timeout": tfrschema.Int64Attribute{
				Description: "An optional timeout till the cluster is ready. The timeout value should be in minutes." +
					" the default value is 60 minutes",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1***REMOVED***, // Timeout must be positive
		***REMOVED***,
	***REMOVED***,
			"ready": tfrschema.BoolAttribute{
				Description: "Whether the cluster is ready",
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
}

func (r *ClusterWaiterResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
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

	r.collection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
}

func (r *ClusterWaiterResource***REMOVED*** Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse***REMOVED*** {
	// Get the plan:
	state := &ClusterWaiterState{}
	diags := req.Plan.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	state, err := r.startPolling(ctx, state***REMOVED***

	if err != nil {
		resp.Diagnostics.AddError("Can't poll cluster state (create resource***REMOVED***", err.Error(***REMOVED******REMOVED***
		return
	}

	// Save the state:
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterWaiterResource***REMOVED*** Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse***REMOVED*** {
	// Do Nothing
}

func (r *ClusterWaiterResource***REMOVED*** Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse***REMOVED*** {
	plan := &ClusterWaiterState{}
	diags := req.Plan.Get(ctx, plan***REMOVED***
	_ = req.Plan.Get(ctx, plan***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}
	state, err := r.startPolling(ctx, plan***REMOVED***

	if err != nil {
		resp.Diagnostics.AddError("Can't poll cluster state (update resource***REMOVED***", err.Error(***REMOVED******REMOVED***
		return
	}

	// Save the state:
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterWaiterResource***REMOVED*** Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse***REMOVED*** {
	resp.State.RemoveResource(ctx***REMOVED***
}

func (r *ClusterWaiterResource***REMOVED*** startPolling(ctx context.Context, state *ClusterWaiterState***REMOVED*** (*ClusterWaiterState, error***REMOVED*** {
	state.Ready = types.BoolValue(false***REMOVED***

	timeout := defaultTimeoutInMinutes
	if !state.Timeout.IsUnknown(***REMOVED*** && !state.Timeout.IsNull(***REMOVED*** {
		timeout = state.Timeout.ValueInt64(***REMOVED***
	}

	// Wait till the cluster is ready:
	object, err := r.retryClusterReadiness(3, 30*time.Second, state.Cluster.ValueString(***REMOVED***, ctx, timeout***REMOVED***
	if err != nil {
		return state, fmt.Errorf(
			"Can't poll state of cluster with identifier '%s': %v",
			state.Cluster.ValueString(***REMOVED***, err,
		***REMOVED***
	}

	state.Ready = types.BoolValue(object.State(***REMOVED*** == cmv1.ClusterStateReady***REMOVED***
	return state, nil
}

func (r *ClusterWaiterResource***REMOVED*** isClusterReady(clusterId string, ctx context.Context, timeout int64***REMOVED*** (*cmv1.Cluster, error***REMOVED*** {
	client := r.collection.Cluster(clusterId***REMOVED***
	var object *cmv1.Cluster
	pollCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout***REMOVED****time.Minute***REMOVED***
	defer cancel(***REMOVED***
	_, err := client.Poll(***REMOVED***.
		Interval(pollingIntervalInMinutes * time.Minute***REMOVED***.
		Predicate(func(getClusterResponse *cmv1.ClusterGetResponse***REMOVED*** bool {
			object = getClusterResponse.Body(***REMOVED***
			tflog.Debug(ctx, "polled cluster state", map[string]interface{}{
				"state": object.State(***REMOVED***,
	***REMOVED******REMOVED***
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
