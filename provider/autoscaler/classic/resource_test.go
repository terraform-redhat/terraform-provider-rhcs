/*
Copyright (c) 2023 Red Hat, Inc.

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

package classic

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/autoscaler"
)

var _ = Describe("Cluster Autoscaler", func() {
	Context("conversion to terraform state", func() {
		It("successfully populates all fields", func() {
			clusterId := "123"
			balIgnLables, _ := types.ListValue(types.StringType, []attr.Value{
				types.StringValue("l1"),
				types.StringValue("l2"),
			})
			autoscalerSpec, err := cmv1.NewClusterAutoscaler().
				BalanceSimilarNodeGroups(true).
				SkipNodesWithLocalStorage(true).
				LogVerbosity(1).
				MaxPodGracePeriod(10).
				PodPriorityThreshold(-10).
				IgnoreDaemonsetsUtilization(true).
				MaxNodeProvisionTime("1h").
				BalancingIgnoredLabels("l1", "l2").
				ResourceLimits(cmv1.NewAutoscalerResourceLimits().
					MaxNodesTotal(10).
					Cores(cmv1.NewResourceRange().
						Min(0).
						Max(1)).
					Memory(cmv1.NewResourceRange().
						Min(2).
						Max(3)).
					GPUS(
						cmv1.NewAutoscalerResourceLimitsGPULimit().
							Type("nvidia").
							Range(cmv1.NewResourceRange().
								Min(0).
								Max(1)),
						cmv1.NewAutoscalerResourceLimitsGPULimit().
							Type("intel").
							Range(cmv1.NewResourceRange().
								Min(2).
								Max(3)),
					)).
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
				Cluster:                     types.StringValue(clusterId),
				BalanceSimilarNodeGroups:    types.BoolValue(true),
				SkipNodesWithLocalStorage:   types.BoolValue(true),
				LogVerbosity:                types.Int64Value(1),
				MaxPodGracePeriod:           types.Int64Value(10),
				PodPriorityThreshold:        types.Int64Value(-10),
				IgnoreDaemonsetsUtilization: types.BoolValue(true),
				MaxNodeProvisionTime:        types.StringValue("1h"),
				BalancingIgnoredLabels:      balIgnLables,
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64Value(10),
					Cores: &autoscaler.ResourceRange{
						Min: types.Int64Value(0),
						Max: types.Int64Value(1),
					},
					Memory: &autoscaler.ResourceRange{
						Min: types.Int64Value(2),
						Max: types.Int64Value(3),
					},
					GPUS: []AutoscalerGPULimit{
						{
							Type: types.StringValue("nvidia"),
							Range: autoscaler.ResourceRange{
								Min: types.Int64Value(0),
								Max: types.Int64Value(1),
							},
						},
						{
							Type: types.StringValue("intel"),
							Range: autoscaler.ResourceRange{
								Min: types.Int64Value(2),
								Max: types.Int64Value(3),
							},
						},
					},
				},
				ScaleDown: &AutoscalerScaleDownConfig{
					Enabled:              types.BoolValue(true),
					UnneededTime:         types.StringValue("2h"),
					UtilizationThreshold: types.StringValue("0.5"),
					DelayAfterAdd:        types.StringValue("3h"),
					DelayAfterDelete:     types.StringValue("4h"),
					DelayAfterFailure:    types.StringValue("5h"),
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
				Cluster:                     types.StringValue(clusterId),
				BalanceSimilarNodeGroups:    types.BoolNull(),
				SkipNodesWithLocalStorage:   types.BoolNull(),
				LogVerbosity:                types.Int64Null(),
				MaxPodGracePeriod:           types.Int64Null(),
				PodPriorityThreshold:        types.Int64Null(),
				IgnoreDaemonsetsUtilization: types.BoolNull(),
				MaxNodeProvisionTime:        types.StringNull(),
				BalancingIgnoredLabels:      types.ListNull(types.StringType),
				ResourceLimits:              nil,
				ScaleDown:                   nil,
			}))
		})
	})

	Context("conversion to an OCM API object", func() {
		It("successfully converts all fields from a terraform state", func() {
			clusterId := "123"
			balIgnLables, _ := types.ListValue(types.StringType, []attr.Value{
				types.StringValue("l1"),
				types.StringValue("l2"),
			})
			state := ClusterAutoscalerState{
				Cluster:                     types.StringValue(clusterId),
				BalanceSimilarNodeGroups:    types.BoolValue(true),
				SkipNodesWithLocalStorage:   types.BoolValue(true),
				LogVerbosity:                types.Int64Value(1),
				MaxPodGracePeriod:           types.Int64Value(10),
				PodPriorityThreshold:        types.Int64Value(-10),
				IgnoreDaemonsetsUtilization: types.BoolValue(true),
				MaxNodeProvisionTime:        types.StringValue("1h"),
				BalancingIgnoredLabels:      balIgnLables,
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64Value(10),
					Cores: &autoscaler.ResourceRange{
						Min: types.Int64Value(0),
						Max: types.Int64Value(1),
					},
					Memory: &autoscaler.ResourceRange{
						Min: types.Int64Value(2),
						Max: types.Int64Value(3),
					},
					GPUS: []AutoscalerGPULimit{
						{
							Type: types.StringValue("nvidia"),
							Range: autoscaler.ResourceRange{
								Min: types.Int64Value(0),
								Max: types.Int64Value(1),
							},
						},
						{
							Type: types.StringValue("intel"),
							Range: autoscaler.ResourceRange{
								Min: types.Int64Value(2),
								Max: types.Int64Value(3),
							},
						},
					},
				},
				ScaleDown: &AutoscalerScaleDownConfig{
					Enabled:              types.BoolValue(true),
					UnneededTime:         types.StringValue("2h"),
					UtilizationThreshold: types.StringValue("0.5"),
					DelayAfterAdd:        types.StringValue("3h"),
					DelayAfterDelete:     types.StringValue("4h"),
					DelayAfterFailure:    types.StringValue("5h"),
				},
			}

			autoscaler, err := clusterAutoscalerStateToObject(&state)
			Expect(err).ToNot(HaveOccurred())

			expectedAutoscaler, err := cmv1.NewClusterAutoscaler().
				BalanceSimilarNodeGroups(true).
				SkipNodesWithLocalStorage(true).
				LogVerbosity(1).
				MaxPodGracePeriod(10).
				PodPriorityThreshold(-10).
				IgnoreDaemonsetsUtilization(true).
				MaxNodeProvisionTime("1h").
				BalancingIgnoredLabels("l1", "l2").
				ResourceLimits(cmv1.NewAutoscalerResourceLimits().
					MaxNodesTotal(10).
					Cores(cmv1.NewResourceRange().
						Min(0).
						Max(1)).
					Memory(cmv1.NewResourceRange().
						Min(2).
						Max(3)).
					GPUS(
						cmv1.NewAutoscalerResourceLimitsGPULimit().
							Type("nvidia").
							Range(cmv1.NewResourceRange().
								Min(0).
								Max(1)),
						cmv1.NewAutoscalerResourceLimitsGPULimit().
							Type("intel").
							Range(cmv1.NewResourceRange().
								Min(2).
								Max(3)),
					)).
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
				Cluster:                     types.StringValue(clusterId),
				BalanceSimilarNodeGroups:    types.BoolNull(),
				SkipNodesWithLocalStorage:   types.BoolNull(),
				LogVerbosity:                types.Int64Null(),
				MaxPodGracePeriod:           types.Int64Null(),
				PodPriorityThreshold:        types.Int64Null(),
				IgnoreDaemonsetsUtilization: types.BoolNull(),
				MaxNodeProvisionTime:        types.StringNull(),
				BalancingIgnoredLabels:      types.ListNull(types.StringType),
				ResourceLimits:              nil,
				ScaleDown:                   nil,
			}

			autoscaler, err := clusterAutoscalerStateToObject(&state)
			Expect(err).ToNot(HaveOccurred())

			expectedAutoscaler, err := cmv1.NewClusterAutoscaler().Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(autoscaler).To(Equal(expectedAutoscaler))
		})
	})
})
