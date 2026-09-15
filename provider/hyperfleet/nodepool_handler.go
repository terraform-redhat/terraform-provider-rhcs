/*
Copyright (c) 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hyperfleet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hfwrappers "github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

// workerInstanceProfileSuffix is appended to the operator roles prefix to form
// the worker IAM instance profile name. It mirrors the naming convention in the
// IAM manifests (`${operator_roles_prefix}-ROSA-Worker-Role`).
const workerInstanceProfileSuffix = "-ROSA-Worker-Role"

// NodePoolHandlerImpl is the concrete implementation of NodePoolHandler
type NodePoolHandlerImpl struct {
	client    hyperfleet.Interface
	accountID string
	callerARN string
}

// NewNodePoolHandler creates a new NodePoolHandler
func NewNodePoolHandler(client hyperfleet.Interface, accountID, callerARN string) NodePoolHandler {
	return &NodePoolHandlerImpl{
		client:    client,
		accountID: accountID,
		callerARN: callerARN,
	}
}

// PreExpand validates inputs and derives computed fields
func (h *NodePoolHandlerImpl) PreExpand(ctx context.Context, input *NodePoolState) diag.Diagnostics {
	var diags diag.Diagnostics

	// Validate required fields
	if input.Name.IsNull() || input.Name.ValueString() == "" {
		diags.AddError("name is required", "NodePool name must be specified")
		return diags
	}

	if input.ClusterName.IsNull() || input.ClusterName.ValueString() == "" {
		diags.AddError("cluster_name is required", "Cluster name must be specified to create a NodePool")
		return diags
	}

	return diags
}

// PostExpand processes inputs after pathbind.Expand to set SDK-specific fields
func (h *NodePoolHandlerImpl) PostExpand(ctx context.Context, input *NodePoolState, obj *v1alpha1.NodePool) diag.Diagnostics {
	var diags diag.Diagnostics

	// Ensure platform.type is set to AWS platform (enum constant)
	if obj.Spec.NodePool.Platform.Type == "" {
		obj.Spec.NodePool.Platform.Type = hypershiftv1beta1.AWSPlatform
	}

	// Compute the worker instance profile from the parent cluster's operator roles prefix
	// This is required for CAPA to correctly assign IAM permissions to worker nodes
	clusterName := input.ClusterName.ValueString()
	if clusterName != "" {
		cluster, err := h.client.HyperfleetV1alpha1().Clusters().Get(ctx, clusterName, hfwrappers.GetOptions{})
		if err != nil {
			diags.AddError(
				"Failed to get parent cluster",
				fmt.Sprintf("Cannot compute worker instance profile: failed to retrieve cluster %q: %v", clusterName, err),
			)
			return diags
		}

		if cluster.Spec.HostedCluster.Platform.AWS == nil {
			diags.AddError(
				"Parent cluster has no AWS configuration",
				fmt.Sprintf("Cannot compute worker instance profile: parent cluster %q has no AWS platform configuration", clusterName),
			)
			return diags
		}

		prefix, _ := prefixAndPartitionFromRolesRef(cluster.Spec.HostedCluster.Platform.AWS.RolesRef)
		if prefix == "" {
			diags.AddError(
				"Cannot derive operator roles prefix",
				fmt.Sprintf("Could not derive the operator roles prefix from parent cluster %q RolesRef. "+
					"The worker instance profile is required; without it the create fails.", clusterName),
			)
			return diags
		}

		// Set the instance profile
		if obj.Spec.NodePool.Platform.AWS == nil {
			obj.Spec.NodePool.Platform.AWS = &hypershiftv1beta1.AWSNodePoolPlatform{}
		}
		obj.Spec.NodePool.Platform.AWS.InstanceProfile = prefix + workerInstanceProfileSuffix
	}

	return diags
}

// PostResponse is called after API operations
func (h *NodePoolHandlerImpl) PostResponse(ctx context.Context, resp *v1alpha1.NodePool) diag.Diagnostics {
	return diag.Diagnostics{}
}

// PostFlatten populates computed fields in the state from the API response
func (h *NodePoolHandlerImpl) PostFlatten(ctx context.Context, state *NodePoolState, resp *v1alpha1.NodePool) {
	// TODO: Populate computed fields from API response
}

// NewNodePoolHandlerImpl creates a new NodePoolHandlerImpl instance.
func NewNodePoolHandlerImpl(client hyperfleet.Interface, accountID string, callerARN string) NodePoolHandler {
	return &NodePoolHandlerImpl{
		client:    client,
		accountID: accountID,
		callerARN: callerARN,
	}
}
