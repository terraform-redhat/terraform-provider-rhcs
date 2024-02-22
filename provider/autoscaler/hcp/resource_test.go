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
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

var _ = Describe("Cluster Autoscaler", func() {
	Context("conversion to terraform state", func() {
		It("successfully populates all fields", func() {
			clusterId := "123"
			autoscalerSpec, err := cmv1.NewClusterAutoscaler().
				MaxPodGracePeriod(10).
				PodPriorityThreshold(-10).
				MaxNodeProvisionTime("1h").
				ResourceLimits(cmv1.NewAutoscalerResourceLimits().
					MaxNodesTotal(10)).
				ScaleDown(cmv1.NewAutoscalerScaleDownConfig().
					Enabled(true).
					UnneededTime("2h").
					UtilizationThreshold("0.5").
					DelayAfterAdd("3h").
					DelayAfterDelete("4h").
					DelayAfterFailure("5h")).Build()

			Expect(err).ToNot(HaveOccurred())

			var state ClusterAutoscalerState
			populateAutoscalerState(autoscalerSpec, clusterId, &state)

			Expect(state).To(Equal(ClusterAutoscalerState{
				Cluster:              types.StringValue(clusterId),
				MaxPodGracePeriod:    types.Int64Value(10),
				PodPriorityThreshold: types.Int64Value(-10),
				MaxNodeProvisionTime: types.StringValue("1h"),
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64Value(10),
				},
			}))
		})

		It("successfully populates empty fields", func() {
			clusterId := "123"
			autoscaler, err := cmv1.NewClusterAutoscaler().Build()
			Expect(err).ToNot(HaveOccurred())

			var state ClusterAutoscalerState
			populateAutoscalerState(autoscaler, clusterId, &state)

			Expect(state).To(Equal(ClusterAutoscalerState{
				Cluster:              types.StringValue(clusterId),
				MaxPodGracePeriod:    types.Int64Null(),
				PodPriorityThreshold: types.Int64Null(),
				MaxNodeProvisionTime: types.StringNull(),
				ResourceLimits:       nil,
			}))
		})
	})

	Context("conversion to an OCM API object", func() {
		It("successfully converts all fields from a terraform state", func() {
			clusterId := "123"
			state := ClusterAutoscalerState{
				Cluster:              types.StringValue(clusterId),
				MaxPodGracePeriod:    types.Int64Value(10),
				PodPriorityThreshold: types.Int64Value(-10),
				MaxNodeProvisionTime: types.StringValue("1h"),
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64Value(10),
				},
			}

			autoscaler, err := clusterAutoscalerStateToObject(&state)
			Expect(err).ToNot(HaveOccurred())

			expectedAutoscaler, err := cmv1.NewClusterAutoscaler().
				MaxPodGracePeriod(10).
				PodPriorityThreshold(-10).
				MaxNodeProvisionTime("1h").
				ResourceLimits(cmv1.NewAutoscalerResourceLimits().
					MaxNodesTotal(10)).
				ScaleDown(cmv1.NewAutoscalerScaleDownConfig().
					Enabled(true).
					UnneededTime("2h").
					UtilizationThreshold("0.5").
					DelayAfterAdd("3h").
					DelayAfterDelete("4h").
					DelayAfterFailure("5h")).Build()

			Expect(err).ToNot(HaveOccurred())

			Expect(autoscaler).To(Equal(expectedAutoscaler))
		})

		It("successfully converts when all state fields are null", func() {
			clusterId := "123"
			state := ClusterAutoscalerState{
				Cluster:              types.StringValue(clusterId),
				MaxPodGracePeriod:    types.Int64Null(),
				PodPriorityThreshold: types.Int64Null(),
				MaxNodeProvisionTime: types.StringNull(),
				ResourceLimits:       nil,
			}

			autoscaler, err := clusterAutoscalerStateToObject(&state)
			Expect(err).ToNot(HaveOccurred())

			expectedAutoscaler, err := cmv1.NewClusterAutoscaler().Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(autoscaler).To(Equal(expectedAutoscaler))
		})
	})
})
