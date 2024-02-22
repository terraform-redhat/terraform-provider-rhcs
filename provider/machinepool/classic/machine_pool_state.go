/*
Copyright (c) 2021 Red Hat, Inc.

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
)

type MachinePoolState struct {
	Cluster                    types.String  `tfsdk:"cluster"`
	ID                         types.String  `tfsdk:"id"`
	MachineType                types.String  `tfsdk:"machine_type"`
	Name                       types.String  `tfsdk:"name"`
	Replicas                   types.Int64   `tfsdk:"replicas"`
	UseSpotInstances           types.Bool    `tfsdk:"use_spot_instances"`
	MaxSpotPrice               types.Float64 `tfsdk:"max_spot_price"`
	AutoScalingEnabled         types.Bool    `tfsdk:"autoscaling_enabled"`
	MinReplicas                types.Int64   `tfsdk:"min_replicas"`
	MaxReplicas                types.Int64   `tfsdk:"max_replicas"`
	Taints                     []Taints      `tfsdk:"taints"`
	Labels                     types.Map     `tfsdk:"labels"`
	MultiAvailabilityZone      types.Bool    `tfsdk:"multi_availability_zone"`
	AvailabilityZone           types.String  `tfsdk:"availability_zone"`
	AvailabilityZones          types.List    `tfsdk:"availability_zones"`
	SubnetID                   types.String  `tfsdk:"subnet_id"`
	SubnetIDs                  types.List    `tfsdk:"subnet_ids"`
	DiskSize                   types.Int64   `tfsdk:"disk_size"`
	AdditionalSecurityGroupIds types.List    `tfsdk:"aws_additional_security_group_ids"`
}

type Taints struct {
	Key          types.String `tfsdk:"key"`
	Value        types.String `tfsdk:"value"`
	ScheduleType types.String `tfsdk:"schedule_type"`
}
