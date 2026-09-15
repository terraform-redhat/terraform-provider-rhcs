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

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

// ClusterHandlerImpl is the concrete implementation of ClusterHandler
type ClusterHandlerImpl struct {
	accountID string
	callerARN string
}

// NewClusterHandler creates a new ClusterHandler
func NewClusterHandler(accountID, callerARN string) ClusterHandler {
	return &ClusterHandlerImpl{
		accountID: accountID,
		callerARN: callerARN,
	}
}

// PreExpand validates inputs and derives cloud_region from availability zone
func (h *ClusterHandlerImpl) PreExpand(ctx context.Context, input *ClusterState) diag.Diagnostics {
	var diags diag.Diagnostics

	// Validate required consumer-only fields
	if input.Operator_roles_prefix.IsNull() || input.Operator_roles_prefix.ValueString() == "" {
		diags.AddError("operator_roles_prefix is required", "")
		return diags
	}

	// Extract and validate availability_zones
	var azs []string
	diags.Append(input.Availability_zones.ElementsAs(ctx, &azs, false)...)
	if diags.HasError() || len(azs) == 0 {
		diags.AddError("availability_zones must not be empty", "")
		return diags
	}

	// Derive cloud_region from first AZ if not set
	if input.Cloud_region.IsNull() || input.Cloud_region.ValueString() == "" {
		input.Cloud_region = types.StringValue(regionFromAZ(azs[0]))
	}

	// Default aws_partition to "aws"
	if input.Aws_partition.IsNull() || input.Aws_partition.ValueString() == "" {
		input.Aws_partition = types.StringValue("aws")
	}

	return diags
}

// PostExpand builds the Platform spec with RolesRef and AWS configuration
func (h *ClusterHandlerImpl) PostExpand(
	ctx context.Context,
	input *ClusterState,
	obj *v1alpha1.Cluster,
) diag.Diagnostics {
	var diags diag.Diagnostics

	prefix := input.Operator_roles_prefix.ValueString()
	partition := input.Aws_partition.ValueString()

	var azs []string
	diags.Append(input.Availability_zones.ElementsAs(ctx, &azs, false)...)
	if diags.HasError() || len(azs) == 0 {
		return diags
	}

	var subnetIDs []string
	diags.Append(input.Aws_subnet_ids.ElementsAs(ctx, &subnetIDs, false)...)
	if diags.HasError() || len(subnetIDs) == 0 {
		return diags
	}

	az := azs[0]
	subnetID := subnetIDs[0]
	vpcID := input.Vpc_id.ValueString()
	region := input.Cloud_region.ValueString()

	// Compute RolesRef
	rolesRef := computeRolesRef(prefix, h.accountID, partition)

	// Build Platform spec with AWS configuration
	awsSpec := &hypershiftv1beta1.AWSPlatformSpec{
		Region:   region,
		RolesRef: rolesRef,
		CloudProviderConfig: &hypershiftv1beta1.AWSCloudProviderConfig{
			VPC:  vpcID,
			Zone: az,
			Subnet: &hypershiftv1beta1.AWSResourceReference{
				ID: &subnetID,
			},
		},
	}

	// Set the complete Platform spec
	obj.Spec.HostedCluster.Platform = v1alpha1.PlatformSpec{
		Type: hypershiftv1beta1.AWSPlatform,
		AWS:  awsSpec,
	}

	return diags
}

// PostResponse is called after API operations
func (h *ClusterHandlerImpl) PostResponse(ctx context.Context, resp *v1alpha1.Cluster) diag.Diagnostics {
	return diag.Diagnostics{}
}

// computeRolesRef builds the RolesRef from operator prefix and account ID
func computeRolesRef(prefix, accountID, partition string) hypershiftv1beta1.AWSRolesRef {
	arn := func(suffix string) string {
		return fmt.Sprintf("arn:%s:iam::%s:role/%s%s", partition, accountID, prefix, suffix)
	}
	return hypershiftv1beta1.AWSRolesRef{
		IngressARN:              arn("-ingress"),
		KubeCloudControllerARN:  arn("-cloud-controller-manager"),
		StorageARN:              arn("-ebs-csi"),
		ImageRegistryARN:        arn("-image-registry"),
		NetworkARN:              arn("-network-config"),
		ControlPlaneOperatorARN: arn("-control-plane-operator"),
		NodePoolManagementARN:   arn("-node-pool-management"),
	}
}

// NewClusterHandlerImpl creates a new ClusterHandlerImpl instance.
func NewClusterHandlerImpl(client hyperfleet.Interface, accountID string, callerARN string) ClusterHandler {
	return &ClusterHandlerImpl{
		accountID: accountID,
		callerARN: callerARN,
	}
}
