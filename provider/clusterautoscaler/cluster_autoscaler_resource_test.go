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

package clusterautoscaler

***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
***REMOVED*** // nolint
***REMOVED***    // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

var _ = Describe("Cluster Autoscaler", func(***REMOVED*** {
	Context("conversion to terraform state", func(***REMOVED*** {
		It("successfully populates all fields", func(***REMOVED*** {
			clusterId := "123"
			balIgnLables, _ := types.ListValue(types.StringType, []attr.Value{
				types.StringValue("l1"***REMOVED***,
				types.StringValue("l2"***REMOVED***,
	***REMOVED******REMOVED***
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
				Cluster:                     types.StringValue(clusterId***REMOVED***,
				BalanceSimilarNodeGroups:    types.BoolValue(true***REMOVED***,
				SkipNodesWithLocalStorage:   types.BoolValue(true***REMOVED***,
				LogVerbosity:                types.Int64Value(1***REMOVED***,
				MaxPodGracePeriod:           types.Int64Value(10***REMOVED***,
				PodPriorityThreshold:        types.Int64Value(-10***REMOVED***,
				IgnoreDaemonsetsUtilization: types.BoolValue(true***REMOVED***,
				MaxNodeProvisionTime:        types.StringValue("1h"***REMOVED***,
				BalancingIgnoredLabels:      balIgnLables,
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64Value(10***REMOVED***,
					Cores: &AutoscalerResourceRange{
						Min: types.Int64Value(0***REMOVED***,
						Max: types.Int64Value(1***REMOVED***,
			***REMOVED***,
					Memory: &AutoscalerResourceRange{
						Min: types.Int64Value(2***REMOVED***,
						Max: types.Int64Value(3***REMOVED***,
			***REMOVED***,
					GPUS: []AutoscalerGPULimit{
						{
							Type: types.StringValue("nvidia"***REMOVED***,
							Range: AutoscalerResourceRange{
								Min: types.Int64Value(0***REMOVED***,
								Max: types.Int64Value(1***REMOVED***,
					***REMOVED***,
				***REMOVED***,
						{
							Type: types.StringValue("intel"***REMOVED***,
							Range: AutoscalerResourceRange{
								Min: types.Int64Value(2***REMOVED***,
								Max: types.Int64Value(3***REMOVED***,
					***REMOVED***,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
				ScaleDown: &AutoscalerScaleDownConfig{
					Enabled:              types.BoolValue(true***REMOVED***,
					UnneededTime:         types.StringValue("2h"***REMOVED***,
					UtilizationThreshold: types.StringValue("0.5"***REMOVED***,
					DelayAfterAdd:        types.StringValue("3h"***REMOVED***,
					DelayAfterDelete:     types.StringValue("4h"***REMOVED***,
					DelayAfterFailure:    types.StringValue("5h"***REMOVED***,
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
				Cluster:                     types.StringValue(clusterId***REMOVED***,
				BalanceSimilarNodeGroups:    types.BoolNull(***REMOVED***,
				SkipNodesWithLocalStorage:   types.BoolNull(***REMOVED***,
				LogVerbosity:                types.Int64Null(***REMOVED***,
				MaxPodGracePeriod:           types.Int64Null(***REMOVED***,
				PodPriorityThreshold:        types.Int64Null(***REMOVED***,
				IgnoreDaemonsetsUtilization: types.BoolNull(***REMOVED***,
				MaxNodeProvisionTime:        types.StringNull(***REMOVED***,
				BalancingIgnoredLabels:      types.ListNull(types.StringType***REMOVED***,
				ResourceLimits:              nil,
				ScaleDown:                   nil,
	***REMOVED******REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	Context("conversion to an OCM API object", func(***REMOVED*** {
		It("successfully converts all fields from a terraform state", func(***REMOVED*** {
			clusterId := "123"
			balIgnLables, _ := types.ListValue(types.StringType, []attr.Value{
				types.StringValue("l1"***REMOVED***,
				types.StringValue("l2"***REMOVED***,
	***REMOVED******REMOVED***
			state := ClusterAutoscalerState{
				Cluster:                     types.StringValue(clusterId***REMOVED***,
				BalanceSimilarNodeGroups:    types.BoolValue(true***REMOVED***,
				SkipNodesWithLocalStorage:   types.BoolValue(true***REMOVED***,
				LogVerbosity:                types.Int64Value(1***REMOVED***,
				MaxPodGracePeriod:           types.Int64Value(10***REMOVED***,
				PodPriorityThreshold:        types.Int64Value(-10***REMOVED***,
				IgnoreDaemonsetsUtilization: types.BoolValue(true***REMOVED***,
				MaxNodeProvisionTime:        types.StringValue("1h"***REMOVED***,
				BalancingIgnoredLabels:      balIgnLables,
				ResourceLimits: &AutoscalerResourceLimits{
					MaxNodesTotal: types.Int64Value(10***REMOVED***,
					Cores: &AutoscalerResourceRange{
						Min: types.Int64Value(0***REMOVED***,
						Max: types.Int64Value(1***REMOVED***,
			***REMOVED***,
					Memory: &AutoscalerResourceRange{
						Min: types.Int64Value(2***REMOVED***,
						Max: types.Int64Value(3***REMOVED***,
			***REMOVED***,
					GPUS: []AutoscalerGPULimit{
						{
							Type: types.StringValue("nvidia"***REMOVED***,
							Range: AutoscalerResourceRange{
								Min: types.Int64Value(0***REMOVED***,
								Max: types.Int64Value(1***REMOVED***,
					***REMOVED***,
				***REMOVED***,
						{
							Type: types.StringValue("intel"***REMOVED***,
							Range: AutoscalerResourceRange{
								Min: types.Int64Value(2***REMOVED***,
								Max: types.Int64Value(3***REMOVED***,
					***REMOVED***,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
				ScaleDown: &AutoscalerScaleDownConfig{
					Enabled:              types.BoolValue(true***REMOVED***,
					UnneededTime:         types.StringValue("2h"***REMOVED***,
					UtilizationThreshold: types.StringValue("0.5"***REMOVED***,
					DelayAfterAdd:        types.StringValue("3h"***REMOVED***,
					DelayAfterDelete:     types.StringValue("4h"***REMOVED***,
					DelayAfterFailure:    types.StringValue("5h"***REMOVED***,
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
				Cluster:                     types.StringValue(clusterId***REMOVED***,
				BalanceSimilarNodeGroups:    types.BoolNull(***REMOVED***,
				SkipNodesWithLocalStorage:   types.BoolNull(***REMOVED***,
				LogVerbosity:                types.Int64Null(***REMOVED***,
				MaxPodGracePeriod:           types.Int64Null(***REMOVED***,
				PodPriorityThreshold:        types.Int64Null(***REMOVED***,
				IgnoreDaemonsetsUtilization: types.BoolNull(***REMOVED***,
				MaxNodeProvisionTime:        types.StringNull(***REMOVED***,
				BalancingIgnoredLabels:      types.ListNull(types.StringType***REMOVED***,
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
