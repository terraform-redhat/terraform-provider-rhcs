// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"fmt"
	"testing"

	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ── regionFromAZ ──────────────────────────────────────────────────────────────

func TestRegionFromAZ(t *testing.T) {
	cases := []struct {
		az   string
		want string
	}{
		{"us-east-1a", "us-east-1"},
		{"us-east-1b", "us-east-1"},
		{"eu-west-2a", "eu-west-2"},
		{"us-gov-east-1a", "us-gov-east-1"},
		{"us-gov-west-1b", "us-gov-west-1"},
		// Local Zone: strip the suffix after the base region
		{"us-east-1-bos-1a", "us-east-1"},
		// Wavelength Zone
		{"us-east-1-wl1-bos-wlz-1", "us-east-1"},
		// No match
		{"", ""},
		{"invalid", ""},
	}
	for _, tc := range cases {
		got := regionFromAZ(tc.az)
		if got != tc.want {
			t.Errorf("regionFromAZ(%q) = %q, want %q", tc.az, got, tc.want)
		}
	}
}

// ── isNotFound ────────────────────────────────────────────────────────────────

func TestIsNotFound(t *testing.T) {
	if !isNotFound(fmt.Errorf("404 not found")) {
		t.Error("expected true for 404 error")
	}
	if isNotFound(fmt.Errorf("500 internal server error")) {
		t.Error("expected false for 500 error")
	}
	if isNotFound(nil) {
		t.Error("expected false for nil error")
	}
}

// ── computeRolesRef ───────────────────────────────────────────────────────────

func TestComputeRolesRef(t *testing.T) {
	ref := computeRolesRef("my-prefix", "123456789012", "aws")
	cases := []struct {
		name string
		arn  string
	}{
		{"ingress", ref.IngressARN},
		{"cloud-controller-manager", ref.KubeCloudControllerARN},
		{"ebs-csi", ref.StorageARN},
		{"image-registry", ref.ImageRegistryARN},
		{"network-config", ref.NetworkARN},
		{"control-plane-operator", ref.ControlPlaneOperatorARN},
		{"node-pool-management", ref.NodePoolManagementARN},
	}
	for _, tc := range cases {
		want := "arn:aws:iam::123456789012:role/my-prefix-" + tc.name
		if tc.arn != want {
			t.Errorf("computeRolesRef %s = %q, want %q", tc.name, tc.arn, want)
		}
	}
}

func TestComputeRolesRef_GovCloud(t *testing.T) {
	ref := computeRolesRef("pfx", "111111111111", "aws-us-gov")
	if ref.IngressARN != "arn:aws-us-gov:iam::111111111111:role/pfx-ingress" {
		t.Errorf("unexpected GovCloud ARN: %s", ref.IngressARN)
	}
}

// ── prefixAndPartitionFromRolesRef ────────────────────────────────────────────

func TestPrefixAndPartitionFromRolesRef_Roundtrip(t *testing.T) {
	cases := []struct {
		prefix    string
		accountID string
		partition string
	}{
		{"my-cluster", "123456789012", "aws"},
		{"hf-e2e-sanity", "999999999999", "aws-us-gov"},
		{"pfx", "000000000001", "aws-cn"},
	}
	for _, tc := range cases {
		ref := computeRolesRef(tc.prefix, tc.accountID, tc.partition)
		gotPrefix, gotPartition := prefixAndPartitionFromRolesRef(ref)
		if gotPrefix != tc.prefix {
			t.Errorf("roundtrip prefix: got %q, want %q", gotPrefix, tc.prefix)
		}
		if gotPartition != tc.partition {
			t.Errorf("roundtrip partition: got %q, want %q", gotPartition, tc.partition)
		}
	}
}

func TestPrefixAndPartitionFromRolesRef_Empty(t *testing.T) {
	prefix, partition := prefixAndPartitionFromRolesRef(hypershiftv1beta1.AWSRolesRef{})
	if prefix != "" || partition != "aws" {
		t.Errorf("empty rolesRef: got prefix=%q partition=%q", prefix, partition)
	}
}

// ── populateState ─────────────────────────────────────────────────────────────

