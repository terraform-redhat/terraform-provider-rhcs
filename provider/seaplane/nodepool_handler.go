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

package seaplane

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

// workerInstanceProfileSuffix is appended to the operator roles prefix to form
// the worker IAM instance profile name. It mirrors the naming convention in the
// IAM manifests (`${operator_roles_prefix}-ROSA-Worker-Role`).
const workerInstanceProfileSuffix = "-ROSA-Worker-Role"

// NodePoolHandlerImpl is the concrete implementation of NodePoolHandler
type NodePoolHandlerImpl struct {
	client    seaplane.Interface
	accountID string
	callerARN string
}

var nodePoolAutoscalingAttributeTypes = map[string]attr.Type{
	"enabled":      types.BoolType,
	"min_replicas": types.Int64Type,
	"max_replicas": types.Int64Type,
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
		diags.AddError("cluster_name is required", "Parent cluster name must be specified to create a NodePool")
		return diags
	}
	if !strings.HasPrefix(input.Name.ValueString(), input.ClusterName.ValueString()+".") {
		diags.AddError(
			"invalid NodePool name",
			fmt.Sprintf(
				"NodePool name %q must start with parent cluster name %q followed by a dot",
				input.Name.ValueString(),
				input.ClusterName.ValueString(),
			),
		)
		return diags
	}

	if input.Cluster_id.IsNull() || input.Cluster_id.ValueString() == "" {
		diags.AddError("cluster_id is required", "Cluster ID must be specified to create a NodePool")
		return diags
	}

	if !input.Autoscaling.IsNull() && !input.Autoscaling.IsUnknown() {
		config := readNodePoolAutoscaling(input.Autoscaling)
		if config == nil {
			diags.AddError("invalid autoscaling configuration", "The autoscaling block has an invalid shape")
			return diags
		}
		if config.Enabled.IsUnknown() || config.MinReplicas.IsUnknown() || config.MaxReplicas.IsUnknown() {
			return diags
		}
		if config.Enabled.IsNull() {
			diags.AddError("invalid autoscaling configuration", "autoscaling.enabled must be specified")
		} else if !config.Enabled.ValueBool() {
			if !config.MinReplicas.IsNull() || !config.MaxReplicas.IsNull() {
				diags.AddError(
					"invalid autoscaling configuration",
					"min_replicas and max_replicas can only be set when autoscaling is enabled",
				)
			}
		} else {
			if !input.Replicas.IsNull() || input.Replicas.IsUnknown() {
				diags.AddError("conflicting nodepool sizing", "replicas cannot be set when autoscaling is enabled")
			}
			if config.MinReplicas.IsNull() || config.MaxReplicas.IsNull() {
				diags.AddError("invalid autoscaling configuration", "enabled autoscaling requires min_replicas and max_replicas")
			} else if config.MinReplicas.ValueInt64() < 0 {
				diags.AddError("invalid autoscaling configuration", "min_replicas must be greater than or equal to zero")
			} else if config.MaxReplicas.ValueInt64() < 1 {
				diags.AddError("invalid autoscaling configuration", "max_replicas must be greater than or equal to one")
			} else if config.MaxReplicas.ValueInt64() < config.MinReplicas.ValueInt64() {
				diags.AddError("invalid autoscaling configuration", "max_replicas must be greater than or equal to min_replicas")
			}
		}
	}

	diags.Append(validateNodeDrainGracePeriod(input)...)
	diags.Append(validateImageType(input)...)

	return diags
}

