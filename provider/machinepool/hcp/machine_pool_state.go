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

type HcpMachinePoolState struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Cluster     types.String `tfsdk:"cluster"`
	Replicas    types.Int64  `tfsdk:"replicas"`
	AutoScaling *AutoScaling `tfsdk:"autoscaling"`

	Taints           []Taints     `tfsdk:"taints"`
	Labels           types.Map    `tfsdk:"labels"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	SubnetID         types.String `tfsdk:"subnet_id"`

	Version        types.String `tfsdk:"version"`
	CurrentVersion types.String `tfsdk:"current_version"`

	UpgradeAcksFor types.String `tfsdk:"upgrade_acknowledgements_for"`

	NodePoolStatus types.Object `tfsdk:"status"`
	AWSNodePool    *AWSNodePool `tfsdk:"aws_node_pool"`
	TuningConfigs  types.List   `tfsdk:"tuning_configs"`
	AutoRepair     types.Bool   `tfsdk:"auto_repair"`
}

type Taints struct {
	Key          types.String `tfsdk:"key"`
	Value        types.String `tfsdk:"value"`
	ScheduleType types.String `tfsdk:"schedule_type"`
}
