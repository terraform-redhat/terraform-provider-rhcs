// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

// ── NodePool state tests ──────────────────────────────────────────────────────

func TestNodePoolState_BasicFields(t *testing.T) {
	subnetID := "subnet-abc"
	autoRepair := true
	replicas := int32(3)
	np := &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster.my-pool",
			Namespace: "account-123456789012",
			UID:       "np-uid-456",
		},
		Spec: v1alpha1.NodePoolSpec{
			AutoRepair: &autoRepair,
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				ClusterName: "my-cluster",
				Replicas:    &replicas,
				Platform: v1alpha1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS: &v1alpha1.AWSNodePoolPlatform{
						InstanceType:    "m5.xlarge",
						InstanceProfile: "my-prefix-ROSA-Worker-Role",
						Subnet:          hypershiftv1beta1.AWSResourceReference{ID: &subnetID},
					},
				},
			},
		},
		Status: v1alpha1.NodePoolStatus{
			Phase: v1alpha1.NodePoolPhaseReady,
		},
	}

	// Verify the nodepool structure matches what the handler expects
	if np.Spec.NodePool.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %q, want my-cluster", np.Spec.NodePool.ClusterName)
	}
	if np.Name != "my-cluster.my-pool" {
		t.Errorf("Name = %q, want my-cluster.my-pool", np.Name)
	}
	if np.Spec.NodePool.Platform.AWS.InstanceType != "m5.xlarge" {
		t.Errorf("InstanceType = %q, want m5.xlarge", np.Spec.NodePool.Platform.AWS.InstanceType)
	}
	if np.Spec.NodePool.Platform.AWS.InstanceProfile != "my-prefix-ROSA-Worker-Role" {
		t.Errorf("InstanceProfile = %q, want my-prefix-ROSA-Worker-Role", np.Spec.NodePool.Platform.AWS.InstanceProfile)
	}
	if *np.Spec.NodePool.Replicas != 3 {
		t.Errorf("Replicas = %d, want 3", *np.Spec.NodePool.Replicas)
	}
	if np.Status.Phase != v1alpha1.NodePoolPhaseReady {
		t.Errorf("Phase = %q, want %q", np.Status.Phase, v1alpha1.NodePoolPhaseReady)
	}
}

func TestNodePoolState_AWSPlatformConfig(t *testing.T) {
	subnetID := "subnet-123"
	volumeSize := int64(100)

	np := &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cluster.pool-aws",
			UID:  "np-789",
		},
		Spec: v1alpha1.NodePoolSpec{
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				ClusterName: "test-cluster",
				Platform: v1alpha1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS: &v1alpha1.AWSNodePoolPlatform{
						InstanceType: "m5.2xlarge",
						RootVolume: &v1alpha1.Volume{
							Size: volumeSize,
						},
						Subnet: hypershiftv1beta1.AWSResourceReference{ID: &subnetID},
					},
				},
			},
		},
	}

	if np.Spec.NodePool.Platform.AWS.RootVolume.Size != 100 {
		t.Errorf("RootVolume.Size = %d, want 100", np.Spec.NodePool.Platform.AWS.RootVolume.Size)
	}
	if *np.Spec.NodePool.Platform.AWS.Subnet.ID != "subnet-123" {
		t.Errorf("Subnet.ID = %q, want subnet-123", *np.Spec.NodePool.Platform.AWS.Subnet.ID)
	}
}

// ── validateMaxUnavailable ────────────────────────────────────────────────────

// maxUnavailableState builds the state the generated schema produces for a node
// pool with the given max_unavailable and replicas.
func maxUnavailableState(maxUnavailable types.String, replicas types.Int64) *NodePoolState {
	return &NodePoolState{MaxUnavailable: maxUnavailable, Replicas: replicas}
}

func TestValidateMaxUnavailable_Valid(t *testing.T) {
	cases := []struct {
		name  string
		state *NodePoolState
	}{
		{"null", maxUnavailableState(types.StringNull(), types.Int64Value(3))},
		{"unknown", maxUnavailableState(types.StringUnknown(), types.Int64Value(3))},
		{"zero", maxUnavailableState(types.StringValue("0"), types.Int64Value(3))},
		{"count below replicas", maxUnavailableState(types.StringValue("2"), types.Int64Value(3))},
		{"count equal to replicas", maxUnavailableState(types.StringValue("3"), types.Int64Value(3))},
		// The replica count is not always known at plan time; the API has the
		// final say in that case.
		{"unknown replicas", maxUnavailableState(types.StringValue("9"), types.Int64Unknown())},
		{"null replicas", maxUnavailableState(types.StringValue("9"), types.Int64Null())},
	}
	for _, tc := range cases {
		if diags := validateMaxUnavailable(tc.state); diags.HasError() {
			t.Errorf("%s: unexpected error: %v", tc.name, diags.Errors())
		}
	}
}

func TestValidateMaxUnavailable_Invalid(t *testing.T) {
	cases := []struct {
		name   string
		state  *NodePoolState
		detail string
	}{
		{
			name:   "exceeds replicas",
			state:  maxUnavailableState(types.StringValue("4"), types.Int64Value(3)),
			detail: "max_unavailable 4 cannot exceed replicas 3",
		},
		{
			name:   "negative",
			state:  maxUnavailableState(types.StringValue("-1"), types.Int64Value(3)),
			detail: "must be a non-negative whole number of nodes",
		},
		{
			name:   "not a number",
			state:  maxUnavailableState(types.StringValue("two"), types.Int64Value(3)),
			detail: `got "two"`,
		},
		// The Platform API accepts percentages; the provider cannot send one
		// until pathbind can round-trip an intstr value. Every percentage-shaped
		// value gets that explanation rather than the generic format error.
		{
			name:   "percentage",
			state:  maxUnavailableState(types.StringValue("30%"), types.Int64Value(3)),
			detail: `percentages are not supported yet, got "30%"`,
		},
		{
			name:   "percentage above 100",
			state:  maxUnavailableState(types.StringValue("150%"), types.Int64Value(3)),
			detail: "percentages are not supported yet",
		},
		{
			name:   "fractional percentage",
			state:  maxUnavailableState(types.StringValue("2.5%"), types.Int64Value(3)),
			detail: "percentages are not supported yet",
		},
		{
			name:   "leading zero",
			state:  maxUnavailableState(types.StringValue("01"), types.Int64Value(3)),
			detail: `got "01"`,
		},
		{
			name:   "too large to be a node count",
			state:  maxUnavailableState(types.StringValue("99999999999999999999"), types.Int64Value(3)),
			detail: "too large to be a number of nodes",
		},
	}
	for _, tc := range cases {
		diags := validateMaxUnavailable(tc.state)
		if !diags.HasError() {
			t.Errorf("%s: expected an error, got none", tc.name)
			continue
		}
		if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, tc.detail) {
			t.Errorf("%s: detail = %q, want it to contain %q", tc.name, detail, tc.detail)
		}
	}
}
