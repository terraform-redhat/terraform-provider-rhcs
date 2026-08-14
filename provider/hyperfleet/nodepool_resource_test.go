// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

// ── populateNodePoolState ─────────────────────────────────────────────────────

func TestPopulateNodePoolState_BasicFields(t *testing.T) {
	subnetID := "subnet-abc"
	autoRepair := true
	np := &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pool",
			Namespace: "cluster-uid-123",
			UID:       "np-uid-456",
		},
		Spec: v1alpha1.NodePoolSpec{
			AutoRepair: &autoRepair,
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				Platform: hypershiftv1beta1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS: &hypershiftv1beta1.AWSNodePoolPlatform{
						InstanceType: "m5.xlarge",
						Subnet:       hypershiftv1beta1.AWSResourceReference{ID: &subnetID},
					},
				},
			},
		},
		Status: v1alpha1.NodePoolStatus{
			Phase: v1alpha1.NodePoolPhaseReady,
		},
	}

	var state NodePoolHyperfleetState
	populateNodePoolState(context.Background(), np, &state)

	if state.ID.ValueString() != "np-uid-456" {
		t.Errorf("ID = %q", state.ID.ValueString())
	}
	if state.Name.ValueString() != "my-pool" {
		t.Errorf("Name = %q", state.Name.ValueString())
	}
	if state.Cluster.ValueString() != "cluster-uid-123" {
		t.Errorf("Cluster = %q", state.Cluster.ValueString())
	}
	if state.Phase.ValueString() != "Ready" {
		t.Errorf("Phase = %q", state.Phase.ValueString())
	}
	if !state.AutoRepair.ValueBool() {
		t.Error("AutoRepair should be true")
	}
	if state.SubnetID.ValueString() != "subnet-abc" {
		t.Errorf("SubnetID = %q", state.SubnetID.ValueString())
	}
	if state.AWSNodePool == nil || state.AWSNodePool.InstanceType.ValueString() != "m5.xlarge" {
		t.Errorf("InstanceType = %v", state.AWSNodePool)
	}
}

func TestPopulateNodePoolState_Replicas(t *testing.T) {
	r := int32(3)
	np := &v1alpha1.NodePool{
		Spec: v1alpha1.NodePoolSpec{
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				Replicas: &r,
				Platform: hypershiftv1beta1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS:  &hypershiftv1beta1.AWSNodePoolPlatform{},
				},
			},
		},
	}

	var state NodePoolHyperfleetState
	populateNodePoolState(context.Background(), np, &state)

	if state.Replicas.ValueInt64() != 3 {
		t.Errorf("Replicas = %d", state.Replicas.ValueInt64())
	}
}

func TestPopulateNodePoolState_Labels(t *testing.T) {
	np := &v1alpha1.NodePool{
		Spec: v1alpha1.NodePoolSpec{
			Labels: map[string]string{"env": "prod"},
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				Platform: hypershiftv1beta1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS:  &hypershiftv1beta1.AWSNodePoolPlatform{},
				},
			},
		},
	}

	var state NodePoolHyperfleetState
	populateNodePoolState(context.Background(), np, &state)

	if state.Labels.IsNull() {
		t.Fatal("Labels should not be null")
	}
	elems := state.Labels.Elements()
	if len(elems) != 1 {
		t.Errorf("Labels len = %d", len(elems))
	}
}

func TestPopulateNodePoolState_DiskSizeAndTags(t *testing.T) {
	np := &v1alpha1.NodePool{
		Spec: v1alpha1.NodePoolSpec{
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				Platform: hypershiftv1beta1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS: &hypershiftv1beta1.AWSNodePoolPlatform{
						InstanceType: "m5.large",
						RootVolume:   &hypershiftv1beta1.Volume{Size: 100},
						ResourceTags: []hypershiftv1beta1.AWSResourceTag{
							{Key: "team", Value: "platform"},
						},
					},
				},
			},
		},
	}

	var state NodePoolHyperfleetState
	populateNodePoolState(context.Background(), np, &state)

	if state.AWSNodePool.DiskSize.ValueInt64() != 100 {
		t.Errorf("DiskSize = %d", state.AWSNodePool.DiskSize.ValueInt64())
	}
	if state.AWSNodePool.Tags.IsNull() {
		t.Fatal("Tags should not be null")
	}
	tagElems := state.AWSNodePool.Tags.Elements()
	if len(tagElems) != 1 {
		t.Errorf("Tags len = %d", len(tagElems))
	}
}

// ── buildNodePoolSpec ─────────────────────────────────────────────────────────

func TestBuildNodePoolSpec_Basic(t *testing.T) {
	plan := &NodePoolHyperfleetState{
		AutoRepair: types.BoolValue(true),
		AWSNodePool: &NPAWSNodePool{
			InstanceType: types.StringValue("m5.xlarge"),
			Tags:         types.MapNull(types.StringType),
			DiskSize:     types.Int64Null(),
		},
		Replicas: types.Int64Value(2),
		Labels:   types.MapNull(types.StringType),
	}

	spec, diags := buildNodePoolSpec(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	if spec.AutoRepair == nil || !*spec.AutoRepair {
		t.Error("AutoRepair should be true")
	}
	if spec.NodePool.Platform.AWS == nil || spec.NodePool.Platform.AWS.InstanceType != "m5.xlarge" {
		t.Errorf("AWS.InstanceType = %v", spec.NodePool.Platform.AWS)
	}
	if spec.NodePool.Replicas == nil || *spec.NodePool.Replicas != 2 {
		t.Errorf("Replicas = %v", spec.NodePool.Replicas)
	}
}

func TestBuildNodePoolSpec_DiskSize(t *testing.T) {
	plan := &NodePoolHyperfleetState{
		AutoRepair: types.BoolValue(false),
		AWSNodePool: &NPAWSNodePool{
			InstanceType: types.StringValue("m5.large"),
			Tags:         types.MapNull(types.StringType),
			DiskSize:     types.Int64Value(120),
		},
		Replicas: types.Int64Value(1),
		Labels:   types.MapNull(types.StringType),
	}

	spec, diags := buildNodePoolSpec(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	if spec.NodePool.Platform.AWS.RootVolume == nil {
		t.Fatal("RootVolume should not be nil")
	}
	if spec.NodePool.Platform.AWS.RootVolume.Size != 120 {
		t.Errorf("RootVolume.Size = %d", spec.NodePool.Platform.AWS.RootVolume.Size)
	}
}
