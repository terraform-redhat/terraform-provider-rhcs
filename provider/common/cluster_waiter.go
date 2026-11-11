// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

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

const (
	pollingIntervalInMinutes = 2
	waitRetryInterval        = 30 * time.Second
)

//go:generate mockgen -source=cluster_waiter.go -package=common -destination=mock_clusterwait.go
type ClusterWait interface {
	WaitForClusterToBeReady(ctx context.Context, clusterId string, waitTimeoutMin int64) (*cmv1.Cluster, error)
	WaitForClusterToBeReadyWithURLs(ctx context.Context, clusterId string, waitTimeoutMin int64) (*cmv1.Cluster, error)
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
		return nil, fmt.Errorf("%s", message)
	}

	tflog.Info(ctx, fmt.Sprintf("WaitForStdComputeNodesToBeReady: Cluster '%s' is expected to initiate with %d worker replicas."+
		" Waiting for the current amount of replicas to reach disired value - timeout %d minutes",
		clusterId, resp.Body().Nodes().Compute(), waitTimeoutMin))

	if resp.Body().Nodes().Compute() == resp.Body().Status().CurrentCompute() {
		tflog.Info(ctx, fmt.Sprintf("WaitForStdComputeNodesToBeReady: Wait done for cluster '%s' with %d/%d", clusterId,
			resp.Body().Nodes().Compute(), resp.Body().Status().CurrentCompute()))
		return resp.Body(), nil
	}

	cluster, err := dw.pollClusterWithRetry(
		ctx,
		clusterId,
		time.Duration(waitTimeoutMin)*time.Minute,
		func(pollCtx context.Context) (*cmv1.Cluster, error) {
			return pollClusterCurrentCompute(clusterId, pollCtx, dw.collection)
		},
	)
	if err != nil {
		return nil, err
	}
	tflog.Info(ctx, fmt.Sprintf("WaitForStdComputeNodesToBeReady: Wait done for cluster '%s' with %d/%d", clusterId,
		cluster.Nodes().Compute(), cluster.Status().CurrentCompute()))
	if cluster.Nodes().Compute() != cluster.Status().CurrentCompute() {
		return cluster, fmt.Errorf("cluster did not reach the desired amount of compute nodes within the timeout")
	}
	return cluster, nil
}

// WaitForClusterToBeReady waits until the cluster state is READY.
func (dw *DefaultClusterWait) WaitForClusterToBeReady(ctx context.Context, clusterId string, waitTimeoutMin int64) (*cmv1.Cluster, error) {
	return dw.waitForClusterToBeReady(ctx, clusterId, waitTimeoutMin, false)
}

// WaitForClusterToBeReadyWithURLs waits until the cluster state is READY and the cluster
// reports both its API URL and its console URL. Readiness and the URLs are checked by the
// same polling loop, at the same interval, within the single waitTimeoutMin budget.
func (dw *DefaultClusterWait) WaitForClusterToBeReadyWithURLs(ctx context.Context, clusterId string,
	waitTimeoutMin int64) (*cmv1.Cluster, error) {
	return dw.waitForClusterToBeReady(ctx, clusterId, waitTimeoutMin, true)
}

// clusterHasAPIAndConsoleURLs reports whether the cluster has non-empty API and console URLs.
func clusterHasAPIAndConsoleURLs(cluster *cmv1.Cluster) bool {
	return cluster.API().URL() != "" && cluster.Console().URL() != ""
}

func (dw *DefaultClusterWait) waitForClusterToBeReady(ctx context.Context, clusterId string, waitTimeoutMin int64,
	requireURLs bool) (*cmv1.Cluster, error) {
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
		return nil, fmt.Errorf("%s", message)
	}
	currentState := resp.Body().State()
	if currentState == cmv1.ClusterStateError || currentState == cmv1.ClusterStateUninstalling {
		message := fmt.Sprintf("Cluster '%s' is in state '%s' and will not become ready", clusterId, currentState)
		tflog.Error(ctx, message)
		return resp.Body(), fmt.Errorf("%s", message)
	}
	if currentState == cmv1.ClusterStateReady {
		if !requireURLs {
			tflog.Info(ctx, fmt.Sprintf("WaitForClusterToBeReady: Cluster '%s' is with state \"READY\"", clusterId))
			return resp.Body(), nil
		}
		if clusterHasAPIAndConsoleURLs(resp.Body()) {
			tflog.Info(ctx, fmt.Sprintf("WaitForClusterToBeReadyWithURLs: Cluster '%s' is with state \"READY\" "+
				"and reports its API and console URLs", clusterId))
			return resp.Body(), nil
		}
	}

	if requireURLs {
		tflog.Info(ctx, fmt.Sprintf("WaitForClusterToBeReadyWithURLs: Cluster '%s' is with state '%s' "+
			"(API URL set: %t, console URL set: %t), Wait for the state to become 'READY' "+
			"with both URLs set with timeout %d minutes",
			clusterId, currentState, resp.Body().API().URL() != "", resp.Body().Console().URL() != "",
			waitTimeoutMin))
	} else {
		tflog.Info(ctx, fmt.Sprintf("WaitForClusterToBeReady: Cluster '%s' is with state '%s', "+
			"Wait for the state to become 'READY' with timeout %d minutes",
			clusterId, currentState, waitTimeoutMin))
	}

	cluster, err := dw.pollClusterWithRetry(
		ctx,
		clusterId,
		time.Duration(waitTimeoutMin)*time.Minute,
		func(pollCtx context.Context) (*cmv1.Cluster, error) {
			return pollClusterState(clusterId, pollCtx, dw.collection, requireURLs)
		},
	)
	if err != nil {
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("WaitForClusterToBeReady: Wait done for cluster '%s' with state '%s'", clusterId, currentState))

	// If Cluster is ready without ERROR
	// Otherwise return with ERROR
	if cluster.State() == cmv1.ClusterStateReady {
		if !requireURLs || clusterHasAPIAndConsoleURLs(cluster) {
			return cluster, nil
		}
		// The poll stops before the deadline when another interval no longer fits in the
		// budget, returning the last cluster seen while its URLs were still missing.
		return cluster, fmt.Errorf("cluster '%s' is in state '%s', but its API URL and console URL "+
			"were not both available before the wait timeout of %d minutes",
			clusterId, cluster.State(), waitTimeoutMin)
	}
	return cluster, fmt.Errorf("cluster '%s' is in state '%s'", clusterId, cluster.State())
}

