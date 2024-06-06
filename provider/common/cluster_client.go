package common

import (
	"context"
	"fmt"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

//go:generate mockgen -source=cluster_client.go -package=common -destination=mock_clusterclient.go
type ClusterClient interface {
	FetchCluster(ctx context.Context, clusterId string) (*cmv1.Cluster, error)
}

type DefaultClusterClient struct {
	client *cmv1.ClustersClient
}

func NewClusterClient(client *cmv1.ClustersClient) ClusterClient {
	return &DefaultClusterClient{client: client}
}

func (c *DefaultClusterClient) FetchCluster(ctx context.Context, clusterId string) (*cmv1.Cluster, error) {
	clusterResource := c.client.Cluster(clusterId)
	clusterResp, err := clusterResource.Get().SendContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Can't find cluster '%s': %v", clusterId, err)
	} else {
		return clusterResp.Body(), nil
	}
}
