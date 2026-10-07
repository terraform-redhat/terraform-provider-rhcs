// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// ── validateClusterTags ───────────────────────────────────────────────────────

// tagsMap builds the types.Map the generated schema produces for the `tags`
// attribute.
func tagsMap(t *testing.T, values map[string]string) types.Map {
	t.Helper()
	elements := make(map[string]attr.Value, len(values))
	for key, value := range values {
		elements[key] = types.StringValue(value)
	}
	tags, diags := types.MapValue(types.StringType, elements)
	if diags.HasError() {
		t.Fatalf("building tags map: %v", diags)
	}
	return tags
}

func TestValidateClusterTags_Valid(t *testing.T) {
	cases := []struct {
		name string
		tags types.Map
	}{
		{"null", types.MapNull(types.StringType)},
		{"unknown", types.MapUnknown(types.StringType)},
		{"empty", tagsMap(t, nil)},
		{"typical", tagsMap(t, map[string]string{"cost-center": "cc-1234", "environment": "production"})},
		// Keys and values may contain the extra characters AWS allows.
		{"allowed characters", tagsMap(t, map[string]string{"kubernetes.io/role": "worker+1@eu=west"})},
		{"maximum", tagsMap(t, generatedTags(maxClusterTags))},
	}
	for _, tc := range cases {
		if diags := validateClusterTags(context.Background(), tc.tags); diags.HasError() {
			t.Errorf("%s: unexpected error: %v", tc.name, diags.Errors())
		}
	}
}

func TestValidateClusterTags_Invalid(t *testing.T) {
	cases := []struct {
		name   string
		tags   types.Map
		detail string
	}{
		{
			name:   "too many tags",
			tags:   tagsMap(t, generatedTags(maxClusterTags+1)),
			detail: "a maximum of 23 tags is supported, got 24",
		},
		{
			name:   "empty value",
			tags:   tagsMap(t, map[string]string{"owner": ""}),
			detail: "tag key or tag value can not be empty",
		},
		{
			name:   "empty key",
			tags:   tagsMap(t, map[string]string{"": "platform"}),
			detail: "tag key or tag value can not be empty",
		},
		{
			name:   "reserved aws prefix",
			tags:   tagsMap(t, map[string]string{"aws:cloudformation:stack-name": "mine"}),
			detail: "reserved for AWS use",
		},
		{
			name:   "invalid key character",
			tags:   tagsMap(t, map[string]string{"owner!": "platform"}),
			detail: "expected a valid user tag key 'owner!'",
		},
		{
			name:   "invalid value character",
			tags:   tagsMap(t, map[string]string{"owner": "platform!"}),
			detail: "expected a valid user tag value for key 'owner'",
		},
	}
	for _, tc := range cases {
		diags := validateClusterTags(context.Background(), tc.tags)
		if !diags.HasError() {
			t.Errorf("%s: expected an error, got none", tc.name)
			continue
		}
		if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, tc.detail) {
			t.Errorf("%s: detail = %q, want it to contain %q", tc.name, detail, tc.detail)
		}
	}
}

func generatedTags(count int) map[string]string {
	tags := make(map[string]string, count)
	for i := range count {
		tags[fmt.Sprintf("tag%d", i)] = "value"
	}
	return tags
}
