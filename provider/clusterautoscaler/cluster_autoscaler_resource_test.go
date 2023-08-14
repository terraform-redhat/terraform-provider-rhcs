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

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

var _ = Describe("Cluster Autoscaler", func() {
	Context("conversion to terraform state", func() {
		It("successfully populates all fields", func() {
			clusterId := "123"
			autoscaler, err := cmv1.NewClusterAutoscaler().
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
			populateAutoscalerState(autoscaler, clusterId, &state)

			Expect(state).To(Equal(ClusterAutoscalerState{
				Cluster:                     types.String{Value: clusterId},
				BalanceSimilarNodeGroups:    types.Bool{Value: true},
				SkipNodesWithLocalStorage:   types.Bool{Value: true},
				LogVerbosity:                types.Int64{Value: 1},
				MaxPodGracePeriod:           types.Int64{Value: 10},
				PodPriorityThreshold:        types.Int64{Value: -10},
				IgnoreDaemonsetsUtilization: types.Bool{Value: true},
				MaxNodeProvisionTime:        types.String{Value: "1h"},
				BalancingIgnoredLabels:      common.StringArrayToList([]string{"l1", "l2"}),
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64{Value: 10},
					Cores: &AutoscalerResourceRange{
						Min: types.Int64{Value: 0},
						Max: types.Int64{Value: 1},
					},
					Memory: &AutoscalerResourceRange{
						Min: types.Int64{Value: 2},
						Max: types.Int64{Value: 3},
					},
					GPUS: []AutoscalerGPULimit{
						{
							Type: types.String{Value: "nvidia"},
							Range: AutoscalerResourceRange{
								Min: types.Int64{Value: 0},
								Max: types.Int64{Value: 1},
							},
						},
						{
							Type: types.String{Value: "intel"},
							Range: AutoscalerResourceRange{
								Min: types.Int64{Value: 2},
								Max: types.Int64{Value: 3},
							},
						},
					},
				},
				ScaleDown: &AutoscalerScaleDownConfig{
					Enabled:              types.Bool{Value: true},
					UnneededTime:         types.String{Value: "2h"},
					UtilizationThreshold: types.String{Value: "0.5"},
					DelayAfterAdd:        types.String{Value: "3h"},
					DelayAfterDelete:     types.String{Value: "4h"},
					DelayAfterFailure:    types.String{Value: "5h"},
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
				Cluster:                     types.String{Value: clusterId},
				BalanceSimilarNodeGroups:    types.Bool{Null: true},
				SkipNodesWithLocalStorage:   types.Bool{Null: true},
				LogVerbosity:                types.Int64{Null: true},
				MaxPodGracePeriod:           types.Int64{Null: true},
				PodPriorityThreshold:        types.Int64{Null: true},
				IgnoreDaemonsetsUtilization: types.Bool{Null: true},
				MaxNodeProvisionTime:        types.String{Null: true},
				BalancingIgnoredLabels:      types.List{Null: true},
				ResourceLimits:              nil,
				ScaleDown:                   nil,
			}))
		})
	})

	Context("conversion to an OCM API object", func() {
		It("successfully converts all fields from a terraform state", func() {
			clusterId := "123"
			state := ClusterAutoscalerState{
				Cluster:                     types.String{Value: clusterId},
				BalanceSimilarNodeGroups:    types.Bool{Value: true},
				SkipNodesWithLocalStorage:   types.Bool{Value: true},
				LogVerbosity:                types.Int64{Value: 1},
				MaxPodGracePeriod:           types.Int64{Value: 10},
				PodPriorityThreshold:        types.Int64{Value: -10},
				IgnoreDaemonsetsUtilization: types.Bool{Value: true},
				MaxNodeProvisionTime:        types.String{Value: "1h"},
				BalancingIgnoredLabels:      common.StringArrayToList([]string{"l1", "l2"}),
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64{Value: 10},
					Cores: &AutoscalerResourceRange{
						Min: types.Int64{Value: 0},
						Max: types.Int64{Value: 1},
					},
					Memory: &AutoscalerResourceRange{
						Min: types.Int64{Value: 2},
						Max: types.Int64{Value: 3},
					},
					GPUS: []AutoscalerGPULimit{
						{
							Type: types.String{Value: "nvidia"},
							Range: AutoscalerResourceRange{
								Min: types.Int64{Value: 0},
								Max: types.Int64{Value: 1},
							},
						},
						{
							Type: types.String{Value: "intel"},
							Range: AutoscalerResourceRange{
								Min: types.Int64{Value: 2},
								Max: types.Int64{Value: 3},
							},
						},
					},
				},
				ScaleDown: &AutoscalerScaleDownConfig{
					Enabled:              types.Bool{Value: true},
					UnneededTime:         types.String{Value: "2h"},
					UtilizationThreshold: types.String{Value: "0.5"},
					DelayAfterAdd:        types.String{Value: "3h"},
					DelayAfterDelete:     types.String{Value: "4h"},
					DelayAfterFailure:    types.String{Value: "5h"},
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
				Cluster:                     types.String{Value: clusterId},
				BalanceSimilarNodeGroups:    types.Bool{Null: true},
				SkipNodesWithLocalStorage:   types.Bool{Null: true},
				LogVerbosity:                types.Int64{Null: true},
				MaxPodGracePeriod:           types.Int64{Null: true},
				PodPriorityThreshold:        types.Int64{Null: true},
				IgnoreDaemonsetsUtilization: types.Bool{Null: true},
				MaxNodeProvisionTime:        types.String{Null: true},
				BalancingIgnoredLabels:      types.List{Null: true},
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
