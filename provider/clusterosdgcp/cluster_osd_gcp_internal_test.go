/*
Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0
*/

// Internal tests cover unexported helpers (buildClusterObject, populateState)
// and exist in the same package so we can call them directly. Each test
// pins a regression for a bug found during the live OSD-GCP smoke test
// described in the PR cover letter; the bug numbers below reference that
// list (e.g. "smoke-test bug #1").

package clusterosdgcp

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// minimalPlan returns a ClusterOsdGcpState wired with the required-attribute
// fields filled in so buildClusterObject can run without nil derefs.
// Optional/Computed fields are left null so each test can opt into setting
// the one attribute it's exercising.
func minimalPlan() *ClusterOsdGcpState {
	return &ClusterOsdGcpState{
		Name:                  types.StringValue("osd-test"),
		CloudRegion:           types.StringValue("us-east1"),
		GCPProjectID:          types.StringValue("test-proj"),
		Product:               types.StringNull(),
		CCSEnabled:            types.BoolValue(true),
		WIFConfigID:           types.StringValue("wif-12345"),
		MultiAZ:               types.BoolValue(false),
		WaitForCreateComplete: types.BoolValue(false),

		// All Object-typed Optional fields default to null so the schema
		// blocks (gcp_network, private_service_connect, security, etc.)
		// don't get populated.
		GCPNetwork:            types.ObjectNull(gcpNetworkObjectType.AttrTypes),
		PrivateServiceConnect: types.ObjectNull(privateServiceConnectObjectType.AttrTypes),
		Security:              types.ObjectNull(securityObjectType.AttrTypes),
		Network:               types.ObjectNull(networkObjectType.AttrTypes),
		Autoscaling:           types.ObjectNull(autoscalingObjectType.AttrTypes),
		Proxy:                 types.ObjectNull(proxyObjectType.AttrTypes),
		AvailabilityZones:     types.ListNull(types.StringType),
		Properties:            types.MapNull(types.StringType),
		AdminCredentials: types.ObjectNull(map[string]attr.Type{
			"username": types.StringType,
			"password": types.StringType,
		}),
	}
}

// Test that buildClusterObject does NOT set BillingModel on the OCM cluster
// body when the user did not provide a value.
//
// Smoke-test bug #1 — billing_model default broke OSD Trial clusters.
// The first live apply against OCM was rejected with
//
//	CLUSTERS-MGMT-400: 'marketplace-gcp' billing_model is not allowed for OSD Trial clusters
//
// because we were hardcoding marketplace-gcp on every CCS cluster. The fix
// (RFC-3) made billing_model a pass-through: forward the user value when
// set, otherwise omit so OCM picks the per-product default (standard for
// osdtrial, marketplace-gcp for osd).
func TestBuildClusterObject_BillingModelPassthrough(t *testing.T) {
	r := &ClusterOsdGcpResource{}

	t.Run("omitted when unset", func(t *testing.T) {
		s := minimalPlan()
		s.BillingModel = types.StringNull()

		var diags diag.Diagnostics
		obj, err := r.buildClusterObject(context.Background(), s, &diags)
		if err != nil {
			t.Fatalf("buildClusterObject: %v", err)
		}
		if diags.HasError() {
			t.Fatalf("buildClusterObject diagnostics: %v", diags)
		}
		if bm, ok := obj.GetBillingModel(); ok {
			t.Errorf("expected no BillingModel on cluster body, got %q", bm)
		}
	})

	t.Run("standard when user sets standard", func(t *testing.T) {
		s := minimalPlan()
		s.BillingModel = types.StringValue("standard")

		var diags diag.Diagnostics
		obj, err := r.buildClusterObject(context.Background(), s, &diags)
		if err != nil {
			t.Fatalf("buildClusterObject: %v", err)
		}
		bm, ok := obj.GetBillingModel()
		if !ok {
			t.Fatal("expected BillingModel set on cluster body, got none")
		}
		if bm != cmv1.BillingModelStandard {
			t.Errorf("expected %q, got %q", cmv1.BillingModelStandard, bm)
		}
	})

	t.Run("marketplace-gcp when user sets marketplace-gcp", func(t *testing.T) {
		s := minimalPlan()
		s.BillingModel = types.StringValue("marketplace-gcp")

		var diags diag.Diagnostics
		obj, err := r.buildClusterObject(context.Background(), s, &diags)
		if err != nil {
			t.Fatalf("buildClusterObject: %v", err)
		}
		bm, ok := obj.GetBillingModel()
		if !ok {
			t.Fatal("expected BillingModel set on cluster body, got none")
		}
		if bm != cmv1.BillingModelMarketplaceGCP {
			t.Errorf("expected %q, got %q", cmv1.BillingModelMarketplaceGCP, bm)
		}
	})
}

