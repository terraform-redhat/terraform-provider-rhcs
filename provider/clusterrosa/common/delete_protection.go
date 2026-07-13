// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const deleteProtectionResourceDescription = "When true, prevents cluster deletion via OCM. " +
	"This attribute can be changed after cluster creation. " +
	"To destroy the cluster, set this to false and apply before running terraform destroy."

const deleteProtectionDatasourceDescription = "Reports whether OCM delete protection is enabled for the cluster."

// DeleteProtectionResourceSchema returns the schema definition for the delete_protection
// resource attribute.
func DeleteProtectionResourceSchema() schema.BoolAttribute {
	return schema.BoolAttribute{
		Description: deleteProtectionResourceDescription,
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	}
}

// DeleteProtectionDatasourceSchema returns the schema definition for the delete_protection
// data source attribute.
func DeleteProtectionDatasourceSchema() schema.BoolAttribute {
	return schema.BoolAttribute{
		Description: deleteProtectionDatasourceDescription,
		Computed:    true,
	}
}

func deleteProtectionFromCluster(object *cmv1.Cluster) (enabled bool, ok bool) {
	if object == nil {
		return false, false
	}
	if dp, present := object.GetDeleteProtection(); present && dp != nil {
		return dp.Enabled(), true
	}
	return false, false
}

// FetchDeleteProtection reads delete protection from the cluster sub-resource endpoint.
func FetchDeleteProtection(ctx context.Context, clusterClient *cmv1.ClusterClient) (bool, error) {
	resp, err := clusterClient.DeleteProtection().Get().SendContext(ctx)
	if err != nil {
		return false, err
	}
	if resp.Body() != nil {
		return resp.Body().Enabled(), nil
	}
	return false, fmt.Errorf("delete protection response body is empty")
}

// ResolveDeleteProtection resolves delete protection from the cluster object, falling back
// to the sub-resource GET when the inline field is absent.
func ResolveDeleteProtection(
	ctx context.Context,
	clusterClient *cmv1.ClusterClient,
	object *cmv1.Cluster,
) (types.Bool, diag.Diagnostics) {
	if enabled, ok := deleteProtectionFromCluster(object); ok {
		return types.BoolValue(enabled), nil
	}

	enabled, err := FetchDeleteProtection(ctx, clusterClient)
	if err != nil {
		return types.BoolValue(false), diag.Diagnostics{
			diag.NewWarningDiagnostic(
				"Can't read delete protection",
				fmt.Sprintf(
					"Could not read delete protection status from the API: %v. "+
						"Assuming delete protection is disabled; run terraform apply again to refresh.",
					err,
				),
			),
		}
	}
	return types.BoolValue(enabled), nil
}

// UpdateDeleteProtection patches the cluster delete protection sub-resource.
func UpdateDeleteProtection(ctx context.Context, clusterClient *cmv1.ClusterClient, enabled bool) error {
	body, err := cmv1.NewDeleteProtection().Enabled(enabled).Build()
	if err != nil {
		return fmt.Errorf("can't build delete protection object: %w", err)
	}
	_, err = clusterClient.DeleteProtection().Update().Body(body).SendContext(ctx)
	if err != nil {
		return fmt.Errorf("can't update cluster delete protection: %w", err)
	}
	return nil
}

// ValidateDeleteAllowed returns an error diagnostic when delete protection is enabled.
func ValidateDeleteAllowed(clusterID string, enabled bool) diag.Diagnostics {
	if enabled {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Can't delete cluster",
				fmt.Sprintf(
					"Delete protection is enabled for cluster '%s'. "+
						"Set delete_protection = false in configuration, run terraform apply, then retry destroy.",
					clusterID,
				),
			),
		}
	}
	return nil
}

// CheckDeleteProtectionEnabled reads the live delete protection status before destroy.
func CheckDeleteProtectionEnabled(
	ctx context.Context,
	clusterID string,
	clusterClient *cmv1.ClusterClient,
) (bool, diag.Diagnostics) {
	enabled, err := FetchDeleteProtection(ctx, clusterClient)
	if err != nil {
		getResp, getErr := clusterClient.Get().SendContext(ctx)
		if getErr == nil && getResp.Body() != nil {
			if inlineEnabled, ok := deleteProtectionFromCluster(getResp.Body()); ok {
				return inlineEnabled, nil
			}
		}
		detail := fmt.Sprintf(
			"Could not verify delete protection status before deleting cluster '%s': %v",
			clusterID, err,
		)
		if getErr != nil {
			detail += fmt.Sprintf("; cluster GET fallback failed: %v", getErr)
		} else if getResp == nil || getResp.Body() == nil {
			detail += "; cluster GET fallback returned an empty response"
		}
		return false, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Can't verify delete protection",
				detail,
			),
		}
	}
	return enabled, nil
}
