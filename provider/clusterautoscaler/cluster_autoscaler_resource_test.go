/*
Copyright (c***REMOVED*** 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/types"
***REMOVED*** // nolint
***REMOVED***    // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

var _ = Describe("Cluster Autoscaler", func(***REMOVED*** {
	Context("conversion to terraform state", func(***REMOVED*** {
		It("successfully populates all fields", func(***REMOVED*** {
			clusterId := "123"
			autoscaler, err := cmv1.NewClusterAutoscaler(***REMOVED***.
				BalanceSimilarNodeGroups(true***REMOVED***.
				SkipNodesWithLocalStorage(true***REMOVED***.
				LogVerbosity(1***REMOVED***.
				MaxPodGracePeriod(10***REMOVED***.
				PodPriorityThreshold(-10***REMOVED***.
				IgnoreDaemonsetsUtilization(true***REMOVED***.
				MaxNodeProvisionTime("1h"***REMOVED***.
				BalancingIgnoredLabels("l1", "l2"***REMOVED***.
				ResourceLimits(cmv1.NewAutoscalerResourceLimits(***REMOVED***.
					MaxNodesTotal(10***REMOVED***.
					Cores(cmv1.NewResourceRange(***REMOVED***.
						Min(0***REMOVED***.
						Max(1***REMOVED******REMOVED***.
					Memory(cmv1.NewResourceRange(***REMOVED***.
						Min(2***REMOVED***.
						Max(3***REMOVED******REMOVED***.
					GPUS(
						cmv1.NewAutoscalerResourceLimitsGPULimit(***REMOVED***.
							Type("nvidia"***REMOVED***.
							Range(cmv1.NewResourceRange(***REMOVED***.
								Min(0***REMOVED***.
								Max(1***REMOVED******REMOVED***,
						cmv1.NewAutoscalerResourceLimitsGPULimit(***REMOVED***.
							Type("intel"***REMOVED***.
							Range(cmv1.NewResourceRange(***REMOVED***.
								Min(2***REMOVED***.
								Max(3***REMOVED******REMOVED***,
					***REMOVED******REMOVED***.
				ScaleDown(cmv1.NewAutoscalerScaleDownConfig(***REMOVED***.
					Enabled(true***REMOVED***.
					UnneededTime("2h"***REMOVED***.
					UtilizationThreshold("0.5"***REMOVED***.
					DelayAfterAdd("3h"***REMOVED***.
					DelayAfterDelete("4h"***REMOVED***.
					DelayAfterFailure("5h"***REMOVED******REMOVED***.Build(***REMOVED***

			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

			var state ClusterAutoscalerState
			populateAutoscalerState(autoscaler, clusterId, &state***REMOVED***

			Expect(state***REMOVED***.To(Equal(ClusterAutoscalerState{
				Cluster:                     types.String{Value: clusterId},
				BalanceSimilarNodeGroups:    types.Bool{Value: true},
				SkipNodesWithLocalStorage:   types.Bool{Value: true},
				LogVerbosity:                types.Int64{Value: 1},
				MaxPodGracePeriod:           types.Int64{Value: 10},
				PodPriorityThreshold:        types.Int64{Value: -10},
				IgnoreDaemonsetsUtilization: types.Bool{Value: true},
				MaxNodeProvisionTime:        types.String{Value: "1h"},
				BalancingIgnoredLabels:      common.StringArrayToList([]string{"l1", "l2"}***REMOVED***,
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64{Value: 10},
					Cores: &AutoscalerResourceRange{
						Min: types.Int64{Value: 0},
						Max: types.Int64{Value: 1},
			***REMOVED***,
					Memory: &AutoscalerResourceRange{
						Min: types.Int64{Value: 2},
						Max: types.Int64{Value: 3},
			***REMOVED***,
					GPUS: []AutoscalerGPULimit{
						{
							Type: types.String{Value: "nvidia"},
							Range: AutoscalerResourceRange{
								Min: types.Int64{Value: 0},
								Max: types.Int64{Value: 1},
					***REMOVED***,
				***REMOVED***,
						{
							Type: types.String{Value: "intel"},
							Range: AutoscalerResourceRange{
								Min: types.Int64{Value: 2},
								Max: types.Int64{Value: 3},
					***REMOVED***,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
				ScaleDown: &AutoscalerScaleDownConfig{
					Enabled:              types.Bool{Value: true},
					UnneededTime:         types.String{Value: "2h"},
					UtilizationThreshold: types.String{Value: "0.5"},
					DelayAfterAdd:        types.String{Value: "3h"},
					DelayAfterDelete:     types.String{Value: "4h"},
					DelayAfterFailure:    types.String{Value: "5h"},
		***REMOVED***,
	***REMOVED******REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("successfully populates empty fields", func(***REMOVED*** {
			clusterId := "123"
			autoscaler, err := cmv1.NewClusterAutoscaler(***REMOVED***.Build(***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

			var state ClusterAutoscalerState
			populateAutoscalerState(autoscaler, clusterId, &state***REMOVED***

			Expect(state***REMOVED***.To(Equal(ClusterAutoscalerState{
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
	***REMOVED******REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	Context("conversion to an OCM API object", func(***REMOVED*** {
		It("successfully converts all fields from a terraform state", func(***REMOVED*** {
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
				BalancingIgnoredLabels:      common.StringArrayToList([]string{"l1", "l2"}***REMOVED***,
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64{Value: 10},
					Cores: &AutoscalerResourceRange{
						Min: types.Int64{Value: 0},
						Max: types.Int64{Value: 1},
			***REMOVED***,
					Memory: &AutoscalerResourceRange{
						Min: types.Int64{Value: 2},
						Max: types.Int64{Value: 3},
			***REMOVED***,
					GPUS: []AutoscalerGPULimit{
						{
							Type: types.String{Value: "nvidia"},
							Range: AutoscalerResourceRange{
								Min: types.Int64{Value: 0},
								Max: types.Int64{Value: 1},
					***REMOVED***,
				***REMOVED***,
						{
							Type: types.String{Value: "intel"},
							Range: AutoscalerResourceRange{
								Min: types.Int64{Value: 2},
								Max: types.Int64{Value: 3},
					***REMOVED***,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
				ScaleDown: &AutoscalerScaleDownConfig{
					Enabled:              types.Bool{Value: true},
					UnneededTime:         types.String{Value: "2h"},
					UtilizationThreshold: types.String{Value: "0.5"},
					DelayAfterAdd:        types.String{Value: "3h"},
					DelayAfterDelete:     types.String{Value: "4h"},
					DelayAfterFailure:    types.String{Value: "5h"},
		***REMOVED***,
	***REMOVED***

			autoscaler, err := clusterAutoscalerStateToObject(&state***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

			expectedAutoscaler, err := cmv1.NewClusterAutoscaler(***REMOVED***.
				BalanceSimilarNodeGroups(true***REMOVED***.
				SkipNodesWithLocalStorage(true***REMOVED***.
				LogVerbosity(1***REMOVED***.
				MaxPodGracePeriod(10***REMOVED***.
				PodPriorityThreshold(-10***REMOVED***.
				IgnoreDaemonsetsUtilization(true***REMOVED***.
				MaxNodeProvisionTime("1h"***REMOVED***.
				BalancingIgnoredLabels("l1", "l2"***REMOVED***.
				ResourceLimits(cmv1.NewAutoscalerResourceLimits(***REMOVED***.
					MaxNodesTotal(10***REMOVED***.
					Cores(cmv1.NewResourceRange(***REMOVED***.
						Min(0***REMOVED***.
						Max(1***REMOVED******REMOVED***.
					Memory(cmv1.NewResourceRange(***REMOVED***.
						Min(2***REMOVED***.
						Max(3***REMOVED******REMOVED***.
					GPUS(
						cmv1.NewAutoscalerResourceLimitsGPULimit(***REMOVED***.
							Type("nvidia"***REMOVED***.
							Range(cmv1.NewResourceRange(***REMOVED***.
								Min(0***REMOVED***.
								Max(1***REMOVED******REMOVED***,
						cmv1.NewAutoscalerResourceLimitsGPULimit(***REMOVED***.
							Type("intel"***REMOVED***.
							Range(cmv1.NewResourceRange(***REMOVED***.
								Min(2***REMOVED***.
								Max(3***REMOVED******REMOVED***,
					***REMOVED******REMOVED***.
				ScaleDown(cmv1.NewAutoscalerScaleDownConfig(***REMOVED***.
					Enabled(true***REMOVED***.
					UnneededTime("2h"***REMOVED***.
					UtilizationThreshold("0.5"***REMOVED***.
					DelayAfterAdd("3h"***REMOVED***.
					DelayAfterDelete("4h"***REMOVED***.
					DelayAfterFailure("5h"***REMOVED******REMOVED***.Build(***REMOVED***

			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

			Expect(autoscaler***REMOVED***.To(Equal(expectedAutoscaler***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("successfully converts when all state fields are null", func(***REMOVED*** {
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
	***REMOVED***

			autoscaler, err := clusterAutoscalerStateToObject(&state***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

			expectedAutoscaler, err := cmv1.NewClusterAutoscaler(***REMOVED***.Build(***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
			Expect(autoscaler***REMOVED***.To(Equal(expectedAutoscaler***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
