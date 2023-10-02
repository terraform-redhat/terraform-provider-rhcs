package common

***REMOVED***
	"context"
***REMOVED***
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

const pollingIntervalInMinutes = 2

func WaitTillClusterReady(ctx context.Context, collection *cmv1.ClustersClient, clusterId string***REMOVED*** error {
	resource := collection.Cluster(clusterId***REMOVED***
	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour***REMOVED***
	defer cancel(***REMOVED***
	_, err := resource.Poll(***REMOVED***.
		Interval(30 * time.Second***REMOVED***.
		Predicate(func(get *cmv1.ClusterGetResponse***REMOVED*** bool {
			return get.Body(***REMOVED***.State(***REMOVED*** == cmv1.ClusterStateReady
***REMOVED******REMOVED***.
		StartContext(pollCtx***REMOVED***
	if err != nil {
		return err
	}
	return nil
}

func pollClusterState(clusterId string, ctx context.Context, timeout int64, clusterCollection *cmv1.ClustersClient***REMOVED*** (*cmv1.Cluster, error***REMOVED*** {
	client := clusterCollection.Cluster(clusterId***REMOVED***
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
		tflog.Error(ctx, "Failed polling cluster state"***REMOVED***
		return nil, err
	}

	return object, nil
}

func RetryClusterReadiness(attempts int, sleep time.Duration, clusterId string, ctx context.Context, timeout int64, clusterCollection *cmv1.ClustersClient***REMOVED*** (*cmv1.Cluster, error***REMOVED*** {
	object, err := pollClusterState(clusterId, ctx, timeout, clusterCollection***REMOVED***
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep***REMOVED***
			return RetryClusterReadiness(attempts, 2*sleep, clusterId, ctx, timeout, clusterCollection***REMOVED***
***REMOVED***
		return nil, fmt.Errorf("polling cluster state failed with error %v", err***REMOVED***
	}

	if object.State(***REMOVED*** == cmv1.ClusterStateError {
		return object, fmt.Errorf("cluster creation failed"***REMOVED***
	}

	return object, nil
}
