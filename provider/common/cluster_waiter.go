package common

import (
	"context"
	"time"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func WaitTillClusterReady(ctx context.Context, collection *cmv1.ClustersClient, clusterId string) error {
	resource := collection.Cluster(clusterId)
	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()
	_, err := resource.Poll().
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
