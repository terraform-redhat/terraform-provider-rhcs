// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import "github.com/hashicorp/terraform-plugin-framework/types"

// NodePoolSeaplaneState holds the Terraform state for an rhcs_nodepool_seaplane resource.
type NodePoolSeaplaneState struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Cluster types.String `tfsdk:"cluster"`

	Replicas types.Int64 `tfsdk:"replicas"`

	Labels types.Map `tfsdk:"labels"`

	SubnetID types.String `tfsdk:"subnet_id"`

	AWSNodePool *NPAWSNodePool `tfsdk:"aws_node_pool"`
	AutoRepair  types.Bool     `tfsdk:"auto_repair"`

	Phase types.String `tfsdk:"phase"`

	IgnoreDeletionError types.Bool `tfsdk:"ignore_deletion_error"`
}

// NPAWSNodePool holds the AWS-specific node pool settings.
type NPAWSNodePool struct {
	InstanceType types.String `tfsdk:"instance_type"`
	Tags         types.Map    `tfsdk:"tags"`
	DiskSize     types.Int64  `tfsdk:"disk_size"`
}
