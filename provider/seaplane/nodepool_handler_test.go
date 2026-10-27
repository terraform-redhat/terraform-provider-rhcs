// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	fakeclientset "github.com/openshift-online/rosa-seaplane-api/clientset/generated/fake"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

type fakeSeaplaneClient struct {
	client spplatform.V1alpha1PublicInterface
}

func (f *fakeSeaplaneClient) SeaplaneV1alpha1() spplatform.V1alpha1PublicInterface {
	return f.client
}

func newFakeSeaplaneClient(objects ...runtime.Object) seaplane.Interface {
	clientset := fakeclientset.NewSimpleClientset(objects...)
	return &fakeSeaplaneClient{
		client: spplatform.NewV1alpha1PublicClient(clientset.V1alpha1Public()),
	}
}

func testNodePoolForPartialUpdate() *v1alpha1.NodePool {
	autoRepair := false
	replicas := int32(1)
	subnetID := "subnet-0abc123"

	return &v1alpha1.NodePool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "NodePool",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-cluster.worker"},
		Spec: v1alpha1.NodePoolSpec{
			AutoRepair: &autoRepair,
			Labels:     map[string]string{"existing": "label"},
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				ClusterName: "test-cluster",
				Replicas:    &replicas,
				Management: hypershiftv1beta1.NodePoolManagement{
					AutoRepair:  false,
					UpgradeType: hypershiftv1beta1.UpgradeTypeReplace,
				},
				Taints: []hypershiftv1beta1.Taint{{
					Key:    "existing",
					Value:  "taint",
					Effect: corev1.TaintEffectNoSchedule,
				}},
				Platform: v1alpha1.NodePoolPlatform{
					Type: hypershiftv1beta1.AWSPlatform,
					AWS: &v1alpha1.AWSNodePoolPlatform{
						InstanceType:    "m5.xlarge",
						InstanceProfile: "prefix-ROSA-Worker-Role",
						Subnet: hypershiftv1beta1.AWSResourceReference{
							ID: &subnetID,
						},
					},
				},
			},
		},
	}
}

func TestNodePoolPostPartialExpandIncludesUnchangedServiceSetFields(t *testing.T) {
	nodePool := testNodePoolForPartialUpdate()
	handler := &NodePoolHandlerImpl{client: newFakeSeaplaneClient(nodePool)}
	document := map[string]any{
		"spec": map[string]any{"autoRepair": true},
	}

	diags := handler.PostPartialExpand(
		context.Background(),
		&NodePoolState{Name: types.StringValue(nodePool.Name)},
		document,
	)
	if diags.HasError() {
		t.Fatalf("PostPartialExpand returned errors: %v", diags.Errors())
	}

	spec := document["spec"].(map[string]any)
	if got := spec["autoRepair"]; got != true {
		t.Errorf("autoRepair = %v, want true", got)
	}
	if got := spec["labels"].(map[string]any)["existing"]; got != "label" {
		t.Errorf("labels[existing] = %v, want label", got)
	}

	nodePoolSpec := spec["nodePool"].(map[string]any)
	management := nodePoolSpec["management"].(map[string]any)
	if got := management["upgradeType"]; got != string(hypershiftv1beta1.UpgradeTypeReplace) {
		t.Errorf("management.upgradeType = %v, want %s", got, hypershiftv1beta1.UpgradeTypeReplace)
	}
	platform := nodePoolSpec["platform"].(map[string]any)
	if got := platform["type"]; got != string(hypershiftv1beta1.AWSPlatform) {
		t.Errorf("platform.type = %v, want %s", got, hypershiftv1beta1.AWSPlatform)
	}
	if got := nodePoolSpec["taints"].([]any); len(got) != 1 {
		t.Errorf("taints length = %d, want 1", len(got))
	}
}

func TestNodePoolPostPartialExpandPreservesExplicitClears(t *testing.T) {
	nodePool := testNodePoolForPartialUpdate()
	handler := &NodePoolHandlerImpl{client: newFakeSeaplaneClient(nodePool)}
	document := map[string]any{
		"spec": map[string]any{
			"labels": map[string]any{},
			"nodePool": map[string]any{
				"taints": []any{},
			},
		},
	}

	diags := handler.PostPartialExpand(
		context.Background(),
		&NodePoolState{Name: types.StringValue(nodePool.Name)},
		document,
	)
	if diags.HasError() {
		t.Fatalf("PostPartialExpand returned errors: %v", diags.Errors())
	}

	spec := document["spec"].(map[string]any)
	if labels, ok := spec["labels"].(map[string]any); !ok || len(labels) != 0 {
		t.Errorf("labels = %#v, want an explicit empty map", spec["labels"])
	}
	nodePoolSpec := spec["nodePool"].(map[string]any)
	if taints, ok := nodePoolSpec["taints"].([]any); !ok || len(taints) != 0 {
		t.Errorf("taints = %#v, want an explicit empty list", nodePoolSpec["taints"])
	}
}

func TestNodePoolPostPartialExpandReplacesTagMaps(t *testing.T) {
	nodePool := testNodePoolForPartialUpdate()
	nodePool.Spec.Tags = map[string]string{"keep": "old", "remove": "old"}
	handler := &NodePoolHandlerImpl{client: newFakeSeaplaneClient(nodePool)}
	document := map[string]any{
		"spec": map[string]any{
			"tags": map[string]any{"keep": "new"},
		},
	}

	diags := handler.PostPartialExpand(
		context.Background(),
		&NodePoolState{Name: types.StringValue(nodePool.Name)},
		document,
	)
	if diags.HasError() {
		t.Fatalf("PostPartialExpand returned errors: %v", diags.Errors())
	}

	spec := document["spec"].(map[string]any)
	if got := spec["tags"]; !reflect.DeepEqual(got, map[string]any{"keep": "new"}) {
		t.Errorf("tags = %#v, want only the updated tag", got)
	}
}
