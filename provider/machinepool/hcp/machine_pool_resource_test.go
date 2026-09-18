/*
Copyright (c) 2024 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func TestHcpMachinePool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HCP Machine Pool Resource Suite")
}

// clusterForNodePoolTests returns a minimal Cluster object with the fields
// that populateState needs (version channel-group, AWS tags).
func clusterForNodePoolTests() *cmv1.Cluster {
	clusterJSON := map[string]any{
		"id":   "test-cluster-123",
		"name": "test-cluster",
		"version": map[string]any{
			"id":            "openshift-v4.16.0",
			"channel_group": "stable",
		},
		"aws": map[string]any{
			"tags": map[string]string{},
		},
	}
	raw, err := json.Marshal(clusterJSON)
	Expect(err).ToNot(HaveOccurred())
	cluster, err := cmv1.UnmarshalCluster(raw)
	Expect(err).ToNot(HaveOccurred())
	return cluster
}

// buildNodePoolJSON returns a base node pool JSON map with common fields set.
func buildNodePoolJSON(instanceProfile string) map[string]any {
	np := map[string]any{
		"id":     "test-pool",
		"subnet": "subnet-abc123",
		"aws_node_pool": map[string]any{
			"instance_type": "m5.xlarge",
		},
		"auto_repair": true,
		"replicas":    2,
		"version": map[string]any{
			"id": "openshift-v4.16.0",
		},
		"availability_zone": "us-east-1a",
	}
	awsNP := np["aws_node_pool"].(map[string]any)
	if instanceProfile != "" {
		awsNP["instance_profile"] = instanceProfile
	}
	return np
}

var _ = Describe("HCP Machine Pool populateState", func() {
	var (
		ctx     context.Context
		cluster *cmv1.Cluster
	)

	BeforeEach(func() {
		ctx = context.Background()
		cluster = clusterForNodePoolTests()
	})

	// Regression test for the no-op PATCH early-return optimization in
	// Cluster Service (rosa-clusters-service PR #59). When the PATCH
	// response omits instance_profile (empty string), the old code would
	// store an empty value in Terraform state, causing "Provider produced
	// inconsistent result after apply". The fix performs a canonical GET
	// after PATCH; this test verifies that populateState from the GET
	// response (which includes instance_profile) correctly stores the
	// real value.
	Context("instance_profile handling on update", func() {
		It("stores empty instance_profile when PATCH response omits it", func() {
			// Simulate a PATCH response with empty instance_profile
			// (the early-return path in Cluster Service).
			npJSON := buildNodePoolJSON("")
			raw, err := json.Marshal(npJSON)
			Expect(err).ToNot(HaveOccurred())

			nodePool, err := cmv1.UnmarshalNodePool(raw)
			Expect(err).ToNot(HaveOccurred())

			state := &HcpMachinePoolState{
				AWSNodePool: &AWSNodePool{
					Tags: types.MapNull(types.StringType),
				},
			}
			err = populateState(ctx, nodePool, state, cluster)
			Expect(err).ToNot(HaveOccurred())
			// With the PATCH response that omits instance_profile,
			// populateState does NOT set instance_profile (because
			// GetInstanceProfile returns ok=false for an absent field).
			// This confirms why using the PATCH response was wrong:
			// the state would retain whatever was there before (or be
			// empty for a fresh state struct).
		})

		It("stores real instance_profile from GET response", func() {
			// Simulate the follow-up GET response that includes
			// the real instance_profile.
			realProfile := "arn:aws:iam::123456789012:instance-profile/test-profile"
			npJSON := buildNodePoolJSON(realProfile)
			raw, err := json.Marshal(npJSON)
			Expect(err).ToNot(HaveOccurred())

			nodePool, err := cmv1.UnmarshalNodePool(raw)
			Expect(err).ToNot(HaveOccurred())

			state := &HcpMachinePoolState{
				AWSNodePool: &AWSNodePool{
					Tags: types.MapNull(types.StringType),
				},
			}
			err = populateState(ctx, nodePool, state, cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(state.AWSNodePool.InstanceProfile.ValueString()).To(
				Equal(realProfile))
		})

		It("retains real instance_profile from GET after PATCH returns empty value", func() {
			// This is the end-to-end regression scenario:
			// 1. PATCH response has empty instance_profile
			// 2. GET response has the real instance_profile
			// 3. State should have the real value after populateState
			//    from the GET response.
			//
			// Before the fix, state was populated from the PATCH
			// response. After the fix, state is populated from the
			// GET response.

			realProfile := "arn:aws:iam::123456789012:instance-profile/test-np-profile"

			// Step 1: Simulate populating from a PATCH response
			// (empty instance_profile — the bug scenario).
			patchJSON := buildNodePoolJSON("")
			patchRaw, err := json.Marshal(patchJSON)
			Expect(err).ToNot(HaveOccurred())
			patchNP, err := cmv1.UnmarshalNodePool(patchRaw)
			Expect(err).ToNot(HaveOccurred())

			patchState := &HcpMachinePoolState{
				AWSNodePool: &AWSNodePool{
					Tags: types.MapNull(types.StringType),
				},
			}
			err = populateState(ctx, patchNP, patchState, cluster)
			Expect(err).ToNot(HaveOccurred())

			// Verify: the PATCH response did NOT populate
			// instance_profile with the real value. This is the bug
			// that the code fix addresses by using GET instead.
			Expect(patchState.AWSNodePool.InstanceProfile.ValueString()).ToNot(
				Equal(realProfile),
				"PATCH response should NOT contain the real instance_profile")

			// Step 2: Simulate populating from the follow-up GET
			// response (contains the real instance_profile).
			getJSON := buildNodePoolJSON(realProfile)
			getRaw, err := json.Marshal(getJSON)
			Expect(err).ToNot(HaveOccurred())
			getNP, err := cmv1.UnmarshalNodePool(getRaw)
			Expect(err).ToNot(HaveOccurred())

			getState := &HcpMachinePoolState{
				AWSNodePool: &AWSNodePool{
					Tags: types.MapNull(types.StringType),
				},
			}
			err = populateState(ctx, getNP, getState, cluster)
			Expect(err).ToNot(HaveOccurred())

			// Verify: the GET response correctly populates
			// instance_profile with the real value.
			Expect(getState.AWSNodePool.InstanceProfile.ValueString()).To(
				Equal(realProfile),
				"GET response must populate instance_profile with the real value")
		})
	})
})
