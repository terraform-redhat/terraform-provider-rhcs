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
)

type ClusterAutoscalerState struct {
	Cluster              types.String              `tfsdk:"cluster"`
	MaxPodGracePeriod    types.Int64               `tfsdk:"max_pod_grace_period"`
	PodPriorityThreshold types.Int64               `tfsdk:"pod_priority_threshold"`
	MaxNodeProvisionTime types.String              `tfsdk:"max_node_provision_time"`
	ResourceLimits       *AutoscalerResourceLimits `tfsdk:"resource_limits"`
}

type AutoscalerResourceLimits struct {
	MaxNodesTotal types.Int64 `tfsdk:"max_nodes_total"`
}
