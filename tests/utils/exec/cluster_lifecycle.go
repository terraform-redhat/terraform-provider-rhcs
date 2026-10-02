// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"fmt"
)

// ClusterReadinessLifecycle contains the provider-specific operations needed by
// shared create/readiness e2e workflows. HCP and HyperFleet adapters can
// implement this without sharing Terraform argument structs or API clients.
type ClusterReadinessLifecycle interface {
	Create() (clusterID string, err error)
	WaitReady(ctx context.Context, clusterID string) error
}

// ClusterReadinessLifecycleFuncs adapts existing test code to the shared
// lifecycle without requiring it to introduce a backend-specific type.
type ClusterReadinessLifecycleFuncs struct {
	CreateFunc    func() (string, error)
	WaitReadyFunc func(context.Context, string) error
}

func (f ClusterReadinessLifecycleFuncs) Create() (string, error) {
	return f.CreateFunc()
}

func (f ClusterReadinessLifecycleFuncs) WaitReady(ctx context.Context, clusterID string) error {
	return f.WaitReadyFunc(ctx, clusterID)
}

// CreateAndWaitReady creates a cluster and waits for the backend-specific
// readiness condition. Teardown remains the caller's responsibility so each
// backend can preserve its own cleanup ordering and grace periods.
func CreateAndWaitReady(ctx context.Context, lifecycle ClusterReadinessLifecycle) (string, error) {
	clusterID, err := lifecycle.Create()
	if err != nil {
		return "", fmt.Errorf("creating cluster: %w", err)
	}
	if clusterID == "" {
		return "", fmt.Errorf("creating cluster: backend returned an empty cluster ID")
	}
	if err := lifecycle.WaitReady(ctx, clusterID); err != nil {
		return clusterID, fmt.Errorf("waiting for cluster %q to become ready: %w", clusterID, err)
	}
	return clusterID, nil
}
