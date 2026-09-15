// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"fmt"
	"testing"

	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
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