// validateImageType checks the operating system image family against the enum
// the Platform API accepts on spec.nodePool.platform.aws.imageType, so a bad
// value fails during apply instead of on the API round trip. The values are
// case-sensitive, which is the most common way to get this wrong.
func validateImageType(input *NodePoolState) diag.Diagnostics {
	var diags diag.Diagnostics

	if input.ImageType.IsNull() || input.ImageType.IsUnknown() {
		return diags
	}

	switch hypershiftv1beta1.ImageType(input.ImageType.ValueString()) {
	case hypershiftv1beta1.ImageTypeLinux, hypershiftv1beta1.ImageTypeWindows:
		return diags
	}

	diags.AddAttributeError(
		path.Root("image_type"), "Invalid node pool image_type",
		fmt.Sprintf("image_type must be %q or %q, got %q",
			hypershiftv1beta1.ImageTypeLinux, hypershiftv1beta1.ImageTypeWindows, input.ImageType.ValueString()),
	)

	return diags
}

// validateNodeDrainGracePeriod checks that node_drain_grace_period is either null, unknown,
// or a valid Go duration string. Examples: "5m", "1h", "30s", "0s".
func validateNodeDrainGracePeriod(input *NodePoolState) diag.Diagnostics {
	var diags diag.Diagnostics

	if input.NodeDrainGracePeriod.IsNull() || input.NodeDrainGracePeriod.IsUnknown() {
		return diags
	}

	value := input.NodeDrainGracePeriod.ValueString()

	// Parse the duration to validate it
	_, err := time.ParseDuration(value)
	if err != nil {
		diags.AddAttributeError(
			path.Root("node_drain_grace_period"),
			"Invalid node_drain_grace_period",
			fmt.Sprintf(
				"node_drain_grace_period must be a valid Go duration string "+
					"(for example \"5m\" for 5 minutes, \"1h\" for 1 hour, \"0s\" for immediate termination), got %q: %v",
				value, err,
			),
		)
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

	if input.Autoscaling.IsNull() {
		obj.Spec.NodePool.AutoScaling = nil
	} else if !input.Autoscaling.IsUnknown() {
		config := readNodePoolAutoscaling(input.Autoscaling)
		if config != nil && !config.Enabled.IsNull() && !config.Enabled.ValueBool() {
			obj.Spec.NodePool.AutoScaling = nil
		}
	}

	// Ensure platform.type is set to AWS platform (enum constant)
	if obj.Spec.NodePool.Platform.Type == "" {
		obj.Spec.NodePool.Platform.Type = hypershiftv1beta1.AWSPlatform
	}

	// Compute the worker instance profile from the parent cluster's operator roles prefix
	// This is required for CAPA to correctly assign IAM permissions to worker nodes
	clusterName := input.ClusterName.ValueString()
	clusterID := input.Cluster_id.ValueString()
	if clusterName != "" {
		cluster, err := h.client.SeaplaneV1alpha1().Clusters().Get(ctx, clusterName, spplatform.GetOptions{})
		if err != nil {
			diags.AddError(
				"Failed to get parent cluster",
				fmt.Sprintf(
					"Cannot compute worker instance profile: failed to retrieve cluster %q: %v",
					clusterName,
					err,
				),
			)
			return diags
		}
		if string(cluster.UID) != clusterID {
			diags.AddError(
				"Parent cluster identity mismatch",
				fmt.Sprintf("cluster_id %q does not match cluster_name %q (UID %q)", clusterID, clusterName, cluster.UID),
			)
			return diags
		}

		if cluster.Spec.HostedCluster.Platform.AWS == nil {
			diags.AddError(
				"Parent cluster has no AWS configuration",
				fmt.Sprintf(
					"Cannot compute worker instance profile: parent cluster %q has no AWS platform configuration",
					clusterName,
				),
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
			obj.Spec.NodePool.Platform.AWS = &v1alpha1.AWSNodePoolPlatform{}
		}
		obj.Spec.NodePool.Platform.AWS.InstanceProfile = prefix + workerInstanceProfileSuffix
	}

	return diags
}

// PostPartialExpand allows the raw partial-update document to be customized.
func (h *NodePoolHandlerImpl) PostPartialExpand(
	ctx context.Context,
	input *NodePoolState,
	document map[string]any,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(document) == 0 {
		return diags
	}
	partialSpec, ok := document["spec"].(map[string]any)
	if !ok || len(partialSpec) == 0 {
		return diags
	}

	current, err := h.client.SeaplaneV1alpha1().NodePools().Get(
		ctx,
		input.Name.ValueString(),
		spplatform.GetOptions{},
	)
	if err != nil {
		diags.AddError("Failed to read NodePool for update", err.Error())
		return diags
	}

	currentSpecJSON, err := json.Marshal(current.Spec)
	if err != nil {
		diags.AddError("Failed to encode NodePool update", err.Error())
		return diags
	}

	currentSpec := map[string]any{}
	if err := json.Unmarshal(currentSpecJSON, &currentSpec); err != nil {
		diags.AddError("Failed to decode NodePool update", err.Error())
		return diags
	}

	mergePartialNodePoolSpec(currentSpec, partialSpec, "spec")
	document["spec"] = currentSpec
	return diags
}

// mergePartialNodePoolSpec overlays changed fields onto the live spec. Sending
// the complete spec keeps required service-set fields unchanged so the API's
// update validation does not treat decoded zero values as modifications.
// Map collections use replacement semantics, matching the API's merge behaviour.
func mergePartialNodePoolSpec(current, update map[string]any, path string) {
	for key, updateValue := range update {
		childPath := path + "." + key
		replaceCollection := childPath == "spec.labels" ||
			childPath == "spec.tags" ||
			childPath == "spec.nodePool.nodeLabels"
		if !replaceCollection {
			currentMap, currentIsMap := current[key].(map[string]any)
			updateMap, updateIsMap := updateValue.(map[string]any)
			if currentIsMap && updateIsMap {
				mergePartialNodePoolSpec(currentMap, updateMap, childPath)
				continue
			}
		}
		current[key] = updateValue
	}
}

// PostResponse is called after API operations
func (h *NodePoolHandlerImpl) PostResponse(ctx context.Context, resp *v1alpha1.NodePool) diag.Diagnostics {
	return diag.Diagnostics{}
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

	if resp.Spec.NodePool.AutoScaling == nil {
		state.Autoscaling = types.ObjectNull(nodePoolAutoscalingAttributeTypes)
		return
	}

	min := types.Int64Null()
	if resp.Spec.NodePool.AutoScaling.Min != nil {
		min = types.Int64Value(int64(*resp.Spec.NodePool.AutoScaling.Min))
	}
	max := types.Int64Value(int64(resp.Spec.NodePool.AutoScaling.Max))
	autoscaling, diags := types.ObjectValue(nodePoolAutoscalingAttributeTypes, map[string]attr.Value{
		"enabled":      types.BoolValue(true),
		"min_replicas": min,
		"max_replicas": max,
	})
	if !diags.HasError() {
		state.Autoscaling = autoscaling
	}
}

type nodePoolAutoscalingConfig struct {
	Enabled     types.Bool
	MinReplicas types.Int64
	MaxReplicas types.Int64
}

func readNodePoolAutoscaling(value types.Object) *nodePoolAutoscalingConfig {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	attrs := value.Attributes()
	enabled, enabledOK := attrs["enabled"].(types.Bool)
	min, minOK := attrs["min_replicas"].(types.Int64)
	max, maxOK := attrs["max_replicas"].(types.Int64)
	if !enabledOK || !minOK || !maxOK {
		return nil
	}
	return &nodePoolAutoscalingConfig{Enabled: enabled, MinReplicas: min, MaxReplicas: max}
}

// NewNodePoolHandlerImpl creates a new NodePoolHandlerImpl instance.
func NewNodePoolHandlerImpl(client seaplane.Interface, accountID string, callerARN string) NodePoolHandler {
	return &NodePoolHandlerImpl{
		client:    client,
		accountID: accountID,
		callerARN: callerARN,
	}
}
