// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
			Name:      "my-pool",
			Namespace: "cluster-uid-123",
			UID:       "np-uid-456",
		},
		Spec: v1alpha1.NodePoolSpec{
			AutoRepair: &autoRepair,
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				ClusterName: "my-cluster",
				Replicas:    &replicas,
				Platform: v1alpha1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS: &hypershiftv1beta1.AWSNodePoolPlatform{
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
	if np.Name != "my-pool" {
		t.Errorf("Name = %q, want my-pool", np.Name)
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
			Name: "pool-aws",
			UID:  "np-789",
		},
		Spec: v1alpha1.NodePoolSpec{
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				ClusterName: "test-cluster",
				Platform: v1alpha1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS: &hypershiftv1beta1.AWSNodePoolPlatform{
						InstanceType: "m5.2xlarge",
						RootVolume: &hypershiftv1beta1.Volume{
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
