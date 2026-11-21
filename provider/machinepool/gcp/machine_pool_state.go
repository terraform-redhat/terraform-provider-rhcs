/*
Copyright (c) 2025 Red Hat, Inc.

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

package gcp

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// autoscalingAttrTypes defines the attribute types for the autoscaling object.
var autoscalingAttrTypes = map[string]attr.Type{
	"min_replicas": types.Int64Type,
	"max_replicas": types.Int64Type,
}

// gcpAttrTypes defines the attribute types for the gcp object.
var gcpAttrTypes = map[string]attr.Type{
	"secure_boot": types.BoolType,
}

// MachinePoolState holds the Terraform state for an OSD-GCP machine pool.
type MachinePoolState struct {
	ID                types.String `tfsdk:"id"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	Name              types.String `tfsdk:"name"`
	InstanceType      types.String `tfsdk:"instance_type"`
	Replicas          types.Int64  `tfsdk:"replicas"`
	AvailabilityZones types.List   `tfsdk:"availability_zones"`
	Labels            types.Map    `tfsdk:"labels"`
	Taints            types.List   `tfsdk:"taints"`
	RootVolumeSize    types.Int64  `tfsdk:"root_volume_size"`

	// Autoscaling and GCP use types.Object to support unknown values during plan.
	Autoscaling types.Object `tfsdk:"autoscaling"`
	GCP         types.Object `tfsdk:"gcp"`
}

// TaintState holds a single taint.
type TaintState struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Effect types.String `tfsdk:"effect"`
}
