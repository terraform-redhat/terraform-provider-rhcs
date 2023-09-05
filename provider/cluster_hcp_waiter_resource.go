package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

//type Waiter interface {
//	Ready() bool
//}

type ClusterHCPWaiterResourceType struct {
}

type ClusterHCPWaiterResource struct {
	collection *cmv1.ClustersClient
}

const (
	hcpDefaultTimeoutInMinutes   = int64(60)
	hcpNonPositiveTimeoutSummary = "Can't poll cluster state with a non-positive timeout"
	hcpNonPositiveTimeoutFormat  = "Can't poll state of cluster with identifier '%s', the timeout that was set is not a positive number"
	hclPollingIntervalInMinutes  = 2
)

func (t *ClusterHCPWaiterResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "Wait Cluster Resource To be Ready when current compute is equals to desired compute",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
			},
			"timeout": {
				Description: "An optional timeout till the cluster current compute is equals to desired compute. The timeout value should be in minutes." +
					" the default value is 60 minutes",
				Type:       types.Int64Type,
				Optional:   true,
				Validators: hcpTimeoutValidators(),
			},
			"ready": {
				Description: "Whether the cluster current compute is equals to desired compute",
				Type:        types.BoolType,
				Computed:    true,
			},
		},
	}
	return
}

func (t *ClusterHCPWaiterResourceType) NewResource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.Resource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation: use it directly when needed.
	parent := p.(*Provider)

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt().V1().Clusters()

	// Create the resource:
	result = &ClusterHCPWaiterResource{
		collection: collection,
	}
	return
}

func (r *ClusterHCPWaiterResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	// Get the plan:
	state := &ClusterHCPWaiterState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	state, err := r.startPolling(ctx, state)

	if err != nil {
		response.Diagnostics.AddError("Can't poll cluster state (create resource)", err.Error())
		return
	}

	// Save the state:
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterHCPWaiterResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse) {
	// Do Nothing
}

func (r *ClusterHCPWaiterResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse) {
	plan := &ClusterHCPWaiterState{}
	diags := request.Plan.Get(ctx, plan)
	_ = request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	state, err := r.startPolling(ctx, plan)

	if err != nil {
		response.Diagnostics.AddError("Can't poll cluster state (update resource)", err.Error())
		return
	}

	// Save the state:
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterHCPWaiterResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse) {
	response.State.RemoveResource(ctx)
}

func (r *ClusterHCPWaiterResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	// Do Nothing
}

func (r *ClusterHCPWaiterResource) startPolling(ctx context.Context, state *ClusterHCPWaiterState) (*ClusterHCPWaiterState, error) {
	state.Ready = types.Bool{
		Value: false,
	}

	timeout := hcpDefaultTimeoutInMinutes
	if !state.Timeout.Unknown && !state.Timeout.Null {
		timeout = state.Timeout.Value
	}

	// Wait till the cluster is ready:
	object, err := r.retryClusterReadiness(3, 30*time.Second, state.Cluster.Value, ctx, timeout)
	if err != nil {
		return state, fmt.Errorf(
			"Can't poll state of cluster with identifier '%s': %v",
			state.Cluster.Value, err,
		)
	}
	isClusterReady := false
	if object.Status().CurrentCompute() == object.Nodes().Compute() {
		isClusterReady = true
	}

	state.Ready = types.Bool{
		Value: isClusterReady,
	}
	return state, nil
}

func (r *ClusterHCPWaiterResource) isClusterReady(clusterId string, ctx context.Context, timeout int64) (*cmv1.Cluster, error) {
	resource := r.collection.Cluster(clusterId)
	var object *cmv1.Cluster
	pollCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cancel()
	_, err := resource.Poll().
		Interval(hclPollingIntervalInMinutes * time.Minute).
		Predicate(func(getClusterResponse *cmv1.ClusterGetResponse) bool {
			object = getClusterResponse.Body()
			tflog.Debug(ctx, fmt.Sprintf("cluster current compute is %d", object.Status().CurrentCompute()))
			switch object.Status().CurrentCompute() {
			case object.Nodes().Compute():
				return true
			}
			return false
		}).
		StartContext(pollCtx)
	if err != nil {
		tflog.Error(ctx, "Can't poll cluster current compute")
		return nil, err
	}

	return object, err
}

func (r *ClusterHCPWaiterResource) retryClusterReadiness(attempts int, sleep time.Duration, clusterId string, ctx context.Context, timeout int64) (*cmv1.Cluster, error) {
	object, err := r.isClusterReady(clusterId, ctx, timeout)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return r.retryClusterReadiness(attempts, 2*sleep, clusterId, ctx, timeout)
		}
		return object, err
	}

	return object, nil
}

func hcpTimeoutValidators() []tfsdk.AttributeValidator {
	errSumm := "Invalid timeout configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Timeout must be a positive number",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				timeout := &types.Int64{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, timeout)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				if !timeout.Null && !timeout.Unknown && timeout.Value <= 0 {
					resp.Diagnostics.AddError(errSumm, "timeout must be positive")
				}
			},
		},
	}
}
