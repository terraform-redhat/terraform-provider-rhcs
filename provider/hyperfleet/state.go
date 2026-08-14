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

package hyperfleet

import "github.com/hashicorp/terraform-plugin-framework/types"

// ClusterHyperfleetState holds the Terraform state for an rhcs_cluster_hyperfleet resource.
type ClusterHyperfleetState struct {
	// id is the Platform API cluster UID assigned on creation (computed).
	ID types.String `tfsdk:"id"`

	// name is the human-readable cluster name used as the resource identifier
	// in all Platform API calls.
	Name types.String `tfsdk:"name"`

	// operator_roles_prefix is the IAM role name prefix used to compute the
	// seven operator role ARNs expected by the Platform API.
	OperatorRolesPrefix types.String `tfsdk:"operator_roles_prefix"`

	// aws_subnet_ids are the worker-node subnets. The first entry is used as
	// the actual cluster subnet.
	AWSSubnetIDs types.List `tfsdk:"aws_subnet_ids"`

	// vpc_id is the VPC that contains the worker subnet.
	VPCID types.String `tfsdk:"vpc_id"`

	// availability_zones are the AZs for worker nodes. The first entry
	// determines the cluster AZ.
	AvailabilityZones types.List `tfsdk:"availability_zones"`

	// expiration_timestamp is an optional RFC3339 timestamp after which the
	// Platform API automatically deletes the cluster.
	ExpirationTimestamp types.String `tfsdk:"expiration_timestamp"`

	// aws_partition overrides the IAM ARN partition used when computing operator
	// role ARNs. Defaults to "aws"; set to "aws-us-gov" for GovCloud.
	AWSPartition types.String `tfsdk:"aws_partition"`

	// --- computed ---

	// creator_arn is populated from the provider's aws_caller_arn on create.
	CreatorARN types.String `tfsdk:"creator_arn"`

	// cloud_region is the AWS region identifier (e.g. 'us-east-1').
	// Derived from availability_zones if not explicitly set.
	CloudRegion types.String `tfsdk:"cloud_region"`

	// phase reflects the Platform API cluster lifecycle phase
	// (WaitingForPlacement, Provisioning, Ready, Deleting).
	Phase types.String `tfsdk:"phase"`

	// api_url is the control-plane endpoint host once the cluster is Ready.
	APIURL types.String `tfsdk:"api_url"`

	// current_version is the running OpenShift version reported by the operator.
	CurrentVersion types.String `tfsdk:"current_version"`

	// management_cluster is the management cluster ID assigned by the operator.
	ManagementCluster types.String `tfsdk:"management_cluster"`

	// oidc_issuer is the OIDC service-account issuer URL for the cluster.
	// Use this to create the IAM OIDC provider and build operator role trust policies.
	OIDCIssuer types.String `tfsdk:"oidc_issuer"`
}
