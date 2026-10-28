// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

// Package awstags prepares day-2 ROSA HCP AWS tag edits for the OCM API.
package awstags

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	commontags "github.com/openshift-online/ocm-common/pkg/aws/tags"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocmerrors "github.com/openshift-online/ocm-sdk-go/errors"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

const (
	// MinimalVersion is the first control-plane version that reconciles AWS tag edits in place.
	MinimalVersion = "4.20.0-0.a"

	// Day2EditDescription documents tag updates in resource schema descriptions.
	Day2EditDescription = "Tags can be updated after creation when the cluster control plane is on " +
		"OpenShift 4.20 or later. When this attribute is set, including to an empty map, Terraform manages " +
		"the complete set of customer tags: tags removed from the configuration or added outside Terraform " +
		"are removed from the AWS resources. When it is not set, existing tags are left unmanaged; removing " +
		"it from a configuration that set it removes the managed tags once. " +
		"Tag values cannot be empty when updating tags."

	reservedKeyPrefix = "api.openshift.com/"
	clustersPath      = "/api/clusters_mgmt/v1/clusters/"
	clusterAWSField   = "aws"
	nodePoolAWSField  = "aws_node_pool"
	userTagsParameter = "fetchUserTagsOnly"
)

// ValidateVersion rejects tag edits on clusters whose control plane is older than 4.20.
func ValidateVersion(rawVersion string) error {
	if rawVersion == "" {
		return fmt.Errorf("cannot edit AWS tags without a control-plane version")
	}
	supported, err := common.IsGreaterThanOrEqual(rawVersion, MinimalVersion)
	if err != nil {
		return fmt.Errorf("cannot check control-plane version '%s' for AWS tag editing: %v", rawVersion, err)
	}
	if !supported {
		return fmt.Errorf("editing AWS tags requires OpenShift version 4.20.0 or greater, "+
			"the cluster control plane is on version '%s'", rawVersion)
	}
	return nil
}

// Validate checks the desired day-2 tag set. Unlike create, HCP day-2 edits reject empty values.
func Validate(tags map[string]string) error {
	if err := commontags.Validate(tags); err != nil {
		return err
	}
	for key, value := range tags {
		if strings.HasPrefix(key, reservedKeyPrefix) {
			return fmt.Errorf("AWS tag key '%s' is reserved", key)
		}
		if value == "" {
			return fmt.Errorf("AWS tag value for key '%s' cannot be empty", key)
		}
	}
	return nil
}

// Diff returns the PATCH tag changes needed to replace current with desired.
// A nil result means that the sets are equal; nil values in the result delete keys.
func Diff(current, desired map[string]string) map[string]*string {
	if commontags.Equal(current, desired) {
		return nil
	}
	diff := commontags.ComputeDiff(current, desired)
	changes := make(map[string]*string, len(diff.Add)+len(diff.Update)+len(diff.Remove))
	for key, value := range diff.Add {
		changes[key] = &value
	}
	for key, value := range diff.Update {
		changes[key] = &value
	}
	for _, key := range diff.Remove {
		changes[key] = nil
	}
	return changes
}

// PoolOverrides removes tags matching inherited cluster values from a node pool tag set.
func PoolOverrides(clusterTags, poolTags map[string]string) map[string]string {
	overrides := make(map[string]string, len(poolTags))
	for key, value := range poolTags {
		if inherited, ok := clusterTags[key]; !ok || inherited != value {
			overrides[key] = value
		}
	}
	return overrides
}

// InheritedMatches returns the sorted desired pool tag keys whose values match cluster tags.
func InheritedMatches(clusterTags, desired map[string]string) []string {
	var matches []string
	for key, value := range desired {
		if inherited, ok := clusterTags[key]; ok && inherited == value {
			matches = append(matches, key)
		}
	}
	sort.Strings(matches)
	return matches
}

// PoolChanges computes node pool tag changes from the effective (merged) pool view.
// Desired keys whose values match the cluster are inherited instead of stored as overrides.
// The effective view cannot distinguish inheritance from a stored override with the same
// value, so such keys are removed with null when no longer desired; OCM ignores null for
// an inherited-only key.
func PoolChanges(clusterTags, effective, desired map[string]string) map[string]*string {
	desiredOverrides := PoolOverrides(clusterTags, desired)
	changes := Diff(PoolOverrides(clusterTags, effective), desiredOverrides)
	for key, inherited := range clusterTags {
		if value, ok := effective[key]; !ok || value != inherited {
			continue
		}
		if _, keep := desiredOverrides[key]; keep {
			continue
		}
		if changes == nil {
			changes = make(map[string]*string)
		}
		changes[key] = nil
	}
	return changes
}

