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

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClusterState struct {
	APIURL             types.String `tfsdk:"api_url"`
	AWSAccessKeyID     types.String `tfsdk:"aws_access_key_id"`
	AWSAccountID       types.String `tfsdk:"aws_account_id"`
	AWSSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	CCSEnabled         types.Bool   `tfsdk:"ccs_enabled"`
	CloudProvider      types.String `tfsdk:"cloud_provider"`
	CloudRegion        types.String `tfsdk:"cloud_region"`
	ComputeMachineType types.String `tfsdk:"compute_machine_type"`
	ComputeNodes       types.Int64  `tfsdk:"compute_nodes"`
	ConsoleURL         types.String `tfsdk:"console_url"`
	HostPrefix         types.Int64  `tfsdk:"host_prefix"`
	ID                 types.String `tfsdk:"id"`
	MachineCIDR        types.String `tfsdk:"machine_cidr"`
	MultiAZ            types.Bool   `tfsdk:"multi_az"`
	Name               types.String `tfsdk:"name"`
	PodCIDR            types.String `tfsdk:"pod_cidr"`
	Properties         types.Map    `tfsdk:"properties"`
	ServiceCIDR        types.String `tfsdk:"service_cidr"`
	State              types.String `tfsdk:"state"`
	Version            types.String `tfsdk:"version"`
	Wait               types.Bool   `tfsdk:"wait"`
}
