package ocmrolelink

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

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
)

const (
	OCMRoleLabel = "sts_ocm_role"
)

var (
	RoleArnRE = regexp.MustCompile(`^arn:(?:aws|aws-us-gov|aws-cn):iam::\d{12}:role\/.+$`)
	// labelMutex protects label read-modify-write operations to prevent
	// concurrent updates from causing lost writes when multiple resources
	// modify the same sts_ocm_role label simultaneously
	labelMutex sync.Mutex
)

type OCMRoleLinkResource struct {
	connection *sdk.Connection
}

var _ resource.ResourceWithConfigure = &OCMRoleLinkResource{}
var _ resource.ResourceWithImportState = &OCMRoleLinkResource{}

func New() resource.Resource {
	return &OCMRoleLinkResource{}
}

func (r *OCMRoleLinkResource) Metadata(ctx context.Context,
	req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ocm_role_link"
}

func (r *OCMRoleLinkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Link an OCM role to an OCM organization",
		Attributes: map[string]schema.Attribute{
			"role_arn": schema.StringAttribute{
				//nolint:lll
				Description: "ARN of the AWS IAM role to link to the organization. The role must exist prior to linking and should follow the OCM role pattern. For more information, follow the documentation for the offering.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						RoleArnRE,
						"role_arn must be a valid AWS IAM role ARN (e.g., arn:aws:iam::123456789012:role/role-name)",
					),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[^,]+$`),
						"role_arn must not contain commas as they are used as delimiters in the OCM role label",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "OCM organization ID (automatically determined from the current user)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *OCMRoleLinkResource) Configure(ctx context.Context,
	req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)
		return
	}

	r.connection = connection
}

func (r *OCMRoleLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OCMRoleLinkState
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleArn := plan.RoleArn.ValueString()

	tflog.Debug(ctx, "Fetching current account organization")
	acctResp, err := r.connection.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get current account",
			fmt.Sprintf("Unable to retrieve current user's organization: %s", err.Error()),
		)
		return
	}
	orgID := acctResp.Body().Organization().ID()
	plan.OrganizationID = types.StringValue(orgID)
	tflog.Debug(ctx, fmt.Sprintf("Using current user's organization: %s", orgID))
	parsedARN, err := arn.Parse(roleArn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid role ARN",
			fmt.Sprintf("The provided role ARN is not valid: %s", err.Error()),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Linking OCM role %s to organization %s", roleArn, orgID))

	err = r.linkRole(ctx, orgID, roleArn, parsedARN.AccountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to link OCM role",
			fmt.Sprintf("Unable to link role %s to organization %s: %s", roleArn, orgID, err.Error()),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully linked role %s to organization %s", roleArn, orgID))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OCMRoleLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OCMRoleLinkState
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := state.OrganizationID.ValueString()
	roleArn := state.RoleArn.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Reading OCM role link: %s for organization %s", roleArn, orgID))

	// Check if current user's organization matches the organization in state
	acctResp, err := r.connection.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Unable to verify organization",
			fmt.Sprintf("Could not retrieve current user's organization to verify it matches state: %s", err.Error()),
		)
	} else {
		currentOrgID := acctResp.Body().Organization().ID()
		if currentOrgID != orgID {
			resp.Diagnostics.AddWarning(
				"Organization mismatch",
				fmt.Sprintf(
					//nolint:lll
					"The OCM role link was created in organization '%s' but the current user belongs to organization '%s'. This resource may not be readable or modifiable with the current credentials.",
					orgID,
					currentOrgID,
				),
			)
		}
	}

	labelsResp, err := r.connection.AccountsMgmt().V1().Organizations().
		Organization(orgID).
		Labels().
		Labels(OCMRoleLabel).
		Get().
		Send()

	if err != nil {
		if labelsResp != nil && labelsResp.Status() == http.StatusNotFound {
			tflog.Warn(ctx, fmt.Sprintf("OCM role label not found for organization %s, removing from state", orgID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OCM role link",
			fmt.Sprintf("Unable to retrieve OCM role links for organization %s: %s", orgID, err.Error()),
		)
		return
	}

	linkedArns := strings.Split(labelsResp.Body().Value(), ",")
	found := false
	for _, linkedArn := range linkedArns {
		if strings.TrimSpace(linkedArn) == roleArn {
			found = true
			break
		}
	}

	if !found {
		tflog.Warn(ctx, fmt.Sprintf(
			"Role %s not found in linked roles for organization %s, removing from state",
			roleArn, orgID))
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Role %s is linked to organization %s", roleArn, orgID))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *OCMRoleLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Updates to link_ocm_role require replacement. This should have been handled by RequiresReplace plan modifiers.",
	)
}

func (r *OCMRoleLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OCMRoleLinkState
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := state.OrganizationID.ValueString()
	roleArn := state.RoleArn.ValueString()

	tflog.Info(ctx, fmt.Sprintf("Unlinking OCM role %s from organization %s", roleArn, orgID))

	// Lock to prevent concurrent label modifications from causing lost updates
	labelMutex.Lock()
	defer labelMutex.Unlock()

	labelsResp, err := r.connection.AccountsMgmt().V1().Organizations().
		Organization(orgID).
		Labels().
		Labels(OCMRoleLabel).
		Get().
		SendContext(ctx)

	if err != nil {
		if labelsResp != nil && labelsResp.Status() == http.StatusNotFound {
			tflog.Info(
				ctx,
				fmt.Sprintf("OCM role label not found for organization %s, considering already unlinked", orgID),
			)
			return
		}
		if labelsResp != nil && labelsResp.Status() == http.StatusForbidden {
			resp.Diagnostics.AddError(
				"Insufficient permissions to unlink OCM role",
				//nolint:lll
				"Only organization admin can unlink OCM roles. Please ask someone with the organization admin role to perform this operation.",
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to get OCM role links",
			fmt.Sprintf("Unable to retrieve OCM role links for organization %s: %s", orgID, err.Error()),
		)
		return
	}

	linkedArns := strings.Split(labelsResp.Body().Value(), ",")
	var newArns []string
	found := false
	for _, linkedArn := range linkedArns {
		trimmedArn := strings.TrimSpace(linkedArn)
		if trimmedArn == roleArn {
			found = true
			continue
		}
		newArns = append(newArns, trimmedArn)
	}

	if !found {
		tflog.Info(ctx, fmt.Sprintf("Role %s not found in linked roles, considering already unlinked", roleArn))
		return
	}

	if len(newArns) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("Deleting OCM role label for organization %s (no remaining roles)", orgID))
		_, err = r.connection.AccountsMgmt().V1().Organizations().
			Organization(orgID).
			Labels().
			Labels(OCMRoleLabel).
			Delete().
			SendContext(ctx)
	} else {
		newValue := strings.Join(newArns, ",")
		tflog.Debug(
			ctx,
			fmt.Sprintf("Updating OCM role label for organization %s with remaining roles: %s", orgID, newValue),
		)
		labelBuilder, buildErr := amsv1.NewLabel().Key(OCMRoleLabel).Value(newValue).Build()
		if buildErr != nil {
			resp.Diagnostics.AddError(
				"Failed to build label",
				fmt.Sprintf("Unable to build OCM role label: %s", buildErr.Error()),
			)
			return
		}

		_, err = r.connection.AccountsMgmt().V1().Organizations().
			Organization(orgID).
			Labels().
			Labels(OCMRoleLabel).
			Update().
			Body(labelBuilder).
			SendContext(ctx)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unlink OCM role",
			fmt.Sprintf("Unable to unlink role %s from organization %s: %s", roleArn, orgID, err.Error()),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully unlinked role %s from organization %s", roleArn, orgID))
}

func (r *OCMRoleLinkResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	roleArn := req.ID

	// Validate ARN format matches schema requirements (partition, service, resource type)
	if !RoleArnRE.MatchString(roleArn) {
		resp.Diagnostics.AddError(
			"Invalid role ARN",
			fmt.Sprintf(
				"The import ID must be a valid AWS IAM role ARN (e.g., arn:aws:iam::123456789012:role/role-name). "+
					"Got: %s", roleArn),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Importing OCM role link: %s", roleArn))

	// Fetch current organization
	acctResp, err := r.connection.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get current account",
			fmt.Sprintf("Unable to retrieve current user's organization: %s", err.Error()),
		)
		return
	}
	orgID := acctResp.Body().Organization().ID()

	// Verify the role is actually linked
	labelsResp, err := r.connection.AccountsMgmt().V1().Organizations().
		Organization(orgID).
		Labels().
		Labels(OCMRoleLabel).
		Get().
		Send()

	if err != nil {
		if labelsResp != nil && labelsResp.Status() == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Role not linked",
				fmt.Sprintf("Role %s is not linked to organization %s", roleArn, orgID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to verify role link",
			fmt.Sprintf("Unable to check if role %s is linked: %s", roleArn, err.Error()),
		)
		return
	}

	// Check if the specific role is in the label and validate AWS account uniqueness
	linkedArns := strings.Split(labelsResp.Body().Value(), ",")
	found := false

	// Parse the ARN to get AWS account ID for validation
	parsedARN, parseErr := arn.Parse(roleArn)
	if parseErr != nil {
		resp.Diagnostics.AddError(
			"Invalid role ARN",
			fmt.Sprintf("The provided role ARN is not valid: %s", parseErr.Error()),
		)
		return
	}

	for _, linkedArn := range linkedArns {
		trimmed := strings.TrimSpace(linkedArn)
		if trimmed == roleArn {
			found = true
			continue
		}

		// Check if another role from the same AWS account is already linked
		if otherParsedARN, err := arn.Parse(trimmed); err == nil {
			if otherParsedARN.AccountID == parsedARN.AccountID {
				resp.Diagnostics.AddError(
					"AWS account conflict detected",
					fmt.Sprintf(
						//nolint:lll
						"Cannot import role '%s' because organization '%s' already has a different role from the same AWS account '%s': '%s'. Only one role per AWS account is allowed per organization.",
						roleArn,
						orgID,
						parsedARN.AccountID,
						trimmed,
					),
				)
				return
			}
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Role not linked",
			fmt.Sprintf("Role %s is not linked to organization %s", roleArn, orgID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_arn"), roleArn)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
}

func (r *OCMRoleLinkResource) checkIfRoleLinked(ctx context.Context, orgID, roleArn,
	awsAccountID string) (bool, []string, error) {
	labelsResp, err := r.connection.AccountsMgmt().V1().Organizations().
		Organization(orgID).
		Labels().
		Labels(OCMRoleLabel).
		Get().
		SendContext(ctx)

	if err != nil {
		if labelsResp != nil && labelsResp.Status() == http.StatusNotFound {
			tflog.Debug(ctx, fmt.Sprintf("No OCM role label found for organization %s", orgID))
			return false, nil, nil
		}
		return false, nil, err
	}

	existingValue := labelsResp.Body().Value()
	rawArns := strings.Split(existingValue, ",")
	var linkedArns []string
	exists := false

	for _, linkedArn := range rawArns {
		trimmed := strings.TrimSpace(linkedArn)
		if trimmed == roleArn {
			exists = true
		}

		parsedARN, parseErr := arn.Parse(trimmed)
		if parseErr == nil && parsedARN.AccountID == awsAccountID && trimmed != roleArn {
			return false, nil, fmt.Errorf(
				//nolint:lll
				"organization '%s' already has role from AWS account %s: %s. Only one role can be linked per AWS account per organization",
				orgID,
				awsAccountID,
				trimmed,
			)
		}

		linkedArns = append(linkedArns, trimmed)
	}

	return exists, linkedArns, nil
}

func (r *OCMRoleLinkResource) linkRole(ctx context.Context, orgID, roleArn, awsAccountID string) error {
	// Lock to prevent concurrent label modifications from causing lost updates
	labelMutex.Lock()
	defer labelMutex.Unlock()

	exists, linkedArns, err := r.checkIfRoleLinked(ctx, orgID, roleArn, awsAccountID)
	if err != nil {
		return err
	}

	if exists {
		tflog.Debug(ctx, fmt.Sprintf("Role %s is already linked, skipping", roleArn))
		return nil
	}

	hasLinkedArns := len(linkedArns) > 0
	newValue := roleArn
	if hasLinkedArns {
		newValue = strings.Join(linkedArns, ",") + "," + roleArn
	}

	labelBuilder, err := amsv1.NewLabel().Key(OCMRoleLabel).Value(newValue).Build()
	if err != nil {
		return fmt.Errorf("failed to build label: %s", err.Error())
	}

	if !hasLinkedArns {
		tflog.Debug(ctx, fmt.Sprintf("Creating new OCM role label for organization %s", orgID))
		_, err = r.connection.AccountsMgmt().V1().Organizations().
			Organization(orgID).
			Labels().
			Add().
			Body(labelBuilder).
			SendContext(ctx)
	} else {
		tflog.Debug(ctx, fmt.Sprintf("Updating existing OCM role label for organization %s", orgID))
		_, err = r.connection.AccountsMgmt().V1().Organizations().
			Organization(orgID).
			Labels().
			Labels(OCMRoleLabel).
			Update().
			Body(labelBuilder).
			SendContext(ctx)
	}

	return err
}
