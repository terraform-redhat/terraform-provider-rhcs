package common

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const pollingIntervalInMinutes = 2

//go:generate mockgen -source=cluster_waiter.go -package=common -destination=mock_clusterwait.go
type ClusterWait interface {
	WaitForClusterToBeReady(ctx context.Context, clusterId string) error
	RetryClusterReadiness(ctx context.Context, clusterId string, attempts int, sleep time.Duration, timeout int64) (*cmv1.Cluster, error)
}

type DefaultClusterWait struct {
	collection *cmv1.ClustersClient
}

func NewClusterWait(collection *cmv1.ClustersClient) ClusterWait {
	return &DefaultClusterWait{collection: collection}
}

func (dw *DefaultClusterWait) RetryClusterReadiness(ctx context.Context, clusterId string, attempts int, sleep time.Duration, timeout int64) (*cmv1.Cluster, error) {
	object, err := pollClusterState(clusterId, ctx, timeout, dw.collection)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return dw.RetryClusterReadiness(ctx, clusterId, attempts, 2*sleep, timeout)
		}
		return nil, fmt.Errorf("polling cluster state failed with error %v", err)
	}

	if object.State() == cmv1.ClusterStateError || object.State() == cmv1.ClusterStateUninstalling {
		return object, fmt.Errorf("cluster '%s' is in state '%s'", clusterId, object.State())
	}

	return object, nil
}

func (dw *DefaultClusterWait) WaitForClusterToBeReady(ctx context.Context, clusterId string) error {
	resource := dw.collection.Cluster(clusterId)
	// We expect the cluster to already exist
	// Try to get it and if result with NotFound error, return error to user
	resp, err := resource.Get().SendContext(ctx)
	if err != nil && resp.Status() == http.StatusNotFound {
		message := fmt.Sprintf("Cluster '%s' not found, error: %v", clusterId, err)
		tflog.Error(ctx, message)
		return fmt.Errorf(message)
	}

	// Errored or uninstalling clusters will never become ready
	if resp.Body().State() == cmv1.ClusterStateError || resp.Body().State() == cmv1.ClusterStateUninstalling {
		message := fmt.Sprintf("Cluster '%s' is in state '%s' and will not become ready", clusterId, resp.Body().State())
		tflog.Error(ctx, message)
		return fmt.Errorf(message)
	}

	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()
	_, err = resource.Poll().
		Interval(30 * time.Second).
		Predicate(func(get *cmv1.ClusterGetResponse) bool {
			return get.Body().State() == cmv1.ClusterStateReady
		}).
		StartContext(pollCtx)
	if err != nil {
		return err
	}
	return nil
}

func pollClusterState(clusterId string, ctx context.Context, timeout int64, clusterCollection *cmv1.ClustersClient) (*cmv1.Cluster, error) {
	client := clusterCollection.Cluster(clusterId)
	var object *cmv1.Cluster
	pollCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cancel()
	_, err := client.Poll().
		Interval(pollingIntervalInMinutes * time.Minute).
		Predicate(func(getClusterResponse *cmv1.ClusterGetResponse) bool {
			object = getClusterResponse.Body()
			tflog.Debug(ctx, "polled cluster state", map[string]interface{}{
				"state": object.State(),
			})
			switch object.State() {
			case cmv1.ClusterStateReady,
				cmv1.ClusterStateError,
				cmv1.ClusterStateUninstalling:
				return true
			}
			return false
		}).
		StartContext(pollCtx)
	if err != nil {
		tflog.Error(ctx, "Failed polling cluster state")
		return nil, err
	}

	return object, nil
}