// GetClusterUserTags returns customer-managed cluster AWS tags, omitting protected tags.
func GetClusterUserTags(ctx context.Context, connection *sdk.Connection, clusterID string) (map[string]string, error) {
	return getUserTags(ctx, connection, clusterPath(clusterID), clusterAWSField)
}

// GetNodePoolUserTags returns customer-managed node pool tags, including inherited cluster tags.
func GetNodePoolUserTags(ctx context.Context, connection *sdk.Connection,
	clusterID, nodePoolID string) (map[string]string, error) {
	return getUserTags(ctx, connection, nodePoolPath(clusterID, nodePoolID), nodePoolAWSField)
}

// PatchCluster sends the typed cluster update and nullable tag changes in one PATCH.
func PatchCluster(ctx context.Context, connection *sdk.Connection, clusterID string,
	cluster *cmv1.Cluster, changes map[string]*string) error {
	var typed bytes.Buffer
	if err := cmv1.MarshalCluster(cluster, &typed); err != nil {
		return fmt.Errorf("failed to encode cluster update: %v", err)
	}
	return patchWithTags(ctx, connection, clusterPath(clusterID), typed.Bytes(), clusterAWSField, changes)
}

// PatchNodePool sends the typed node pool update and nullable tag changes in one PATCH.
func PatchNodePool(ctx context.Context, connection *sdk.Connection, clusterID string,
	nodePool *cmv1.NodePool, changes map[string]*string) error {
	var typed bytes.Buffer
	if err := cmv1.MarshalNodePool(nodePool, &typed); err != nil {
		return fmt.Errorf("failed to encode node pool update: %v", err)
	}
	return patchWithTags(ctx, connection, nodePoolPath(clusterID, nodePool.ID()), typed.Bytes(),
		nodePoolAWSField, changes)
}

func clusterPath(clusterID string) string {
	return clustersPath + url.PathEscape(clusterID)
}

func nodePoolPath(clusterID, nodePoolID string) string {
	return clusterPath(clusterID) + "/node_pools/" + url.PathEscape(nodePoolID)
}

// getUserTags requests customer-only AWS tags and omits null deletion markers from the response.
func getUserTags(ctx context.Context, connection *sdk.Connection, path, field string) (map[string]string, error) {
	response, err := connection.Get().Path(path).Parameter(userTagsParameter, true).SendContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read AWS tags: %v", err)
	}
	if err := responseError(response); err != nil {
		return nil, fmt.Errorf("failed to read AWS tags: %v", err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(response.Bytes(), &object); err != nil {
		return nil, fmt.Errorf("failed to decode AWS tags response: %v", err)
	}
	var awsFields struct {
		Tags map[string]*string `json:"tags"`
	}
	if len(object[field]) > 0 {
		if err := json.Unmarshal(object[field], &awsFields); err != nil {
			return nil, fmt.Errorf("failed to decode AWS tags: %v", err)
		}
	}
	tags := make(map[string]string, len(awsFields.Tags))
	for key, value := range awsFields.Tags {
		if value != nil {
			tags[key] = *value
		}
	}
	return tags, nil
}

// patchWithTags inserts nullable tag changes into a typed JSON update and sends one PATCH.
// The typed SDK builders cannot encode null map values, which OCM uses to delete tags.
func patchWithTags(ctx context.Context, connection *sdk.Connection, path string, typed []byte,
	field string, changes map[string]*string) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(typed, &object); err != nil {
		return fmt.Errorf("failed to decode typed update: %v", err)
	}
	var awsFields map[string]json.RawMessage
	if len(object[field]) > 0 {
		if err := json.Unmarshal(object[field], &awsFields); err != nil {
			return fmt.Errorf("failed to decode typed AWS update: %v", err)
		}
	}
	if awsFields == nil {
		awsFields = make(map[string]json.RawMessage)
	}
	tagsJSON, err := json.Marshal(changes)
	if err != nil {
		return fmt.Errorf("failed to encode AWS tag changes: %v", err)
	}
	awsFields["tags"] = tagsJSON
	object[field], err = json.Marshal(awsFields)
	if err != nil {
		return fmt.Errorf("failed to encode AWS update: %v", err)
	}
	body, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed to encode update: %v", err)
	}
	response, err := connection.Patch().Path(path).
		Parameter(userTagsParameter, true).
		Header("Content-Type", "application/json").
		Bytes(body).
		SendContext(ctx)
	if err != nil {
		return err
	}
	return responseError(response)
}

// responseError preserves the API validation reason when a raw request fails.
func responseError(response *sdk.Response) error {
	status := response.Status()
	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		return nil
	}
	serviceErr, err := ocmerrors.UnmarshalErrorStatus(response.Bytes(), status)
	if err == nil && serviceErr != nil && serviceErr.Reason() != "" {
		return fmt.Errorf("status is %d, %s", status, serviceErr.Reason())
	}
	return fmt.Errorf("status is %d", status)
}
