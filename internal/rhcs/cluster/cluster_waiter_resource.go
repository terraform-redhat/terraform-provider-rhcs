package cluster

import (
	"context"
	"fmt"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/cluster/clusterschema"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

type Waiter interface {
	Ready() bool
}

const (
	NonPositiveTimeoutSummary = "Can't poll cluster state with a non-positive timeout"
	NonPositiveTimeoutFormat  = "Can't poll state of cluster with identifier '%s', the timeout that was set is not a positive number"
	PollingIntervalInMinutes  = 2
	DefaultAttempts           = 3
	DefaultSleepInSecond      = 30
)

func ResourceClusterWaiter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterWaiterCreate,
		ReadContext:   resourceClusterWaiterRead,
		UpdateContext: resourceClusterWaiterUpdate,
		DeleteContext: resourceClusterWaiterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: clusterschema.ClusterWaiterFields(),
	}
}

func resourceClusterWaiterCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin create()")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	clusterState := clusterschema.ExpandClusterWaiterFromResourceData(resourceData)

	return pollingClusterStateAndUpdateResourceData(ctx, clusterState, clusterCollection, resourceData)
}

func pollingClusterStateAndUpdateResourceData(ctx context.Context, clusterState *clusterschema.ClusterWaiterState,
	clusterCollection *cmv1.ClustersClient, resourceData *schema.ResourceData) diag.Diagnostics {
	timeout := int64(clusterschema.DefaultTimeoutInMinutes)
	if clusterState.Timeout > 0 {
		timeout = clusterState.Timeout
	}
	state, err := common.RetryClusterReadiness(ctx, DefaultAttempts, DefaultSleepInSecond*time.Second,
		clusterCollection, clusterState.Cluster, timeout)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't poll cluster state (create resource)",
				Detail:   err.Error(),
			}}
	}
	clusterIsReady := false
	if state == cmv1.ClusterStateReady {
		clusterIsReady = true
	}
	clusterState.Ready = clusterIsReady

	clusterWaiterToResourceData(clusterState, resourceData)
	return nil
}

func clusterWaiterToResourceData(clusterState *clusterschema.ClusterWaiterState, resourceData *schema.ResourceData) {
	resourceData.SetId(clusterState.Cluster)
	resourceData.Set("ready", clusterState.Ready)
}

func resourceClusterWaiterRead(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin read()")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	clusterState := clusterschema.ExpandClusterWaiterFromResourceData(resourceData)

	get, err := clusterCollection.Cluster(clusterState.Cluster).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		resourceID := clusterState.Cluster
		summary := fmt.Sprintf("cluster (%s) not found, removing from state", resourceID)
		tflog.Warn(ctx, summary)
		resourceData.SetId("")
		return []diag.Diagnostic{
			{
				Severity: diag.Warning,
				Summary:  summary,
				Detail: fmt.Sprintf(
					"cluster (%s) not found, removing from state",
					resourceID),
			}}
	} else if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't find cluster",
				Detail: fmt.Sprintf(
					"Can't find cluster with identifier '%s': %v",
					clusterState.Cluster, err),
			}}
	}
	object := get.Body()

	clusterIsReady := false
	if object.State() == cmv1.ClusterStateReady {
		clusterIsReady = true
	}
	clusterState.Ready = clusterIsReady

	clusterWaiterToResourceData(clusterState, resourceData)
	return nil
}

func resourceClusterWaiterUpdate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin update()")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	clusterState := clusterschema.ExpandClusterWaiterFromResourceData(resourceData)

	return pollingClusterStateAndUpdateResourceData(ctx, clusterState, clusterCollection, resourceData)
}

func resourceClusterWaiterDelete(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin delete()")
	resourceData.SetId("")
	return nil
}
