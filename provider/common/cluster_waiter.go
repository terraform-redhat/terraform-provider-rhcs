package common

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const pollingIntervalInMinutes = 2

//go:generate mockgen -source=cluster_waiter.go -package=common -destination=mock_clusterwait.go
type ClusterWait interface {
	WaitForClusterToBeReady(ctx context.Context, clusterId string, waitTimeoutMin int64) (*cmv1.Cluster, error)
	WaitForStdComputeNodesToBeReady(ctx context.Context, clusterId string, waitTimeoutMin int64) (*cmv1.Cluster, error)
}

type DefaultClusterWait struct {
	collection *cmv1.ClustersClient
	connection *sdk.Connection
}

func NewClusterWait(collection *cmv1.ClustersClient, connection *sdk.Connection) ClusterWait {
	return &DefaultClusterWait{
		collection: collection,
		connection: connection,
	}
}

func (dw *DefaultClusterWait) WaitForStdComputeNodesToBeReady(ctx context.Context, clusterId string, waitTimeoutMin int64) (*cmv1.Cluster, error) {
	resource := dw.collection.Cluster(clusterId)
	resp, err := resource.Get().SendContext(ctx)
	if err != nil && resp.Status() == http.StatusNotFound {
		message := fmt.Sprintf("Failed to get Cluster '%s', with error: %v", clusterId, err)
		tflog.Error(ctx, message)
		return nil, fmt.Errorf(message)
	}

	tflog.Info(ctx, fmt.Sprintf("WaitForStdComputeNodesToBeReady: Cluster '%s' is expected to initiate with %d worker replicas."+
		" Waiting for the current amount of replicas to reach disired value - timeout %d minutes",
		clusterId, resp.Body().Nodes().Compute(), waitTimeoutMin))

	if resp.Body().Nodes().Compute() == resp.Body().Status().CurrentCompute() {
		tflog.Info(ctx, fmt.Sprintf("WaitForStdComputeNodesToBeReady: Wait done for cluster '%s' with %d/%d", clusterId,
			resp.Body().Nodes().Compute(), resp.Body().Status().CurrentCompute()))
		return resp.Body(), nil
	}

	backoffAttempts := 3
	backoffSleep := 30 * time.Second
	var cluster *cmv1.Cluster
	for cluster == nil {
		tflog.Debug(ctx, fmt.Sprintf("Updating tokens for cluster %s", clusterId))
		dw.connection.Tokens()
		cluster, err = pollClusterCurrentCompute(clusterId, ctx, waitTimeoutMin, dw.collection)
		if err != nil {
			backoffAttempts--
			if backoffAttempts == 0 {
				return nil, fmt.Errorf("polling cluster state failed with error %v", err)
			}
			time.Sleep(backoffSleep)
		}
	}
	tflog.Info(ctx, fmt.Sprintf("WaitForStdComputeNodesToBeReady: Wait done for cluster '%s' with %d/%d", clusterId,
		cluster.Nodes().Compute(), cluster.Status().CurrentCompute()))
	if cluster.Nodes().Compute() != cluster.Status().CurrentCompute() {
		return cluster, fmt.Errorf("cluster did not reach the desired amount of compute nodes within the timeout")
	}
	return cluster, nil
}

func (dw *DefaultClusterWait) WaitForClusterToBeReady(ctx context.Context, clusterId string, waitTimeoutMin int64) (*cmv1.Cluster, error) {
	resource := dw.collection.Cluster(clusterId)

	// First try to get the cluster and check its state
	// Return an error in case:
	// * Cluster not found
	// * Cluster found but its state is "ERROR" or "UNINSTALLING" (will never become to "READY")
	// In case the state is "READY" return the cluster
	resp, err := resource.Get().SendContext(ctx)
	if err != nil && resp.Status() == http.StatusNotFound {
		message := fmt.Sprintf("Failed to get Cluster '%s', with error: %v", clusterId, err)
		tflog.Error(ctx, message)
		return nil, fmt.Errorf(message)
	}
	currentState := resp.Body().State()
	if currentState == cmv1.ClusterStateError || currentState == cmv1.ClusterStateUninstalling {
		message := fmt.Sprintf("Cluster '%s' is in state '%s' and will not become ready", clusterId, currentState)
		tflog.Error(ctx, message)
		return resp.Body(), fmt.Errorf(message)
	}
	if currentState == cmv1.ClusterStateReady {
		tflog.Info(ctx, fmt.Sprintf("WaitForClusterToBeReady: Cluster '%s' is with state \"READY\"", clusterId))
		return resp.Body(), nil
	}

	tflog.Info(ctx, fmt.Sprintf("WaitForClusterToBeReady: Cluster '%s' is with state '%s', Wait for the state to become 'READY' with timeout %d minutes",
		clusterId, currentState, waitTimeoutMin))

	backoffAttempts := 3
	backoffSleep := 30 * time.Second
	var cluster *cmv1.Cluster
	for cluster == nil {
		tflog.Debug(ctx, fmt.Sprintf("Updating tokens for cluster %s", clusterId))
		dw.connection.Tokens()
		cluster, err = pollClusterState(clusterId, ctx, waitTimeoutMin, dw.collection)
		if err != nil {
			backoffAttempts--
			if backoffAttempts == 0 {
				return nil, fmt.Errorf("polling cluster state failed with error %v", err)
			}
			time.Sleep(backoffSleep)
		}
	}

	tflog.Info(ctx, fmt.Sprintf("WaitForClusterToBeReady: Wait done for cluster '%s' with state '%s'", clusterId, currentState))

	// If Cluster is ready without ERROR
	// Otherwise return with ERROR
	if cluster.State() == cmv1.ClusterStateReady {
		return cluster, nil
	}
	return cluster, fmt.Errorf("cluster '%s' is in state '%s'", clusterId, cluster.State())
}

func pollClusterCurrentCompute(clusterId string, ctx context.Context, timeout int64, clusterCollection *cmv1.ClustersClient) (*cmv1.Cluster, error) {
	client := clusterCollection.Cluster(clusterId)
	var object *cmv1.Cluster
	pollCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cancel()
	_, err := client.Poll().
		Interval(pollingIntervalInMinutes * time.Minute).
		Predicate(func(getClusterResponse *cmv1.ClusterGetResponse) bool {
			object = getClusterResponse.Body()
			tflog.Debug(ctx, "polled cluster compute", map[string]interface{}{
				"currentCompute": object.Status().CurrentCompute(),
			})
			switch object.Status().CurrentCompute() {
			case object.Nodes().Compute():
				return true
			}
			return false
		}).
		StartContext(pollCtx)
	if err != nil {
		tflog.Error(ctx, "Failed polling cluster compute")
		return nil, err
	}

	return object, nil
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

func ValidateTimeout(timeOut *int64, defaultTimeout int64) (*int64, error) {
	if timeOut == nil {
		return &defaultTimeout, nil
	}
	if *timeOut <= 0 {
		return nil, fmt.Errorf("timeout must be greater than 0 minutes")
	}
	return timeOut, nil
}
