// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"

	generated "github.com/terraform-redhat/terraform-provider-rhcs/provider/hyperfleet/generated"
)

func TestNodePoolClearUpdateDocumentPreservesFieldsAndIncludesEmptyCollections(t *testing.T) {
	nodePool := &v1alpha1.NodePool{
		Spec: v1alpha1.NodePoolSpec{
			AutoRepair: func(value bool) *bool { return &value }(true),
			Labels:     map[string]string{"team": "platform"},
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				Replicas: func(value int32) *int32 { return &value }(3),
			},
		},
	}

	document, err := nodePoolClearUpdateDocument(nodePool, true, true)
	if err != nil {
		t.Fatalf("nodePoolClearUpdateDocument() error = %v", err)
	}

	spec, ok := document["spec"].(map[string]any)
	if !ok {
		t.Fatalf("spec = %#v, want map[string]any", document["spec"])
	}
	if autoRepair, ok := spec["autoRepair"].(bool); !ok || !autoRepair {
		t.Errorf("spec.autoRepair = %#v, want true", spec["autoRepair"])
	}
	nodePoolSpec, ok := spec["nodePool"].(map[string]any)
	if !ok {
		t.Fatalf("spec.nodePool = %#v, want map[string]any", spec["nodePool"])
	}
	if replicas, ok := nodePoolSpec["replicas"].(float64); !ok || replicas != 3 {
		t.Errorf("spec.nodePool.replicas = %#v, want 3", nodePoolSpec["replicas"])
	}
	if labels, ok := spec["labels"].(map[string]any); !ok || len(labels) != 0 {
		t.Errorf("spec.labels = %#v, want empty map", spec["labels"])
	}
	if taints, ok := nodePoolSpec["taints"].([]any); !ok || len(taints) != 0 {
		t.Errorf("spec.nodePool.taints = %#v, want empty slice", nodePoolSpec["taints"])
	}
}

func TestNodePoolShouldClearCollectionsOnlyWhenRemovingExistingValues(t *testing.T) {
	taintObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"effect": types.StringType,
			"key":    types.StringType,
			"value":  types.StringType,
		},
	}
	labels := types.MapValueMust(
		types.StringType,
		map[string]attr.Value{"team": types.StringValue("platform")},
	)
	taints := types.ListValueMust(taintObjectType, nil)
	plan := &generated.NodePoolState{
		Labels: types.MapNull(types.StringType),
		Taints: types.ListNull(taintObjectType),
	}
	priorState := &generated.NodePoolState{Labels: labels, Taints: taints}

	clearLabels, clearTaints := nodePoolShouldClearCollections(plan, priorState)
	if !clearLabels || !clearTaints {
		t.Errorf(
			"nodePoolShouldClearCollections() = (%t, %t), want (true, true)",
			clearLabels,
			clearTaints,
		)
	}

	clearLabels, clearTaints = nodePoolShouldClearCollections(plan, plan)
	if clearLabels || clearTaints {
		t.Errorf(
			"nodePoolShouldClearCollections() with null prior state = (%t, %t), want (false, false)",
			clearLabels,
			clearTaints,
		)
	}
}