func (dw *DefaultClusterWait) pollClusterWithRetry(
	ctx context.Context,
	clusterId string,
	waitTimeout time.Duration,
	poll func(context.Context) (*cmv1.Cluster, error),
) (*cmv1.Cluster, error) {
	pollCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	// The OCM SDK retries only selected HTTP and transport failures. Preserve the
	// provider's operation-level retry for errors returned by Poll, including
	// transient network failures not covered by the SDK. All attempts share one
	// deadline, retaining the existing retry behavior while bounding the total wait.
	backoffAttempts := 3
	for {
		tflog.Debug(pollCtx, fmt.Sprintf("Updating tokens for cluster %s", clusterId))
		// Preserve the explicit token refresh required for GovCloud/FedRAMP, where
		// token lifetimes can be shorter than cluster provisioning.
		if _, _, err := dw.connection.TokensContext(pollCtx); err != nil {
			return nil, fmt.Errorf("refreshing tokens for cluster %s: %w", clusterId, err)
		}
		cluster, err := poll(pollCtx)
		if err == nil {
			return cluster, nil
		}

		backoffAttempts--
		if ctxErr := pollCtx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("polling cluster state failed: %w", ctxErr)
		}
		if backoffAttempts == 0 {
			return nil, fmt.Errorf("polling cluster state failed: %w", err)
		}
		select {
		case <-time.After(waitRetryInterval):
		case <-pollCtx.Done():
			return nil, fmt.Errorf("polling cluster state failed: %w", pollCtx.Err())
		}
	}
}

func pollClusterCurrentCompute(
	clusterId string,
	ctx context.Context,
	clusterCollection *cmv1.ClustersClient,
) (*cmv1.Cluster, error) {
	client := clusterCollection.Cluster(clusterId)
	var object *cmv1.Cluster
	_, err := client.Poll().
		Interval(pollingIntervalInMinutes * time.Minute).
		Predicate(func(getClusterResponse *cmv1.ClusterGetResponse) bool {
			object = getClusterResponse.Body()
			tflog.Debug(ctx, "polled cluster compute", map[string]any{
				"currentCompute": object.Status().CurrentCompute(),
			})
			switch object.Status().CurrentCompute() {
			case object.Nodes().Compute():
				return true
			}
			return false
		}).
		StartContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed polling cluster compute: %v", err))
		return nil, err
	}

	return object, nil
}

// pollClusterState polls until the cluster reaches a terminal state (READY, ERROR or
// UNINSTALLING). When requireURLs is true, READY is only accepted once the cluster also
// reports both its API URL and its console URL.
func pollClusterState(
	clusterId string,
	ctx context.Context,
	clusterCollection *cmv1.ClustersClient,
	requireURLs bool,
) (*cmv1.Cluster, error) {
	client := clusterCollection.Cluster(clusterId)
	var object *cmv1.Cluster
	_, err := client.Poll().
		Interval(pollingIntervalInMinutes * time.Minute).
		Predicate(func(getClusterResponse *cmv1.ClusterGetResponse) bool {
			object = getClusterResponse.Body()
			fields := map[string]any{
				"state": object.State(),
			}
			if requireURLs {
				fields["api_url_set"] = object.API().URL() != ""
				fields["console_url_set"] = object.Console().URL() != ""
			}
			tflog.Debug(ctx, "polled cluster state", fields)
			switch object.State() {
			case cmv1.ClusterStateReady:
				return !requireURLs || clusterHasAPIAndConsoleURLs(object)
			case cmv1.ClusterStateError,
				cmv1.ClusterStateUninstalling:
				return true
			}
			return false
		}).
		StartContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed polling cluster state: %v", err))
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
