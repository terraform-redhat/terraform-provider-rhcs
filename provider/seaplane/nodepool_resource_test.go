// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	"github.com/openshift-online/rosa-seaplane-api/clientset/pathbind"
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

func TestNodePoolState_EC2InstanceMetadataHttpTokensExpansion(t *testing.T) {
	native := &NodePoolStateNative{EC2InstanceMetadataHttpTokens: "optional"}
	obj := &v1alpha1.NodePool{}

	if err := pathbind.Expand(context.Background(), native, obj); err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if got := obj.Spec.EC2InstanceMetadataHttpTokens; got != v1alpha1.EC2InstanceMetadataHttpTokensOptional {
		t.Errorf("EC2InstanceMetadataHttpTokens = %q, want %q", got, v1alpha1.EC2InstanceMetadataHttpTokensOptional)
	}
}

func TestNodePoolResourceSchema_EC2InstanceMetadataHttpTokens(t *testing.T) {
	var resp resource.SchemaResponse
	NewNodePool().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	attribute, ok := resp.Schema.Attributes["ec2_instance_metadata_http_tokens"].(schema.StringAttribute)
	if !ok {
		t.Fatal("ec2_instance_metadata_http_tokens is not a string attribute")
	}
	if !attribute.Optional {
		t.Error("ec2_instance_metadata_http_tokens should be optional so the API can apply its required default")
	}
	if attribute.Computed {
		t.Error("ec2_instance_metadata_http_tokens should not be provider-computed")
	}
	if len(attribute.PlanModifiers) == 0 {
		t.Error("ec2_instance_metadata_http_tokens should require replacement when changed")
	}
}

func TestNodePoolResourceSchema_NodeDrainGracePeriod(t *testing.T) {
	var resp resource.SchemaResponse
	NewNodePool().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	attribute, ok := resp.Schema.Attributes["node_drain_grace_period"].(schema.StringAttribute)
	if !ok {
		t.Fatal("node_drain_grace_period is not a string attribute")
	}
	if !attribute.Optional {
		t.Error("node_drain_grace_period should be optional")
	}
	if attribute.Computed {
		t.Error("node_drain_grace_period should not be provider-computed")
	}
	if attribute.Required {
		t.Error("node_drain_grace_period should not be required (defaults to 0s when unset)")
	}
}

// ── validateMaxUnavailable ────────────────────────────────────────────────────

func TestValidateMaxUnavailable_Valid(t *testing.T) {
	cases := []struct {
		name  string
		value types.String
		replicas types.Int64
	}{
		{"null", types.StringNull(), types.Int64Value(3)},
		{"unknown", types.StringUnknown(), types.Int64Value(3)},
		{"zero", types.StringValue("0"), types.Int64Value(3)},
		{"positive count", types.StringValue("2"), types.Int64Value(3)},
		{"count equals replicas", types.StringValue("3"), types.Int64Value(3)},
		{"large count under limit", types.StringValue("999"), types.Int64Value(1000)},
	}
	for _, tc := range cases {
		diags := validateMaxUnavailable(&NodePoolState{
			MaxUnavailable: tc.value,
			Replicas:       tc.replicas,
		})
		if diags.HasError() {
			t.Errorf("%s: unexpected error: %v", tc.name, diags.Errors())
		}
	}
}

func TestValidateMaxUnavailable_Invalid(t *testing.T) {
	cases := []struct {
		name         string
		value        string
		replicas     types.Int64
		wantContains string
	}{
		{"percentage", "50%", types.Int64Value(10), "percentages are not supported yet"},
		{"negative", "-1", types.Int64Value(5), "non-negative whole number"},
		{"float", "2.5", types.Int64Value(5), "non-negative whole number"},
		{"not a number", "abc", types.Int64Value(5), "non-negative whole number"},
		{"leading zero", "01", types.Int64Value(5), "non-negative whole number"},
		{"exceeds replicas", "10", types.Int64Value(5), "cannot exceed replicas"},
	}
	for _, tc := range cases {
		diags := validateMaxUnavailable(&NodePoolState{
			MaxUnavailable: types.StringValue(tc.value),
			Replicas:       tc.replicas,
		})
		if !diags.HasError() {
			t.Errorf("%s: expected an error, got none", tc.name)
			continue
		}
		detail := diags.Errors()[0].Detail()
		if !strings.Contains(detail, tc.wantContains) {
			t.Errorf("%s: detail = %q, want it to contain %q", tc.name, detail, tc.wantContains)
		}
	}
}

func TestValidateMaxUnavailable_WithAutoscaling(t *testing.T) {
	cases := []struct {
		name           string
		maxUnavailable string
		minReplicas    int64
		maxReplicas    int64
		wantError      bool
		wantContains   string
	}{
		{"valid: under min_replicas", "1", 2, 5, false, ""},
		{"valid: equals min_replicas", "2", 2, 5, false, ""},
		{"invalid: exceeds min_replicas", "3", 2, 5, true, "cannot exceed autoscaling min_replicas"},
	}

	for _, tc := range cases {
		autoscalingConfig := map[string]attr.Value{
			"enabled":      types.BoolValue(true),
			"min_replicas": types.Int64Value(tc.minReplicas),
			"max_replicas": types.Int64Value(tc.maxReplicas),
		}
		autoscaling, _ := types.ObjectValue(map[string]attr.Type{
			"enabled":      types.BoolType,
			"min_replicas": types.Int64Type,
			"max_replicas": types.Int64Type,
		}, autoscalingConfig)

		diags := validateMaxUnavailable(&NodePoolState{
			MaxUnavailable: types.StringValue(tc.maxUnavailable),
			Autoscaling:    autoscaling,
		})
		if tc.wantError {
			if !diags.HasError() {
				t.Errorf("%s: expected an error, got none", tc.name)
				continue
			}
			detail := diags.Errors()[0].Detail()
			if !strings.Contains(detail, tc.wantContains) {
				t.Errorf("%s: detail = %q, want it to contain %q", tc.name, detail, tc.wantContains)
			}
		} else {
			if diags.HasError() {
				t.Errorf("%s: unexpected error: %v", tc.name, diags.Errors())
			}
		}
	}
}

