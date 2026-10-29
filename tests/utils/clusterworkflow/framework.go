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
// Cleanup is deliberately separate because HCP and Seaplane have different
// dependency ordering and grace periods.
type Backend struct {
	Lifecycle exec.ClusterLifecycle
	Fetcher   ClusterIDFetcher
}

// SeaplaneOptions customizes a Seaplane cluster create workflow. Empty values
// leave the corresponding cluster argument unset.
type SeaplaneOptions struct {
	DNSReservationID string
	BaseDomainPrefix *string
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
	case constants.SEAPLANE.String():
		return newSeaplaneBackend(profile, workspace)
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for shared cluster workflow", clusterType.String())
	}
}

// NewSeaplaneWithOptions creates a Seaplane backend with optional DNS
// reservation settings. It is separate from New so HCP callers do not need to
// understand Seaplane-only cluster fields.
func NewSeaplaneWithOptions(
	profile profilehandler.ProfileHandler,
	workspace string,
	options SeaplaneOptions,
) (*Backend, error) {
	if profile.Profile().GetClusterType() != constants.SEAPLANE {
		return nil, fmt.Errorf("seaplane options require a seaplane profile")
	}
	return newSeaplaneBackendWithOptions(profile, workspace, options)
}

// NewFetcher creates only the existing-cluster lookup strategy. It does not
// initialize Terraform resources, so it is safe to use from BeforeSuite.
func NewFetcher(profile profilehandler.ProfileHandler) (ClusterIDFetcher, error) {
	clusterType := profile.Profile().GetClusterType()
	switch clusterType.String() {
	case constants.ROSA_HCP.String():
		return newHCPBackend(profile, "").Fetcher, nil
	case constants.SEAPLANE.String():
		return newSeaplaneFetcher(profile)
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for cluster ID lookup", clusterType.String())
	}
}

// NewExisting creates a backend-neutral lifecycle for an already-provisioned
// cluster. Unlike New, it does not create or apply any supporting resources.
// It is intended for day-2 workflows which operate on the cluster selected by
// the shared test-suite CLUSTER_ID.
func NewExisting(profile profilehandler.ProfileHandler) (*Backend, error) {
	clusterType := profile.Profile().GetClusterType()
	switch clusterType.String() {
	case constants.ROSA_HCP.String():
		return newHCPBackend(profile, ""), nil
	case constants.SEAPLANE.String():
		return newExistingSeaplaneBackend(profile)
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for existing cluster workflow", clusterType.String())
	}
}

// NewDestroyer creates the backend-specific destroy strategy without applying
// or recreating any Terraform resources.
func NewDestroyer(profile profilehandler.ProfileHandler, token, workspace string) (ClusterDestroyer, error) {
	clusterType := profile.Profile().GetClusterType()
	switch clusterType.String() {
	case constants.ROSA_HCP.String():
		return hcpDestroyer{profile: profile, token: token}, nil
	case constants.SEAPLANE.String():
		return newSeaplaneDestroyer(profile, workspace)
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for destroy workflow", clusterType.String())
	}
}

// NewNodePool creates the backend-neutral node-pool lifecycle for the selected
// profile.
func NewNodePool(
	profile profilehandler.ProfileHandler,
	clusterID, nodePoolName string,
) (exec.NodePoolLifecycle, error) {
	clusterType := profile.Profile().GetClusterType()
	switch clusterType.String() {
	case constants.ROSA_HCP.String():
		return newHCPNodePoolReplica(profile, clusterID, nodePoolName)
	case constants.SEAPLANE.String():
		clusterName := config.GetRHCSClusterName()
		if clusterName == "" {
			clusterName = profile.Profile().GetName()
		}
		return NewSeaplaneNodePoolReplica(profile, clusterName, nodePoolName, clusterID, profile.Profile().GetName())
	default:
		return nil, fmt.Errorf("unsupported cluster type %q for nodepool workflow", clusterType.String())
	}
}