// Test that populateState preserves the user's input-only fields when OCM's
// Get response omits them. Before the fix, populateState unconditionally
// assigned nil/null to state.GCPEncryptionKey (and GCPNetwork) whenever
// OCM didn't return the field, which tripped Terraform's plan-after-apply
// consistency check with
//
//	Provider produced inconsistent result after apply
//	.gcp_encryption_key: was cty.ObjectVal(...), but now null
//
// Smoke-test bug #4 — populateState wiped input-only fields when OCM
// omitted them. Fix in RFC-8 / RFC-10: only overwrite the state attribute
// when OCM actually returned a value.
func TestPopulateState_PreservesInputOnlyFields(t *testing.T) {
	r := &ClusterOsdGcpResource{}

	planEncryption := &GCPEncryptionKeyState{
		KmsKeyServiceAccount: types.StringValue("sa@proj.iam.gserviceaccount.com"),
		KeyLocation:          types.StringValue("us-east1"),
		KeyName:              types.StringValue("test-key"),
		KeyRing:              types.StringValue("test-ring"),
	}
	gcpNetworkPlan, _ := types.ObjectValue(gcpNetworkObjectType.AttrTypes, map[string]attr.Value{
		"vpc_name":             types.StringValue("test-vpc"),
		"vpc_project_id":       types.StringValue("host-proj"),
		"compute_subnet":       types.StringValue("test-compute"),
		"control_plane_subnet": types.StringValue("test-cp"),
	})

	state := &ClusterOsdGcpState{
		GCPEncryptionKey: planEncryption,
		GCPNetwork:       gcpNetworkPlan,
	}

	// Build a minimal cmv1.Cluster that does NOT include GCPEncryptionKey
	// or GCPNetwork — the realistic OCM Get response shape.
	cluster, err := cmv1.NewCluster().
		ID("cluster-abc").
		Name("osd-test").
		Region(cmv1.NewCloudRegion().ID("us-east1")).
		Product(cmv1.NewProduct().ID("osdtrial")).
		MultiAZ(false).
		State(cmv1.ClusterStateReady).
		Build()
	if err != nil {
		t.Fatalf("build cluster: %v", err)
	}

	if err := r.populateState(context.Background(), cluster, state); err != nil {
		t.Fatalf("populateState: %v", err)
	}

	if state.GCPEncryptionKey == nil ||
		state.GCPEncryptionKey.KeyName.ValueString() != "test-key" {
		t.Errorf("populateState wiped GCPEncryptionKey (was set in state, omitted by OCM); got %+v",
			state.GCPEncryptionKey)
	}

	attrs := state.GCPNetwork.Attributes()
	if attrs == nil {
		t.Fatal("populateState wiped GCPNetwork (set in state, omitted by OCM)")
	}
	if v, _ := attrs["vpc_name"].(types.String); v.ValueString() != "test-vpc" {
		t.Errorf("populateState wiped GCPNetwork.vpc_name; got %q", v.ValueString())
	}
}