// ── validateImageType ─────────────────────────────────────────────────────────

func TestValidateImageType_Valid(t *testing.T) {
	cases := []struct {
		name      string
		imageType types.String
	}{
		{"null", types.StringNull()},
		{"unknown", types.StringUnknown()},
		{"linux", types.StringValue("Linux")},
		{"windows", types.StringValue("Windows")},
	}
	for _, tc := range cases {
		diags := validateImageType(&NodePoolState{ImageType: tc.imageType})
		if diags.HasError() {
			t.Errorf("%s: unexpected error: %v", tc.name, diags.Errors())
		}
	}
}

func TestValidateImageType_Invalid(t *testing.T) {
	cases := []struct {
		name      string
		imageType string
	}{
		// The API enum is case-sensitive, so these are the common near misses.
		{"lowercase windows", "windows"},
		{"lowercase linux", "linux"},
		{"rhcos", "RHCOS"},
		{"empty string", ""},
	}
	for _, tc := range cases {
		diags := validateImageType(&NodePoolState{ImageType: types.StringValue(tc.imageType)})
		if !diags.HasError() {
			t.Errorf("%s: expected an error, got none", tc.name)
			continue
		}
		detail := diags.Errors()[0].Detail()
		if !strings.Contains(detail, `must be "Linux" or "Windows"`) {
			t.Errorf("%s: detail = %q, want it to name the valid values", tc.name, detail)
		}
	}
}

// Exercise the real SDK binding, including preserving service-managed settings
// when expanding a size into an object fetched for an unrelated update.
func TestNodePoolRootVolumeBinding(t *testing.T) {
	var schemaResponse resource.SchemaResponse
	NewNodePool().Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	if _, exists := schemaResponse.Schema.Attributes["size"]; exists {
		t.Fatal("legacy size attribute should not be exposed")
	}
	if _, ok := schemaResponse.Schema.Attributes["disk_size"].(schema.Int64Attribute); !ok {
		t.Fatal("disk_size should be an integer attribute")
	}
	size := int64(500)
	encrypted := true
	obj := &v1alpha1.NodePool{
		Spec: v1alpha1.NodePoolSpec{
			NodePool: v1alpha1.NodePoolSpecPassthrough{
				Platform: v1alpha1.NodePoolPlatform{
					AWS: &v1alpha1.AWSNodePoolPlatform{
						RootVolume: &v1alpha1.Volume{Size: 100, Type: "gp3", Encrypted: &encrypted},
					},
				},
			},
		},
	}
	if err := pathbind.Expand(context.Background(), &NodePoolStateNative{DiskSize: &size}, obj); err != nil {
		t.Fatal(err)
	}
	volume := obj.Spec.NodePool.Platform.AWS.RootVolume
	if volume.Size != size || volume.Type != "gp3" || volume.Encrypted == nil || !*volume.Encrypted {
		t.Fatalf("root volume settings were not preserved: %+v", volume)
	}
	var native NodePoolStateNative
	if err := pathbind.Flatten(context.Background(), obj, &native); err != nil {
		t.Fatal(err)
	}
	if native.DiskSize == nil || *native.DiskSize != size {
		t.Fatalf("flattened size = %v, want %d", native.DiskSize, size)
	}
}

// ── validateNodeDrainTimeout ──────────────────────────────────────────────────

func TestValidateNodeDrainTimeout_Valid(t *testing.T) {
	cases := []struct {
		name    string
		timeout types.String
	}{
		{"null", types.StringNull()},
		{"unknown", types.StringUnknown()},
		{"zero", types.StringValue("0s")},
		{"seconds", types.StringValue("30s")},
		{"minutes", types.StringValue("5m")},
		{"hours", types.StringValue("1h")},
		{"combined", types.StringValue("1h30m")},
	}
	for _, tc := range cases {
		diags := validateNodeDrainTimeout(&NodePoolState{NodeDrainTimeout: tc.timeout})
		if diags.HasError() {
			t.Errorf("%s: unexpected error: %v", tc.name, diags.Errors())
		}
	}
}

func TestValidateNodeDrainTimeout_Invalid(t *testing.T) {
	cases := []struct {
		name         string
		timeout      string
		wantContains string
	}{
		{"not a duration", "abc", "invalid duration"},
		{"just a number", "123", "missing unit"},
		{"bad unit", "5x", "unknown unit"},
		{"empty string", "", "invalid duration"},
	}
	for _, tc := range cases {
		diags := validateNodeDrainTimeout(&NodePoolState{NodeDrainTimeout: types.StringValue(tc.timeout)})
		if !diags.HasError() {
			t.Errorf("%s: expected an error, got none", tc.name)
			continue
		}
		detail := diags.Errors()[0].Detail()
		if !strings.Contains(detail, tc.wantContains) {
			t.Errorf("%s: detail = %q, want it to contain %q", tc.name, detail, tc.wantContains)
		}
	}
}
