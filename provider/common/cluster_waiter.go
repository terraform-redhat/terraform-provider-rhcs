package common

***REMOVED***
	"context"
	"time"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

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