func TestPopulateState_BasicFields(t *testing.T) {
	cluster := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
			UID:  "uid-abc",
		},
		Status: v1alpha1.ClusterStatus{
			Phase:   v1alpha1.ClusterPhaseReady,
			Version: "4.17.0",
			ControlPlaneEndpoint: hypershiftv1beta1.APIEndpoint{
				Host: "api.my-cluster.example.com",
			},
		},
		Spec: v1alpha1.ClusterSpec{
			HostedCluster: v1alpha1.HostedClusterSpecPassthrough{
				IssuerURL: "https://oidc.example.com/issuer",
			},
		},
	}

	var state ClusterHyperfleetState
	populateState(cluster, "arn:aws:iam::123:user/test", "us-east-1", &state)

	if state.ID.ValueString() != "uid-abc" {
		t.Errorf("ID = %q", state.ID.ValueString())
	}
	if state.Name.ValueString() != "my-cluster" {
		t.Errorf("Name = %q", state.Name.ValueString())
	}
	if state.Phase.ValueString() != string(v1alpha1.ClusterPhaseReady) {
		t.Errorf("Phase = %q", state.Phase.ValueString())
	}
	if state.APIURL.ValueString() != "api.my-cluster.example.com" {
		t.Errorf("APIURL = %q", state.APIURL.ValueString())
	}
	if state.OIDCIssuer.ValueString() != "https://oidc.example.com/issuer" {
		t.Errorf("OIDCIssuer = %q", state.OIDCIssuer.ValueString())
	}
	if state.CreatorARN.ValueString() != "arn:aws:iam::123:user/test" {
		t.Errorf("CreatorARN = %q", state.CreatorARN.ValueString())
	}
}

func TestPopulateState_ExpirationNull(t *testing.T) {
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			HostedCluster: v1alpha1.HostedClusterSpecPassthrough{},
		},
	}
	var state ClusterHyperfleetState
	populateState(cluster, "", "", &state)
	if !state.ExpirationTimestamp.IsNull() {
		t.Errorf("ExpirationTimestamp should be null when not set")
	}
}

func TestPopulateState_AWSPlatformFields(t *testing.T) {
	subnetID := "subnet-456"
	cluster := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "aws-cluster"},
		Spec: v1alpha1.ClusterSpec{
			HostedCluster: v1alpha1.HostedClusterSpecPassthrough{
				Platform: hypershiftv1beta1.PlatformSpec{
					AWS: &hypershiftv1beta1.AWSPlatformSpec{
						Region: "us-west-2",
						CloudProviderConfig: &hypershiftv1beta1.AWSCloudProviderConfig{
							VPC:  "vpc-123",
							Zone: "us-west-2a",
							Subnet: &hypershiftv1beta1.AWSResourceReference{
								ID: &subnetID,
							},
						},
						RolesRef: computeRolesRef("my-prefix", "123456789012", "aws"),
					},
				},
			},
		},
	}
	var state ClusterHyperfleetState
	populateState(cluster, "arn:aws:iam::123:user/test", "us-east-1", &state)

	if state.CloudRegion.ValueString() != "us-west-2" {
		t.Errorf("CloudRegion = %q, want us-west-2", state.CloudRegion.ValueString())
	}
	if state.VPCID.ValueString() != "vpc-123" {
		t.Errorf("VPCID = %q, want vpc-123", state.VPCID.ValueString())
	}
	azElems := state.AvailabilityZones.Elements()
	if len(azElems) != 1 || azElems[0].String() != `"us-west-2a"` {
		t.Errorf("AvailabilityZones = %v, want [us-west-2a]", azElems)
	}
	subnetElems := state.AWSSubnetIDs.Elements()
	if len(subnetElems) != 1 || subnetElems[0].String() != `"subnet-456"` {
		t.Errorf("AWSSubnetIDs = %v, want [subnet-456]", subnetElems)
	}
	if state.OperatorRolesPrefix.ValueString() != "my-prefix" {
		t.Errorf("OperatorRolesPrefix = %q, want my-prefix", state.OperatorRolesPrefix.ValueString())
	}
	if state.AWSPartition.ValueString() != "aws" {
		t.Errorf("AWSPartition = %q, want aws", state.AWSPartition.ValueString())
	}
}

func TestPopulateState_PlacementRef(t *testing.T) {
	cluster := &v1alpha1.Cluster{
		Status: v1alpha1.ClusterStatus{
			PlacementRef: &v1alpha1.PlacementReference{
				ManagementCluster: "mc01",
			},
		},
	}
	var state ClusterHyperfleetState
	populateState(cluster, "", "", &state)
	if state.ManagementCluster.ValueString() != "mc01" {
		t.Errorf("ManagementCluster = %q, want mc01", state.ManagementCluster.ValueString())
	}
}
