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
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// PreExpand validates inputs and derives computed fields
func (h *NodePoolHandlerImpl) PreExpand(ctx context.Context, input *NodePoolState) diag.Diagnostics {
	var diags diag.Diagnostics

	// Validate required fields
	if input.Name.IsNull() || input.Name.ValueString() == "" {
		diags.AddError("name is required", "NodePool name must be specified")
		return diags
	}

	if input.Cluster_id.IsNull() || input.Cluster_id.ValueString() == "" {
		diags.AddError("cluster_id is required", "Cluster ID must be specified to create a NodePool")
		return diags
	}

	return diags
}

// PostExpand processes inputs after pathbind.Expand to set SDK-specific fields
func (h *NodePoolHandlerImpl) PostExpand(
	ctx context.Context,
	input *NodePoolState,
	obj *v1alpha1.NodePool,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// Ensure platform.type is set to AWS platform (enum constant)
	if obj.Spec.NodePool.Platform.Type == "" {
		obj.Spec.NodePool.Platform.Type = hypershiftv1beta1.AWSPlatform
	}

	// Compute the worker instance profile from the parent cluster's operator roles prefix
	// This is required for CAPA to correctly assign IAM permissions to worker nodes
	clusterID := input.Cluster_id.ValueString()
	if clusterID != "" {
		cluster, err := h.client.HyperfleetV1alpha1().Clusters().Get(ctx, clusterID, hfwrappers.GetOptions{})
		if err != nil {
			diags.AddError(
				"Failed to get parent cluster",
				fmt.Sprintf(
					"Cannot compute worker instance profile: failed to retrieve cluster %q: %v",
					clusterID,
					err,
				),
			)
			return diags
		}

		// Set the nodepool namespace to match the cluster's namespace (cluster-<clusterID>)
		// The API requires the namespace in the format "cluster-<uuid>"
		obj.SetNamespace(fmt.Sprintf("cluster-%s", clusterID))

		if cluster.Spec.HostedCluster.Platform.AWS == nil {
			diags.AddError(
				"Parent cluster has no AWS configuration",
				fmt.Sprintf(
					"Cannot compute worker instance profile: parent cluster %q has no AWS platform configuration",
					clusterID,
				),
			)
			return diags
		}

		prefix, _ := prefixAndPartitionFromRolesRef(cluster.Spec.HostedCluster.Platform.AWS.RolesRef)
		if prefix == "" {
			diags.AddError(
				"Cannot derive operator roles prefix",
				fmt.Sprintf("Could not derive the operator roles prefix from parent cluster %q RolesRef. "+
					"The worker instance profile is required; without it the create fails.", clusterID),
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

// Namespace derives the K8s namespace ("cluster-<uuid>") for this NodePool from
// state.Cluster_id. Used by Read/Delete/ImportState, which only have access to
// Terraform state (no SDK object) before calling the API. Must be side-effect-free.
func (h *NodePoolHandlerImpl) Namespace(ctx context.Context, state *NodePoolState) string {
	return fmt.Sprintf("cluster-%s", state.Cluster_id.ValueString())
}

// PostFlatten populates computed fields in the state from the API response
func (h *NodePoolHandlerImpl) PostFlatten(ctx context.Context, state *NodePoolState, resp *v1alpha1.NodePool) {
	if resp == nil {
		return
	}
	// Populate Phase from status (consumer-only field, not mapped via pathbind)
	// Default to "Provisioning" if not yet set by the controller
	if resp.Status.Phase != "" {
		state.Phase = types.StringValue(string(resp.Status.Phase))
	} else {
		state.Phase = types.StringValue(string(v1alpha1.NodePoolPhaseProvisioning))
	}
}

// NewNodePoolHandlerImpl creates a new NodePoolHandlerImpl instance.
func NewNodePoolHandlerImpl(client hyperfleet.Interface, accountID string, callerARN string) NodePoolHandler {
	return &NodePoolHandlerImpl{
		client:    client,
		accountID: accountID,
		callerARN: callerARN,
	}
}
