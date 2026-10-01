// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"fmt"
)

// NodePoolReplicaLifecycle is the backend-neutral portion of the day-2
// replica workflow. Backend strategies own Terraform arguments and API types.
type NodePoolReplicaLifecycle interface {
	Create(replicas int) error
	WaitReady(ctx context.Context) error
	UpdateReplicas(replicas int) error
	Replicas(ctx context.Context) (int, error)
	Destroy() error
	WaitDeleted(ctx context.Context) error
}

// ScaleNodePool verifies the mutable fixed-replica operation shared by HCP and
// HyperFleet. Destruction is intentionally caller-owned so tests can register
// it with DeferCleanup and apply backend-specific deletion-wait policy.
func ScaleNodePool(ctx context.Context, lifecycle NodePoolReplicaLifecycle, initial, updated int) error {
	if err := lifecycle.Create(initial); err != nil {
		return err
	}
	if err := lifecycle.WaitReady(ctx); err != nil {
		return err
	}
	if err := lifecycle.UpdateReplicas(updated); err != nil {
		return err
	}
	actual, err := lifecycle.Replicas(ctx)
	if err != nil {
		return err
	}
	if actual != updated {
		return fmt.Errorf("nodepool replicas: got %d, want %d", actual, updated)
	}
	return nil
}
