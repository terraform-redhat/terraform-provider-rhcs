/*
Copyright (c) 2026 Red Hat, Inc.

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

package ocmrole

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

const ocmRoleLabelKey = "sts_ocm_role"

var (
	roleARNRegex = regexp.MustCompile(
		`^arn:(?:aws|aws-us-gov|aws-cn):iam::\d{12}:role/.+$`,
	)
	labelMutex = common.NewMutexKV()
)

type RosaOCMRoleLinkResource struct {
	currentAccountClient *amsv1.CurrentAccountClient
	organizationsClient  *amsv1.OrganizationsClient
}

var _ resource.ResourceWithConfigure = &RosaOCMRoleLinkResource{}
var _ resource.ResourceWithImportState = &RosaOCMRoleLinkResource{}

func New() resource.Resource {
	return &RosaOCMRoleLinkResource{}
}

func (r *RosaOCMRoleLinkResource) Metadata(
	ctx context.Context, req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_rosa_ocm_role_link"
}

func (r *RosaOCMRoleLinkResource) Schema(
	ctx context.Context, req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Links an externally created AWS IAM OCM role " +
			"to an OCM organization. The AWS IAM role, its " +
			"permission policy, and trust relationship should be " +
			"created using the AWS provider. This resource manages " +
			"only the OCM-side link.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier (the role ARN).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_arn": schema.StringAttribute{
				Description: "The ARN of the AWS IAM OCM role to link.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						roleARNRegex,
						"must be a valid AWS IAM role ARN "+
							"(e.g. arn:aws:iam::123456789012:role/name)",
					),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[^,]+$`),
						"must not contain commas",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RosaOCMRoleLinkResource) Configure(
	ctx context.Context, req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *sdk.Connection, got: %T.",
				req.ProviderData,
			),
		)
		return
	}

	r.currentAccountClient = connection.AccountsMgmt().V1().CurrentAccount()
	r.organizationsClient = connection.AccountsMgmt().V1().Organizations()
}

func (r *RosaOCMRoleLinkResource) Create(
	ctx context.Context, req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	plan := &RosaOCMRoleLinkState{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := r.resolveOrgID(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to resolve OCM organization",
			fmt.Sprintf(
				"Could not determine the current organization: %v",
				err,
			),
		)
		return
	}

	roleARN := plan.RoleARN.ValueString()

	parsedARN, err := arn.Parse(roleARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid role ARN",
			fmt.Sprintf("Could not parse role ARN: %v", err),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf(
		"Attempting to acquire lock for organization %s", orgID,
	))
	labelMutex.Lock(orgID)
	defer labelMutex.Unlock(orgID)
	tflog.Debug(ctx, fmt.Sprintf(
		"Acquired lock for organization %s", orgID,
	))

	existing, hasLabel, _, err := r.findOCMRoleLabelWithStatus(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to check existing OCM role link",
			fmt.Sprintf("Could not read organization labels: %v", err),
		)
		return
	}

	if hasLabel {
		if containsARN(existing, roleARN) {
			tflog.Info(ctx, fmt.Sprintf(
				"Role %s is already linked to org %s",
				roleARN, orgID,
			))
			resp.Diagnostics.AddError(
				"OCM role already linked",
				fmt.Sprintf(
					"The role ARN %q is already linked to "+
						"organization %q. Use terraform import "+
						"instead.",
					roleARN, orgID,
				),
			)
			return
		}

		if conflict := findSameAccountARN(
			existing, parsedARN.AccountID, roleARN,
		); conflict != "" {
			tflog.Info(ctx, fmt.Sprintf(
				"AWS account %s conflict: existing %s vs new %s",
				parsedARN.AccountID, conflict, roleARN,
			))
			resp.Diagnostics.AddError(
				"AWS account already has a linked role",
				fmt.Sprintf(
					"Organization %q already has role %q from "+
						"AWS account %s. Only one role per AWS "+
						"account is allowed per organization.",
					orgID, conflict, parsedARN.AccountID,
				),
			)
			return
		}

		newValue := appendARN(existing, roleARN)
		label, buildErr := amsv1.NewLabel().
			Key(ocmRoleLabelKey).Value(newValue).Build()
		if buildErr != nil {
			resp.Diagnostics.AddError(
				"Failed to build label",
				fmt.Sprintf("%v", buildErr),
			)
			return
		}
		_, err = r.organizationsClient.Organization(orgID).
			Labels().Label(ocmRoleLabelKey).
			Update().Body(label).SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OCM role link",
				fmt.Sprintf(
					"Could not add role ARN to existing label: %v",
					err,
				),
			)
			return
		}
	} else {
		label, buildErr := amsv1.NewLabel().
			Key(ocmRoleLabelKey).Value(roleARN).Build()
		if buildErr != nil {
			resp.Diagnostics.AddError(
				"Failed to build label",
				fmt.Sprintf("%v", buildErr),
			)
			return
		}
		_, err = r.organizationsClient.Organization(orgID).
			Labels().Add().Body(label).SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to link OCM role",
				fmt.Sprintf(
					"Could not create organization label: %v", err,
				),
			)
			return
		}
	}

	tflog.Info(ctx, fmt.Sprintf(
		"Linked role %s to organization %s", roleARN, orgID,
	))

	plan.ID = types.StringValue(roleARN)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RosaOCMRoleLinkResource) Read(
	ctx context.Context, req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	state := &RosaOCMRoleLinkState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleARN := state.RoleARN.ValueString()

	orgID, err := r.resolveOrgID(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to resolve OCM organization",
			fmt.Sprintf(
				"Could not determine organization: %v", err,
			),
		)
		return
	}

	labelValue, hasLabel, listStatus, err := r.findOCMRoleLabelWithStatus(ctx, orgID)
	if err != nil {
		if listStatus == http.StatusNotFound {
			tflog.Warn(ctx, fmt.Sprintf(
				"Organization %s not found (HTTP 404), removing from state",
				orgID,
			))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OCM role link",
			fmt.Sprintf("Could not read organization labels: %v", err),
		)
		return
	}

	if !hasLabel || !containsARN(labelValue, roleARN) {
		tflog.Warn(ctx, fmt.Sprintf(
			"OCM role link for %s not found in org %s, removing from state",
			roleARN, orgID,
		))
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(roleARN)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *RosaOCMRoleLinkResource) Update(
	ctx context.Context, req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"The OCM role link resource does not support in-place "+
			"updates. Delete and recreate the resource instead.",
	)
}

func (r *RosaOCMRoleLinkResource) Delete(
	ctx context.Context, req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	state := &RosaOCMRoleLinkState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleARN := state.RoleARN.ValueString()

	orgID, err := r.resolveOrgID(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to resolve OCM organization",
			fmt.Sprintf(
				"Could not determine organization: %v", err,
			),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf(
		"Attempting to acquire lock for organization %s", orgID,
	))
	labelMutex.Lock(orgID)
	defer labelMutex.Unlock(orgID)
	tflog.Debug(ctx, fmt.Sprintf(
		"Acquired lock for organization %s", orgID,
	))

	labelValue, hasLabel, listStatus, err := r.findOCMRoleLabelWithStatus(ctx, orgID)
	if err != nil {
		if listStatus == http.StatusNotFound {
			tflog.Info(ctx, fmt.Sprintf(
				"Organization %s not found during delete, treating as already unlinked",
				orgID,
			))
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OCM role link for deletion",
			fmt.Sprintf("Could not read organization labels: %v", err),
		)
		return
	}

	if !hasLabel {
		tflog.Info(ctx, fmt.Sprintf(
			"Label %s not found in org %s, treating as already unlinked",
			ocmRoleLabelKey, orgID,
		))
		return
	}

	if !containsARN(labelValue, roleARN) {
		tflog.Info(ctx, fmt.Sprintf(
			"Role %s not found in org %s, treating as already unlinked",
			roleARN, orgID,
		))
		return
	}

	newValue := removeARN(labelValue, roleARN)
	if newValue == "" {
		deleteResp, deleteErr := r.organizationsClient.Organization(orgID).
			Labels().Label(ocmRoleLabelKey).
			Delete().SendContext(ctx)
		if deleteErr != nil {
			if deleteResp != nil &&
				deleteResp.Status() == http.StatusNotFound {
				tflog.Info(ctx, fmt.Sprintf(
					"Label already deleted for org %s", orgID,
				))
				return
			}
			resp.Diagnostics.AddError(
				"Failed to unlink OCM role",
				fmt.Sprintf(
					"Could not delete organization label: %v",
					deleteErr,
				),
			)
			return
		}
	} else {
		label, buildErr := amsv1.NewLabel().
			Key(ocmRoleLabelKey).Value(newValue).Build()
		if buildErr != nil {
			resp.Diagnostics.AddError(
				"Failed to build label",
				fmt.Sprintf("%v", buildErr),
			)
			return
		}
		_, err = r.organizationsClient.Organization(orgID).
			Labels().Label(ocmRoleLabelKey).
			Update().Body(label).SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to unlink OCM role",
				fmt.Sprintf(
					"Could not update organization label: %v", err,
				),
			)
			return
		}
	}

	tflog.Info(ctx, fmt.Sprintf(
		"Unlinked role %s from organization %s", roleARN, orgID,
	))
}

func (r *RosaOCMRoleLinkResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	roleARN := req.ID

	if !roleARNRegex.MatchString(roleARN) {
		resp.Diagnostics.AddError(
			"Invalid role ARN format",
			fmt.Sprintf(
				"Import ID must be a valid AWS IAM role ARN "+
					"(e.g. arn:aws:iam::123456789012:role/name). "+
					"Got: %s", roleARN,
			),
		)
		return
	}
	if strings.Contains(roleARN, ",") {
		resp.Diagnostics.AddError(
			"Invalid role ARN format",
			"Import ID must not contain commas.",
		)
		return
	}

	orgID, err := r.resolveOrgID(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to resolve organization during import",
			fmt.Sprintf(
				"Could not determine the current organization: %v",
				err,
			),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf(
		"Importing OCM role link: org=%s arn=%s", orgID, roleARN,
	))

	labelValue, hasLabel, _, err := r.findOCMRoleLabelWithStatus(
		ctx, orgID,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to verify role link during import",
			fmt.Sprintf(
				"Could not read organization labels for %s: %v",
				orgID, err,
			),
		)
		return
	}

	if !hasLabel || !containsARN(labelValue, roleARN) {
		resp.Diagnostics.AddError(
			"Role not linked",
			fmt.Sprintf(
				"Role %s is not linked to organization %s",
				roleARN, orgID,
			),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("role_arn"), roleARN)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("id"), roleARN)...)
}

func (r *RosaOCMRoleLinkResource) resolveOrgID(
	ctx context.Context,
) (string, error) {
	accountResp, err := r.currentAccountClient.Get().SendContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current account: %w", err)
	}
	body, ok := accountResp.GetBody()
	if !ok {
		return "", fmt.Errorf("empty response from current account API")
	}
	orgID := body.Organization().ID()
	if orgID == "" {
		return "", fmt.Errorf("current account has no organization")
	}
	return orgID, nil
}

// findOCMRoleLabelWithStatus returns the label value, whether it exists,
// the HTTP status code from the list call (0 if no error), and any error.
func (r *RosaOCMRoleLinkResource) findOCMRoleLabelWithStatus(
	ctx context.Context, orgID string,
) (string, bool, int, error) {
	listResp, err := r.organizationsClient.Organization(orgID).
		Labels().List().SendContext(ctx)
	if err != nil {
		status := 0
		if listResp != nil {
			status = listResp.Status()
		}
		return "", false, status, err
	}

	var labelValue string
	hasLabel := false
	listResp.Items().Each(func(label *amsv1.Label) bool {
		if label.Key() == ocmRoleLabelKey {
			labelValue = label.Value()
			hasLabel = true
			return false
		}
		return true
	})
	return labelValue, hasLabel, listResp.Status(), nil
}

// findSameAccountARN checks if any ARN in the comma-separated list belongs
// to the same AWS account as the given accountID but is a different ARN.
func findSameAccountARN(
	commaSeparated, accountID, excludeARN string,
) string {
	for a := range strings.SplitSeq(commaSeparated, ",") {
		trimmed := strings.TrimSpace(a)
		if trimmed == "" || trimmed == excludeARN {
			continue
		}
		parsed, parseErr := arn.Parse(trimmed)
		if parseErr == nil && parsed.AccountID == accountID {
			return trimmed
		}
	}
	return ""
}

func containsARN(commaSeparated, roleARN string) bool {
	for a := range strings.SplitSeq(commaSeparated, ",") {
		if strings.TrimSpace(a) == roleARN {
			return true
		}
	}
	return false
}

func appendARN(commaSeparated, roleARN string) string {
	if commaSeparated == "" {
		return roleARN
	}
	return commaSeparated + "," + roleARN
}

func removeARN(commaSeparated, roleARN string) string {
	var remaining []string
	for a := range strings.SplitSeq(commaSeparated, ",") {
		trimmed := strings.TrimSpace(a)
		if trimmed != "" && trimmed != roleARN {
			remaining = append(remaining, trimmed)
		}
	}
	return strings.Join(remaining, ",")
}
