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
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
)

// maxClusterTags is the number of customer AWS tags the Platform API accepts on
// spec.tags. HyperShift caps platform.aws.resourceTags at 25 entries and the
// operator injects 2 system tags (red-hat-managed and
// kubernetes.io/cluster/<id>), leaving 23 for the customer.
const maxClusterTags = 23

// awsReservedTagPrefix marks tag keys AWS reserves for its own use.
const awsReservedTagPrefix = "aws:"

// Character classes the ROSA CLI enforces on --tags, so a tag accepted by one
// path is accepted by the other.
var (
	clusterTagKeyRE   = regexp.MustCompile(`^[\pL\pZ\pN_.:/=+\-@]{1,128}$`)
	clusterTagValueRE = regexp.MustCompile(`^[\pL\pZ\pN_.:/=+\-@]{0,256}$`)
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
	var azDiags diag.Diagnostics
	azs, azDiags = awsBundleStringList(ctx, input.Aws, "availability_zones")
	diags.Append(azDiags...)
	if diags.HasError() || len(azs) == 0 {
		diags.AddError("availability_zones must not be empty", "")
		return diags
	}
	if awsBundleStringValue(input.Aws, "aws_partition") == "" {
		var bundleDiags diag.Diagnostics
		input.Aws, bundleDiags = awsBundleWithString(input.Aws, "aws_partition", "aws")
		diags.Append(bundleDiags...)
		if diags.HasError() {
			return diags
		}
	}

	// Derive cloud_region from first AZ if not set
	if input.Cloud_region.IsNull() || input.Cloud_region.ValueString() == "" {
		input.Cloud_region = types.StringValue(regionFromAZ(azs[0]))
	}

	diags.Append(validateClusterTags(ctx, input.Tags)...)

	return diags
}

// validateClusterTags checks the customer AWS tags against the limits the
// Platform API enforces on spec.tags so a bad value fails during apply instead
// of on the API round trip. Wording follows the ROSA CLI --tags validation.
func validateClusterTags(ctx context.Context, tags types.Map) diag.Diagnostics {
	var diags diag.Diagnostics

	if tags.IsNull() || tags.IsUnknown() {
		return diags
	}

	var values map[string]string
	diags.Append(tags.ElementsAs(ctx, &values, false)...)
	if diags.HasError() {
		return diags
	}

	addError := func(format string, args ...any) {
		diags.AddAttributeError(path.Root("tags"), "Invalid cluster AWS tags", fmt.Sprintf(format, args...))
	}

	if len(values) > maxClusterTags {
		addError("a maximum of %d tags is supported, got %d", maxClusterTags, len(values))
		return diags
	}

	// Sorted so repeated plans report the same tag first.
	for _, key := range slices.Sorted(maps.Keys(values)) {
		value := values[key]
		switch {
		case key == "" || value == "":
			addError("invalid tag format, tag key or tag value can not be empty")
		case strings.HasPrefix(strings.ToLower(key), awsReservedTagPrefix):
			addError("invalid tag key '%s': keys starting with '%s' are reserved for AWS use",
				key, awsReservedTagPrefix)
		case !clusterTagKeyRE.MatchString(key):
			addError("expected a valid user tag key '%s' matching %s", key, clusterTagKeyRE.String())
		case !clusterTagValueRE.MatchString(value):
			addError("expected a valid user tag value for key '%s' matching %s", key, clusterTagValueRE.String())
		}
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
	partition := awsBundleStringValue(input.Aws, "aws_partition")
	if partition == "" {
		partition = "aws"
	}

	var azs []string
	var azDiags diag.Diagnostics
	azs, azDiags = awsBundleStringList(ctx, input.Aws, "availability_zones")
	diags.Append(azDiags...)
	if diags.HasError() || len(azs) == 0 {
		return diags
	}

	var subnetIDs []string
	var subnetDiags diag.Diagnostics
	subnetIDs, subnetDiags = awsBundleStringList(ctx, input.Aws, "aws_subnet_ids")
	diags.Append(subnetDiags...)
	if diags.HasError() || len(subnetIDs) == 0 {
		return diags
	}

	az := azs[0]
	subnetID := subnetIDs[0]
	vpcID := awsBundleStringValue(input.Aws, "vpc_id")
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

func awsBundleStringValue(bundle types.Object, name string) string {
	if bundle.IsNull() || bundle.IsUnknown() {
		return ""
	}
	value, ok := bundle.Attributes()[name].(types.String)
	if !ok || value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
}

func awsBundleStringList(ctx context.Context, bundle types.Object, name string) ([]string, diag.Diagnostics) {
	if bundle.IsNull() || bundle.IsUnknown() {
		return nil, nil
	}
	value, ok := bundle.Attributes()[name].(types.List)
	if !ok || value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	var result []string
	return result, value.ElementsAs(ctx, &result, false)
}

func awsBundleWithString(bundle types.Object, name, value string) (types.Object, diag.Diagnostics) {
	if bundle.IsNull() || bundle.IsUnknown() {
		return bundle, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid AWS bundle", "AWS bundle is null or unknown")}
	}
	attributes := bundle.Attributes()
	attributes[name] = types.StringValue(value)
	typesByName := make(map[string]attr.Type, len(attributes))
	for attributeName, attributeValue := range attributes {
		typesByName[attributeName] = attributeValue.Type(context.Background())
	}
	return types.ObjectValue(typesByName, attributes)
}

// PostResponse is called after API operations
func (h *ClusterHandlerImpl) PostResponse(ctx context.Context, resp *v1alpha1.Cluster) diag.Diagnostics {
	return diag.Diagnostics{}
}

// PostFlatten populates computed fields in the state from the API response
func (h *ClusterHandlerImpl) PostFlatten(ctx context.Context, state *ClusterState, resp *v1alpha1.Cluster) {
	if resp == nil {
		return
	}
	// Populate Phase from status (consumer-only field, not mapped via pathbind)
	// Default to "WaitingForPlacement" if not yet set by the controller
	if resp.Status.Phase != "" {
		state.Phase = types.StringValue(string(resp.Status.Phase))
	} else {
		state.Phase = types.StringValue(string(v1alpha1.ClusterPhaseWaitingForPlacement))
	}
	// Populate API URL from control plane endpoint host
	// Default to empty string if not yet computed by the controller
	if resp.Status.ControlPlaneEndpoint.Host != "" {
		state.Api_url = types.StringValue(resp.Status.ControlPlaneEndpoint.Host)
	} else {
		state.Api_url = types.StringValue("")
	}
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
