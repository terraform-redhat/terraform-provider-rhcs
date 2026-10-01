// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

// Package clusterworkflow contains backend adapters for shared cluster e2e
// workflows.
package clusterworkflow

import (
	"context"
	"fmt"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

// Backend is a shared cluster workflow plus backend-specific teardown.
// Cleanup is deliberately separate because HCP and HyperFleet have different
// dependency ordering and grace periods.
type Backend struct {
	Lifecycle exec.ClusterReadinessLifecycle
	Fetcher   ClusterIDFetcher
}

// ClusterIDFetcher obtains the existing cluster selected by the shared
// CLUSTER_ID environment variable, using the API belonging to the profile.
type ClusterIDFetcher interface {
	FetchClusterID(ctx context.Context) (string, error)
}

// ClusterDestroyer removes an existing cluster and its backend-owned
// dependencies. It is separate from the create/readiness lifecycle so create
// and day-2 tests can retain their cluster until the destroy phase.
type ClusterDestroyer interface {
	Destroy(ctx context.Context) error
}

// New creates the backend selected by the profile's cluster_type.
func New(profile profilehandler.ProfileHandler, token, workspace string) (*Backend, error) {
	clusterType := profile.Profile().GetClusterType()
	switch clusterType.String() {
	case constants.ROSA_HCP.String():
		return newHCPBackend(profile, token), nil
	case constants.HYPERFLEET.String():
		return newHyperFleetBackend(profile, workspace)
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for shared cluster workflow", clusterType.String())
	}
}

// NewFetcher creates only the existing-cluster lookup strategy. It does not
// initialize Terraform resources, so it is safe to use from BeforeSuite.
func NewFetcher(profile profilehandler.ProfileHandler) (ClusterIDFetcher, error) {
	clusterType := profile.Profile().GetClusterType()
	switch clusterType.String() {
	case constants.ROSA_HCP.String():
		return newHCPBackend(profile, "").Fetcher, nil
	case constants.HYPERFLEET.String():
		return newHyperFleetFetcher(profile)
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for cluster ID lookup", clusterType.String())
	}
}

// NewDestroyer creates the backend-specific destroy strategy without applying
// or recreating any Terraform resources.
func NewDestroyer(profile profilehandler.ProfileHandler, token, workspace string) (ClusterDestroyer, error) {
	clusterType := profile.Profile().GetClusterType()
	switch clusterType.String() {
	case constants.ROSA_HCP.String():
		return hcpDestroyer{profile: profile, token: token}, nil
	case constants.HYPERFLEET.String():
		return newHyperFleetDestroyer(profile, workspace)
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for destroy workflow", clusterType.String())
	}
}

// NewNodePoolReplica creates the fixed-replica day-2 strategy for the selected
// profile backend.
func NewNodePoolReplica(
	profile profilehandler.ProfileHandler,
	clusterID, nodePoolName string,
) (exec.NodePoolReplicaLifecycle, error) {
	clusterType := profile.Profile().GetClusterType()
	switch clusterType.String() {
	case constants.ROSA_HCP.String():
		return newHCPNodePoolReplica(profile, clusterID, nodePoolName)
	case constants.HYPERFLEET.String():
		clusterName := config.GetRHCSClusterName()
		if clusterName == "" {
			clusterName = profile.Profile().GetName()
		}
		return NewHyperFleetNodePoolReplica(clusterName, nodePoolName, clusterID, profile.Profile().GetName())
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for nodepool replica workflow", clusterType.String())
	}
}
