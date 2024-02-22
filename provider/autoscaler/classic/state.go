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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/autoscaler"
)

type ClusterAutoscalerState struct {
	Cluster                     types.String               `tfsdk:"cluster"`
	BalanceSimilarNodeGroups    types.Bool                 `tfsdk:"balance_similar_node_groups"`
	SkipNodesWithLocalStorage   types.Bool                 `tfsdk:"skip_nodes_with_local_storage"`
	LogVerbosity                types.Int64                `tfsdk:"log_verbosity"`
	MaxPodGracePeriod           types.Int64                `tfsdk:"max_pod_grace_period"`
	PodPriorityThreshold        types.Int64                `tfsdk:"pod_priority_threshold"`
	IgnoreDaemonsetsUtilization types.Bool                 `tfsdk:"ignore_daemonsets_utilization"`
	MaxNodeProvisionTime        types.String               `tfsdk:"max_node_provision_time"`
	BalancingIgnoredLabels      types.List                 `tfsdk:"balancing_ignored_labels"`
	ResourceLimits              *AutoscalerResourceLimits  `tfsdk:"resource_limits"`
	ScaleDown                   *AutoscalerScaleDownConfig `tfsdk:"scale_down"`
}

type AutoscalerResourceLimits struct {
	MaxNodesTotal types.Int64               `tfsdk:"max_nodes_total"`
	Cores         *autoscaler.ResourceRange `tfsdk:"cores"`
	Memory        *autoscaler.ResourceRange `tfsdk:"memory"`
	GPUS          []AutoscalerGPULimit      `tfsdk:"gpus"`
}

type AutoscalerGPULimit struct {
	Type  types.String             `tfsdk:"type"`
	Range autoscaler.ResourceRange `tfsdk:"range"`
}

type AutoscalerScaleDownConfig struct {
	Enabled              types.Bool   `tfsdk:"enabled"`
	UnneededTime         types.String `tfsdk:"unneeded_time"`
	UtilizationThreshold types.String `tfsdk:"utilization_threshold"`
	DelayAfterAdd        types.String `tfsdk:"delay_after_add"`
	DelayAfterDelete     types.String `tfsdk:"delay_after_delete"`
	DelayAfterFailure    types.String `tfsdk:"delay_after_failure"`
}
