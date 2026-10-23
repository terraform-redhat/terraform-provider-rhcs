// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"fmt"
)

// NodePoolSpec contains the node-pool fields supported by the shared
// create/update workflow. Nil fields are omitted during Update and therefore
// preserve their current value.
type NodePoolSpec struct {
	Replicas                      *int
	Autoscaling                   *NodePoolAutoscaling
	EC2InstanceMetadataHttpTokens *string
	Tags                          *map[string]string
}

// NodePoolState is the backend-neutral representation returned by Read.
type NodePoolState struct {
	Replicas                      *int
	Autoscaling                   *NodePoolAutoscalingState
	EC2InstanceMetadataHttpTokens string
	Tags                          map[string]string
}

type NodePoolAutoscalingState struct {
	Enabled     bool
	MinReplicas int
	MaxReplicas int
}

// NodePoolLifecycle is the common node-pool lifecycle for HCP and Seaplane.
// The backend adapters retain ownership of Terraform and API details.
type NodePoolLifecycle interface {
	Create(spec NodePoolSpec) error
	Update(spec NodePoolSpec) error
	WaitReady(ctx context.Context) error
	Read(ctx context.Context) (NodePoolState, error)
	WaitTagsApplied(ctx context.Context, tags map[string]string) error
	Destroy() error
	WaitDeleted(ctx context.Context) error
}

// ScaleNodePool verifies the mutable fixed-replica operation shared by HCP and
// Seaplane. Destruction remains caller-owned.
func ScaleNodePool(ctx context.Context, lifecycle NodePoolLifecycle, initial, updated int) error {
	if err := lifecycle.Create(NodePoolSpec{Replicas: &initial}); err != nil {
		return err
	}
	if err := lifecycle.WaitReady(ctx); err != nil {
		return err
	}
	if err := lifecycle.Update(NodePoolSpec{Replicas: &updated}); err != nil {
		return err
	}
	state, err := lifecycle.Read(ctx)
	if err != nil {
		return err
	}
	if state.Replicas == nil || *state.Replicas != updated {
		actual := 0
		if state.Replicas != nil {
			actual = *state.Replicas
		}
		return fmt.Errorf("nodepool replicas: got %d, want %d", actual, updated)
	}
	return nil
}
